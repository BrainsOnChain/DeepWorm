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
extern int16_t Worm_state(Worm* worm, const uint16_t id);
*/
import "C"
import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
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
	cworm   *C.Worm
	Mu      *sync.Mutex
	Address *common.Address
}

func NewWorm() *Worm {
	return &Worm{
		cworm:   C.Worm_Worm(),
		Mu:      &sync.Mutex{},
		Address: nil,
	}
}

func (w *Worm) StateServe(efetcher *eventFetcher) error {
	http.HandleFunc("/set", func(rw http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))

		efetcher.Mu.Lock()
		if efetcher.Address != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		efetcher.Address = &address
		efetcher.Mu.Unlock()

		w.Mu.Lock()
		w.Address = &address
		w.Mu.Unlock()

		rw.Write([]byte("done"))
	})

	http.HandleFunc("/leftmuscle", func(rw http.ResponseWriter, r *http.Request) {
		w.Mu.Lock()
		state := C.Worm_getLeftMuscle(w.cworm)
		w.Mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	http.HandleFunc("/rightmuscle", func(rw http.ResponseWriter, r *http.Request) {
		w.Mu.Lock()
		state := C.Worm_getRightMuscle(w.cworm)
		w.Mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	http.HandleFunc("/state", func(rw http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		id_int, err := strconv.ParseInt(id, 10, 16)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Mu.Lock()
		state := C.Worm_state(w.cworm, C.uint16_t(id_int))
		w.Mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	return http.ListenAndServe(":8080", nil)
}

func (w *Worm) Run(pfetcher *priceFetcher, efetcher *eventFetcher) {
	defer C.Worm_destroy(w.cworm) // Ensure proper cleanup
	var p position

	// privateKeyBytes, err := ioutil.ReadFile("./secp.sec")
	privateKeyBytes, err := ioutil.ReadFile("/app/secp.sec")
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
	currentPrice := <-pfetcher.priceChan

	for {
		select {
		case price := <-pfetcher.priceChan:
			{
				priceChange := math.Abs(price.priceUSD - currentPrice.priceUSD)
				currentPrice = price

				if priceChange == 0 { // no change in price no worm movement
					continue
				}

				// Magnify the price change to simulate worm movement
				intChange := int(priceChange * magnification)
				zap.S().Infow("price change", "change", intChange)

				w.Mu.Lock()
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
				w.Mu.Unlock()

				client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
				if err != nil {
					zap.S().Errorw("Failed to connect to the Ethereum rpc", "error", err)
					continue
				}

				contractAddress := *w.Address

				calldata := fmt.Sprintf("0x6faeae2b%s%s%s%s%s%s",
					encode(adX),
					encode(adY),
					encodeInt(time.Now().Unix()),
					encode(leftMuscle),
					encode(rightMuscle),
					encode(currentPrice.priceUSD*magnification))
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
		case address := <-efetcher.eventChan:
			{
				intChange := 10
				w.Mu.Lock()
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
				w.Mu.Unlock()

				client, err := ethclient.Dial("https://api.hyperliquid-testnet.xyz/evm")
				if err != nil {
					zap.S().Errorw("Failed to connect to the Ethereum rpc", "error", err)
					continue
				}

				contractAddress := *w.Address

				calldata := fmt.Sprintf("0xce4f76ca%s%s%s%s%s%s",
					encode(adX),
					encode(adY),
					encodeInt(time.Now().Unix()),
					encode(leftMuscle),
					encode(rightMuscle),
					address)
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
