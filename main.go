package main

import (
	"context"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load(".env")
	if (err != nil){
		panic("Could't load env variables")
	}
	server := http.Server{Addr: mustGetEnv("ADDRESS")}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     mustGetEnv("REDIS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: 0,
	})

	ctx := context.WithValue(context.Background(), "redis", redisClient)
	ctx = mustInitCrypto(ctx)

	http.HandleFunc("INFO /*", contextHandler(false, corsHandler, ctx))
	http.HandleFunc("OPTIONS /*", contextHandler(false, corsHandler, ctx))

	http.HandleFunc("GET /api/auth/signin", contextHandler(false, signinHandler, ctx))
	http.HandleFunc("GET /api/auth/signout", contextHandler(true, signoutHandler, ctx))
	http.HandleFunc("GET /api/auth", contextHandler(true, authHandler, ctx))

	http.HandleFunc("GET /api/search/all", contextHandler(false, queryReposHandler, ctx))
	http.HandleFunc("GET /api/search/star", contextHandler(true, queryReposHandler, ctx))
	http.HandleFunc("GET /api/search/watch", contextHandler(true, queryReposHandler, ctx))
	http.HandleFunc("GET /api/search/own", contextHandler(true, queryReposHandler, ctx))

	http.HandleFunc("GET /api/issues", contextHandler(false, queryIssuesHandler, ctx))
	http.HandleFunc("GET /api/issue_details", contextHandler(false, queryIssuesHandler, ctx))

	http.HandleFunc("GET /api/comments", contextHandler(false, queryCommentsHandler, ctx))

	http.HandleFunc("POST /api/comments/new", contextHandler(true, createCommentHandler, ctx))

	http.HandleFunc("GET /*", SPAHandler)
	err = server.ListenAndServe()
	println(err.Error())
}


