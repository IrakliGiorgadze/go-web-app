package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/IrakliGiorgadze/go-web-app/controllers"
	"github.com/IrakliGiorgadze/go-web-app/migrations"
	"github.com/IrakliGiorgadze/go-web-app/models"
	"github.com/IrakliGiorgadze/go-web-app/templates"
	"github.com/IrakliGiorgadze/go-web-app/views"

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

	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	cfg.PSQL = models.DefaultPostgresConfig()

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	cfg.CSRF.Key = "4Kp7j9LsR2TtVxNnQpZbCvFr5GhXwYzD"
	cfg.CSRF.Secure = false

	cfg.Server.Address = ":3000"

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Set up the DB
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}

	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			log.Println("Cannot close DB connection", err)
		}
	}(db)

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Set up services
	userService := &models.UserService{
		DB: db,
	}

	sessionService := &models.SessionService{
		DB: db,
	}

	pwResetService := &models.PasswordResetService{
		DB: db,
	}

	galleryService := &models.GalleryService{
		DB: db,
	}

	emailService := models.NewEmailService(cfg.SMTP)

	// Set up middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)

	// Set up controllers
	// Users Controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
	))

	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS,
		"forgot-pw.gohtml", "tailwind.gohtml",
	))

	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS,
		"check-your-email.gohtml", "tailwind.gohtml",
	))

	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS,
		"reset-pw.gohtml", "tailwind.gohtml",
	))

	// Galleries Controllers
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}

	galleriesC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"galleries/new.gohtml", "tailwind.gohtml",
	))

	galleriesC.Templates.Edit = views.Must(views.ParseFS(
		templates.FS,
		"galleries/edit.gohtml", "tailwind.gohtml",
	))

	galleriesC.Templates.Index = views.Must(views.ParseFS(
		templates.FS,
		"galleries/index.gohtml", "tailwind.gohtml",
	))

	galleriesC.Templates.Show = views.Must(views.ParseFS(
		templates.FS,
		"galleries/show.gohtml", "tailwind.gohtml",
	))

	// Set up router and routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"home.gohtml", "tailwind.gohtml",
	))))

	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(
		templates.FS,
		"contact.gohtml", "tailwind.gohtml",
	))))

	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(
		templates.FS,
		"faq.gohtml", "tailwind.gohtml",
	))))

	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/users", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)

	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})

	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images", galleriesC.UploadImage)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
