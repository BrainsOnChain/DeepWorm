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
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
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
	log       *zap.Logger
	cworm     *C.Worm
	mu        *sync.Mutex
	address   *common.Address
	ethclient *ethclient.Client
}

func NewWorm(log *zap.Logger, ethclient *ethclient.Client) *Worm {
	return &Worm{
		log:       log,
		cworm:     C.Worm_Worm(),
		mu:        &sync.Mutex{},
		ethclient: ethclient,
	}
}

func (w *Worm) StateServe(efetcher *eventFetcher) error {
	http.HandleFunc("/set", func(rw http.ResponseWriter, r *http.Request) {
		address := common.HexToAddress(r.URL.Query().Get("address"))

		efetcher.mu.Lock()
		if efetcher.address != nil {
			efetcher.mu.Unlock()
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		efetcher.address = &address
		efetcher.mu.Unlock()

		w.mu.Lock()
		w.address = &address
		w.mu.Unlock()

		rw.Write([]byte("done"))
	})

	http.HandleFunc("/leftmuscle", func(rw http.ResponseWriter, r *http.Request) {
		w.mu.Lock()
		state := C.Worm_getLeftMuscle(w.cworm)
		w.mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	http.HandleFunc("/rightmuscle", func(rw http.ResponseWriter, r *http.Request) {
		w.mu.Lock()
		state := C.Worm_getRightMuscle(w.cworm)
		w.mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	http.HandleFunc("/state", func(rw http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		id_int, err := strconv.ParseInt(id, 10, 16)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		w.mu.Lock()
		state := C.Worm_state(w.cworm, C.uint16_t(id_int))
		w.mu.Unlock()

		rw.Write([]byte(strconv.Itoa(int(state))))
	})

	return http.ListenAndServe(":8080", nil)
}

func (w *Worm) Run(pfetcher *priceFetcher, efetcher *eventFetcher) error {
	defer C.Worm_destroy(w.cworm)
	var p position

	privateKeyBytes, err := os.ReadFile("/app/secp.sec")
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("failed to get public key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	w.log.Sugar().Infow("Using address", "from", fromAddress.Hex())

	currentPrice := <-pfetcher.priceChan

	for {
		select {
		case price := <-pfetcher.priceChan:
			priceChange := math.Abs(price.priceUSD - currentPrice.priceUSD)
			currentPrice = price

			if priceChange == 0 {
				continue
			}

			intChange := int(priceChange * magnification)
			w.log.Sugar().Infow("price change", "change", intChange)

			w.mu.Lock()
			adX, adY := 0.0, 0.0
			var leftMuscle, rightMuscle float64
			for i := 0; i < intChange; i++ {
				dX, dY := stimulateWorm(w, &p)
				adX += dX
				adY += dY
			}
			w.mu.Unlock()

			calldata := buildCalldata(adX, adY, leftMuscle, rightMuscle, encode(currentPrice.priceUSD*magnification))

			if err := w.sendTransaction(fromAddress, privateKey, calldata); err != nil {
				w.log.Error("failed to send transaction", zap.Error(err))
			}

		case address := <-efetcher.eventChan:
			intChange := 10
			w.mu.Lock()
			adX, adY := 0.0, 0.0
			var leftMuscle, rightMuscle float64
			for i := 0; i < intChange; i++ {
				dX, dY := stimulateWorm(w, &p)
				adX += dX
				adY += dY
			}
			w.mu.Unlock()

			calldata := buildCalldata(adX, adY, leftMuscle, rightMuscle, address.Hex()[2:])

			if err := w.sendTransaction(fromAddress, privateKey, calldata); err != nil {
				w.log.Error("failed to send transaction", zap.Error(err))
			}
		}
	}
}

func stimulateWorm(w *Worm, p *position) (float64, float64) {
	if rand.Intn(100) < 80 {
		C.Worm_chemotaxis(w.cworm)
	} else {
		C.Worm_noseTouch(w.cworm)
	}

	leftMuscle := float64(C.Worm_getLeftMuscle(w.cworm))
	rightMuscle := float64(C.Worm_getRightMuscle(w.cworm))

	angle, magnitude := movement(leftMuscle, rightMuscle)
	dX, dY := p.update(angle, magnitude)

	return dX, dY
}

func buildCalldata(adX, adY, leftMuscle, rightMuscle float64, encodedVal string) string {
	return fmt.Sprintf("0x6faeae2b%s%s%s%s%s%s",
		encode(adX),
		encode(adY),
		encodeInt(time.Now().Unix()),
		encode(leftMuscle),
		encode(rightMuscle),
		encodedVal,
	)
}

func (w *Worm) sendTransaction(fromAddress common.Address, privateKey *ecdsa.PrivateKey, calldata string) error {
	data := common.FromHex(calldata)
	contractAddress := *w.address

	nonce, err := w.ethclient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := w.ethclient.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
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
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = w.ethclient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	w.log.Sugar().Infow("Transaction sent", "hash", signedTx.Hash().Hex())
	return nil
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

// update updates the position based on the angle and magnitude. It returns the
// distance moved in the X and Y directions.
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
