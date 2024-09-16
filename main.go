package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1> Welcome to my page</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1> Contact page </h1> Email at <a href=\"mailto:tnmykhot@gmail.com\"> tnmykhot@gmail.com </a> to get in touch")
}

type Router struct{}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	default:
		//w.WriteHeader(http.StatusNotFound)
		//fmt.Fprint(w, "Page not found")
		http.Error(w, "page not found", http.StatusNotFound)
	}
}
func main() {
	var router Router
	fmt.Println("Starting the server on :3000..")
	http.ListenAndServe(":3000", router)
}
