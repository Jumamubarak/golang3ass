package main

import (
	"context"
	"net/http"
	"week4/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.UserInfo) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.UserInfo {
	user, ok := r.Context().Value(userContextKey).(*data.UserInfo)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
