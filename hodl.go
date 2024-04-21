package main

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	var priv, calldata string
	fmt.Print("私钥: ")
	fmt.Scanln(&priv)
	fmt.Print("字节码: ")
	fmt.Scanln(&calldata)

	if !strings.HasPrefix(calldata, "0x") || calldata[:10] != "0xdc336230" || len(calldata) != 458 {
		fmt.Println("字节码格式错误")
		return
	}

	var maxAmt int
	fmt.Print("数量: ")
	fmt.Scanln(&maxAmt)

	client, err := ethclient.Dial("https://young-solitary-cherry.bsc.quiknode.pro/")
	if err != nil {
		fmt.Println("无法连接到以太坊客户端:", err)
		return
	}

	privateKey, err := crypto.HexToECDSA(priv)
	if err != nil {
		fmt.Println("私钥无效:", err)
		return
	}

	account := crypto.PubkeyToAddress(privateKey.PublicKey)

	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		fmt.Println("获取余额错误:", err)
		return
	}

	requiredBalance := big.NewInt(182 * 10_000_000_000_000 * int64(maxAmt))
	requiredBalance.Add(requiredBalance, big.NewInt(110_000*1_000_010_000*int64(maxAmt)))

	if balance.Cmp(requiredBalance) < 0 {
		fmt.Printf("余额不足, 需要 %f BNB, 当前余额: %f BNB\n",
			new(big.Float).Quo(new(big.Float).SetInt(requiredBalance), big.NewFloat(10_000_000_000_000_000_000)),
			new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(10_000_000_000_000_000_000)),
		)
		return
	}

	nonce, err := client.PendingNonceAt(context.Background(), account)
	if err != nil {
		fmt.Println("获取Nonce错误:", err)
		return
	}

	fmt.Println("Nonce:", nonce)

	currentNonce := nonce
	upperLim := nonce + uint64(maxAmt)
	for currentNonce < upperLim {
		targetNonce := currentNonce + 200
		if targetNonce > upperLim {
			targetNonce = upperLim
		}
		for i := currentNonce; i < targetNonce; i++ {
			tx := types.NewTransaction(
				i,
				common.HexToAddress("0x1832e00DfF829547E1F564f92401C2886F3236b4"),
				big.NewInt(182*10_000_000_000_000),
				220_000,
				big.NewInt(1_000_010_000),
				common.FromHex(calldata),
			)

			signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(56)), privateKey)
			if err != nil {
				fmt.Println("签名错误:", err)
				continue
			}

			err = client.SendTransaction(context.Background(), signedTx)
			if err != nil {
				fmt.Println("交易错误:", err)
				continue
			}

			fmt.Println("交易", i-nonce, signedTx.Hash().Hex())
		}
		tempNonce := nonce
		for targetNonce != upperLim && tempNonce < targetNonce-10 {
			fmt.Printf("打包中... %d/%d 被打包\n", tempNonce-nonce, targetNonce-10-nonce)
			time.Sleep(3 * time.Second)
			tempNonce, err = client.PendingNonceAt(context.Background(), account)
			if err != nil {
				fmt.Println("获取Nonce错误:", err)
				return
			}
		}
		if targetNonce == upperLim {
			break
		}
		currentNonce, err = client.PendingNonceAt(context.Background(), account)
		if err != nil {
			fmt.Println("获取Nonce错误:", err)
			return
		}
	}
}
