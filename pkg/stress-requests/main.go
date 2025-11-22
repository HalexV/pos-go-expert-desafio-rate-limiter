package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type GenerateTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type GenerateTokenBody struct {
	MaxReqsBySec   uint `json:"max_reqs_by_sec"`
	BlockTimeBySec uint `json:"block_time_by_sec"`
}

func main() {

	const REQUESTS uint = 25

	wg := &sync.WaitGroup{}

	for range REQUESTS {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := GenerateTokenBody{
				MaxReqsBySec:   10,
				BlockTimeBySec: 60,
			}

			b, err := json.Marshal(body)
			if err != nil {
				panic(err)
			}

			req, err := http.NewRequest("POST", "http://localhost:8080/generate_token", bytes.NewBuffer(b))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Println("Erro: ", err)
				panic(err)
			}

			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error to read response body from webserver")
				panic(err)
			}

			var result GenerateTokenResponse
			if err := json.Unmarshal(respBody, &result); err != nil {
				log.Println("Error to unmarshal json")
				panic(err)
			}

			for {
				req, err = http.NewRequest("GET", "http://localhost:8080/rate-limit", nil)
				if err != nil {
					panic(err)
				}
				req.Header.Set("Api-key", result.AccessToken)

				resp, err = http.DefaultClient.Do(req)
				if err != nil {
					log.Println("Erro: ", err)
					panic(err)
				}

				println(resp.Status)

				time.Sleep(time.Duration(50) * time.Millisecond)
			}

		}()
	}

	wg.Wait()

}
