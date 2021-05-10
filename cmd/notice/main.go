package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

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
	ch, ok := v.(chan chan *payload)
	if !ok {
		ch = make(chan chan *payload)
		channels.Store(key, ch)
	}

	if r.Method == "GET" {
		w.(http.Flusher).Flush()
		select {
		case chTmp := <-ch:
			ct := (<-chTmp).meta
			if ct != "" {
				w.Header().Set("Content-Type", ct)
			}
			id := httpd.ClientIP(r) + "\t" + r.URL.RawQuery + "\t" + r.UserAgent()
			chTmp <- &payload{meta: id, writer: w}
			<-chTmp
		case <-ctx.Done():
			log.Println(ctx)
		}
	} else {
		time.Sleep(0)
		mime := r.Header.Get("Content-Type")
		todos := make([]chan *payload, 0, 1)
		writers := []io.Writer{}
		subscribers := []string{}
	For:
		for {
			chTmp := make(chan *payload)
			select {
			case ch <- chTmp:
				chTmp <- &payload{meta: mime}
				todos = append(todos, chTmp)
				peer := <-chTmp
				writers = append(writers, peer.writer)
				subscribers = append(subscribers, peer.meta)
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

		t0 := time.Now()
		written, err := io.Copy(io.MultiWriter(writers...), r.Body)
		if err != nil {
			log.Println("error copy", written, err)
		}

		for _, chTmp := range todos {
			close(chTmp)
		}

		fmt.Fprintf(w, "%v size:%v cost:%v error:%v\n",
			len(subscribers), written, time.Since(t0), err)
		for _, s := range subscribers {
			fmt.Fprintln(w, s)
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
}
