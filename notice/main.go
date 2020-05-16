package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/lwzm/httpd"
)

type payload struct {
	body io.Writer
	meta string
}

var channels = make(map[string](chan chan payload))
var mutex = sync.Mutex{}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	_, single := r.URL.Query()["single"]
	ctx := r.Context()
	mutex.Lock()
	ch, ok := channels[key]
	if !ok {
		ch = make(chan chan payload)
		channels[key] = ch
	}
	mutex.Unlock()

	if r.Method == "GET" {
		select {
		case chTmp := <-ch:
			w.Header().Set("Content-Type", (<-chTmp).meta)
			chTmp <- payload{meta: httpd.ClientIP(r) + "\t" + r.UserAgent(), body: w}
			<-chTmp
		case <-ctx.Done():
			log.Println(ctx)
		}
	} else {
		mime := r.Header.Get("Content-Type")
		todos := make([]chan payload, 0)
		writers := []io.Writer{}
	For:
		for {
			chTmp := make(chan payload)
			select {
			case ch <- chTmp:
				chTmp <- payload{meta: mime}
				todos = append(todos, chTmp)
				peer := <-chTmp
				writers = append(writers, peer.body)
				fmt.Fprintln(w, peer.meta)
			default:
				break For
			}
			if single {
				break
			}
		}

		if len(todos) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		n, err := io.Copy(io.MultiWriter(writers...), r.Body)
		if err != nil {
			log.Println("copy Request.Body to MultiWriter", n, err)
		}

		for _, chTmp := range todos {
			close(chTmp)
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
}
