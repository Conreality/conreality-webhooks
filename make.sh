#!/bin/bash
GOOS=linux go build github_webhook.go
zip -9 github_webhook.zip ./github_webhook
