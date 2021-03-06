package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func Test_handler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	e := httpexpect.New(t, server.URL)
	e.POST("/").Expect().Status(http.StatusNotFound).Body().Empty()

	go func() {
		e.GET("/t1").Expect().Status(http.StatusOK)
	}()
	time.Sleep(time.Millisecond)
	e.POST("/t1").Expect().Status(http.StatusOK).Text().Contains("\t")

	go func() {
		e.GET("/json").Expect().
			Status(http.StatusOK).
			JSON().Number().Equal(1)
	}()
	time.Sleep(time.Millisecond)
	e.POST("/json").WithJSON(1).Expect().Status(http.StatusOK)

	e.POST("/has-no-one-get").Expect().Status(http.StatusNotFound)

	n := 30

	for i := 0; i < n; i++ {
		go func() {
			e.GET("/multi-get").Expect().Status(http.StatusOK).Text().Equal("foobar")
		}()
	}
	time.Sleep(time.Millisecond * 100)
	e.POST("/multi-get").WithText("foobar").
		Expect().Text().Contains("\t")
	s := e.POST("/multi-get").WithQueryString("broadcast").WithText("foobar").
		Expect().Text().Raw()
	assert.Equal(t, n-1, strings.Count(s, "\n"))

	for idx := 0; idx < 20; idx++ {
		for i := 0; i < n; i++ {
			go func() {
				e.GET("/multi-get").Expect().Status(http.StatusOK).Text().Equal("foobar")
			}()
		}
		time.Sleep(time.Millisecond * 100)
		s := e.POST("/multi-get").WithQueryString("broadcast").WithText("foobar").
			Expect().Text().Raw()
		assert.Equal(t, n, strings.Count(s, "\n"))
	}

	for i := 0; i < n; i++ {
		go func(i int) {
			e.GET(fmt.Sprintf("/%v", i)).Expect().Status(http.StatusOK)
		}(i)
	}
	time.Sleep(time.Millisecond)
	for i := 0; i < n; i++ {
		e.POST(fmt.Sprintf("/%v", i)).Expect().Status(http.StatusOK)
	}

	time.Sleep(time.Millisecond * 10)
}
func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &http.Client{Timeout: time.Second / 10}
	_, err := client.Get(server.URL)
	assert.Error(t, http.ErrHandlerTimeout, err)
}

type stream struct {
	count int
	max   int
	piece []byte
}

func (st *stream) Read(p []byte) (int, error) {
	if st.count < st.max {
		// time.Sleep(time.Second / 1000)
		n := copy(p, st.piece)
		st.count++
		return n, nil
	}
	return 0, io.EOF
}

func TestIOError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	e := httpexpect.New(t, server.URL)

	go func() {
		e.GET("/").Expect().Body().Length().Equal(6000)
	}()

	time.Sleep(time.Millisecond)

	e.POST("/").
		WithChunked(&stream{max: 3000, piece: []byte{65, 66}}).
		Expect().Status(http.StatusOK)

	time.Sleep(time.Millisecond * 10)
}
