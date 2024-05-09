package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"hash"
)

// GenerateKeys EllipticCurve public and private keys
func GenerateKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	pubKeyCurve := elliptic.P256()
	var err error
	privKey, err := ecdsa.GenerateKey(pubKeyCurve, rand.Reader)

	return privKey, &privKey.PublicKey, err
}

// EncodePrivate private key
func EncodePrivate(privKey *ecdsa.PrivateKey) (string, error) {

	encoded, err := x509.MarshalECPrivateKey(privKey)

	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})

	return string(pemEncoded), nil
}

// EncodePublic public key
func EncodePublic(pubKey *ecdsa.PublicKey) (string, error) {

	encoded, err := x509.MarshalPKIXPublicKey(pubKey)

	if err != nil {
		return "", err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded})

	return string(pemEncodedPub), nil
}

// DecodePrivate private key
func DecodePrivate(pemEncodedPriv string) (*ecdsa.PrivateKey, error) {
	blockPriv, _ := pem.Decode([]byte(pemEncodedPriv))

	x509EncodedPriv := blockPriv.Bytes

	privateKey, err := x509.ParseECPrivateKey(x509EncodedPriv)

	return privateKey, err
}

// DecodePublic public key
func DecodePublic(pemEncodedPub string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))

	x509EncodedPub := blockPub.Bytes

	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey, err
}

func Sign(message []byte, privKey *ecdsa.PrivateKey) (string, error) {
	var h hash.Hash
	h = md5.New()
	h.Write(message)
	signature, err := ecdsa.SignASN1(rand.Reader, privKey, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(signature), nil
}

func Verify(message []byte, sign string, pubKey *ecdsa.PublicKey) bool {
	var h hash.Hash
	h = md5.New()
	h.Write(message)
	signature, err := hex.DecodeString(sign)
	if err != nil {
		return false
	}

	return ecdsa.VerifyASN1(pubKey, h.Sum(nil), signature)
}
