package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type User struct {
	AccessToken          string
	Login                 string
	Id                    string
}

func (this *User) Save(ctx context.Context) error {
	redis := getRedis(ctx)
	err := redis.Set(ctx, fmt.Sprintf("%s:login", this.Id), this.Login, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (this *User) Load(ctx context.Context, id string) error {

	redis := getRedis(ctx)
	login, err := redis.Get(ctx, fmt.Sprintf("%s:login", id)).Result()

	if err != nil {
		return err
	}
	if login == "" {
		return fmt.Errorf("Not Found")
	}

	accessToken, err := redis.Get(ctx, fmt.Sprintf("%s:access_token", id)).Result()
	this.Id = id
	this.Login = login
	this.AccessToken = accessToken

	if this.AccessToken == "" {
		return fmt.Errorf("No access token")
	}
	return nil
}

func (this *User) Register(code string) error {
	resp, err := http.Post(fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", mustGetEnv("CLIENT_ID"), mustGetEnv("CLIENT_SECRET"), code), "text/plain", strings.NewReader(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	access, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	for _, param := range strings.Split(string(access), "&") {
		parts := strings.Split(param, "=")
		key := parts[0]
		value := parts[1]
		if key == "access_token" {
			this.AccessToken = value
		}
	}
	return nil
}

func getUser(w http.ResponseWriter, r *http.Request, ctx context.Context) (User, error) {
	user := User{}

	authCookie, err := r.Cookie("Authorization")
	if err != nil {
		return user, fmt.Errorf("No Authorization Cookie")
	}

	auth := authCookie.Value
	if auth == "" {
		return user, fmt.Errorf("Invalid Authorization Cookie")
	}

	token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodRSA)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return ctx.Value("public_rsa_key"), nil
	})
	if err != nil {
		return user, fmt.Errorf("Invalid JWT token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["id"] == "" {
		return user, fmt.Errorf("JWT token doesn't contain id")
	}

	id, ok := claims["id"].(string)
	if !ok {
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
