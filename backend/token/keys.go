package token

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	database "github.com/DEVunderdog/transcript-generator-backend/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/pbkdf2"
)

const iter = 100000

type JWTKeyResponse struct {
	PublicKey  string
	PrivateKey []byte
}

func generateRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, &privateKey.PublicKey, nil
}

func encryptPrivateKey(privateKey *rsa.PrivateKey, passphrase []byte) ([]byte, error) {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	blockBytes := pem.EncodeToMemory(block)

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := pbkdf2.Key(passphrase, salt, iter, 32, sha256.New)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := gcm.Seal(nil, nonce, blockBytes, nil)

	result := make([]byte, len(salt)+len(nonce)+len(cipherText))

	copy(result, salt)
	copy(result[len(salt):], nonce)
	copy(result[len(salt)+len(nonce):], cipherText)

	return result, nil
}

func decryptPrivateKey(encryptedKey []byte, passphrase []byte) (*rsa.PrivateKey, error) {

	if len(encryptedKey) < 16+12 {
		return nil, errors.New("invalid encrypted key format")
	}

	salt := encryptedKey[:16]
	nonce := encryptedKey[16:28]
	cipherText := encryptedKey[28:]

	key := pbkdf2.Key(passphrase, salt, iter, 32, sha256.New)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)

	if err != nil {
		return nil, err
	}

	pemBytes, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBytes)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func encodePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	pubASN1, err := x509.MarshalPKIXPublicKey(publicKey)

	if err != nil {
		return nil, err
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return pubBytes, nil
}

func generateAndStoreKeys(passphrase string, store database.Store, ctx context.Context, keyPurpose string) error {
	privateKey, publicKey, err := generateRSAKeys()

	if err != nil {
		return err
	}

	encryptedPrivateKey, err := encryptPrivateKey(privateKey, []byte(passphrase))
	if err != nil {
		return err
	}

	publicKeyPEM, err := encodePublicKey(publicKey)
	if err != nil {
		return err
	}

	args := database.CreateEncryptionKeysParams{
		PublicKey:  string(publicKeyPEM),
		PrivateKey: encryptedPrivateKey,
		IsActive: pgtype.Bool{
			Valid: true,
			Bool:  true,
		},
		Purpose: keyPurpose,
	}

	_, err = store.CreateEncryptionKeys(ctx, args)
	if err != nil {
		return err
	}

	return nil
}

func InitializeJWTKeys(passphrase string, store database.Store, ctx context.Context, keyPurpose string) error {
	count, err := store.CountEncryptionKeys(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		return generateAndStoreKeys(passphrase, store, ctx, keyPurpose)
	}

	return nil
}

func GetKeyBasedOnPurpose(ctx context.Context, store database.Store, purpose string) (*JWTKeyResponse, error) {
	jwtStruct, err := store.GetActiveKeyBasedOnPurpose(ctx, purpose)
	if err != nil {
		return nil, err
	}

	data := &JWTKeyResponse{
		PublicKey:  jwtStruct.PublicKey,
		PrivateKey: jwtStruct.PrivateKey,
	}

	return data, nil
}

func GetPrivateKey(key []byte, passphrase []byte) (*rsa.PrivateKey, error) {
	return decryptPrivateKey(key, passphrase)
}

func GetPublicKey(key []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(key)

	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("not an RSA based public key")
	}
}
