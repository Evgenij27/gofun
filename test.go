package main

import "fmt"
import "sync"


func gen(done <-chan bool, nums ...int) <-chan int {
    out := make(chan int)
    go func() {
		defer close(out)
        for _, n := range nums {
            select {
			case out <- n:
			case <- done:
				return
			}
        }
        
    }()
    return out
}

func sq(done <-chan bool, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
			if n % 2 == 0 {
				select {
				case out <- n * n:
				case <- done:
					return
				}
			}
        }
        close(out)
    }()
    return out
}

func merge(done <- chan bool, cs ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    // Start an output goroutine for each input channel in cs.  output
    // copies values from c to out until c is closed, then calls wg.Done.
    output := func(c <-chan int) {
		defer wg.Done()
        for n := range c {
            select {
			case out <- n:
			case <- done:
				return
			}
        }
       
    }
    wg.Add(len(cs))
    for _, c := range cs {
        go output(c)
    }

    // Start a goroutine to close out once all the output goroutines are
    // done.  This must start after the wg.Add call.
    go func() {
        wg.Wait()
        close(out)
    }()
    return out
}

func main() {
	done := make(chan bool)
    in := gen(done, 3, 3, 5, 7, 9)

    // Distribute the sq work across two goroutines that both read from in.
    c1 := sq(done, in)
    c2 := sq(done, in)

    // Consume the merged output from c1 and c2.
    for n := range merge(done, c1, c2) {
        fmt.Println(n) // 4 then 9, or 9 then 4
    }
}

