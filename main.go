package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/IrakliGiorgadze/go-web-app/controllers"
	"github.com/IrakliGiorgadze/go-web-app/migrations"
	"github.com/IrakliGiorgadze/go-web-app/models"
	"github.com/IrakliGiorgadze/go-web-app/templates"
	"github.com/IrakliGiorgadze/go-web-app/views"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

const (
	webPort = "3000"
)

func main() {
	// Set up the DB
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
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
	userService := models.UserService{
		DB: db,
	}

	sessionService := models.SessionService{
		DB: db,
	}

	// Set up middleware
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	csrfKey := "4Kp7j9LsR2TtVxNnQpZbCvFr5GhXwYzD"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false))

	// Set up controllers
	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		"signup.gohtml", "tailwind.gohtml",
	))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS,
		"signin.gohtml", "tailwind.gohtml",
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

	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start the server
	fmt.Println("Starting the server on port:", webPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", webPort), r)
}
