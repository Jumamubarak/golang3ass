package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	//router.HandlerFunc(http.MethodGet, "/module/:id", app.getModule)
	router.HandlerFunc(http.MethodPost, "/login", app.loginUser)
	router.HandlerFunc(http.MethodPost, "/register", app.createUser)
	router.HandlerFunc(http.MethodGet, "/activation", app.authenticateUser)

	return router
}

//app.requireActivatedUser()
