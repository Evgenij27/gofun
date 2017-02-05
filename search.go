package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Request struct {
	Site []string `json:"site"`
	Find string   `json:"find"`
}

type Response struct {
	Founded string `json:"founded"`
}

func f(done chan bool, url string, str string) <-chan string {
	out := make(chan string)

	go func() {
		fmt.Printf("START %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("error ", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("error ", err)
		}
		if strings.Contains(string(body), str) {
			fmt.Printf("Finded %s at %s\n", str, url)
			select {
			case out <- url:
			case <-done:
				fmt.Printf("%s closed at DONE\n", url)
				return
			}
		}
		close(out)
	}()
	return out
}

func merge(done chan bool, chnls []<-chan string) <-chan string {
	var wg sync.WaitGroup

	out := make(chan string)

	output := func(c <-chan string) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(chnls))
	for _, c := range chnls {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	r := gin.Default()

	r.POST("/ping", func(c *gin.Context) {

		fmt.Println("--- START ---")
		done := make(chan bool)
		defer close(done)
		var req Request
		if c.BindJSON(&req) == nil {
			urls := req.Site
			find := req.Find
			chnls := make([]<-chan string, len(urls))
			for index, url := range urls {
				chnl := f(done, url, find)
				chnls[index] = chnl
			}

			resChnl := merge(done, chnls)

			s := <-resChnl
			if len(s) > 0 {
				resp := &Response{s}
				c.JSON(http.StatusOK, resp)
			} else {
				c.Status(http.StatusNoContent)
			}
		} else {
			c.Status(http.StatusBadRequest)
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
