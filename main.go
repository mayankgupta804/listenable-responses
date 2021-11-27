package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Response struct {
	Data chan string
	Err  chan error
	Done chan struct{}
}

func NewResponse() Response {
	resp := Response{}
	resp.Done = make(chan struct{})
	resp.Data = make(chan string, 5)
	resp.Err = make(chan error, 5)
	return resp
}

func (r *Response) WaitTillComplete() {
	<-r.Done
}

func (r *Response) Subscribe(hooks ...func(response *Response)) {
	for _, hook := range hooks {
		go func(hook func(response *Response)) {
			hook(r)
		}(hook)
	}
}

type Processor struct{}

func (p *Processor) Execute() Response {
	resp := NewResponse()
	// simulating a long running task
	go func() {
		for i := 0; i < 10; i++ {
			// push results to data channel of response
			if i%2 == 0 {
				resp.Data <- fmt.Sprintf("%d -> hello", i)
			} else {
				resp.Err <- fmt.Errorf("%d -> error", i)
			}
		}
		close(resp.Err)
		close(resp.Data)
	}()
	time.AfterFunc(10*time.Second, func() {
		close(resp.Done)
	})
	return resp
}

func main() {
	log.Println("Starting process...")
	var withListener string
	if len(os.Args) > 1 {
		withListener = os.Args[1]
	}
	p := Processor{}
	resp := p.Execute()
	log.Println("Process started...")

	if withListener == "true" {
		resultListener := func(response *Response) {
			for data := range response.Data {
				log.Println(data)
			}
		}

		errorListener := func(response *Response) {
			for err := range response.Err {
				log.Println(err)
			}
		}
		resp.Subscribe(resultListener)
		resp.Subscribe(errorListener)
	}

	resp.WaitTillComplete()
	log.Println("Process over...")
}
