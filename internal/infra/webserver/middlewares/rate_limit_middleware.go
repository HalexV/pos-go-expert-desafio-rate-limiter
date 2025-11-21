package middlewares

import (
	"fmt"
	"net"
	"net/http"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/entity/limit_entity"
	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/infra/database/redis/limit"
	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/usecase"
	"github.com/go-chi/jwtauth"
)

type RepositoryStrategy string

const (
	StrategyUnknown RepositoryStrategy = ""
	StrategyRedis   RepositoryStrategy = "redis"
)

type RateLimitMiddleware struct {
	ipRateLimit      bool
	ipMaxReqsBySec   int32
	ipBlockTimeBySec int32
	tokenRateLimit   bool
	limitUseCase     *usecase.LimitUseCase
}

func (rtlt *RateLimitMiddleware) ReturnRateLimitHandler() func(next http.Handler) http.Handler {

	if rtlt.ipRateLimit && rtlt.tokenRateLimit {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				var id string
				var reqsBySec, blockTimeBySec int32

				_, claims, _ := jwtauth.FromContext(r.Context())

				if len(claims) > 0 {

					jwtSub, ok := claims["sub"].(string)
					if !ok {
						panic("jwt sub property does not exist")
					}
					id = jwtSub

					jwtMaxReqsBySec, ok := claims["maxReqsBySec"].(float64)
					if !ok {
						panic("jwt maxReqsBySec property does not exist")
					}
					reqsBySec = int32(jwtMaxReqsBySec)

					jwtBlockTimeBySec, ok := claims["blockTimeBySec"].(float64)
					if !ok {
						panic("jwt blockTimeBySec property does not exist")
					}

					blockTimeBySec = int32(jwtBlockTimeBySec)

				} else {
					ip, _, err := net.SplitHostPort(r.RemoteAddr)
					if err == nil {
						id = ip
					} else {
						id = r.RemoteAddr
					}

					reqsBySec = rtlt.ipMaxReqsBySec
					blockTimeBySec = rtlt.ipBlockTimeBySec

				}

				result, err := rtlt.limitUseCase.Execute(r.Context(), usecase.LimitInputDTO{
					Id:             id,
					ReqsBySec:      reqsBySec,
					BlockTimeBySec: blockTimeBySec,
				})
				if err != nil {
					fmt.Printf("Erro no limit use case: %s\n", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !result.Pass {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
					return
				}

				next.ServeHTTP(w, r)
			})
		}
	}

	if rtlt.tokenRateLimit {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				var id string
				var reqsBySec, blockTimeBySec int32

				_, claims, _ := jwtauth.FromContext(r.Context())

				if len(claims) > 0 {

					jwtSub, ok := claims["sub"].(string)
					if !ok {
						panic("jwt sub property does not exist")
					}
					id = jwtSub

					jwtMaxReqsBySec, ok := claims["maxReqsBySec"].(int32)
					if !ok {
						panic("jwt maxReqsBySec property does not exist")
					}
					reqsBySec = jwtMaxReqsBySec

					jwtBlockTimeBySec, ok := claims["blockTimeBySec"].(int32)
					if !ok {
						panic("jwt blockTimeBySec property does not exist")
					}

					blockTimeBySec = jwtBlockTimeBySec
				} else {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("something wrong with your token"))
					return
				}

				result, err := rtlt.limitUseCase.Execute(r.Context(), usecase.LimitInputDTO{
					Id:             id,
					ReqsBySec:      reqsBySec,
					BlockTimeBySec: blockTimeBySec,
				})
				if err != nil {
					fmt.Printf("Erro no limit use case: %s\n", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if !result.Pass {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
					return
				}

				next.ServeHTTP(w, r)
			})
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var id string
			var reqsBySec, blockTimeBySec int32

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				id = ip
			} else {
				id = r.RemoteAddr
			}

			reqsBySec = rtlt.ipMaxReqsBySec
			blockTimeBySec = rtlt.ipBlockTimeBySec

			result, err := rtlt.limitUseCase.Execute(r.Context(), usecase.LimitInputDTO{
				Id:             id,
				ReqsBySec:      reqsBySec,
				BlockTimeBySec: blockTimeBySec,
			})
			if err != nil {
				fmt.Printf("Erro no limit use case: %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !result.Pass {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type RateLimitMiddlewareBuilder struct {
	ipRateLimit        bool
	ipMaxReqsBySec     int32
	ipBlockTimeBySec   int32
	tokenRateLimit     bool
	repositoryStrategy RepositoryStrategy
	limitRepository    limit_entity.LimitEntityRepository
}

func NewRateLimitMiddlewareBuilder() *RateLimitMiddlewareBuilder {
	return &RateLimitMiddlewareBuilder{}
}

func (b *RateLimitMiddlewareBuilder) WithRateLimitByIP(ipMaxReqsBySec int32, ipBlockTimeBySec int32) *RateLimitMiddlewareBuilder {
	b.ipRateLimit = true
	b.ipMaxReqsBySec = ipMaxReqsBySec
	b.ipBlockTimeBySec = ipBlockTimeBySec

	return b
}

func (b *RateLimitMiddlewareBuilder) WithRateLimitByToken() *RateLimitMiddlewareBuilder {
	b.tokenRateLimit = true

	return b
}

func (b *RateLimitMiddlewareBuilder) WithRedis(host string, port string) *RateLimitMiddlewareBuilder {

	if b.repositoryStrategy != StrategyUnknown {
		panic("Strategy já selecionada!")
	}

	b.repositoryStrategy = StrategyRedis
	b.limitRepository = limit.NewRedisLimitRepository(host, port)

	return b

}

func (b *RateLimitMiddlewareBuilder) Build() func(next http.Handler) http.Handler {
	if b.repositoryStrategy == StrategyUnknown {
		panic("Nenhuma strategy válida selecionada!")
	}

	rateLimitMiddleware := RateLimitMiddleware{
		ipRateLimit:      b.ipRateLimit,
		ipMaxReqsBySec:   b.ipMaxReqsBySec,
		ipBlockTimeBySec: b.ipBlockTimeBySec,
		tokenRateLimit:   b.tokenRateLimit,
		limitUseCase:     usecase.NewLimitUseCase(b.limitRepository),
	}

	return rateLimitMiddleware.ReturnRateLimitHandler()
}
