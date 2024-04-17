package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type Viewer struct {
	Data struct {
		Viewer struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"viewer"`
	} `json:"data"`
}

func (this *User) RefreshInfo() error {
	reqbody, err := makeValidRequestBody(QueryUserInfo)
	if err != nil {
		panic("Bad Constant User Request!")
	}
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", reqbody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+this.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode/100 != 2 {
		return err
	}
	defer resp.Body.Close()
	viewer := Viewer{}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &viewer)
	if err != nil {
		return err
	}
	this.Login = viewer.Data.Viewer.Login
	this.Id = viewer.Data.Viewer.ID
	return nil
}

func (this *User) JWT(ctx context.Context) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id": this.Id,
	})
	return token.SignedString(ctx.Value("private_rsa_key"))
}

func getAuthorization(ctx context.Context) (string, string) {
	token, ok := ctx.Value("authorization").(string)
	if !ok {
		panic("NO AUTHORIZATION IN CONTEXT")
	}
	login, ok := ctx.Value("login").(string)
	if !ok {
		panic("NO AUTHORIZATION IN CONTEXT")
	}
	return token, login
}

func authenticate(w http.ResponseWriter, r *http.Request, ctx context.Context) (string, string, error) {
	user, err := getUser(w, r, ctx)
	if err != nil {
		return "Bearer " + mustGetEnv("DEFAULT_TOKEN"), "", nil
	}
	token := user.AccessToken
	return "Bearer " + token, user.Login, nil
}
