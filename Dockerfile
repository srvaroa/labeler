FROM alpine:3.17.2

LABEL "com.github.actions.name"="Condition-based Pull Request labeller" \
          "com.github.actions.description"="Automatically label pull requests based on rules" \
          "com.github.actions.icon"="award" \
          "com.github.actions.color"="blue" \
          "maintainer"="Galo Navarro <anglorvaroa@gmail.com>" \
          "repository"="https://github.com/srvaroa/labeler"

WORKDIR /
ARG ASSET_URL=https://github.com/srvaroa/labeler/releases/download/v1.8.1/action.tar.gz
RUN wget -q -O- $ASSET_URL | tar xzvf -
ENTRYPOINT ["/action"]
