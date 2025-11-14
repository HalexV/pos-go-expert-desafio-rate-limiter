package limit

import (
	"context"
	"errors"
	"sync"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/entity/limit_entity"
)

type InMemoryLimitRepository struct {
	Db    map[string]*limit_entity.Limit
	Mutex *sync.Mutex
}

func NewInMemoryLimitRepository() *InMemoryLimitRepository {
	return &InMemoryLimitRepository{
		Db:    make(map[string]*limit_entity.Limit),
		Mutex: &sync.Mutex{},
	}
}

func (imdb *InMemoryLimitRepository) CreateLimit(ctx context.Context, limit *limit_entity.Limit) error {
	imdb.Mutex.Lock()
	defer imdb.Mutex.Unlock()

	imdb.Db[limit.Id] = &limit_entity.Limit{
		Id:      limit.Id,
		FreeAt:  limit.FreeAt,
		LastAt:  limit.LastAt,
		Counter: limit.Counter,
	}
	return nil
}

func (imdb *InMemoryLimitRepository) GetLimitById(ctx context.Context, id string) (*limit_entity.Limit, error) {
	println("getlimit: Iniciando")
	imdb.Mutex.Lock()
	defer imdb.Mutex.Unlock()

	println("getlimit: Antes de acessar o map")
	limit, ok := imdb.Db[id]
	println("getlimit: Depois de acessar o map")
	if !ok {
		println("getlimit: Limit não encontrado")
		return nil, nil
	}

	println("getlimit: Retornando o limit")
	return &limit_entity.Limit{
		Id:      limit.Id,
		FreeAt:  limit.FreeAt,
		LastAt:  limit.LastAt,
		Counter: limit.Counter,
	}, nil
}

func (imdb *InMemoryLimitRepository) UpdateLimitById(ctx context.Context, id string, newLimit *limit_entity.Limit) error {
	println("repository update começando")
	imdb.Mutex.Lock()
	defer imdb.Mutex.Unlock()

	limit, ok := imdb.Db[id]
	if !ok {
		println("repository limit não encontrado")
		return errors.New("limit not found")
	}

	*limit = limit_entity.Limit{
		Id:      newLimit.Id,
		FreeAt:  newLimit.FreeAt,
		LastAt:  newLimit.LastAt,
		Counter: newLimit.Counter,
	}

	println("repository update terminando")

	return nil
}
