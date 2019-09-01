package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
)

func handler(w http.ResponseWriter, r *http.Request) {
	var formatted, err = httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Fprint(w, err)
	}

	fmt.Println(string(formatted))
}

var port = flag.Uint("p", 10255, "port to listen to")

func main() {
	flag.Parse()

	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}
