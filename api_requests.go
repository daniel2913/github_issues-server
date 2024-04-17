package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func relayRequest(w http.ResponseWriter, query string, ctx context.Context) error {
	reqbody, err := makeValidRequestBody(query)
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", reqbody)
	if err != nil {
		return err
	}

	auth, _ := getAuthorization(ctx)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	hasErrors := bytes.HasPrefix(body, []byte(`{"errors":[`))
	if hasErrors {
		w.WriteHeader(resp.StatusCode)
		return nil
	}
	_, err = w.Write(body)
	if err != nil {
		return err
	}
	w.WriteHeader(resp.StatusCode)
	w.Header().Add("Cache-Control", "max-age=600")
	w.Header().Add("Content-Type", "application/json")
	return nil
}

var (
	QueryRepos = `query {
    search(query: "%s", type: REPOSITORY, first: 10, after: %s) {
      edges {
        node {
          ... on Repository` + selectRepoFields + `
			}
		}` + pageInfo + `
	}
}`

	QueryPersonalRepos = `{user(login: "%s") {
		%s(first: 10, after:%s) {
      nodes` + selectRepoFields + pageInfo + `
		}
	}
	}
	`
	QueryIssue = `query {
		repository(owner: "%s", name: "%s") {
			issue(number: %s) {
				author {
					login
					avatarUrl
					url
				}
				id
				titleHTML
				bodyHTML
				createdAt
			}
  	}
	}`
	QueryIssues = `query {repository(owner: "%s", name: "%s") {
	issues(first: 10, orderBy:{field:CREATED_AT, direction: DESC}, after: %s, states: OPEN) {
		nodes {
			titleHTML
			number
			url
			createdAt
			author {
				login
				avatarUrl
				url
			}
			comments{
				totalCount
			}
		}
		pageInfo {
			hasNextPage
			endCursor
		}
	}
	}}`
	QueryComments = `query {
  repository(owner: "%s", name: "%s") {
    issue(number: %s) {
	comments(first: 10 after:%s) {
        nodes {
          bodyHTML
					createdAt
          author {
            login
						avatarUrl
						url
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}`
	selectRepoFields = `{
            name
						url
            issues {
              totalCount
            }
            stargazers {
              totalCount
            }
						descriptionHTML
						owner {
							id,
							login,
							url,
							avatarUrl
						}
      }`

	pageInfo = `pageInfo {
				endCursor
				hasNextPage
			}`
	PostComment = `mutation {
	addComment(input: {subjectId: "%s", clientMutationId: "%s" body: "%s"}) {
		commentEdge {
			node {
				id
			}
		}
	}
}`
	QueryUserInfo = `query{viewer {id login}}`
)
