package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func executeTemplate(w http.ResponseWriter, filepath string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl, err := template.ParseFiles(filepath)
	if err != nil {
		log.Printf("Parsing %s template: %v", filepath, err)
		http.Error(w, "There was an error in parsing the template", http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, nil)
	if err != nil {
		log.Printf("Executing %s template: %v", filepath, err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tplPath := "templates/home.gohtml"
	executeTemplate(w, tplPath)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	tplPath := "templates/contact.gohtml"
	executeTemplate(w, tplPath)
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	tplPath := "templates/faq.gohtml"
	executeTemplate(w, tplPath)
}

func galleryIDHandler(w http.ResponseWriter, r *http.Request) {
	// Working with URL parameters
	itemID := chi.URLParam(r, "galleryID")
	fmt.Fprintf(w, "Item id is %s", itemID)
}

func main() {

	r := chi.NewRouter()
	// r.Use(middleware.Logger) // Use it across the app
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
