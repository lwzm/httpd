package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/lwzm/httpd"
)

func init() {

	chchs := make(map[string](chan chan payload))
	chs := make(map[string](chan payload))
	mutex := sync.Mutex{}

	var timeout time.Duration = time.Second
	{
		n, _ := strconv.Atoi(os.Getenv("TIMEOUT"))
		if n == 0 {
			n = 600
		}
		timeout *= time.Duration(n)
	}

	release := func(key string, chTmp chan payload) {
		time.Sleep(timeout)
		mutex.Lock()
		defer mutex.Unlock()
		if chTmp != chs[key] {
			return
		}
		log.Println("timeout release:", key)
		select {
		case <-chTmp: // check closed
		default:
			chTmp <- payload{status: http.StatusRequestTimeout}
		}
		delete(chs, key)
		poolPutter <- key
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()
		mutex.Lock()
		ch, ok := chchs[key]
		if !ok {
			ch = make(chan chan payload)
			chchs[key] = ch
		}
		mutex.Unlock()
		ctx := r.Context()
		if r.Method == "GET" {
			select {
			case chTmp := <-ch:
				foo := <-chTmp
				loc := <-poolGetter
				w.Header().Set("Location", loc)
				foo.transport(w)
				mutex.Lock()
				chs[loc] = chTmp
				mutex.Unlock()
				go release(loc, chTmp)
			case <-ctx.Done():
				log.Println("gone away:", r)
			}
		} else {
			chTmp := make(chan payload)
			defer close(chTmp)
			select {
			case ch <- chTmp:
				chTmp <- payloadNew(r)
				select {
				case bar := <-chTmp:
					bar.transport(w)
				case <-ctx.Done():
					log.Println("gone away:", r)
				}
			default:
				w.WriteHeader(http.StatusTooManyRequests)
			}
		}
	})

	http.HandleFunc("/@/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RequestURI()
		mutex.Lock()
		chTmp, ok := chs[key]
		if !ok {
			mutex.Unlock()
			w.WriteHeader(http.StatusNotFound)
			return
		}
		delete(chs, key)
		mutex.Unlock()
		poolPutter <- key
		select {
		case <-chTmp: // check closed
			w.WriteHeader(http.StatusGone)
		default:
			chTmp <- payloadNew(r)
			<-chTmp
		}
	})
}

func main() {
	httpd.Start()
}
