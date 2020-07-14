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
	writer io.Writer
	meta   string
}

var channels = sync.Map{}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	_, broadcast := r.URL.Query()["broadcast"]
	ctx := r.Context()
	v, _ := channels.Load(key)
	ch, ok := v.(chan chan payload)
	if !ok {
		ch = make(chan chan payload)
		channels.Store(key, ch)
	}

	if r.Method == "GET" {
		select {
		case chTmp := <-ch:
			meta := (<-chTmp).meta
			if meta != "" {
				w.Header().Set("Content-Type", meta)
			}
			chTmp <- payload{meta: httpd.ClientIP(r) + "\t" + r.UserAgent(), writer: w}
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
				writers = append(writers, peer.writer)
				fmt.Fprintln(w, peer.meta)
			default:
				break For
			}
			if !broadcast {
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
			if err != io.ErrUnexpectedEOF {
				fmt.Fprintln(w, "\n"+err.Error())
			}
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
