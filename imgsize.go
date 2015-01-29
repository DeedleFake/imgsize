package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleMain(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "This is a test.")
}

func main() {
	http.HandleFunc("/", handleMain)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
