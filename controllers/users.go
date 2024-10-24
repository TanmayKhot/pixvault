package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/TanmayKhot/pixvault/context"
	"github.com/TanmayKhot/pixvault/models"
)

type Users struct {
	// This struct will contain all objects of type Template organized
	Templates struct {
		New            Template
		SignIn         Template
		UserProfile    Template
		ForgotPassword Template
		CheckYourEmail Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
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
	http.Redirect(w, r, "users/me", http.StatusFound)
	fmt.Fprintf(w, "User authenticated", user)
}

// Finding the user using Cookie token + session + DB query
/*
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
	//fmt.Fprintf(w, "Email cookie: %s\n", user.Email)
	// ---------------------------------
	var data struct {
		Email string
	}
	data.Email = user.Email
	u.Templates.UserProfile.Execute(w, r, data)
	// ------------------
}
*/
// CurrentUser function using Context.
// SetUser and RequireUser middleware are required.
func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	var data struct {
		Email string
	}
	data.Email = user.Email
	u.Templates.UserProfile.Execute(w, r, data)

}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {

	// 1. Read current session token
	token, err := readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
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

type UserMiddleware struct {
	SessionService *models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := readCookie(r, CookieSession)
		if err != nil {
			// Cannot lookup the user's cookie/session so we continue with the next handler
			// For example, the user could be looking at our Home page
			// And we do not need a session tracking with cookies for that. It can be processed without sesison
			next.ServeHTTP(w, r)
			return
		}
		user, err := umw.SessionService.User(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()                // Get the context from request
		ctx = context.WithUser(ctx, user) // Update the context by adding a new user to UserKey
		r = r.WithContext(ctx)            // This updates the old request with new request which has user context
		next.ServeHTTP(w, r)              // Pass updated request, with context having user info to next handler
	})

}

// Some pages must require a user to be present in the request context (such as viewing or editing galleries)
// If user is not present in the context, we redirect them to signin
// If it is present then we call the next handlerfunc()
func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ForgotPassword handler
func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	// This method is used to get input (email) from the user
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

// Process the input for ForgotPassword
func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")

	// Create the passwordreset token
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		// TODO: Handle other cases in future
		// For eg: user email doesn't exist
		fmt.Println("Password Reset error token creation: ", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	vals := url.Values{
		"token": {pwReset.Token},
	}
	// Make the URL here configurable
	resetURL := "https://www.pixvault.com/reset-pw?" + vals.Encode()
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	u.Templates.CheckYourEmail.Execute(w, r, data)
}
