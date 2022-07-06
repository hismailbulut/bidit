package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type Account struct {
	User          any            `json:"user"`
	ProfileImgURL string         `json:"profile_img_url"`
	Address       common.Address `json:"address"`
	Config        string         `json:"config"`
}

type Transaction struct {
	BlockHash        string  `json:"block_hash"`
	BlockNumber      string  `json:"block_number"`
	FromAccount      Account `json:"from_account"`
	ID               int     `json:"id"`
	Timestamp        string  `json:"timestamp"`
	ToAccount        Account `json:"to_account"`
	TransactionHash  string  `json:"transaction_hash"`
	TransactionIndex string  `json:"transaction_index"`
}

type AssetContract struct {
	Address                     common.Address `json:"address"`
	AssetContractType           string         `json:"asset_contract_type"`
	CreatedDate                 string         `json:"created_date"`
	Name                        string         `json:"name"`
	NftVersion                  string         `json:"nft_version"`
	OpenseaVersion              any            `json:"opensea_version"`
	Owner                       any            `json:"owner"`
	SchemaName                  string         `json:"schema_name"`
	Symbol                      string         `json:"symbol"`
	TotalSupply                 any            `json:"total_supply"`
	Description                 string         `json:"description"`
	ExternalLink                string         `json:"external_link"`
	ImageURL                    string         `json:"image_url"`
	DefaultToFiat               bool           `json:"default_to_fiat"`
	DevBuyerFeeBasisPoints      int            `json:"dev_buyer_fee_basis_points"`
	DevSellerFeeBasisPoints     int            `json:"dev_seller_fee_basis_points"`
	OnlyProxiedTransfers        bool           `json:"only_proxied_transfers"`
	OpenseaBuyerFeeBasisPoints  int            `json:"opensea_buyer_fee_basis_points"`
	OpenseaSellerFeeBasisPoints int            `json:"opensea_seller_fee_basis_points"`
	BuyerFeeBasisPoints         int            `json:"buyer_fee_basis_points"`
	SellerFeeBasisPoints        int            `json:"seller_fee_basis_points"`
	PayoutAddress               any            `json:"payout_address"`
}

type PaymentToken struct {
	ID       int             `json:"id"`
	Symbol   string          `json:"symbol"`
	Address  common.Address  `json:"address"`
	ImageURL string          `json:"image_url"`
	Name     string          `json:"name"`
	Decimals int             `json:"decimals"`
	EthPrice decimal.Decimal `json:"eth_price"`
	UsdPrice decimal.Decimal `json:"usd_price"`
}
