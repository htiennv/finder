package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

const fileSaved = "./saved.txt"

func main() {
	file, err := os.Open(fileSaved)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	client, err := ethclient.Dial("https://eth-mainnet.public.blastapi.io")
	if err != nil {
		panic(err)
	}

	for {
		w, err := GenerateWallet2()
		if err != nil {
			panic(err)
		}
		slog.Info("GenerateWallet2", "address", w.Address)
		bal, err := GetBalance(client, common.HexToAddress(w.Address), 5)
		if err != nil {
			panic(err)
		}
		slog.Info("GetBalance", "address", w.Address, "bal", bal.Int64())
		if bal.Cmp(big.NewInt(1)) == 1 {
			if err := SaveResult(file, *w); err != nil {
				panic(err)
			}
		}
	}
}

type walletResult struct {
	Address  string
	Mnemonic string
}

// GetBalance retrieves the balance of an Ethereum address with retry logic.
func GetBalance(client *ethclient.Client, address common.Address, maxRetries int) (*big.Int, error) {
	var balance *big.Int
	var err error
	retries := 0
	for retries <= maxRetries {
		balance, err = client.BalanceAt(context.Background(), address, nil)
		if err == nil {
			return balance, nil
		}
		time.Sleep(1500 * time.Millisecond)
		retries += 1
	}

	return nil, fmt.Errorf("failed to get balance after %d attempts: %v", retries, err)
}

func SaveResult(file *os.File, w walletResult) error {
	dataString := fmt.Sprintf("Address: %s\nMnemonic: %s\n\n", w.Address, w.Mnemonic)
	_, err := file.Write([]byte(dataString))
	if err != nil {
		return err
	}
	slog.Info("SaveResult done")
	return nil
}

func GenerateWallet2() (*walletResult, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err

	}
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	return &walletResult{
		Address:  account.Address.Hex(),
		Mnemonic: mnemonic,
	}, nil
}
