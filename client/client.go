package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type GetCotacaoApiDTO struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Print("Server took too much time to respond\n")
			panic(err)
		}

		panic(err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		panic("request failed")
	}

	var getCotacaoApiDTO GetCotacaoApiDTO
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, &getCotacaoApiDTO)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	_, err = file.WriteString("Dólar: " + getCotacaoApiDTO.Bid + "\n")
}
