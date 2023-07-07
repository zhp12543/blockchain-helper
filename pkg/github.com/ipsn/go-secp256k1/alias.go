package secp256k1

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var (
	ErrInvalidMsgLen       = secp256k1.ErrInvalidMsgLen
	ErrInvalidSignatureLen = secp256k1.ErrInvalidSignatureLen
	ErrInvalidRecoveryID   = secp256k1.ErrInvalidRecoveryID
	ErrInvalidKey          = secp256k1.ErrInvalidKey
	ErrInvalidPubkey       = secp256k1.ErrInvalidPubkey
	ErrSignFailed          = secp256k1.ErrSignFailed
	ErrRecoverFailed       = secp256k1.ErrRecoverFailed
)

func CompressPubkey(x, y *big.Int) []byte {
	return secp256k1.CompressPubkey(x, y)
}

func DecompressPubkey(pubkey []byte) (x, y *big.Int) {
	return secp256k1.DecompressPubkey(pubkey)
}

func RecoverPubkey(msg []byte, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(msg, sig)
}

func Sign(msg []byte, seckey []byte) ([]byte, error) {
	return secp256k1.Sign(msg, seckey)
}

func VerifySignature(pubkey, msg, signature []byte) bool {
	return secp256k1.VerifySignature(pubkey, msg, signature)
}

type BitCurve = secp256k1.BitCurve

func S256() *BitCurve {
	return secp256k1.S256()
}
