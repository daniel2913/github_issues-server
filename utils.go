package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
)

func getRedis(ctx context.Context) *redis.Client {
	redisClient, ok := ctx.Value("redis").(*redis.Client)
	if !ok {
		panic("Redis not available!")
	}
	return redisClient
}

func contextHandler(userOnly bool, fn func(w http.ResponseWriter, r *http.Request, ctx context.Context), ctx context.Context) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		token, login, err := authenticate(w, r, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userOnly && login == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		ctx = context.WithValue(ctx, "authorization", token)
		ctx = context.WithValue(ctx, "login", login)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		fn(w, r, ctx)
		if userOnly {
			w.Header().Set("Cache-Control", "no-store")
		}
	}
}

func mustInitCrypto(ctx context.Context) context.Context {
	keyBytes, err := os.ReadFile(mustGetEnv("PRIVATE_KEY"))
	if err != nil {
		panic("Couldn't read private key")
	}
	pubBytes, err := os.ReadFile(mustGetEnv("PUBLIC_KEY"))
	if err != nil {
		panic("Couldn't read public key")
	}

	privateRsaKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		panic(fmt.Sprintf("Couldn't parse private key"))
	}
	publicRsaKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		panic(fmt.Sprintf("Couldn't parse public key"))
	}
	ctx = context.WithValue(ctx, "private_rsa_key", privateRsaKey)
	return context.WithValue(ctx, "public_rsa_key", publicRsaKey)
}

func getQuerySearchAfter(r *http.Request) (string, string) {
	search := r.URL.Query().Get("query")
	after := r.URL.Query().Get("after")
	if after == "" || after == "null" {
		after = "null"
	} else {
		after = fmt.Sprintf(`"%s"`, after)
	}
	return search, after
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Key %s not found in env", key))
	}
	return val
}

func makeValidRequestBody(request string) (*bytes.Buffer, error) {
	jsonQuery, err := json.Marshal(map[string]string{"query": request})
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonQuery), nil
}
