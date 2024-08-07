FROM golang:1.22 AS install
WORKDIR /go/src/app
COPY e2e/go.mod e2e/go.mod
COPY e2e/go.sum e2e/go.sum
COPY go.* .
RUN go mod download \
  && go install github.com/mitranim/gow@latest

FROM install AS build
COPY gen gen
COPY internal internal
COPY main.go main.go
RUN CGO_ENABLED=0 go build -o /go/bin/app github.com/basemind-ai/monorepo/services/$BUILD_TARGET \
  && chmod +x /go/bin/app

# hadolint ignore=DL3007
FROM gcr.io/distroless/static-debian12:latest AS app
COPY --from=build /go/bin/app /
CMD ["/app"]
