package main

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed client/*
var app embed.FS

func SPAHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=24939360")
	if len(strings.Split(r.URL.Path, ".")) > 1 {
		http.ServeFileFS(w, r, app, "client"+r.URL.Path)
		return
	}
	http.ServeFileFS(w, r, app, "client/index.html")
}
