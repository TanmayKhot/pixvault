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

	// Setup the database connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.OpenDBConnection(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//err = models.MigrateFS(db, migrations.FS, ".")
	err = models.Migrate(db, "migrations")
	if err != nil {
		panic(err)
	}

	// Setup services
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	// Setup middleware
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	// Adding CSRF security middleware
	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrf_middleware := csrf.Protect(
		[]byte(csrfKey),
		csrf.Secure(false), // Will update at the time of deployment
	)

	// Set up controllers
	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.UserProfile = views.Must(views.ParseFS(
		templates.FS, "userprofile.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS, "forgot-pw.gohtml", "tailwind.gohtml"))
	// Set up router and routes
	r := chi.NewRouter()

	// These middleware are used everywhere.
	r.Use(csrf_middleware)
	r.Use(umw.SetUser)

	// Setup for routes
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQhandler(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
	//r.Get("/users/me", controllers.FAQhandler(views.Must(views.ParseFS(templates.FS, "userprofile.gohtml", "tailwind.gohtml"))))
	r.Get("/users/me", usersC.CurrentUser)
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)

	// Start the server
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
