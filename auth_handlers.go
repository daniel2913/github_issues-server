package main

import (
	"context"
	"net/http"
	"time"
)

func authHandler(w http.ResponseWriter, _ *http.Request, ctx context.Context) {
	query := `query{ viewer {name url avatarUrl login}}`
	err := relayRequest(w, query, ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func signinHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newUser := User{}

	err := newUser.Register(code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = newUser.RefreshInfo()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	err = newUser.Save(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := newUser.JWT(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		MaxAge:   60 * 60 * 24,
		Name:     "Authorization",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})
}

func signoutHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.SetCookie(w, &http.Cookie{
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
		Name:     "Authorization",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
	})
}

