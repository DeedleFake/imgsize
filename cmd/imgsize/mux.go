package main

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"text/template"
)

type Mux struct {
	*http.ServeMux

	Err *template.Template
}

func (m *Mux) Handle(p string, h http.Handler) {
	m.ServeMux.Handle(p, &errorHandler{h, m.Err})
}

func (m *Mux) HandleFunc(p string, h http.HandlerFunc) {
	m.ServeMux.HandleFunc(p, (&errorHandler{h, m.Err}).ServeHTTP)
}

type bufWriter struct {
	http.ResponseWriter

	buf bytes.Buffer
}

var bwPool = sync.Pool{
	New: func() interface{} {
		return new(bufWriter)
	},
}

func (bw *bufWriter) Write(data []byte) (int, error) {
	return bw.buf.Write(data)
}

func (bw *bufWriter) Flush() (err error) {
	_, err = bw.buf.WriteTo(bw.ResponseWriter)
	return
}

func (bw *bufWriter) Contains(data []byte) bool {
	return bytes.Contains(bw.buf.Bytes(), data)
}

type errorHandler struct {
	h    http.Handler
	tmpl *template.Template
}

func (h *errorHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic: %v", err)

			rw.WriteHeader(http.StatusInternalServerError)
			err := h.tmpl.Execute(rw, err)
			if err != nil {
				panic(err)
			}
		}
	}()

	bw := bwPool.Get().(*bufWriter)
	bw.ResponseWriter = rw
	defer func() {
		err := bw.Flush()
		if err != nil {
			panic(err)
		}

		bw.buf.Reset()
		bwPool.Put(bw)
	}()

	h.h.ServeHTTP(bw, req)

	if bw.Contains([]byte("404 page not found")) {
		log.Printf("404 caught: %v", req)

		rw.WriteHeader(http.StatusNotFound)
		err := h.tmpl.Execute(rw, bw)
		if err != nil {
			panic(err)
		}
	}
}
