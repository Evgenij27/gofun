package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Request struct {
	Site       []string
	SearchText string
}

type Response struct {
	FoundAtSite string
}

func readClient(uri string) (string, error) {
	resp, err := http.Get(uri)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func doSearch(what string, uri string) string {
	fmt.Println("Starting searching for ", uri)
	body, err := readClient(uri)
	var res string
	if err != nil {
		fmt.Printf("Error during reading %s\n", err)
		panic(err)
	}
	if strings.Contains(body, what) {
		res = uri
	}
	return res
}

func search(ctx *gin.Context) {
	// Create context
	cwt, cancel :=
		context.WithTimeout(context.Background(), 30000*time.Millisecond)
	defer cancel()
	// Read input
	var req Request
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Bad")
		return
	}

	fmt.Println("request is ", req)
	ch := make(chan string)
	for _, uri := range req.Site {
		go func(uri string) {
			select {
			case ch <- doSearch(req.SearchText, uri):
			case <-cwt.Done():
				return
			}
		}(uri)
	}

	msg := <-ch
	if len(msg) > 0 {
		resp := Response{FoundAtSite: msg}
		ctx.JSON(http.StatusOK, resp)
	} else {
		ctx.Status(http.StatusNoContent)
	}
	fmt.Println("Writing response to the client")
}

func main() {
	router := gin.Default()
	router.POST("/ping", search)
	router.Run(":8080")
}
