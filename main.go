package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type Smbls struct {
	Symbols []struct {
		Symbol string `json:"symbol"`
	} `json:"symbols"`
}

type Price struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func GetPrice(symbol string, ch chan map[int]Price) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=" + symbol)

	if err != nil {
		fmt.Printf("GetPrice: error to get %v\n", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("GetPrice: get status code %v\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("GetPrice: error reading answer %v\n", err)
		return
	}

	var price Price

	err = json.Unmarshal(body, &price)
	if err != nil {
		fmt.Printf("GetPrice: error unmarshal %v\n", err)
		return
	}

	answer := make(map[int]Price)

	answer[0] = price

	ch <- answer
}

func GetExchangeSymbols() ([]string, error) {
	resp, err := http.Get("https://binance.com/api/v3/exchangeInfo")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Status code:%d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var answer Smbls

	err = json.Unmarshal(body, &answer)
	if err != nil {
		return nil, err
	}

	totalLength := len(answer.Symbols)
	if totalLength > 5 {
		totalLength = 5
	}

	var result []string

	for i := 0; i < totalLength; i++ {
		result = append(result, answer.Symbols[i].Symbol)
	}

	return result, nil
}

func main() {
	wg := sync.WaitGroup{}

	symbols, err := GetExchangeSymbols()
	if err != nil {
		fmt.Printf("Error to get symbols, %v", err)
		os.Exit(1)
	}

	ch := make(chan map[int]Price)

	for _, symbol := range symbols {
		wg.Add(1)
		go GetPrice(symbol, ch)
	}

	go func() {
		for p := range ch {
			fmt.Println(p[0].Symbol, p[0].Price)
			wg.Done()
		}
	}()

	wg.Wait()
}
