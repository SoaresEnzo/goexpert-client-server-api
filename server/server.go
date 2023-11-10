package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PriceApiResponse struct {
	Usdbrl Usdbrl `json:"USDBRL"`
}

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp" gorm:"primaryKey"`
	CreateDate string `json:"create_date"`
}

type BrlPriceDB struct {
	Timestamp string `gorm:"primaryKey"`
	Bid       string
}

type GetCotacaoApiDTO struct {
	Bid string `json:"bid"`
}

var db *gorm.DB

const databaseTimeout = 10 * time.Millisecond

const apiTimeout = 200 * time.Millisecond

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&BrlPriceDB{})
	if err != nil {
		return
	}
	http.HandleFunc("/cotacao", handler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	price, err := requestCurrentDollarPrice()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	dbEntity := BrlPriceDB{Bid: price.Usdbrl.Bid, Timestamp: price.Usdbrl.Timestamp}
	db.WithContext(ctx).Save(&dbEntity)
	err = ctx.Err()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Print("Database took too much time to respond\n")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	responseDto := GetCotacaoApiDTO{Bid: price.Usdbrl.Bid}
	response, err := json.Marshal(responseDto)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		return
	}
}

func requestCurrentDollarPrice() (*PriceApiResponse, error) {
	priceApiUrl := "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, priceApiUrl, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Print("API call took too much time to respond - https://economia.awesomeapi.com.br/json/last/USD-BRL \n")
		}
		return nil, err
	}
	defer res.Body.Close()

	j, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var responseBody PriceApiResponse
	err = json.Unmarshal(j, &responseBody)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &responseBody, nil
}
