package main

import (
	"fmt"
	"net/http"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/configs"
	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/infra/webserver/handlers"
	myMiddlewares "github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/infra/webserver/middlewares"
	jwtcustomverifiers "github.com/HalexV/pos-go-expert-desafio-rate-limiter/pkg/jwt-custom-verifiers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	rateLimitMiddleware := myMiddlewares.NewRateLimitMiddlewareBuilder()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.WithValue("jwt", configs.TokenAuth))
	r.Use(middleware.WithValue("jwtExpiresIn", configs.JWTExpiresIn))

	r.Route("/rate-limit", func(r chi.Router) {
		r.Use(jwtauth.Verify(configs.TokenAuth, jwtcustomverifiers.VerifyApiKeyHeader))
		// r.Use(jwtauth.Authenticator)
		r.Use(
			rateLimitMiddleware.
				WithRateLimitByIP(configs.IpMaxReqsBySec, configs.IpBlockTimeBySec).
				WithRateLimitByToken().
				WithRedis(configs.RedisHost, configs.RedisPort).
				Build())
		r.Get("/", handlers.NewAnyHandler().GetAny)
	})

	r.Post("/generate_token", handlers.NewJWTAPIKeyHandler().CreateJWTAPIKey)

	http.ListenAndServe(fmt.Sprintf(":%s", configs.WebServerPort), r)
}
