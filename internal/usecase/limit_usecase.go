package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/entity/limit_entity"
)

type LimitInputDTO struct {
	Id             string
	ReqsBySec      int32
	BlockTimeBySec int32
}

type LimitOutputDTO struct {
	pass bool
}

type MapLimitValue struct {
	Data  *limit_entity.Limit
	Mutex *sync.Mutex
}

const TIMER_DURATION time.Duration = 10 * time.Second

type LimitUseCase struct {
	LimitRepository   limit_entity.LimitEntityRepository
	CacheLimit        map[string]*MapLimitValue
	CacheLimitClearWG *sync.WaitGroup
	UseCaseMutex      *sync.Mutex
	UseCaseWG         *sync.WaitGroup
	timer             *time.Timer
	ClearMutex        *sync.RWMutex
}

func NewLimitUseCase(LimitRepository limit_entity.LimitEntityRepository) *LimitUseCase {
	limitUseCase := &LimitUseCase{
		LimitRepository:   LimitRepository,
		CacheLimit:        make(map[string]*MapLimitValue),
		CacheLimitClearWG: &sync.WaitGroup{},
		UseCaseMutex:      &sync.Mutex{},
		UseCaseWG:         &sync.WaitGroup{},
		timer:             time.NewTimer(TIMER_DURATION),
		ClearMutex:        &sync.RWMutex{},
	}

	limitUseCase.triggerUpdateAndClearRoutine(context.Background())

	return limitUseCase
}

func (l *LimitUseCase) triggerUpdateAndClearRoutine(ctx context.Context) {
	go func() {

		for {
			select {
			case <-l.timer.C:
				println("Ativei o clear")
				// Espera todas as execuções que já começaram do execute terminem
				l.UseCaseWG.Wait()

				l.ClearMutex.Lock()
				println("Verificando se o cache é zero")
				if len(l.CacheLimit) == 0 {
					println("o cache é zero")
					l.ClearMutex.Unlock()
					continue
				}
				l.ClearMutex.Unlock()

				println("O cache não é zero")

				l.ClearMutex.Lock()
				println("Processando o cache")
				for k, v := range l.CacheLimit {
					if err := l.LimitRepository.UpdateLimitById(ctx, v.Data.Id, v.Data); err != nil {
						fmt.Printf("Erro ao atualizar registro de ID %s\n", v.Data.Id)
					}
					delete(l.CacheLimit, k)
				}
				println("Terminou o processamento do cache")
				l.ClearMutex.Unlock()

				l.timer.Reset(TIMER_DURATION)

			}
		}
	}()
}

func (l *LimitUseCase) Execute(ctx context.Context, input LimitInputDTO) (LimitOutputDTO, error) {
	println("Execute: começou")
	defer println("Execute: terminou")
	l.ClearMutex.RLock()
	defer l.ClearMutex.RUnlock()

	// Comecei a executar
	l.UseCaseWG.Add(1)
	defer l.UseCaseWG.Done()

	// Trava por conta da hipótese do limit value não estar no cache
	l.UseCaseMutex.Lock()

	mapLimitValue, ok := l.CacheLimit[input.Id]

	// Não está no cache
	if !ok {
		limitData, err := l.LimitRepository.GetLimitById(ctx, input.Id)
		if err != nil {
			l.UseCaseMutex.Unlock()
			return LimitOutputDTO{pass: false}, err
		}
		// Not found, create
		if limitData == nil {
			newLimitData := &limit_entity.Limit{
				Id:      input.Id,
				FreeAt:  nil,
				LastAt:  time.Now(),
				Counter: 1,
			}
			err = l.LimitRepository.CreateLimit(ctx, newLimitData)
			if err != nil {
				l.UseCaseMutex.Unlock()
				return LimitOutputDTO{pass: false}, err
			}
			l.CacheLimit[input.Id] = &MapLimitValue{
				Data:  newLimitData,
				Mutex: &sync.Mutex{},
			}

			l.UseCaseMutex.Unlock()
			return LimitOutputDTO{pass: true}, nil
		}

		// Não está no cache mas está no repository
		mapLimitValue = &MapLimitValue{
			Data: &limit_entity.Limit{
				Id:      limitData.Id,
				FreeAt:  limitData.FreeAt,
				LastAt:  limitData.LastAt,
				Counter: limitData.Counter,
			},
			Mutex: &sync.Mutex{},
		}
		l.CacheLimit[input.Id] = mapLimitValue
	}
	l.UseCaseMutex.Unlock()

	mapLimitValue.Mutex.Lock()
	defer mapLimitValue.Mutex.Unlock()
	// Está com bloqueio
	if mapLimitValue.Data.FreeAt != nil {
		// Já passou o tempo de bloqueio
		if mapLimitValue.Data.FreeAt.Before(time.Now()) {
			*l.CacheLimit[input.Id].Data = limit_entity.Limit{
				Id:      mapLimitValue.Data.Id,
				FreeAt:  nil,
				LastAt:  time.Now(),
				Counter: 1,
			}

			return LimitOutputDTO{pass: true}, nil
		}
		// Não passou o tempo de bloqueio
		return LimitOutputDTO{pass: false}, nil
	}

	// Passou um segundo sem requisição
	if time.Since(mapLimitValue.Data.LastAt) > time.Second {

		*l.CacheLimit[input.Id].Data = limit_entity.Limit{
			Id:      mapLimitValue.Data.Id,
			FreeAt:  nil,
			LastAt:  time.Now(),
			Counter: 1,
		}

		return LimitOutputDTO{pass: true}, nil
	}

	// Atingiu o máximo de requisições por segundo
	if mapLimitValue.Data.Counter+1 > input.ReqsBySec {
		t := time.Now().Add(time.Duration(input.BlockTimeBySec) * time.Second)

		*l.CacheLimit[input.Id].Data = limit_entity.Limit{
			Id:      mapLimitValue.Data.Id,
			FreeAt:  &t,
			LastAt:  time.Now(),
			Counter: 1,
		}

		return LimitOutputDTO{pass: false}, nil
	}

	// Incrementa o counter e ok
	*l.CacheLimit[input.Id].Data = limit_entity.Limit{
		Id:      mapLimitValue.Data.Id,
		FreeAt:  nil,
		LastAt:  time.Now(),
		Counter: mapLimitValue.Data.Counter + 1,
	}

	return LimitOutputDTO{pass: true}, nil

}
