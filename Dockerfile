FROM golang:1.12-alpine

LABEL "com.github.actions.name"="Condition-based Pull Request labeller" \
          "com.github.actions.description"="Automatically label pull requests based on rules" \
          "com.github.actions.icon"="award" \
          "com.github.actions.color"="blue" \
          "maintainer"="Galo Navarro <anglorvaroa@gmail.com>" \
          "repository"="https://github.com/srvaroa/labeler"

RUN apk add --no-cache git

WORKDIR /go/src/app
COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org
RUN go build -o action ./cmd
ENTRYPOINT ["/go/src/app/action"]
