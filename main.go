package main

import (
	"fmt"
	"net/http"

	"github.com/IrakliGiorgadze/go-web-app/controllers"
	"github.com/IrakliGiorgadze/go-web-app/templates"
	"github.com/IrakliGiorgadze/go-web-app/views"

	"github.com/go-chi/chi/v5"
)

const (
	webPort = "3000"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "home.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "contact.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml"))))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on port:", webPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", webPort), r)
}
