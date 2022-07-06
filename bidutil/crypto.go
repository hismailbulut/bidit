package bidutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

const HASH_COUNT = 256

func Sha256Hash(data []byte, times int) []byte {
	hash := sha256.Sum256(data)
	for i := 0; i < times-1; i++ {
		hash = sha256.Sum256(hash[:])
	}
	return hash[:]
}

func PasswordHash(p []byte) []byte {
	return Sha256Hash(p, HASH_COUNT)
}

func AesEncrypt(data []byte, key []byte) ([]byte, error) {
	if len(key) != 256/8 {
		return nil, errors.New("key must be 256 bits")
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func AesDecrypt(data []byte, key []byte) ([]byte, error) {
	if len(key) != 256/8 {
		return nil, errors.New("key must be 256 bits")
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("data length is less than nonce size")
	}
	nonce, cipherdata := data[:nonceSize], data[nonceSize:]
	decrypted, err := gcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func Encrypt(data []byte, key []byte) ([]byte, error) {
	var err error
	for i := 0; i < HASH_COUNT; i++ {
		keyHash := sha256.Sum256(key)
		key = keyHash[:]
		data, err = AesEncrypt(data, key)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

func Decrypt(data []byte, key []byte) ([]byte, error) {
	var err error
	for i := 0; i < HASH_COUNT; i++ {
		keyHash := Sha256Hash(key, HASH_COUNT-i)
		data, err = AesDecrypt(data, keyHash)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
