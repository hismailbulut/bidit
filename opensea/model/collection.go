package model

type CollectionResponse struct {
	Success    bool       `json:"success"`
	Collection Collection `json:"collection"`
}

type Collection struct {
	Editors                     []string        `json:"editors"`
	PaymentTokens               []PaymentToken  `json:"payment_tokens"`
	PrimaryAssetContracts       []AssetContract `json:"primary_asset_contracts"`
	Traits                      any             `json:"traits"`
	Stats                       Stats           `json:"stats"`
	BannerImageURL              string          `json:"banner_image_url"`
	ChatURL                     string          `json:"chat_url"`
	CreatedDate                 string          `json:"created_date"`
	DefaultToFiat               bool            `json:"default_to_fiat"`
	Description                 string          `json:"description"`
	DevBuyerFeeBasisPoints      string          `json:"dev_buyer_fee_basis_points"`
	DevSellerFeeBasisPoints     string          `json:"dev_seller_fee_basis_points"`
	DiscordURL                  string          `json:"discord_url"`
	DisplayData                 any             `json:"display_data"`
	ExternalURL                 string          `json:"external_url"`
	Featured                    bool            `json:"featured"`
	FeaturedImageURL            string          `json:"featured_image_url"`
	Hidden                      bool            `json:"hidden"`
	SafelistRequestStatus       string          `json:"safelist_request_status"`
	ImageURL                    string          `json:"image_url"`
	IsSubjectToWhitelist        bool            `json:"is_subject_to_whitelist"`
	LargeImageURL               string          `json:"large_image_url"`
	MediumUsername              string          `json:"medium_username"`
	Name                        string          `json:"name"`
	OnlyProxiedTransfers        bool            `json:"only_proxied_transfers"`
	OpenseaBuyerFeeBasisPoints  string          `json:"opensea_buyer_fee_basis_points"`
	OpenseaSellerFeeBasisPoints string          `json:"opensea_seller_fee_basis_points"`
	PayoutAddress               string          `json:"payout_address"`
	RequireEmail                bool            `json:"require_email"`
	ShortDescription            string          `json:"short_description"`
	Slug                        string          `json:"slug"`
	TelegramURL                 string          `json:"telegram_url"`
	TwitterUsername             string          `json:"twitter_username"`
	InstagramUsername           string          `json:"instagram_username"`
	WikiURL                     string          `json:"wiki_url"`
	IsNsfw                      bool            `json:"is_nsfw"`
}

type Stats struct {
	OneDayVolume          float64 `json:"one_day_volume"`
	OneDayChange          float64 `json:"one_day_change"`
	OneDaySales           float64 `json:"one_day_sales"`
	OneDayAveragePrice    float64 `json:"one_day_average_price"`
	SevenDayVolume        float64 `json:"seven_day_volume"`
	SevenDayChange        float64 `json:"seven_day_change"`
	SevenDaySales         float64 `json:"seven_day_sales"`
	SevenDayAveragePrice  float64 `json:"seven_day_average_price"`
	ThirtyDayVolume       float64 `json:"thirty_day_volume"`
	ThirtyDayChange       float64 `json:"thirty_day_change"`
	ThirtyDaySales        float64 `json:"thirty_day_sales"`
	ThirtyDayAveragePrice float64 `json:"thirty_day_average_price"`
	TotalVolume           float64 `json:"total_volume"`
	TotalSales            float64 `json:"total_sales"`
	TotalSupply           float64 `json:"total_supply"`
	Count                 float64 `json:"count"`
	NumOwners             float64 `json:"num_owners"`
	AveragePrice          float64 `json:"average_price"`
	NumReports            float64 `json:"num_reports"`
	MarketCap             float64 `json:"market_cap"`
	FloorPrice            float64 `json:"floor_price"`
}
