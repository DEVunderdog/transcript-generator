package token

import (
	"crypto/rsa"
)

type TokenMaker struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func NewTokenMaker(publicKey *rsa.PublicKey, privateKey *rsa.PrivateKey) *TokenMaker {
	return &TokenMaker{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

