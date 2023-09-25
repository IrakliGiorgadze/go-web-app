package controllers

import (
	"net/http"
)

func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}

func FAQ(tpl Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   string
	}{
		{
			Question: "What are two things you can never eat for breakfast?",
			Answer:   "Lunch and Dinner.",
		},
		{
			Question: "What is always coming but never arrives?",
			Answer:   "Tomorrow",
		},
		{
			Question: "What gets wetter the more it dries?",
			Answer:   `A towel`,
		},
		{
			Question: "What can be broken but never held?",
			Answer:   `A promise`,
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, questions)
	}
}
