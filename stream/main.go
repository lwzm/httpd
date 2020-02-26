package main

import (
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/lwzm/httpd"
)

var channels = make(map[string](chan chan io.Reader))
var mutex = sync.Mutex{}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.RequestURI()
	ctx := r.Context()
	mutex.Lock()
	ch, ok := channels[key]
	if !ok {
		ch = make(chan chan io.Reader)
		channels[key] = ch
	}
	mutex.Unlock()
	if r.Method == "GET" {
		select {
		case chTmp := <-ch:
			body := <-chTmp
			n, err := io.Copy(w, body)
			if err != nil {
				log.Printf("written %v, %v", n, err)
			}
			close(chTmp)
		case <-ctx.Done():
			log.Print(ctx)
		}
	} else {
		chTmp := make(chan io.Reader)
		select {
		case ch <- chTmp:
			chTmp <- r.Body
			// if flusher, ok := w.(http.Flusher); ok {
			// 	flusher.Flush()
			// }
			<-chTmp
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
	/*
		curl -v -H Expect: --limit-rate 10k -T /dev/zero localhost:1111/z
		curl -v localhost:1111/z >/dev/null
	*/
}
