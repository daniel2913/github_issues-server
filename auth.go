package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
)

type Viewer struct {
	Data struct {
		Viewer struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"viewer"`
	} `json:"data"`
}

func (this *User) RefreshInfo(ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	reqbody, err := makeValidRequestBody(QueryUserInfo)

	if err != nil {
		reqctx.status = http.StatusInternalServerError
		return err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", reqbody)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		log.Error().Err(err).Msg("Error while creating user info request")
		return err
	}

	req.Header.Set("Authorization", "Bearer "+this.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode/100 != 2 {
		reqctx.status = http.StatusUnauthorized
		log.Error().Err(err).Msg("Error while making user info request")
		return err
	}
	defer resp.Body.Close()
	viewer := Viewer{}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		log.Error().Err(err).Msg("Bad response from user info request")
		return err
	}
	err = json.Unmarshal(bytes, &viewer)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		log.Error().Err(err).Msg("Bad response json from user info request")
		return err
	}
	this.Login = viewer.Data.Viewer.Login
	this.Id = viewer.Data.Viewer.ID
	log.Debug().Msgf("Got user info for %s", this.Login)
	return nil
}

func (this *User) JWT(ctx context.Context) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"id": this.Id,
	})
	return token.SignedString(ctx.Value("private_rsa_key"))
}

func authenticate(w http.ResponseWriter, r *http.Request, ctx context.Context) (string, string, error) {
	user, err := getUser(w, r, ctx)
	if err != nil {
		return "Bearer " + mustGetEnv("DEFAULT_TOKEN"), "", nil
	}
	token := user.AccessToken
	return "Bearer " + token, user.Login, nil
}
