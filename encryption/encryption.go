package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func EncryptAmount(amount []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	encryptedAmount, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, amount, nil)
	if err != nil {
		return nil, err
	}
	return encryptedAmount, nil
}

func DecryptAmount(encryptedAmount []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	decryptedAmount, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAmount, nil)
	if err != nil {
		return nil, err
	}
	return decryptedAmount, nil
}

func ExportPublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
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

func ImportPublicKeyFromPEM(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key type is not RSA")
	}

	return rsaPub, nil
}