package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
)

type User struct {
	AccessToken string
	Login       string
	Id          string
}

func (this *User) Save(ctx context.Context) error {
	redis := mustGetRedis(ctx)
	err := redis.Set(ctx, fmt.Sprintf("%s:login", this.Id), this.Login, 0).Err()
	if err != nil {
		log.Error().Err(err).Str("id", this.Id).Str("login", this.Login).Msg("Error while saving user login to redis database")
		return err
	}
	err = redis.Set(ctx, fmt.Sprintf("%s:access_token", this.Id), this.AccessToken, 0).Err()
	if err != nil {
		log.Error().Err(err).Str("id", this.Id).Int("access token length", len(this.AccessToken)).Msg("Error while saving user token to redis database")
		return err
	}
	return nil
}

func (this *User) Load(ctx context.Context, id string) error {

	redis := mustGetRedis(ctx)
	login, err := redis.Get(ctx, fmt.Sprintf("%s:login", id)).Result()
	if err != nil {
		log.Error().Err(err).Msgf("Error while accessing user %s in database", id)
		return err
	}
	if login == "" {
		log.Debug().Msgf("User %s not found in database", id)
		return fmt.Errorf("Not Found")
	}

	accessToken, err := redis.Get(ctx, fmt.Sprintf("%s:access_token", id)).Result()
	this.Id = id
	this.Login = login
	this.AccessToken = accessToken

	if this.AccessToken == "" {
		log.Debug().Msgf("User %s doesn't have an access token saved", this.Id)
		return fmt.Errorf("No access token")
	}
	return nil
}

func (u *User) Register(code string, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	resp, err := http.Post(fmt.Sprintf(
		"https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s",
		mustGetEnv("CLIENT_ID"),
		mustGetEnv("CLIENT_SECRET"),
		code,
	), "text/plain", strings.NewReader(""))
	if err != nil {
		log.Error().Err(err).Str("code length", code).Msgf("Error while getting access token for %s", u.Id)
		reqctx.status = http.StatusUnauthorized
		return err
	}
	if (resp.StatusCode/100 != 2){
		reqctx.status = http.StatusUnauthorized
		return fmt.Errorf("Not Authorized")
	}
	defer resp.Body.Close()
	access, err := io.ReadAll(resp.Body)
	if err != nil {
		reqctx.status = http.StatusServiceUnavailable
		log.Error().Err(err).Msg("Bad response from access token endpoint")
		return err
	}
	for _, param := range strings.Split(string(access), "&") {
		parts := strings.Split(param, "=")
		key := parts[0]
		value := parts[1]
		if key == "access_token" {
			u.AccessToken = value
		}
	}
	if u.AccessToken == "" {
		err = fmt.Errorf("Bad Server Response")
		reqctx.status = http.StatusInternalServerError
		log.Error().Err(err).Msg("No access token in authentication endpoint response")
		return err
	}
	return nil
}

func getUser(_ http.ResponseWriter, r *http.Request, ctx context.Context) (User, error) {
	user := User{}

	authCookie, err := r.Cookie("Authorization")
	if err != nil {
		return user, fmt.Errorf("No Authorization Cookie")
	}

	auth := authCookie.Value
	if auth == "" {
		log.Debug().Msgf("Connection from %s has empty authorization cookie", r.RemoteAddr)
		return user, fmt.Errorf("Invalid Authorization Cookie")
	}

	token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			log.Debug().Interface("algorithm", token.Header["alg"]).Msgf("Connection from %s has unexpected jwt signing method", r.RemoteAddr)
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return ctx.Value("public_rsa_key"), nil
	})
	if err != nil {
		log.Debug().Msgf("Connection from %s has invalid jwt token", r.RemoteAddr)
		return user, fmt.Errorf("Invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["id"] == "" {
		log.Error().Msgf("Connection from %s has invalid claims in jwt token", r.RemoteAddr)
		return user, fmt.Errorf("JWT token doesn't contain id")
	}

	id, ok := claims["id"].(string)
	if !ok {
		log.Error().Msgf("Connection from %s has no id in jwt token", r.RemoteAddr)
		return user, fmt.Errorf("Invalid JWT token")
	}

	err = user.Load(ctx, id)
	if err != nil {
		return user, err
	}

	if user.AccessToken == "" {
		return user, fmt.Errorf("User doesn't have auth token")
	}

	return user, nil
}
