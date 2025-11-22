FROM golang:1.24.3 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY configs ./configs
COPY internal ./internal
COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /webserver ./cmd/server

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -failfast -run ^TestLimitUseCaseTestSuite$ ./internal/usecase

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=run-test-stage /webserver /webserver

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/webserver"]