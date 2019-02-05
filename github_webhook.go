/* This is free and unencumbered software released into the public domain. */

// github_webhook posts commits to https://twitter.com/ConrealityCode
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
)

const maxTweetLength = 280

type PingRequest struct {
	// no fields of importance
}

type PushRequest struct {
	HeadCommit PushCommit   `json:"head_commit"`
	Commits    []PushCommit `json:"commits"`
}

type PushCommit struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Message   string    `json:"message"`
	Author    GitAuthor `json:"author"`
	Committer GitAuthor `json:"committer"`
}

type GitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func parsePingRequest(request events.APIGatewayProxyRequest) (*PingRequest, error) {
	var payload PingRequest
	err := json.Unmarshal([]byte(request.Body), &payload)
	if err != nil {
		return nil, errors.Wrap(err, "parsePingRequest failed")
	}
	return &payload, nil
}

func parsePushRequest(request events.APIGatewayProxyRequest) (*PushRequest, error) {
	var payload PushRequest
	err := json.Unmarshal([]byte(request.Body), &payload)
	if err != nil {
		return nil, errors.Wrap(err, "parsePushRequest failed")
	}
	return &payload, nil
}

func postTweet(tweetBody string) error {
	config := oauth1.NewConfig(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_SECRET"))
	token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	_, _, err := client.Statuses.Update(tweetBody, nil)
	if err != nil {
		return errors.Wrap(err, "postTweet failed")
	}
	return nil
}

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	event := request.Headers["X-GitHub-Event"]

	log.Print("Handling request with X-GitHub-Event=", event)
	log.Print(request)

	if event == "ping" {
		payload, err := parsePingRequest(request)
		log.Print(payload)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    map[string]string{"content-type": "text/plain"},
				Body:       fmt.Sprintf("Failed to parse a ping request:\n%s\n", err),
			}, nil
		}
		return events.APIGatewayProxyResponse{Body: "OK\n", StatusCode: 200}, nil
	}

	if event == "push" {
		payload, err := parsePushRequest(request)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    map[string]string{"content-type": "text/plain"},
				Body:       fmt.Sprintf("Failed to parse a push request:\n%s\n", err),
			}, nil
		}
		commit := payload.HeadCommit
		commitMessage := commit.Message
		fullMessage := []rune(commitMessage)
		maxMessageLength := maxTweetLength - len(commit.ID) - len(commit.Author.Name) - len(commit.URL) - 8
		if len(fullMessage) > maxMessageLength {
			commitMessage = string(fullMessage[:maxMessageLength]) + "\u2026"
		}
		tweetBody := fmt.Sprintf("%s by %s: %s\n%s\n", commit.ID, commit.Author.Name, commitMessage, commit.URL)
		log.Printf(tweetBody)
		postTweet(tweetBody)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    map[string]string{"content-type": "text/plain"},
			Body:       tweetBody,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    map[string]string{"content-type": "text/plain"},
		Body:       fmt.Sprintf("Failed to grok the X-GitHub-Event header: %s\n", event),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
