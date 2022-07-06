package main

import (
	"bidit/opensea/model"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func FindHighestOffer(asset *model.Asset) (decimal.Decimal, common.Address) {
	highest := decimal.NewFromInt(0)
	maker := common.Address{}
	for i := 0; i < len(asset.Orders); i++ {
		order := asset.Orders[i]
		expr := time.Unix(int64(order.ExpirationTime), 0)
		if order.Side == 0 && order.SaleKind == 0 && expr.After(time.Now()) && order.PaymentTokenContract.Symbol == "WETH" {
			price, _ := decimal.NewFromString(order.CurrentPrice)
			// Assert(err == nil, "FindHighestOffer:", err, "CurrentPrice:", order.CurrentPrice)
			if price.GreaterThan(highest) {
				highest = price
				maker = order.Maker.Address
			}
		}
	}
	return highest, maker
}
