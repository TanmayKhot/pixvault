package controllers

import (
	"fmt"
	"net/http"

	"github.com/TanmayKhot/pixvault/models"
)

type Users struct {
	// This tstruct will contain all objects of type Template organized
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
}

// Render the webpage for signup, get user inputs
func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

// Read the data from the user and create a new user
func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	// UserService is used to create the new user and insert in DB
	user, err := u.UserService.Create(email, password)
	if err != nil {
		fmt.Println("Error creating user %w", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Create a new session and cookie for the user created
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "signin", http.StatusFound)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "users/me", http.StatusFound)
}

// Render the signin page, get user input
func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

// Validate user inputs for Signin
func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")

	// User service authenticates the credentials with the DB
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// If credentials are valid, create a new session and setup a cookie
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSession, session.Token)

	fmt.Fprintf(w, "User authenticated", user)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	token, err := readCookie(r, CookieSession)
	// If we don't have a cookie for the user then we ask them to login again, create a new cookie
	// to track the new session
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	user, err := u.SessionService.User(token)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	fmt.Fprintf(w, "Email cookie: %s\n", user.Email)
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {

	// 1. Read current session token
	token, err := readCookie(r, CookieSession)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// 2. Delete the current session
	err = u.SessionService.Delete(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	// 3. Delete the user's cookie
	deleteCookie(w, CookieSession)

	// 4. Redirect the user to signin page after logging them out
	http.Redirect(w, r, "/signin", http.StatusFound)
}
