package src

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	solAddr  = "So11111111111111111111111111111111111111112"
	WormAddr = "DwDtUqBZJtbRpdjsFw3N7YKB5epocSru25BGcVhfcYtg"

	// Jupiter API
	host          = "https://api.jup.ag"
	priceEndpoint = "/price/v2"

	priceFetchInterval = 4 * time.Second
)

type priceFetcher struct {
	client    *http.Client
	ticker    *time.Ticker
	coinAddr  string
	priceChan chan price
}

func NewPriceFetcher(ca string) *priceFetcher {
	return &priceFetcher{
		client:    &http.Client{},
		ticker:    time.NewTicker(priceFetchInterval),
		coinAddr:  ca,
		priceChan: make(chan price),
	}
}

func (pf *priceFetcher) Fetch() error {
	url := fmt.Sprintf("%s%s?ids=%s,%s", host, priceEndpoint, pf.coinAddr, solAddr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// fetch the endpoint on every tick
	for range pf.ticker.C {
		resp, err := pf.client.Do(req)
		if err != nil {
			return fmt.Errorf("error fetching price: %w", err)
		}
		defer resp.Body.Close()

		type respObj struct {
			Data map[string]struct {
				Price string `json:"price"`
			} `json:"data"`
			TimeTake float64 `json:"timeTaken"`
		}

		var data respObj
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}

		pf.priceChan <- newPrice(pf.coinAddr, data.Data[pf.coinAddr].Price)
	}

	return nil
}

func newPrice(mint, priceUSD string) price {
	p, err := strconv.ParseFloat(priceUSD, 64)
	if err != nil {
		fmt.Println("Error converting price to float64:", err)
		return price{}
	}

	return price{
		mint:     mint,
		priceUSD: p,
		ts:       time.Now(),
	}
}

type price struct {
	mint     string
	priceUSD float64
	ts       time.Time
}
