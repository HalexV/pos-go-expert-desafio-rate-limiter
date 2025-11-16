package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/infra/database/redis/limit"
)

type LimitUseCaseRedisTestSuite struct {
	suite.Suite
	LimitRepository *limit.RedisLimitRepository
	Sut             *LimitUseCase
}

func (suite *LimitUseCaseRedisTestSuite) SetupTest() {
	LimitRepository := limit.NewRedisLimitRepository()
	suite.Sut = NewLimitUseCase(LimitRepository)
	suite.LimitRepository = LimitRepository
}

func (suite *LimitUseCaseRedisTestSuite) TearDownTest() {
	err := suite.LimitRepository.Rdb.FlushDB(context.Background()).Err()
	if err != nil {
		panic(err)
	}
}

func (suite *LimitUseCaseRedisTestSuite) TearDownSuite() {
	suite.LimitRepository.Rdb.Close()
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_one_single_request() {
	output, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             "IP",
		ReqsBySec:      5,
		BlockTimeBySec: 5,
	})
	suite.Nil(err)
	suite.True(output.pass)
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_one_single_request_and_cachelimit_has_one() {
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

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_two_same_requests_and_after_ten_seconds_update_limit_on_repository() {
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

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_block_after_five_same_requests_in_a_second() {
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
	suite.Nil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	output2, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output2.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(2), suite.Sut.CacheLimit[myID].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	output3, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output3.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(3), suite.Sut.CacheLimit[myID].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	output4, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output4.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(4), suite.Sut.CacheLimit[myID].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	output5, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.True(output5.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(5), suite.Sut.CacheLimit[myID].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	output6, err := suite.Sut.Execute(context.Background(), LimitInputDTO{
		Id:             myID,
		ReqsBySec:      int32(reqsBySec),
		BlockTimeBySec: int32(blockTimeBySec),
	})
	suite.Nil(err)
	suite.False(output6.pass)
	suite.Equal(myID, suite.Sut.CacheLimit[myID].Data.Id)
	suite.Equal(int32(1), suite.Sut.CacheLimit[myID].Data.Counter)
	suite.NotNil(suite.Sut.CacheLimit[myID].Data.FreeAt)

	time.Sleep(15 * time.Second)

	suite.Equal(0, len(suite.Sut.CacheLimit))

	myLimit, err := suite.LimitRepository.GetLimitById(context.Background(), myID)
	suite.Nil(err)
	suite.Equal(myID, myLimit.Id)
	suite.Equal(int32(1), myLimit.Counter)
	suite.NotNil(myLimit.FreeAt)
	suite.IsType(time.Time{}, myLimit.LastAt)
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_third_request_after_one_second() {
	limitInput := LimitInputDTO{
		Id:             "IP",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	output1, err := suite.Sut.Execute(context.Background(), limitInput)
	suite.Nil(err)
	suite.True(output1.pass)
	suite.Equal(1, len(suite.Sut.CacheLimit))
	suite.Equal(limitInput.Id, suite.Sut.CacheLimit[limitInput.Id].Data.Id)
	suite.Nil(suite.Sut.CacheLimit[limitInput.Id].Data.FreeAt)

	output2, err := suite.Sut.Execute(context.Background(), limitInput)
	suite.Nil(err)
	suite.True(output2.pass)
	suite.Equal(limitInput.Id, suite.Sut.CacheLimit[limitInput.Id].Data.Id)
	suite.Equal(int32(2), suite.Sut.CacheLimit[limitInput.Id].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[limitInput.Id].Data.FreeAt)

	time.Sleep(1 * time.Second)

	output3, err := suite.Sut.Execute(context.Background(), limitInput)
	suite.Nil(err)
	suite.True(output3.pass)
	suite.Equal(limitInput.Id, suite.Sut.CacheLimit[limitInput.Id].Data.Id)
	suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput.Id].Data.Counter)
	suite.Nil(suite.Sut.CacheLimit[limitInput.Id].Data.FreeAt)

}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_three_concurrent_requests_and_cachelimit_has_three() {
	limitInput1 := LimitInputDTO{
		Id:             "IP.01",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput2 := LimitInputDTO{
		Id:             "IP.02",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput3 := LimitInputDTO{
		Id:             "IP.03",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	testWG := &sync.WaitGroup{}

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output1, err := suite.Sut.Execute(context.Background(), limitInput1)
		suite.Nil(err)
		suite.True(output1.pass)
		suite.Equal(limitInput1.Id, suite.Sut.CacheLimit[limitInput1.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput1.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output2, err := suite.Sut.Execute(context.Background(), limitInput2)
		suite.Nil(err)
		suite.True(output2.pass)
		suite.Equal(limitInput2.Id, suite.Sut.CacheLimit[limitInput2.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput2.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {

		defer testWG.Done()

		output3, err := suite.Sut.Execute(context.Background(), limitInput3)
		suite.Nil(err)
		suite.True(output3.pass)
		suite.Equal(limitInput3.Id, suite.Sut.CacheLimit[limitInput3.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput3.Id].Data.Counter)
	}()

	testWG.Wait()

	suite.Equal(3, len(suite.Sut.CacheLimit))

}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_three_concurrent_requests_and_after_ten_seconds_update_limit_on_repository() {
	limitInput1 := LimitInputDTO{
		Id:             "IP.01",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput2 := LimitInputDTO{
		Id:             "IP.02",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput3 := LimitInputDTO{
		Id:             "IP.03",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	testWG := &sync.WaitGroup{}

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output1, err := suite.Sut.Execute(context.Background(), limitInput1)
		suite.Nil(err)
		suite.True(output1.pass)
		suite.Equal(limitInput1.Id, suite.Sut.CacheLimit[limitInput1.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput1.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output2, err := suite.Sut.Execute(context.Background(), limitInput2)
		suite.Nil(err)
		suite.True(output2.pass)
		suite.Equal(limitInput2.Id, suite.Sut.CacheLimit[limitInput2.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput2.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {

		defer testWG.Done()

		output3, err := suite.Sut.Execute(context.Background(), limitInput3)
		suite.Nil(err)
		suite.True(output3.pass)
		suite.Equal(limitInput3.Id, suite.Sut.CacheLimit[limitInput3.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput3.Id].Data.Counter)
	}()

	testWG.Wait()

	suite.Equal(3, len(suite.Sut.CacheLimit))

	time.Sleep(15 * time.Second)

	suite.Equal(0, len(suite.Sut.CacheLimit))

	myLimit1, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput1.Id)
	suite.Nil(err)
	suite.Equal(limitInput1.Id, myLimit1.Id)
	suite.Equal(int32(1), myLimit1.Counter)
	suite.IsType(&time.Time{}, myLimit1.FreeAt)
	suite.IsType(time.Time{}, myLimit1.LastAt)

	myLimit2, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput2.Id)
	suite.Nil(err)
	suite.Equal(limitInput2.Id, myLimit2.Id)
	suite.Equal(int32(1), myLimit2.Counter)
	suite.IsType(&time.Time{}, myLimit2.FreeAt)
	suite.IsType(time.Time{}, myLimit2.LastAt)

	myLimit3, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput3.Id)
	suite.Nil(err)
	suite.Equal(limitInput3.Id, myLimit3.Id)
	suite.Equal(int32(1), myLimit3.Counter)
	suite.IsType(&time.Time{}, myLimit3.FreeAt)
	suite.IsType(time.Time{}, myLimit3.LastAt)
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_pass_three_concurrent_requests_and_after_ten_seconds_update_limit_on_repository_and_pass_again_three_concurrent_requests() {
	limitInput1 := LimitInputDTO{
		Id:             "IP.01",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput2 := LimitInputDTO{
		Id:             "IP.02",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput3 := LimitInputDTO{
		Id:             "IP.03",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	testWG := &sync.WaitGroup{}

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output1, err := suite.Sut.Execute(context.Background(), limitInput1)
		suite.Nil(err)
		suite.True(output1.pass)
		suite.Equal(limitInput1.Id, suite.Sut.CacheLimit[limitInput1.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput1.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output2, err := suite.Sut.Execute(context.Background(), limitInput2)
		suite.Nil(err)
		suite.True(output2.pass)
		suite.Equal(limitInput2.Id, suite.Sut.CacheLimit[limitInput2.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput2.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {

		defer testWG.Done()

		output3, err := suite.Sut.Execute(context.Background(), limitInput3)
		suite.Nil(err)
		suite.True(output3.pass)
		suite.Equal(limitInput3.Id, suite.Sut.CacheLimit[limitInput3.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput3.Id].Data.Counter)
	}()

	testWG.Wait()

	suite.Equal(3, len(suite.Sut.CacheLimit))

	time.Sleep(15 * time.Second)

	suite.Equal(0, len(suite.Sut.CacheLimit))

	myLimit1, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput1.Id)
	suite.Nil(err)
	suite.Equal(limitInput1.Id, myLimit1.Id)
	suite.Equal(int32(1), myLimit1.Counter)
	suite.IsType(&time.Time{}, myLimit1.FreeAt)
	suite.IsType(time.Time{}, myLimit1.LastAt)

	myLimit2, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput2.Id)
	suite.Nil(err)
	suite.Equal(limitInput2.Id, myLimit2.Id)
	suite.Equal(int32(1), myLimit2.Counter)
	suite.IsType(&time.Time{}, myLimit2.FreeAt)
	suite.IsType(time.Time{}, myLimit2.LastAt)

	myLimit3, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput3.Id)
	suite.Nil(err)
	suite.Equal(limitInput3.Id, myLimit3.Id)
	suite.Equal(int32(1), myLimit3.Counter)
	suite.IsType(&time.Time{}, myLimit3.FreeAt)
	suite.IsType(time.Time{}, myLimit3.LastAt)

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output1, err := suite.Sut.Execute(context.Background(), limitInput1)
		suite.Nil(err)
		suite.True(output1.pass)
		suite.Equal(limitInput1.Id, suite.Sut.CacheLimit[limitInput1.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput1.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output2, err := suite.Sut.Execute(context.Background(), limitInput2)
		suite.Nil(err)
		suite.True(output2.pass)
		suite.Equal(limitInput2.Id, suite.Sut.CacheLimit[limitInput2.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput2.Id].Data.Counter)
	}()

	testWG.Add(1)
	go func() {

		defer testWG.Done()

		output3, err := suite.Sut.Execute(context.Background(), limitInput3)
		suite.Nil(err)
		suite.True(output3.pass)
		suite.Equal(limitInput3.Id, suite.Sut.CacheLimit[limitInput3.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput3.Id].Data.Counter)
	}()

	testWG.Wait()

	suite.Equal(3, len(suite.Sut.CacheLimit))
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_clean_cachelimit_after_ten_seconds_even_with_zero_requests() {

	time.Sleep(11 * time.Second)

	suite.Equal(0, len(suite.Sut.CacheLimit))
}

func (suite *LimitUseCaseRedisTestSuite) TestLimitUseCase_Should_lock_requests_while_clear_is_running_and_must_pass_normal_rate_request() {
	limitInput1 := LimitInputDTO{
		Id:             "IP.01",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput2 := LimitInputDTO{
		Id:             "IP.02",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	limitInput3 := LimitInputDTO{
		Id:             "IP.03",
		ReqsBySec:      2,
		BlockTimeBySec: 5,
	}

	testWG := &sync.WaitGroup{}

	stop := make(chan struct{})

	time.Sleep(9 * time.Second)

	go func() {

		for {
			select {
			case <-stop:
				return
			default:
				suite.Sut.Execute(context.Background(), limitInput2)
			}
		}

	}()

	go func() {

		for {
			select {
			case <-stop:
				return
			default:
				suite.Sut.Execute(context.Background(), limitInput3)
			}
		}

	}()

	time.Sleep(5 * time.Second)

	testWG.Add(1)
	go func() {
		defer testWG.Done()

		output1, err := suite.Sut.Execute(context.Background(), limitInput1)
		suite.Nil(err)
		suite.True(output1.pass)
		suite.Equal(limitInput1.Id, suite.Sut.CacheLimit[limitInput1.Id].Data.Id)
		suite.Equal(int32(1), suite.Sut.CacheLimit[limitInput1.Id].Data.Counter)
	}()

	testWG.Wait()

	stop <- struct{}{}
	stop <- struct{}{}

	suite.Equal(3, len(suite.Sut.CacheLimit))

	myLimit1, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput1.Id)
	suite.Nil(err)
	suite.Equal(limitInput1.Id, myLimit1.Id)
	suite.Equal(int32(1), myLimit1.Counter)
	suite.Nil(myLimit1.FreeAt)

	myLimit2, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput2.Id)
	suite.Nil(err)
	suite.Equal(limitInput2.Id, myLimit2.Id)
	suite.Equal(int32(1), myLimit2.Counter)
	suite.NotNil(myLimit2.FreeAt)

	myLimit3, err := suite.LimitRepository.GetLimitById(context.Background(), limitInput3.Id)
	suite.Nil(err)
	suite.Equal(limitInput3.Id, myLimit3.Id)
	suite.Equal(int32(1), myLimit3.Counter)
	suite.NotNil(myLimit3.FreeAt)
}

func TestLimitUseCaseRedisTestSuite(t *testing.T) {
	suite.Run(t, new(LimitUseCaseRedisTestSuite))
}
