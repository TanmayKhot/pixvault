package main

import (
	"fmt"
	"net/http"

	"github.com/TanmayKhot/pixvault/controllers"
	"github.com/TanmayKhot/pixvault/models"
	"github.com/TanmayKhot/pixvault/templates"
	"github.com/TanmayKhot/pixvault/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {

	r := chi.NewRouter()

	homeTpl := views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))
	r.Get("/", controllers.StaticHandler(homeTpl))

	contactTpl := views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))
	r.Get("/contact", controllers.StaticHandler(contactTpl))

	faqTpl := views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))
	r.Get("/faq", controllers.FAQhandler(faqTpl))

	cfg := models.DefaultPostgresConfig()
	db, err := models.OpenDBConnection(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	// New user Sign up
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)

	// User Sign in
	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Get("/users/me", usersC.CurrentUser)
	r.Post("/signout", usersC.ProcessSignOut)

	// ---------------------------
	usersC.Templates.UserProfile = views.Must(views.ParseFS(templates.FS, "userprofile.gohtml", "tailwind.gohtml"))
	r.Get("/users/me", usersC.CurrentUser)

	// ---------------------------

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	fmt.Println("Starting the server on :3000..")

	// Adding CSRF security
	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrf_middleware := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(false), // Will update at the time of deployment
	)
	http.ListenAndServe(":3000", csrf_middleware(r))
}
