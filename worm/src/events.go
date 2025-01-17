package src

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type eventFetcher struct {
	log       *zap.Logger
	ticker    *time.Ticker
	eventChan chan common.Hash
	address   *common.Address
	ethclient *ethclient.Client
	mu        *sync.Mutex
}

func NewEventFetcher(log *zap.Logger, ethclient *ethclient.Client) *eventFetcher {
	return &eventFetcher{
		log:       log,
		ticker:    time.NewTicker(2 * time.Second),
		eventChan: make(chan common.Hash),
		ethclient: ethclient,
		mu:        &sync.Mutex{},
	}
}

func fetchLatestBlock(client *ethclient.Client) (*big.Int, error) {
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}
	return big.NewInt(int64(latestBlock)), nil
}

func (ef *eventFetcher) Fetch() error {
	latestBlock, err := fetchLatestBlock(ef.ethclient)
	if err != nil {
		return fmt.Errorf("failed to fetch latest block: %w", err)
	}

	for range ef.ticker.C {
		newLatestBlock, err := fetchLatestBlock(ef.ethclient)
		if err != nil {
			ef.log.Sugar().Errorw("Failed to fetch latest block", "error", err)
			continue
		}

		contractAddress := *ef.address
		topic := common.HexToHash("0x655d1e7b93108e2de8f400b1c7a9720d149068ab024d30642da3af0345db848c")

		query := ethereum.FilterQuery{
			FromBlock: latestBlock,
			ToBlock:   newLatestBlock,
			Addresses: []common.Address{contractAddress},
			Topics:    [][]common.Hash{{topic}},
		}

		logs, err := ef.ethclient.FilterLogs(context.Background(), query)
		if err != nil {
			ef.log.Sugar().Errorw("Failed to fetch latest block", "error", err)
			continue
		}

		latestBlock.Add(newLatestBlock, big.NewInt(1))

		if len(logs) > 0 {
			address := logs[0].Topics[1]
			ef.log.Sugar().Infow("event trigger", "address", address)
			ef.eventChan <- address
		}
	}

	return nil
}
