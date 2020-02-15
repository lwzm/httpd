package main

import (
	"io"
	"log"
	"net/http"

	"github.com/lwzm/httpd"
)

func init() {
	channels := make(map[string](chan chan io.Reader))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()
		ctx := r.Context()
		ch, ok := channels[key]
		if !ok {
			ch = make(chan chan io.Reader)
			channels[key] = ch
		}
		if r.Method == "GET" {
			select {
			case chTmp := <-ch:
				body := <-chTmp
				n, err := io.Copy(w, body)
				if err != nil {
					log.Println("written", n, err)
				}
				close(chTmp)
			case <-ctx.Done():
				log.Println(ctx)
			}
		} else {
			chTmp := make(chan io.Reader)
			select {
			case ch <- chTmp:
				chTmp <- r.Body
				<-chTmp
			default:
				w.WriteHeader(404)
			}
		}
	})
}

func main() {
	httpd.Start()
}
