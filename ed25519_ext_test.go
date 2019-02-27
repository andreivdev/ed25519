// Copyright 2019 Spacemesh Authors
// ed25519 extensions tests

package ed25519

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/spacemeshos/ed25519/internal/edwards25519"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestInvertModLOne(t *testing.T) {
	var x, xInv, z [32]byte
	x[0] = byte(1)
	InvertModL(&xInv, &x)
	assert.Equal(t, "1", ToInt(xInv[:]).String())
	edwards25519.ScMulAdd(&x, &x, &xInv, &z)
	outVal := ToInt(x[:])
	assert.Equal(t, "1", outVal.String(), "expected 0 * 0 == 0")
}

func TestInvertModL2(t *testing.T) {
	var x, xInv, z [32]byte
	x[0] = byte(2)
	InvertModL(&xInv, &x)
	fmt.Printf("Hex string: 0x%s\n", hex.EncodeToString(xInv[:]))
	fmt.Printf("Int value: %s\n", ToInt(xInv[:]).String())

	edwards25519.ScMulAdd(&x, &x, &xInv, &z)

	outVal := ToInt(x[:])
	fmt.Printf("Hex string: 0x%s\n", hex.EncodeToString(x[:]))
	fmt.Printf("Int value: %s\n", outVal.String())

	// checking that we actually got the inverse - result should be 1.
	assert.Equal(t, "1", outVal.String(), "expected x * xInv == 1")
}

func TestInvertModLRnd(testing *testing.T) {

	var t, tinv, z, out [32]byte

	for i := 1; i < 10000; i++ {
		n, err := rand.Read(t[:])
		assert.NoError(testing, err, "no system entropy")
		assert.Equal(testing, 32, n, "expected 32 bytes of entropy")
		InvertModL(&tinv, &t)
		edwards25519.ScMulAdd(&out, &t, &tinv, &z)
		assert.Equal(testing, "1", ToInt(out[:]).String(), "expected t * tinv to equal 1")
	}
}

// ToInt returns a big int with the value of 256^0*b[0]+256^1*b[1]+...+256^31*b[len(b)-1]
// b must be a non-empty bytes slice. ToInt is a test helper function.
func ToInt(b []byte) *big.Int {
	res := big.NewInt(0)
	mul := big.NewInt(0)
	c := big.NewInt(256)
	t := big.NewInt(0)
	data := big.NewInt(0)
	l := len(b)

	for i := 0; i < l; i++ {

		// 256^i
		mul = mul.Exp(c, big.NewInt(int64(i)), nil)

		// res[i] = 256^i * b[i]
		data.SetUint64(uint64(b[i]))
		t = t.Mul(data, mul)
		res = res.Add(res, t)
	}
	return res
}

func TestPublicKeyExtraction(t *testing.T) {
	var zero zeroReader
	public, private, _ := GenerateKey(zero)
	message := []byte("test message")
	sig := SignExt(private, message)

	public1, err := ExtractPublicKey(message, sig)

	assert.NoError(t, err)
	assert.EqualValues(t, public, public1, "expected same public key")

	wrongMessage := []byte("wrong message")
	public2, err := ExtractPublicKey(wrongMessage, sig)

	assert.NoError(t, err)
	if bytes.Compare(public, public2) == 0 {
		t.Errorf("expected different public keys")
	}
}

func TestSignVerifyExt(t *testing.T) {
	var zero zeroReader
	public, private, _ := GenerateKey(zero)

	message := []byte("test message")
	sig := SignExt(private, message)
	if !Verify(public, message, sig) {
		t.Errorf("valid signature rejected")
	}

	wrongMessage := []byte("wrong message")
	if Verify(public, wrongMessage, sig) {
		t.Errorf("signature of different message accepted")
	}
}

func BenchmarkPublicKeyExtraction(b *testing.B) {
	var zero zeroReader
	_, priv, err := GenerateKey(zero)
	if err != nil {
		b.Fatal(err)
	}
	message := []byte("Hello, world!")
	sig := SignExt(priv, message)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractPublicKey(message, sig)
	}
}

func BenchmarkSigningExt(b *testing.B) {
	var zero zeroReader
	_, priv, err := GenerateKey(zero)
	if err != nil {
		b.Fatal(err)
	}
	message := []byte("Hello, world!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SignExt(priv, message)
	}
}

func BenchmarkVerificationExt(b *testing.B) {
	var zero zeroReader
	pub, priv, err := GenerateKey(zero)
	if err != nil {
		b.Fatal(err)
	}
	message := []byte("Hello, world!")
	signature := SignExt(priv, message)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(pub, message, signature)
	}
}
