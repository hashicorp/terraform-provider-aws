package brainpool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
)

func testCurve(t *testing.T, curve elliptic.Curve) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("test")
	r, s, err := ecdsa.Sign(rand.Reader, priv, msg)
	if err != nil {
		t.Fatal(err)
	}

	if !ecdsa.Verify(&priv.PublicKey, msg, r, s) {
		t.Fatal("signature didn't verify.")
	}
}

func TestP256t1(t *testing.T) {
	testCurve(t, P256t1())
}

func TestP256r1(t *testing.T) {
	testCurve(t, P256r1())
}

func TestP384t1(t *testing.T) {
	testCurve(t, P384t1())
}

func TestP384r1(t *testing.T) {
	testCurve(t, P384r1())
}

func TestP512t1(t *testing.T) {
	testCurve(t, P512t1())
}

func TestP512r1(t *testing.T) {
	testCurve(t, P512r1())
}
