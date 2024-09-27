package controllers

import (
	"fmt"
	"net/http"
)

type Users struct {
	// This tstruct will contain all objects of type Template organized
	Templates struct {
		New Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, data)

}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse the form submission", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, "<p>Email: %s</p>", r.FormValue("email"))
	fmt.Fprint(w, "<p>Password: %s</p>", r.FormValue("password"))

}
