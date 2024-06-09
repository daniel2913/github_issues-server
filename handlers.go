package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func corsHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	reqctx.status = http.StatusNoContent
	return nil
}

func queryCommentsHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	qp := r.URL.Query()
	owner := qp.Get("owner")
	repoName := qp.Get("repo_name")
	issueNumber := qp.Get("issue_number")
	after := qp.Get("after")
	if owner == "" || repoName == "" || issueNumber == "" {
		err := fmt.Errorf("Bad request")
		log.Debug().Msgf("Request from %s didn't provide query params", r.RemoteAddr)
		reqctx.status = http.StatusBadRequest
		return err
	}

	if after == "" || after == "null" {
		after = "null"
	} else {
		after = fmt.Sprintf(`"%s"`, after)
	}
	query := fmt.Sprintf(QueryComments, owner, repoName, issueNumber, after)
	w.Header().Set("Cache-Control", "no-store")
	return relayRequest(w, query, ctx)
}

func queryIssuesHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	qp := r.URL.Query()
	owner := qp.Get("owner")
	repoName := qp.Get("repo_name")
	issueNumber := qp.Get("issue_number")
	after := qp.Get("after")
	if owner == "" || repoName == "" {
		reqctx.status = http.StatusBadRequest
		log.Debug().Msgf("Request from %s didn't provide query params", r.RemoteAddr)
		return fmt.Errorf("Bad Request")
	}
	if after == "" || after == "null" {
		after = "null"
	} else {
		after = fmt.Sprintf(`"%s"`, after)
	}
	query := ""
	if issueNumber == "" {
		query = fmt.Sprintf(QueryIssues, owner, repoName, after)
	} else {
		query = fmt.Sprintf(QueryIssue, owner, repoName, issueNumber)
	}
	return relayRequest(w, query, ctx)
}

func queryReposHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	params := strings.Split(r.URL.Path, "/")
	param := params[len(params)-1]
	query := ""
	search, after := getQuerySearchAfter(r)
	if search == "" {
		search = "stars:>1"
	}

	if param == "all" {
		query = fmt.Sprintf(QueryRepos, search, after)
	} else {
		filter := ""
		switch param {
		case "watch":
			filter = "watching"
		case "star":
			filter = "starredRepositories"
		case "own":
			filter = "repositories"
		default:
			log.Debug().Msgf("Request from %s didn't provide query params", r.RemoteAddr)
			reqctx.status = http.StatusBadRequest
			return fmt.Errorf("Bad request")
		}
		login := reqctx.login
		query = fmt.Sprintf(QueryPersonalRepos, login, filter, after)
	}
	return relayRequest(w, query, ctx)
}

func createCommentHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	reqctx := mustGetReqContext(ctx)
	subj := r.URL.Query().Get("subject_id")
	body := r.FormValue("body")
	login := reqctx.login
	if subj == "" || body == "" {
		log.Debug().Msgf("Request from %s didn't provide comment parameters", r.RemoteAddr)
		reqctx.status = http.StatusBadRequest
		return fmt.Errorf("Bad request")
	}
	return relayRequest(w, fmt.Sprintf(PostComment, subj, login, body), ctx)
}
