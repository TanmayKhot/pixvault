package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1> Welcome to my page</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1> Contact page </h1> Email at <a href=\"mailto:tnmykhot@gmail.com\"> tnmykhot@gmail.com </a> to get in touch")
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<h1>FAQ Page</h1>
  <ul>
	<li>
	  <b>Questions 1?</b>
	  Answer 1.
	</li>
	<li>
	  <b>Question 2?</b>
	  Answer 2
	</li>
	<li>
	  <b>Question 3?</b>
	  Answer 3
	</li>
  </ul>
  `)
}

func galleryIDHandler(w http.ResponseWriter, r *http.Request) {
	// Working with URL parameters
	itemID := chi.URLParam(r, "galleryID")
	fmt.Fprintf(w, "Item id is %s", itemID)
}

func main() {

	r := chi.NewRouter()
	// r.Use(middleware.Logger) Use it across the app
	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)

	//r.Get("/gallery/{galleryID}", galleryIDHandler)
	// Use middleware logger only for 1 path
	r.With(middleware.Logger).Get("/gallery/{galleryID}", galleryIDHandler)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	fmt.Println("Starting the server on :3000..")
	http.ListenAndServe(":3000", r)
}
