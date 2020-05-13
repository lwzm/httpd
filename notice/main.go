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
	body io.Reader
	meta string
}

func (data *payload) transport(w http.ResponseWriter) {
	if s := data.meta; s != "" {
		w.Header().Set("Content-Type", s)
	}
	if f := data.body; f != nil {
		n, err := io.Copy(w, f)
		if err != nil {
			log.Println("written", n, err)
		}
	}
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
			data := <-chTmp
			data.transport(w)
			chTmp <- payload{meta: httpd.ClientIP(r) + "\t" + r.UserAgent()}
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
				r, w := io.Pipe()
				chTmp <- payload{r, mime}
				todos = append(todos, chTmp)
				writers = append(writers, w)
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
		for _, w := range writers {
			w.(io.Closer).Close()
		}

		for _, chTmp := range todos {
			fmt.Fprintln(w, (<-chTmp).meta)
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
}
