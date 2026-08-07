FROM golang:1.24 AS build
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /app main.go

FROM scratch
COPY --from=build /app /app
ENTRYPOINT ["/app"]
