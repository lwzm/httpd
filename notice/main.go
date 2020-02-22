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
	mime string
}

func (data *payload) transport(w http.ResponseWriter) {
	if s := data.mime; s != "" {
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

func responseError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(500)
	fmt.Fprint(w, err)
}

var pool = stack{make([]string, 100), sync.Mutex{}}
var channels = make(map[string](chan chan payload))

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	_, all := r.URL.Query()["all"]
	ctx := r.Context()
	ch, ok := channels[key]
	if !ok {
		ch = make(chan chan payload)
		channels[key] = ch
	}
	if r.Method == "GET" {
		select {
		case chTmp := <-ch:
			data := <-chTmp
			data.transport(w)
			close(chTmp)
		case <-ctx.Done():
			log.Println(ctx)
		}
	} else {
		n := 0
		if all {
			// use temp file
			mime := r.Header.Get("Content-Type")
			tmp := pool.pop()
			defer pool.push(tmp)
			f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE, 0600)
			defer os.Remove(tmp)
			if err != nil {
				responseError(err, w)
				return
			}
			defer f.Close()
			{
				_, err := io.Copy(f, r.Body)
				if err != nil {
					responseError(err, w)
					return
				}
			}
			for {
				chTmp := make(chan payload)
				select {
				case ch <- chTmp:
					f, err := os.OpenFile(tmp, os.O_RDONLY, 0)
					if err != nil {
						responseError(err, w)
						return
					}
					chTmp <- payload{f, mime}
					n++
				default:
					goto end
				}
			}
		end:
		} else {
			chTmp := make(chan payload)
			select {
			case ch <- chTmp:
				chTmp <- newPayload(r)
				<-chTmp
				n++
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
		fmt.Fprint(w, n)
	}
}

func main() {
	http.HandleFunc("/", handler)
	httpd.Start()
}
