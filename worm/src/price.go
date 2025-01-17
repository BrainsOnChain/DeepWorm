package src

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	solAddr  = "So11111111111111111111111111111111111111112"
	WormAddr = "DwDtUqBZJtbRpdjsFw3N7YKB5epocSru25BGcVhfcYtg"

	// Jupiter API
	host          = "https://api.jup.ag"
	priceEndpoint = "/price/v2"

	priceFetchInterval = 2 * time.Second
)

type priceFetcher struct {
	log       *zap.Logger
	client    *http.Client
	ticker    *time.Ticker
	coinAddr  string
	priceChan chan price
}

func NewPriceFetcher(log *zap.Logger, ca string) *priceFetcher {
	return &priceFetcher{
		log:       log,
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
			pf.log.Sugar().Errorw("error fetching price", "error", err)
			continue
		}

		type respObj struct {
			Data map[string]struct {
				Price string `json:"price"`
			} `json:"data"`
			TimeTake float64 `json:"timeTaken"`
		}

		var data respObj
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			resp.Body.Close()
			pf.log.Sugar().Errorw("error decoding response", "error", err)
			continue
		}
		resp.Body.Close()

		pf.priceChan <- newPrice(pf.coinAddr, data.Data[pf.coinAddr].Price)
	}

	return nil
}

func newPrice(mint, priceUSD string) price {
	p, err := strconv.ParseFloat(priceUSD, 64)
	if err != nil {
		zap.S().Errorw("Error converting price to float64", "error", err)
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
