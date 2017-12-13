// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openpgp

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"hash"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/keybase/go-crypto/openpgp/packet"
	"github.com/keybase/go-crypto/rsa"
)

func TestSignDetached(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	out := bytes.NewBuffer(nil)
	message := bytes.NewBufferString(signedInput)
	err := DetachSign(out, kring[0], message, nil)
	if err != nil {
		t.Error(err)
	}

	testDetachedSignature(t, kring, out, signedInput, "check", testKey1KeyId)
}

func TestSignTextDetached(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	out := bytes.NewBuffer(nil)
	message := bytes.NewBufferString(signedInput)
	err := DetachSignText(out, kring[0], message, nil)
	if err != nil {
		t.Error(err)
	}

	testDetachedSignature(t, kring, out, signedInput, "check", testKey1KeyId)
}

func TestSignDetachedDSA(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(dsaTestKeyPrivateHex))
	out := bytes.NewBuffer(nil)
	message := bytes.NewBufferString(signedInput)
	err := DetachSign(out, kring[0], message, nil)
	if err != nil {
		t.Error(err)
	}

	testDetachedSignature(t, kring, out, signedInput, "check", testKey3KeyId)
}

type TestRSASigner struct {
	hash.Hash
	PublicKeyId uint64
	PrivateKey  *rsa.PrivateKey
}

func (s *TestRSASigner) KeyId() uint64 {
	return s.PublicKeyId
}

func (s *TestRSASigner) PublicKeyAlgo() packet.PublicKeyAlgorithm {
	return packet.PubKeyAlgoRSA
}

func (s *TestRSASigner) Sign(sig *packet.Signature) (err error) {
	digest := s.Sum(nil)

	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, s.PrivateKey, sig.Hash, digest)
	if err != nil {
		return
	}

	sig.RSASignature = packet.FromBytes(sigBytes)

	return
}

func TestSignWithSigner(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	if err != nil {
		t.Error(err)
	}

	signerSubkey, ok := kring[0].signingKey(time.Now())
	if !ok {
		t.Error("couldn't get signer subkey")
	}

	keyId := signerSubkey.PrivateKey.KeyId
	privateKey := signerSubkey.PrivateKey.PrivateKey.(*rsa.PrivateKey)

	signer := &TestRSASigner{
		PublicKeyId: keyId,
		PrivateKey:  privateKey,
		Hash:        crypto.SHA256.New(),
	}

	out := bytes.NewBuffer(nil)
	message := bytes.NewBufferString(signedInput)
	err = SignWithSigner(signer, out, message, packet.SigTypeBinary, nil)
	if err != nil {
		t.Error(err)
	}

	testDetachedSignature(t, kring, out, signedInput, "check", testKey1KeyId)
}

type TestErrorSigner struct {
	TestRSASigner
}

func (s *TestErrorSigner) Sign(sig *packet.Signature) error {
	return TestErrorSignerError("error from TestErrorSigner.Sign")
}

type TestErrorSignerError string

func (e TestErrorSignerError) Error() string {
	return string(e)
}

func TestSignerCanReturnErrors(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	if err != nil {
		t.Error(err)
	}

	signerSubkey, ok := kring[0].signingKey(time.Now())
	if !ok {
		t.Error("couldn't get signer subkey")
	}

	keyId := signerSubkey.PrivateKey.KeyId
	privateKey := signerSubkey.PrivateKey.PrivateKey.(*rsa.PrivateKey)

	signer := &TestErrorSigner{
		TestRSASigner: TestRSASigner{
			PublicKeyId: keyId,
			PrivateKey:  privateKey,
			Hash:        crypto.SHA256.New(),
		},
	}

	out := bytes.NewBuffer(nil)
	message := bytes.NewBufferString(signedInput)
	err = SignWithSigner(signer, out, message, packet.SigTypeBinary, nil)
	if err == nil {
		t.Error("expecting error from TestErrorSigner.Sign")
	}

	_, isTestErrorSignerError := err.(TestErrorSignerError)
	if !isTestErrorSignerError {
		t.Error("was expecting error returned from TestErrorSigner.Sign")
	}
}

func TestNewEntity(t *testing.T) {
	if testing.Short() {
		return
	}

	// Check bit-length with no config.
	e, err := NewEntity("Test User", "test", "test@example.com", nil)
	if err != nil {
		t.Errorf("failed to create entity: %s", err)
		return
	}
	bl, err := e.PrimaryKey.BitLength()
	if err != nil {
		t.Errorf("failed to find bit length: %s", err)
	}
	if int(bl) != defaultRSAKeyBits {
		t.Errorf("BitLength %v, expected %v", defaultRSAKeyBits)
	}

	// Check bit-length with a config.
	cfg := &packet.Config{RSABits: 1024}
	e, err = NewEntity("Test User", "test", "test@example.com", cfg)
	if err != nil {
		t.Errorf("failed to create entity: %s", err)
		return
	}
	bl, err = e.PrimaryKey.BitLength()
	if err != nil {
		t.Errorf("failed to find bit length: %s", err)
	}
	if int(bl) != cfg.RSABits {
		t.Errorf("BitLength %v, expected %v", bl, cfg.RSABits)
	}

	w := bytes.NewBuffer(nil)
	if err := e.SerializePrivate(w, nil); err != nil {
		t.Errorf("failed to serialize entity: %s", err)
		return
	}
	serialized := w.Bytes()

	el, err := ReadKeyRing(w)
	if err != nil {
		t.Errorf("failed to reparse entity: %s", err)
		return
	}

	if len(el) != 1 {
		t.Errorf("wrong number of entities found, got %d, want 1", len(el))
	}

	w = bytes.NewBuffer(nil)
	if err := e.SerializePrivate(w, nil); err != nil {
		t.Errorf("failed to serialize entity second time: %s", err)
		return
	}

	if !bytes.Equal(w.Bytes(), serialized) {
		t.Errorf("results differed")
	}
}

func TestSymmetricEncryption(t *testing.T) {
	buf := new(bytes.Buffer)
	plaintext, err := SymmetricallyEncrypt(buf, []byte("testing"), nil, nil)
	if err != nil {
		t.Errorf("error writing headers: %s", err)
		return
	}
	message := []byte("hello world\n")
	_, err = plaintext.Write(message)
	if err != nil {
		t.Errorf("error writing to plaintext writer: %s", err)
	}
	err = plaintext.Close()
	if err != nil {
		t.Errorf("error closing plaintext writer: %s", err)
	}

	md, err := ReadMessage(buf, nil, func(keys []Key, symmetric bool) ([]byte, error) {
		return []byte("testing"), nil
	}, nil)
	if err != nil {
		t.Errorf("error rereading message: %s", err)
	}
	messageBuf := bytes.NewBuffer(nil)
	_, err = io.Copy(messageBuf, md.UnverifiedBody)
	if err != nil {
		t.Errorf("error rereading message: %s", err)
	}
	if !bytes.Equal(message, messageBuf.Bytes()) {
		t.Errorf("recovered message incorrect got '%s', want '%s'", messageBuf.Bytes(), message)
	}
}

var testEncryptionTests = []struct {
	keyRingHex string
	isSigned   bool
}{
	{
		testKeys1And2PrivateHex,
		false,
	},
	{
		testKeys1And2PrivateHex,
		true,
	},
	{
		dsaElGamalTestKeysHex,
		false,
	},
	{
		dsaElGamalTestKeysHex,
		true,
	},
}

func TestEncryption(t *testing.T) {
	for i, test := range testEncryptionTests {
		kring, _ := ReadKeyRing(readerFromHex(test.keyRingHex))

		passphrase := []byte("passphrase")
		for _, entity := range kring {
			if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
				err := entity.PrivateKey.Decrypt(passphrase)
				if err != nil {
					t.Errorf("#%d: failed to decrypt key", i)
				}
			}
			for _, subkey := range entity.Subkeys {
				if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
					err := subkey.PrivateKey.Decrypt(passphrase)
					if err != nil {
						t.Errorf("#%d: failed to decrypt subkey", i)
					}
				}
			}
		}

		var signed *Entity
		if test.isSigned {
			signed = kring[0]
		}

		buf := new(bytes.Buffer)
		w, err := Encrypt(buf, kring[:1], signed, nil /* no hints */, nil)
		if err != nil {
			t.Errorf("#%d: error in Encrypt: %s", i, err)
			continue
		}

		const message = "testing"
		_, err = w.Write([]byte(message))
		if err != nil {
			t.Errorf("#%d: error writing plaintext: %s", i, err)
			continue
		}
		err = w.Close()
		if err != nil {
			t.Errorf("#%d: error closing WriteCloser: %s", i, err)
			continue
		}

		md, err := ReadMessage(buf, kring, nil /* no prompt */, nil)
		if err != nil {
			t.Errorf("#%d: error reading message: %s", i, err)
			continue
		}

		testTime, _ := time.Parse("2006-01-02", "2013-07-01")
		if test.isSigned {
			signKey, _ := kring[0].signingKey(testTime)
			expectedKeyId := signKey.PublicKey.KeyId
			if md.SignedByKeyId != expectedKeyId {
				t.Errorf("#%d: message signed by wrong key id, got: %d, want: %d", i, *md.SignedBy, expectedKeyId)
			}
			if md.SignedBy == nil {
				t.Errorf("#%d: failed to find the signing Entity", i)
			}
		}

		plaintext, err := ioutil.ReadAll(md.UnverifiedBody)
		if err != nil {
			t.Errorf("#%d: error reading encrypted contents: %s", i, err)
			continue
		}

		encryptKey, _ := kring[0].encryptionKey(testTime)
		expectedKeyId := encryptKey.PublicKey.KeyId
		if len(md.EncryptedToKeyIds) != 1 || md.EncryptedToKeyIds[0] != expectedKeyId {
			t.Errorf("#%d: expected message to be encrypted to %v, but got %#v", i, expectedKeyId, md.EncryptedToKeyIds)
		}

		if string(plaintext) != message {
			t.Errorf("#%d: got: %s, want: %s", i, string(plaintext), message)
		}

		if test.isSigned {
			if md.SignatureError != nil {
				t.Errorf("#%d: signature error: %s", i, md.SignatureError)
			}
			if md.Signature == nil {
				t.Error("signature missing")
			}
		}
	}
}
