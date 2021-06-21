FROM golang:1.16-alpine

LABEL "com.github.actions.name"="Condition-based Pull Request labeller"
LABEL "com.github.actions.description"="Automatically label pull requests based on rules"
LABEL "com.github.actions.icon"="award"
LABEL "com.github.actions.color"="blue"
LABEL "maintainer"="Galo Navarro <anglorvaroa@gmail.com>"
LABEL "repository"="https://github.com/srvaroa/labeler"

RUN apk add --no-cache git

WORKDIR /go/src/app
COPY . .
ENV GO111MODULE=on
ENV GO15VENDOREXPERIMENT=1
RUN go build -o action ./cmd
ENTRYPOINT ["/go/src/app/action"]
