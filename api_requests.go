package main

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func relayRequest(w http.ResponseWriter, query string, ctx context.Context) error {
	reqbody, err := makeValidRequestBody(query)
	reqctx := mustGetReqContext(ctx)
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", reqbody)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		log.Error().Err(err).Str("GraphQL request", query).Msgf("Error while creating request to data endpoint")
		return err
	}
	auth := reqctx.token

	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		reqctx.status = http.StatusServiceUnavailable
		log.Error().Err(err).Str("GraphQL request", query).Msgf("Error while making request to data endpoint")
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		reqctx.status = resp.StatusCode
		log.Error().Err(err).Str("GraphQL request", query).Msgf("Bad response from data endpoint")
		return err
	}
	hasErrors := bytes.HasPrefix(body, []byte(`{"errors":[`))
	if hasErrors {
		log.Debug().Str("API Response", string(body)).Str("GraphQL request", query).Msgf("Request to data endpoint returned an error")
		reqctx.status = resp.StatusCode
		return nil
	}
	_, err = w.Write(body)
	if err != nil {
		reqctx.status = http.StatusInternalServerError
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	reqctx.cache = "max-age=600"
	reqctx.status = -1
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
	}}
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
