package src

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type eventFetcher struct {
	ticker    *time.Ticker
	eventChan chan common.Hash
	Mu        *sync.Mutex
	Address   *common.Address
}

func NewEventFetcher() *eventFetcher {
	return &eventFetcher{
		ticker:    time.NewTicker(2 * time.Second),
		eventChan: make(chan common.Hash),
		Mu:        &sync.Mutex{},
		Address:   nil,
	}
}

func fetchLatestBlock() (*big.Int, error) {
	client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
	if err != nil {
		zap.S().Errorw("Failed to connect to the Ethereum rpc", "error", err)
		return nil, err
	}

	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		zap.S().Errorw("Failed to get latest block", "error", err)
		return nil, err
	}

	return big.NewInt(int64(latestBlock)), nil
}

func (ef *eventFetcher) Fetch() error {
	latestBlock, err := fetchLatestBlock()
	if err != nil {
		zap.S().Errorw("Failed to fetch latest block", "error", err)
		return err
	}
	<-ef.ticker.C
	for range ef.ticker.C {
		newLatestBlock, err := fetchLatestBlock()
		if err != nil {
			zap.S().Errorw("Failed to fetch latest block", "error", err)
			continue
		}

		client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
		if err != nil {
			zap.S().Errorw("Failed to connect to the Ethereum rpc", "error", err)
			continue
		}

		contractAddress := *ef.Address
		topic := common.HexToHash("0x655d1e7b93108e2de8f400b1c7a9720d149068ab024d30642da3af0345db848c")

		query := ethereum.FilterQuery{
			FromBlock: latestBlock,
			ToBlock:   newLatestBlock,
			Addresses: []common.Address{
				contractAddress,
			},
			Topics: [][]common.Hash{{
				topic,
			}},
		}

		logs, err := client.FilterLogs(context.Background(), query)
		if err != nil {
			zap.S().Errorw("Failed to fetch latest block", "error", err)
			continue
		}

		latestBlock.Set(newLatestBlock)

		if len(logs) > 0 {
			address := logs[0].Topics[1]
			zap.S().Infow("event trigger", "address", address)
			ef.eventChan <- address
		}
	}

	return nil
}
