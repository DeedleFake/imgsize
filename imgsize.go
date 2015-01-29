package main

import (
	"fmt"
	"net/http"
)

func handleMain(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "This is a test.")
}

func main() {
	http.HandleFunc("/", handleMain)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}
