package bidutil

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Private Key

type PrivateKey struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
}

func NewPrivateKey(privateKeyHex string) (PrivateKey, error) {
	privateKey := PrivateKey{}
	err := privateKey.Set(privateKeyHex)
	if err != nil {
		return PrivateKey{}, err
	}
	return privateKey, nil
}

func (privateKey *PrivateKey) Set(privateKeyHex string) error {

	if !(len(privateKeyHex) == 64 || len(privateKeyHex) == 66) {
		return errors.New("Private key must be 32 bytes")
	}
	// Fix 0x if not specified
	if privateKeyHex[:2] != "0x" {
		privateKeyHex = "0x" + privateKeyHex
	}

	privateKeyBytes, err := hexutil.Decode(privateKeyHex)
	if err != nil {
		return err
	}

	if len(privateKeyBytes) != 32 {
		return errors.New("Private key must be 32 bytes")
	}

	privateKey.privateKey, err = crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return err
	}

	var ok bool
	privateKey.publicKey, ok = privateKey.privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return errors.New("publicKey is not type of ecdsa public key")
	}

	privateKey.address = crypto.PubkeyToAddress(*privateKey.publicKey)

	return nil
}

func (privateKey *PrivateKey) Sign(digestHash []byte) ([]byte, error) {
	if len(digestHash) != 32 {
		return nil, errors.New("Sign requires 32 byte hash.")
	}
	signature, err := crypto.Sign(digestHash, privateKey.privateKey)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// func (privateKey *PrivateKey) Verify(digestHash, signature []byte) bool {
//     publicKeyBytes := crypto.FromECDSAPub(privateKey.publicKey)
//     signatureNoRecoverID := signature[:len(signature)-1] // remove recovery ID
//     return crypto.VerifySignature(publicKeyBytes, digestHash, signatureNoRecoverID)
// }

func (privateKey *PrivateKey) PublicAddress() common.Address {
	return privateKey.address
}
