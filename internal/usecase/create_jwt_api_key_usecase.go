package usecase

import "github.com/HalexV/pos-go-expert-desafio-rate-limiter/pkg/entity"

type CreateJWTAPIKeyInputDTO struct {
	MaxReqsBySec   int32 `json:"max_reqs_by_sec"`
	BlockTimeBySec int32 `json:"block_time_by_sec"`
}

type CreateJWTAPIKeyOutputDTO struct {
	ID             entity.ID `json:"id"`
	MaxReqsBySec   int32     `json:"max_reqs_by_sec"`
	BlockTimeBySec int32     `json:"block_time_by_sec"`
}

type CreateJWTAPIKeyUseCase struct{}

func NewCreateJWTAPIKeyUseCase() *CreateJWTAPIKeyUseCase {
	return &CreateJWTAPIKeyUseCase{}
}

func (c *CreateJWTAPIKeyUseCase) Execute(input CreateJWTAPIKeyInputDTO) (CreateJWTAPIKeyOutputDTO, error) {

	dto := CreateJWTAPIKeyOutputDTO{
		ID:             entity.NewID(),
		MaxReqsBySec:   input.MaxReqsBySec,
		BlockTimeBySec: input.BlockTimeBySec,
	}

	return dto, nil
}
