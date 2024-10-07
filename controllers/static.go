package controllers

import (
	"net/http"
)

type Static struct {
	Template Template
}

func (static Static) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	static.Template.Execute(w, r, nil)
}

func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}

func FAQhandler(tpl Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   string
	}{
		{
			Question: "Question 1 from FAQHandler",
			Answer:   "Answer",
		},
		{
			Question: "Question 2 from FAQHandler",
			Answer:   "Answer",
		},
		{
			Question: "Question 3 from FAQHandler",
			Answer:   "Answer",
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, questions)
	}
}
