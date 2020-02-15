package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

type payload struct {
	body   io.Reader
	mime   string
	status int
}

func (data *payload) transport(w http.ResponseWriter) {
	if s := data.mime; s != "" {
		w.Header().Set("Content-Type", s)
	}
	if n := data.status; n != 0 {
		w.WriteHeader(n)
	}
	if data.body != nil {
		n, err := io.Copy(w, data.body)
		if err != nil {
			log.Println("written", n, err)
		}
	}
}

func payloadNew(r *http.Request) payload {
	code, _ := strconv.Atoi(r.Header.Get("Status-Code"))
	return payload{
		r.Body,
		r.Header.Get("Content-Type"),
		code,
	}
}
