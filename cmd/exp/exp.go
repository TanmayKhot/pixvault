package main

import (
	"html/template"
	"net/http"
)

type User struct {
	Name     string
	Bio      string
	Age      int
	Friends  []string
	Details  map[string]string
	Admin    bool
	Loggedin bool
}

func handler(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}

	users := []User{
		{
			Name:    "John Smith",
			Bio:     `Haha! you have been h4x0r3d`,
			Age:     17,
			Friends: []string{"A", "B", "C"},
			Details: map[string]string{
				"City":   "New York",
				"Age":    "25",
				"School": "NYU",
			},
			Admin:    false,
			Loggedin: true,
		},
		{
			Name:    "Tommy Holfiger",
			Bio:     `hey`,
			Age:     37,
			Friends: []string{"X", "Y", "Z"},
			Details: map[string]string{
				"City":   "LA",
				"Age":    "35",
				"School": "UCLA",
			},
			Admin:    true,
			Loggedin: true,
		},
	}
	err = t.Execute(w, users)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

}
