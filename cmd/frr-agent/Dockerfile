FROM golang:1.17 as dist
ARG GIT_SHA=unknown
ARG GIT_BRANCH=unknown
ARG GIT_TAG=unknown
ARG BUILD_DATE=unknown
WORKDIR /opt
COPY ./ ./
RUN CGO_ENABLED=0 go build -o ./bin/frr-agent -ldflags "\
  -X github.com/slankdev/vtyang/pkg/util.gitSHA=$GIT_SHA \
  -X github.com/slankdev/vtyang/pkg/util.gitBranch=$GIT_BRANCH \
  -X github.com/slankdev/vtyang/pkg/util.gitTag=$GIT_TAG \
  -X github.com/slankdev/vtyang/pkg/util.buildDate=$BUILD_DATE \
  " ./cmd/frr-agent/main.go

FROM busybox
LABEL org.opencontainers.image.source https://github.com/slankdev/vtyang
COPY --from=dist /opt/bin/frr-agent /usr/bin/
