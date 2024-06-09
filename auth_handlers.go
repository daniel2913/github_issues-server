package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func authHandler(w http.ResponseWriter, _ *http.Request, ctx context.Context) error {
	query := `query{ viewer {name url avatarUrl login}}`
	return relayRequest(w, query, ctx)
}

func signinHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	code := r.URL.Query().Get("code")
	reqctx := mustGetReqContext(ctx)
	if code == "" {
		reqctx.status = http.StatusBadRequest
		log.Debug().Msgf("Connection from %s did not provide code for authentication", r.RemoteAddr)
		return fmt.Errorf("Invalid request")
	}

	newUser := User{}

	err := newUser.Register(code, ctx)
	if err != nil {
		return err
	}

	err = newUser.RefreshInfo(ctx)
	if err != nil {
		return err
	}

	err = newUser.Save(ctx)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		return err
	}

	token, err := newUser.JWT(ctx)

	if err != nil {
		reqctx.status = http.StatusInternalServerError
		return err
	}
	http.SetCookie(w, &http.Cookie{
		MaxAge:   60 * 60 * 24,
		Name:     "Authorization",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})
	reqctx.status = http.StatusOK
	return nil
}

func signoutHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	http.SetCookie(w, &http.Cookie{
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
		Name:     "Authorization",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
	})
	reqctx.status = http.StatusOK
	return nil
}

