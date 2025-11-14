package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/infra/database/in_memory/limit"
)

type LimitUseCaseTestSuite struct {
	suite.Suite
	LimitRepository *limit.InMemoryLimitRepository
	Sut             *LimitUseCase
}

func (suite *LimitUseCaseTestSuite) SetupTest() {
	LimitRepository := limit.NewInMemoryLimitRepository()
	suite.Sut = NewLimitUseCase(LimitRepository)
	suite.LimitRepository = LimitRepository
}

func (suite *LimitUseCaseTestSuite) TestLimitUseCase_Should_pass_one_single_request() {
	output, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             "IP",
		ReqsBySec:      5,
		BlockTimeBySec: 5,
	})
	suite.Nil(err)
	suite.True(output.pass)
}

func (suite *LimitUseCaseTestSuite) TestLimitUseCase_Should_pass_one_single_request_and_cachelimit_has_one() {
	myID := "IP"
	output, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      5,
		BlockTimeBySec: 5,
	})
	suite.Nil(err)
	suite.True(output.pass)
	suite.Equal(1, len(suite.Sut.CacheLimit))
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)

}

func (suite *LimitUseCaseTestSuite) TestLimitUseCase_Should_pass_two_same_requests_and_after_ten_seconds_update_limit_on_repository() {
	myID := "IP"
	reqsBySec := 5
	blockTimeBySec := 5

	output1, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output1.pass)
	suite.Equal(1, len(suite.Sut.CacheLimit))
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)

	output2, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output2.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(2), suite.Sut.CacheLimit[myID].Data.Counter)

	time.Sleep(15 * time.Second)

	suite.Equal(0, len(suite.Sut.CacheLimit))

	myLimit, err := suite.LimitRepository.GetLimitById(context.Background(), myID)
	suite.Nil(err)
	suite.Equal(myID, myLimit.Id)
	suite.Equal(int32(2), myLimit.Counter)
	suite.Nil(myLimit.FreeAt)
	suite.IsType(time.Time{}, myLimit.LastAt)
}

func TestLimitUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(LimitUseCaseTestSuite))
}
