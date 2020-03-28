package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/lwzm/httpd"
)

type payload struct {
	body io.ReadCloser
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
		f.Close()
	}
}

func newPayload(r *http.Request) payload {
	return payload{
		r.Body,
		r.Header.Get("Content-Type"),
	}
}

type stack struct {
	data  []string
	mutex sync.Mutex
}

func (t *stack) push(s string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.data = append(t.data, s)
}

func (t *stack) pop() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	l := t.data
	n := len(l) - 1
	t.data = l[:n]
	if l[n] != "" {
		return l[n]
	}
	return fmt.Sprintf(".%v.tmp", n)
}

var pool = stack{make([]string, 100), sync.Mutex{}}
var channels = make(map[string](chan chan payload))
var mutex = sync.Mutex{}

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	_, all := r.URL.Query()["all"]
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
		if all {
			// use temp file
			mime := r.Header.Get("Content-Type")
			tmp := pool.pop()
			defer pool.push(tmp)
			f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if _, err := io.Copy(f, r.Body); err != nil {
				log.Fatal(err)
			}
			todos := make([]chan payload, 0)
			for {
				chTmp := make(chan payload)
				select {
				case ch <- chTmp:
					f, err := os.OpenFile(tmp, os.O_RDONLY, 0)
					if err != nil {
						log.Fatal(err)
					}
					chTmp <- payload{f, mime}
					todos = append(todos, chTmp)
				default:
					goto end
				}
			}
		end:
			for _, chTmp := range todos {
				fmt.Fprintln(w, (<-chTmp).meta)
			}
		} else {
			chTmp := make(chan payload)
			select {
			case ch <- chTmp:
				chTmp <- newPayload(r)
				fmt.Fprintln(w, (<-chTmp).meta)
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
}
