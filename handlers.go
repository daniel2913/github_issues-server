package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)



func corsHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	return
}

func queryCommentsHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	qp := r.URL.Query()
	owner := qp.Get("owner")
	repoName := qp.Get("repo_name")
	issueNumber := qp.Get("issue_number")
	after := qp.Get("after")
	if owner == "" || repoName == "" || issueNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if after == "" || after == "null" {
		after = "null"
	} else {
		after = fmt.Sprintf(`"%s"`, after)
	}
	query := fmt.Sprintf(QueryComments, owner, repoName, issueNumber, after)
	err := relayRequest(w, query, ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Cache-Control", "no-store")
}

func queryIssuesHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	qp := r.URL.Query()
	owner := qp.Get("owner")
	repoName := qp.Get("repo_name")
	issueNumber := qp.Get("issue_number")
	after := qp.Get("after")
	if owner == "" || repoName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
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
	err := relayRequest(w, query, ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func queryReposHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, login := getAuthorization(ctx)
		query = fmt.Sprintf(QueryPersonalRepos, login, filter, after)
	}
	err := relayRequest(w, query, ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func createCommentHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	subj := r.URL.Query().Get("subject_id")
	body := r.FormValue("body")
	_, login := getAuthorization(ctx)
	if subj == "" || body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := relayRequest(w, fmt.Sprintf(PostComment, subj, login, body), ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
