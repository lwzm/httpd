package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_handler(t *testing.T) {
	test := func(method, target string, body io.Reader, status int, text string) {
		req := httptest.NewRequest(method, target, body)
		w := httptest.NewRecorder()
		handler(w, req)
		resp := w.Result()
		if n := resp.StatusCode; n != status {
			t.Error(method, target, status, n)
		}
		if text != "" {
			body, _ := ioutil.ReadAll(resp.Body)
			if s := string(body); s != text {
				t.Error(method, target, text, s)
			}
		}
	}

	test("POST", "/", nil, http.StatusNotFound, "")

	go test("GET", "/", nil, http.StatusOK, "")
	time.Sleep(time.Millisecond)
	test("POST", "/", nil, http.StatusOK, "1")

	test("POST", "/?all", nil, http.StatusOK, "0")

	n := 100
	for i := 0; i < n; i++ {
		go test("GET", "/", nil, http.StatusOK, "foobar")
	}
	time.Sleep(time.Millisecond)

	body := strings.NewReader("foobar")
	test("POST", "/", body, http.StatusOK, "1")
	body.Seek(0, io.SeekStart)
	test("POST", "/?all", body, http.StatusOK, fmt.Sprint(n-1))
}
