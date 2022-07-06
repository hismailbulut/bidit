package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type Asset struct {
	ID                      int           `json:"id"`
	NumSales                int           `json:"num_sales"`
	BackgroundColor         any           `json:"background_color"`
	ImageURL                string        `json:"image_url"`
	ImagePreviewURL         string        `json:"image_preview_url"`
	ImageThumbnailURL       string        `json:"image_thumbnail_url"`
	ImageOriginalURL        string        `json:"image_original_url"`
	AnimationURL            any           `json:"animation_url"`
	AnimationOriginalURL    any           `json:"animation_original_url"`
	Name                    string        `json:"name"`
	Description             any           `json:"description"`
	ExternalLink            string        `json:"external_link"`
	AssetContract           AssetContract `json:"asset_contract"`
	Permalink               string        `json:"permalink"`
	Collection              Collection    `json:"collection"`
	Decimals                any           `json:"decimals"`
	TokenMetadata           string        `json:"token_metadata"`
	IsNsfw                  bool          `json:"is_nsfw"`
	Owner                   Account       `json:"owner"`
	SellOrders              []Order       `json:"sell_orders"`
	SeaportSellOrders       any           `json:"seaport_sell_orders"`
	Creator                 Account       `json:"creator"`
	Traits                  []any         `json:"traits"`
	LastSale                LastSale      `json:"last_sale"`
	TopBid                  any           `json:"top_bid"`
	ListingDate             any           `json:"listing_date"`
	IsPresale               bool          `json:"is_presale"`
	TransferFeePaymentToken any           `json:"transfer_fee_payment_token"`
	TransferFee             any           `json:"transfer_fee"`
	RelatedAssets           []any         `json:"related_assets"`
	Orders                  []Order       `json:"orders"`
	Auctions                []any         `json:"auctions"`
	SupportsWyvern          bool          `json:"supports_wyvern"`
	TopOwnerships           []any         `json:"top_ownerships"`
	Ownership               any           `json:"ownership"`
	HighestBuyerCommitment  any           `json:"highest_buyer_commitment"`
	TokenID                 string        `json:"token_id"`
}

type AssetMini struct {
	ID      decimal.Decimal `json:"id"`
	Address common.Address  `json:"address"`
}
