package opensea

import (
	"bidit/bidutil"
	"bidit/opensea/model"
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/shopspring/decimal"
)

//go:embed abi.json
var WyvernABIJson string

// Order type used when bidding, post request
type BidOrder struct {
	// included in both hash and json
	Exchange           common.Address  `json:"exchange"`
	Maker              common.Address  `json:"maker"`
	Taker              common.Address  `json:"taker"`
	MakerRelayerFee    decimal.Decimal `json:"makerRelayerFee"`
	TakerRelayerFee    decimal.Decimal `json:"takerRelayerFee"`
	MakerProtocolFee   decimal.Decimal `json:"makerProtocolFee"`
	TakerProtocolFee   decimal.Decimal `json:"takerProtocolFee"`
	MakerReferrerFee   decimal.Decimal `json:"makerReferrerFee"` // Not in hash
	FeeRecipient       common.Address  `json:"feeRecipient"`
	FeeMethod          uint8           `json:"feeMethod"`
	Side               uint8           `json:"side"`
	SaleKind           uint8           `json:"saleKind"`
	Target             common.Address  `json:"target"`
	HowToCall          uint8           `json:"howToCall"`
	Calldata           string          `json:"calldata"`
	ReplacementPattern string          `json:"replacementPattern"`
	StaticTarget       common.Address  `json:"staticTarget"`
	StaticExtradata    string          `json:"staticExtradata"`
	PaymentToken       common.Address  `json:"paymentToken"`
	Quantity           string          `json:"quantity"` // Not in hash
	BasePrice          decimal.Decimal `json:"basePrice"`
	Extra              decimal.Decimal `json:"extra"`
	ListingTime        decimal.Decimal `json:"listingTime"`
	ExpirationTime     decimal.Decimal `json:"expirationTime"`
	Salt               decimal.Decimal `json:"salt"`
	Metadata           model.Metadata  `json:"metadata"` // Not in hash
	V                  int             `json:"v"`        // Signature
	R                  string          `json:"r"`        // Signature
	S                  string          `json:"s"`        // Signature
	Nonce              decimal.Decimal `json:"nonce"`
}

var (
	ExchangeAddressMainnet        = common.HexToAddress("0x0000000000000000000000000000000000000000") // TODO
	ExchangeAddressRinkeby        = common.HexToAddress("0xdd54d660178b28f6033a953b0e55073cfa7e3744")
	WethAddressMainnet            = common.HexToAddress("0x0000000000000000000000000000000000000000") // TODO
	WethAddressRinkeby            = common.HexToAddress("0xc778417e063141139fce010982780140aa0cd5ab")
	FeeRecipientAddress           = common.HexToAddress("0x5b3256965e7c3cf26e11fcaf296dfc8807c01073")
	MerkleValidatorAddressMainnet = common.HexToAddress("0xbaf2127b49fc93cbca6269fade0f7f31df4c88a7")
	MerkleValidatorAddressRinkeby = common.HexToAddress("0x45b594792a5cdc008d0de1c1d69faa3d16b3ddc1")
)

func MakeOffer(asset *model.Asset, collection *model.Collection, privKey *bidutil.PrivateKey, price decimal.Decimal, duration time.Duration) error {

	order := BidOrder{}

	openseaGuard.Lock()
	isTestnet := openseaClient.testnet
	openseaGuard.Unlock()

	if isTestnet {
		order.Exchange = ExchangeAddressRinkeby
		order.Target = MerkleValidatorAddressRinkeby
		order.PaymentToken = WethAddressRinkeby
	} else {
		order.Exchange = ExchangeAddressMainnet
		order.Target = MerkleValidatorAddressMainnet
		order.PaymentToken = WethAddressMainnet
	}

	order.Maker = privKey.PublicAddress()
	order.Taker = common.Address{} // 0 address

	// Calculate Fees
	{
		openseaBuyerFee, err := decimal.NewFromString(collection.OpenseaBuyerFeeBasisPoints)
		if err != nil {
			return err
		}
		devBuyerFee, err := decimal.NewFromString(collection.DevBuyerFeeBasisPoints)
		if err != nil {
			return err
		}
		order.MakerRelayerFee = openseaBuyerFee.Add(devBuyerFee)

		openseaSellerFee, err := decimal.NewFromString(collection.OpenseaSellerFeeBasisPoints)
		if err != nil {
			return err
		}
		devSellerFee, err := decimal.NewFromString(collection.DevSellerFeeBasisPoints)
		if err != nil {
			return err
		}
		order.TakerRelayerFee = openseaSellerFee.Add(devSellerFee)

		order.MakerProtocolFee = decimal.NewFromInt(0)
		order.TakerProtocolFee = decimal.NewFromInt(0)
		order.MakerRelayerFee = decimal.NewFromInt(0)
	}

	// These are static for bid order
	order.FeeRecipient = FeeRecipientAddress
	order.FeeMethod = 1
	order.Side = 0
	order.SaleKind = 0
	order.HowToCall = 1

	// Generate calldata
	{
		parsedAbi, err := abi.JSON(strings.NewReader(WyvernABIJson))
		if err != nil {
			return err
		}

		from := common.Address{}
		to := privKey.PublicAddress()
		token := asset.AssetContract.Address
		tokenID, ok := new(big.Int).SetString(asset.TokenID, 10)
		if !ok {
			return errors.New("TokenID field of asset structure does not contain valid integer.")
		}

		calldata, err := parsedAbi.Pack("matchERC721UsingCriteria", from, to, token, tokenID, [32]byte{}, [][32]byte{})
		if err != nil {
			return err
		}
		order.Calldata = hexutil.Encode(calldata)
		// TODO: Hardcoded but still works
		order.ReplacementPattern = "0x00000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

		if len(order.Calldata) != len(order.ReplacementPattern) {
			return errors.New("Length of calldata and replacementPattern is not equal")
		}
	}

	order.StaticTarget = common.Address{} // TODO
	order.StaticExtradata = "0x"          // TODO
	order.Quantity = "1"                  // Hardcoded
	order.BasePrice = price
	order.Extra = decimal.NewFromInt(0)
	order.ListingTime = decimal.NewFromInt(time.Now().Unix())
	order.ExpirationTime = decimal.NewFromInt(time.Now().Add(duration).Unix())

	// Generate 256 bit cryptographically secure pseudorandom salt
	{
		buf := make([]byte, 32)
		rand.Read(buf)
		saltBigInt := new(big.Int).SetBytes(buf)
		order.Salt = decimal.NewFromBigInt(saltBigInt, 0)
	}

	// Generate metadata
	order.Metadata.Asset.Address = asset.AssetContract.Address // NOTE May not
	order.Metadata.Asset.ID, _ = decimal.NewFromString(asset.TokenID)
	order.Metadata.Schema = "ERC721" // Hardcoded

	order.Nonce = decimal.NewFromInt(0) // TODO

	// Sign order with private key
	signature, err := order.Sign(privKey)
	if err != nil {
		return err
	}

	order.R = hexutil.Encode(signature[:32])
	order.S = hexutil.Encode(signature[32:64])
	order.V = 27 + int(signature[64])

	orderJson, err := json.Marshal(order)
	if err != nil {
		return err
	}

	// Post
	err = postJson([]string{"wyvern", "v1", "orders", "post"}, bidutil.Params{}, orderJson, DEFAULT_RETRY_COUNT, DEFAULT_COOLDOWN)
	if err != nil {
		return err
	}

	return nil
}

func (order *BidOrder) Sign(key *bidutil.PrivateKey) ([]byte, error) {

	data := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": {
				{Name: "exchange", Type: "address"},
				{Name: "maker", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "makerRelayerFee", Type: "uint256"},
				{Name: "takerRelayerFee", Type: "uint256"},
				{Name: "makerProtocolFee", Type: "uint256"},
				{Name: "takerProtocolFee", Type: "uint256"},
				{Name: "feeRecipient", Type: "address"},
				{Name: "feeMethod", Type: "uint8"},
				{Name: "side", Type: "uint8"},
				{Name: "saleKind", Type: "uint8"},
				{Name: "target", Type: "address"},
				{Name: "howToCall", Type: "uint8"},
				{Name: "calldata", Type: "bytes"},
				{Name: "replacementPattern", Type: "bytes"},
				{Name: "staticTarget", Type: "address"},
				{Name: "staticExtradata", Type: "bytes"},
				{Name: "paymentToken", Type: "address"},
				{Name: "basePrice", Type: "uint256"},
				{Name: "extra", Type: "uint256"},
				{Name: "listingTime", Type: "uint256"},
				{Name: "expirationTime", Type: "uint256"},
				{Name: "salt", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
			},
		},
		Domain: apitypes.TypedDataDomain{
			Name:              "Wyvern Exchange Contract",
			Version:           "2.3",
			ChainId:           math.NewHexOrDecimal256(ChainID()),
			VerifyingContract: order.Exchange.Hex(),
		},
		PrimaryType: "Order",
		Message: apitypes.TypedDataMessage{
			"exchange":           order.Exchange.Hex(),
			"maker":              order.Maker.Hex(),
			"taker":              order.Taker.Hex(),
			"makerRelayerFee":    order.MakerRelayerFee.String(),
			"takerRelayerFee":    order.TakerRelayerFee.String(),
			"makerProtocolFee":   order.MakerProtocolFee.String(),
			"takerProtocolFee":   order.TakerProtocolFee.String(),
			"feeRecipient":       order.FeeRecipient.String(),
			"feeMethod":          fmt.Sprint(order.FeeMethod),
			"side":               fmt.Sprint(order.Side),
			"saleKind":           fmt.Sprint(order.SaleKind),
			"target":             order.Target.Hex(),
			"howToCall":          fmt.Sprint(order.HowToCall),
			"calldata":           order.Calldata,
			"replacementPattern": order.ReplacementPattern,
			"staticTarget":       order.StaticTarget.Hex(),
			"staticExtradata":    order.StaticExtradata,
			"paymentToken":       order.PaymentToken.Hex(),
			"basePrice":          order.BasePrice.String(),
			"extra":              order.Extra.String(),
			"listingTime":        order.ListingTime.String(),
			"expirationTime":     order.ExpirationTime.String(),
			"salt":               order.Salt.String(),
			"nonce":              order.Nonce.String(),
		},
	}

	domainSeparator, err := data.HashStruct("EIP712Domain", data.Domain.Map())
	if err != nil {
		return nil, err
	}

	typedDataHash, err := data.HashStruct(data.PrimaryType, data.Message)
	if err != nil {
		return nil, err
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))

	dataHash := crypto.Keccak256(rawData)

	signature, err := key.Sign(dataHash)
	if err != nil {
		return nil, err
	}

	return signature, nil
}
