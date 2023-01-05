ARG GOLANG_VER=latest
ARG ALPINE_VER=latest

FROM golang:${GOLANG_VER} as builder
WORKDIR /go/src/app
COPY go.* *.go ./
COPY cmd cmd/
ENV CGO_ENABLED 0
ARG ACTIONLINT_VER=
RUN go build -v -ldflags "-s -w -X github.com/rhysd/actionlint.version=${ACTIONLINT_VER}" ./cmd/actionlint

FROM koalaman/shellcheck-alpine:stable as shellcheck

FROM alpine:${ALPINE_VER}
COPY --from=builder /go/src/app/actionlint /usr/local/bin/
COPY --from=shellcheck /bin/shellcheck /usr/local/bin/shellcheck
RUN apk add --no-cache py3-pyflakes
USER guest
ENTRYPOINT ["/usr/local/bin/actionlint"]
