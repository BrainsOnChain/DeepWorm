package src

/*
#cgo CFLAGS: -I../nematoduino
#cgo LDFLAGS: -L.. -lnematoduino
#include "Worm.h"
#include <stdlib.h>

// Declare the C functions based on Worm.h
extern Worm* Worm_Worm();
extern void Worm_destroy(Worm* worm);
extern void Worm_chemotaxis(Worm* worm);
extern void Worm_noseTouch(Worm* worm);
extern int Worm_getLeftMuscle(Worm* worm);
extern int Worm_getRightMuscle(Worm* worm);
*/
import "C"
import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

const (
	magnification = 1_000_000
)

type Worm struct {
	cworm *C.Worm
}

func NewWorm() *Worm {
	return &Worm{
		cworm: C.Worm_Worm(),
	}
}

func (w *Worm) Run(fetcher *priceFetcher, mu *sync.Mutex) {
	defer C.Worm_destroy(w.cworm) // Ensure proper cleanup
	var p position

	privateKeyBytes, err := os.ReadFile("/app/secp.sec")
	if err != nil {
		zap.S().Errorw("Failed to read private key file", "error", err)
		return
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		zap.S().Errorw("Failed to parse private key", "error", err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		zap.S().Error("Failed to get public key")
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	zap.S().Infow("Using address", "from", fromAddress.Hex())

	// Fetch the price of worm and compare to previous price
	currentPrice := <-fetcher.priceChan
	for price := range fetcher.priceChan {
		priceChange := math.Abs(price.priceUSD - currentPrice.priceUSD)
		currentPrice = price

		if priceChange == 0 { // no change in price no worm movement
			continue
		}

		// Magnify the price change to simulate worm movement
		intChange := int(priceChange * magnification)
		zap.S().Infow("price change", "change", intChange)

		adX, adY := 0.0, 0.0
		var leftMuscle, rightMuscle float64
		for i := 0; i < intChange; i++ {
			// 80% change of chemotaxis and 20% of nose touch for each cycle
			if rand.Intn(100) < 80 {
				C.Worm_chemotaxis(w.cworm)
			} else {
				C.Worm_noseTouch(w.cworm)
			}

			leftMuscle, rightMuscle = float64(C.Worm_getLeftMuscle(w.cworm)), float64(C.Worm_getRightMuscle(w.cworm))

			angle, magnitude := movement(leftMuscle, rightMuscle)

			dX, dY := p.update(angle, magnitude)
			adX += dX
			adY += dY
		}

		client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
		if err != nil {
			zap.S().Errorw("Failed to connect to the Ethereum rpc", "error", err)
			continue
		}

		contractAddress := common.HexToAddress("0x7A129762332B8f4c6Ed4850c17B218C89e78854d")

		calldata := fmt.Sprintf("0x6faeae2b%s%s%s%s%s%s",
			encode(adX),
			encode(adY),
			encodeInt(time.Now().Unix()),
			encode(leftMuscle),
			encode(rightMuscle),
			encode(currentPrice.priceUSD*magnification),
		)
		data := common.FromHex(calldata)

		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			zap.S().Errorw("Failed to get nonce", "error", err)
			continue
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			zap.S().Errorw("Failed to get gas price", "error", err)
			continue
		}

		tx := types.NewTransaction(
			nonce,
			contractAddress,
			big.NewInt(0),
			300000,
			gasPrice,
			data,
		)

		chainID := big.NewInt(998)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			zap.S().Errorw("Failed to sign transaction", "error", err)
			continue
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			zap.S().Errorw("Failed to send transaction", "error", err)
			continue
		}

		zap.S().Infow("Transaction sent", "hash", signedTx.Hash().Hex())
	}
}

func encodeInt(num int64) string {
	bignum := new(big.Int).SetInt64(int64(num))
	if num < 0 {
		bignum.Add(bignum, new(big.Int).Lsh(big.NewInt(1), 256))
	}

	return fmt.Sprintf("%064x", bignum)
}

func encode(num float64) string {
	bignum := new(big.Int).SetInt64(int64(num))
	if num < 0 {
		bignum.Add(bignum, new(big.Int).Lsh(big.NewInt(1), 256))
	}

	return fmt.Sprintf("%064x", bignum)
}

// -----------------------------------------------------------------------------
// Position and movement functions

type position struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Direction float64 `json:"direction"`
}

func (p *position) update(angle, magnitude float64) (float64, float64) {
	p.Direction += angle
	if p.Direction < 0 {
		p.Direction += 360
	} else if p.Direction >= 360 {
		p.Direction -= 360
	}

	dX, dY := magnitude*math.Cos(p.Direction*math.Pi/180), magnitude*math.Sin(p.Direction*math.Pi/180)

	// Update the position based on the Direction
	p.X += dX
	p.Y += dY

	return dX, dY
}

// movement outputs the movement in the form of angle and magnitude based on the
// left and right muscle activity.
func movement(left, right float64) (float64, float64) {
	// Calculate the angle and magnitude
	angle := (right - left) / 2
	magnitude := (right + left) / 2

	return angle, magnitude
}
