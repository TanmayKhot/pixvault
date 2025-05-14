package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/TanmayKhot/pixvault/controllers"
	"github.com/TanmayKhot/pixvault/models"
	"github.com/TanmayKhot/pixvault/templates"
	"github.com/TanmayKhot/pixvault/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load() // Check if there could be an error reading config data from env
	if err != nil {
		return cfg, err
	}

	// TODO: Read the PSQL values from an ENV variable
	cfg.PSQL = models.DefaultPostgresConfig()

	// TODO: Setup SMTP
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	postStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(postStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// TODO: Read the CSRF values from an ENV variable
	cfg.CSRF.Key = "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	cfg.CSRF.Secure = false

	// TODO: Read the server values from an ENV variable
	cfg.Server.Address = ":3000"

	return cfg, nil
}

func main() {

	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup the database connection
	db, err := models.OpenDBConnection(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//err = models.MigrateFS(db, migrations.FS, ".")
	err = models.Migrate(db, "migrations")
	if err != nil {
		panic(err)
	}

	// Setup CSRF middleware
	csrf_middleware := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"), // Using a fixed path will ensure that the CSRF cookie value is read at all page addresses, even sub pages within gallery
	)

	// Setup services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	emailSigninService := &models.EmailSigninService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	galleryService := &models.GalleryService{
		DB: db,
	}

	// Setup middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// Set up controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
		EmailSigninService:   emailSigninService,
	}
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}

	// Parse templates
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "signin.gohtml", "tailwind.gohtml"))
	usersC.Templates.UserProfile = views.Must(views.ParseFS(
		templates.FS, "userprofile.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS, "forgot-pw.gohtml", "tailwind.gohtml"))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS, "check-your-email.gohtml", "tailwind.gohtml"))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS, "reset-pw.gohtml", "tailwind.gohtml"))
	usersC.Templates.SigninWithEmail = views.Must(views.ParseFS(
		templates.FS, "signin-with-email.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "galleries/new.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(
		templates.FS, "galleries/edit.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Index = views.Must(views.ParseFS(
		templates.FS, "galleries/index.gohtml", "tailwind.gohtml"))
	galleriesC.Templates.Show = views.Must(views.ParseFS(
		templates.FS, "galleries/show.gohtml", "tailwind.gohtml"))

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
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	r.Route("/users/me", func(r chi.Router) { //We want to restrict access to this page only to signed in users. Without that, anyone can open this page
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/signin-with-email", usersC.EmailSignin)
	r.Post("/signin-with-email", usersC.ProcessEmailSignin)
	r.Get("/email-signin", usersC.VerifyEmailSignin)
	r.Route("/galleries", func(r chi.Router) { //We want to restrict access to this page only to signed in users. Without that, anyone can open this page
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}/access", galleriesC.ChangeAccess)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
	})

	// Start the server
	fmt.Println("Starting the server on %s...", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
