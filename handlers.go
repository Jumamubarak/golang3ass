package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"week4/internal/data"
	"week4/pkg/utils"
)

func (app *application) createUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Fname    string `json:"fname"`
		Sname    string `json:"sname"`
		Email    string `json:"email"`
		Password string `json:"password"`
		UserRole string `json:"user_role"`
		Version  int    `json:"version"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	user := &data.UserInfo{
		Fname:        input.Fname,
		Sname:        input.Sname,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		UserRole:     input.UserRole,
		Version:      input.Version,
	}

	userInfoFromDb, err2 := app.models.UserInfo.CreateUser(user)
	if err2 != nil {
		app.serverErrorResponse(w, r, err2)
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": userInfoFromDb}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	user, err := app.models.UserInfo.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			app.notFoundResponse(w, r)
		}
		app.serverErrorResponse(w, r, err)
	}

	if !user.Activated {
		app.unauthorizedResponse(w, r)
	}

	if utils.CheckPasswordHash(user.PasswordHash, input.Password) {
		app.badCredentialsResponse(w, r)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": "authenticated!"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) authenticateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	activationCode := r.URL.Query().Get("activationCode")
	if userID == "" || activationCode == "" {
		http.Error(w, "id and activationCode are required", http.StatusBadRequest)
		return
	}

	id, _ := strconv.ParseInt(userID, 10, 64)
	user, err3 := app.models.UserInfo.GetByID(id)
	if err3 != nil {
		if errors.Is(sql.ErrNoRows, err3) {
			app.notFoundResponse(w, r)
		}
		app.serverErrorResponse(w, r, err3)
	}

	if activationCode == user.Token.Hash {
		user.Activated = true
		_, err := app.models.UserInfo.Update(user)
		if err != nil {
			app.serverErrorResponse(w, r, err)
		}
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"user": "token is correct!"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
