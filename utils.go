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
	"github.com/rs/zerolog/log"
)

func mustGetRedis(ctx context.Context) *redis.Client {
	redisClient, ok := ctx.Value("redis").(*redis.Client)
	if !ok {
		log.Panic().Msg("Redis not available!")
	}
	return redisClient
}

type RequestContext struct {
	status int
	cache  string
	login  string
	token  string
}

func mustGetReqContext(ctx context.Context) *RequestContext {
	context, ok := ctx.Value("pContext").(*RequestContext)
	if !ok {
		log.Panic().Interface("context", ctx).Msgf("Error reading request context")
	}
	return context
}

func contextHandler(userOnly bool, fn func(w http.ResponseWriter, r *http.Request, ctx context.Context) error, ctx context.Context) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		reqctx := RequestContext{status: http.StatusTeapot, cache: "no-store"}

		ctx = context.WithValue(ctx, "pContext", &reqctx)

		token, login, err := authenticate(w, r, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		reqctx.login = login
		reqctx.token = token

		if userOnly && login == "" {
			log.Debug().Msgf("Unauthenticated request to %s from %s", r.RequestURI, r.RemoteAddr)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err = fn(w, r, ctx)

		if reqctx.status == http.StatusTeapot {
			log.Warn().Msgf("Request to %s %s returned default status code", r.Method, r.RequestURI)
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if userOnly {
			reqctx.cache = "no-store"
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")
		if reqctx.status != -1 {
			w.WriteHeader(reqctx.status)
		}
		w.Header().Set("Cache-Control", reqctx.cache)
	}
}

func mustInitCrypto(ctx context.Context) context.Context {
	keyBytes, err := os.ReadFile(mustGetEnv("PRIVATE_KEY"))
	if err != nil {
		log.Panic().Msg("Couldn't read private key")
	}
	pubBytes, err := os.ReadFile(mustGetEnv("PUBLIC_KEY"))
	if err != nil {
		log.Panic().Msg("Couldn't read public key")
	}

	privateRsaKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		log.Panic().Msg("Couldn't parse private key")
	}
	publicRsaKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		log.Panic().Msg("Couldn't parse public key")
	}
	log.Info().Msg("Loaded rsa keys")
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
		log.Panic().Msgf("Key %s not found in env", key)
	}
	return val
}

func makeValidRequestBody(request string) (*bytes.Buffer, error) {
	jsonQuery, err := json.Marshal(map[string]string{"query": request})
	if err != nil {

		log.Error().Err(err).Msg("Error when making request body")
		return nil, err
	}
	return bytes.NewBuffer(jsonQuery), nil
}
