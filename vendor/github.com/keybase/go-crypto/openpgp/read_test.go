// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openpgp

import (
	"bytes"
	_ "crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/keybase/go-crypto/openpgp/armor"
	"github.com/keybase/go-crypto/openpgp/errors"
	"github.com/keybase/go-crypto/openpgp/packet"
)

func readerFromHex(s string) io.Reader {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic("readerFromHex: bad input")
	}
	return bytes.NewBuffer(data)
}

func TestReadKeyRing(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(testKeys1And2Hex))
	if err != nil {
		t.Error(err)
		return
	}
	if len(kring) != 2 || uint32(kring[0].PrimaryKey.KeyId) != 0xC20C31BB || uint32(kring[1].PrimaryKey.KeyId) != 0x1E35246B {
		t.Errorf("bad keyring: %#v", kring)
	}
}

func TestRereadKeyRing(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(testKeys1And2Hex))
	if err != nil {
		t.Errorf("error in initial parse: %s", err)
		return
	}
	out := new(bytes.Buffer)
	err = kring[0].Serialize(out)
	if err != nil {
		t.Errorf("error in serialization: %s", err)
		return
	}
	kring, err = ReadKeyRing(out)
	if err != nil {
		t.Errorf("error in second parse: %s", err)
		return
	}

	if len(kring) != 1 || uint32(kring[0].PrimaryKey.KeyId) != 0xC20C31BB {
		t.Errorf("bad keyring: %#v", kring)
	}
}

func TestReadPrivateKeyRing(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	if err != nil {
		t.Error(err)
		return
	}
	if len(kring) != 2 || uint32(kring[0].PrimaryKey.KeyId) != 0xC20C31BB || uint32(kring[1].PrimaryKey.KeyId) != 0x1E35246B || kring[0].PrimaryKey == nil {
		t.Errorf("bad keyring: %#v", kring)
	}
}

func TestReadDSAKey(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(dsaTestKeyHex))
	if err != nil {
		t.Error(err)
		return
	}
	if len(kring) != 1 || uint32(kring[0].PrimaryKey.KeyId) != 0x0CCC0360 {
		t.Errorf("bad parse: %#v", kring)
	}
}

func TestDSAHashTruncatation(t *testing.T) {
	// dsaKeyWithSHA512 was generated with GnuPG and --cert-digest-algo
	// SHA512 in order to require DSA hash truncation to verify correctly.
	_, err := ReadKeyRing(readerFromHex(dsaKeyWithSHA512))
	if err != nil {
		t.Error(err)
	}
}

func TestGetKeyById(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))

	keys := kring.KeysById(0xa34d7e18c20c31bb, nil)
	if len(keys) != 1 || keys[0].Entity != kring[0] {
		t.Errorf("bad result for 0xa34d7e18c20c31bb: %#v", keys)
	}

	keys = kring.KeysById(0xfd94408d4543314f, nil)
	if len(keys) != 1 || keys[0].Entity != kring[0] {
		t.Errorf("bad result for 0xa34d7e18c20c31bb: %#v", keys)
	}
}

func checkSignedMessage(t *testing.T, signedHex, expected string) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))

	md, err := ReadMessage(readerFromHex(signedHex), kring, nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if !md.IsSigned || md.SignedByKeyId != 0xa34d7e18c20c31bb || md.SignedBy == nil || md.IsEncrypted || md.IsSymmetricallyEncrypted || len(md.EncryptedToKeyIds) != 0 || md.IsSymmetricallyEncrypted || md.MultiSig {
		t.Errorf("bad MessageDetails: %#v", md)
	}

	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Errorf("error reading UnverifiedBody: %s", err)
	}
	if string(contents) != expected {
		t.Errorf("bad UnverifiedBody got:%s want:%s", string(contents), expected)
	}
	if md.SignatureError != nil || md.Signature == nil {
		t.Errorf("failed to validate: %s", md.SignatureError)
	}
}

func TestSignedMessage(t *testing.T) {
	checkSignedMessage(t, signedMessageHex, signedInput)
}

func TestTextSignedMessage(t *testing.T) {
	checkSignedMessage(t, signedTextMessageHex, signedTextInput)
}

// The reader should detect "compressed quines", which are compressed
// packets that expand into themselves and cause an infinite recursive
// parsing loop.
// The packet in this test case comes from Taylor R. Campbell at
// http://mumble.net/~campbell/misc/pgp-quine/
func TestCampbellQuine(t *testing.T) {
	md, err := ReadMessage(readerFromHex(campbellQuine), nil, nil, nil)
	if md != nil {
		t.Errorf("Reading a compressed quine should not return any data: %#v", md)
	}
	structural, ok := err.(errors.StructuralError)
	if !ok {
		t.Fatalf("Unexpected class of error: %T", err)
	}
	if !strings.Contains(string(structural), "too many layers of packets") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

var signedEncryptedMessageTests = []struct {
	keyRingHex       string
	messageHex       string
	signedByKeyId    uint64
	encryptedToKeyId uint64
}{
	{
		testKeys1And2PrivateHex,
		signedEncryptedMessageHex,
		0xa34d7e18c20c31bb,
		0x2a67d68660df41c7,
	},
	{
		dsaElGamalTestKeysHex,
		signedEncryptedMessage2Hex,
		0x33af447ccd759b09,
		0xcf6a7abcd43e3673,
	},
}

func TestSignedEncryptedMessage(t *testing.T) {
	for i, test := range signedEncryptedMessageTests {
		expected := "Signed and encrypted message\n"
		kring, _ := ReadKeyRing(readerFromHex(test.keyRingHex))
		prompt := func(keys []Key, symmetric bool) ([]byte, error) {
			if symmetric {
				t.Errorf("prompt: message was marked as symmetrically encrypted")
				return nil, errors.ErrKeyIncorrect
			}

			if len(keys) == 0 {
				t.Error("prompt: no keys requested")
				return nil, errors.ErrKeyIncorrect
			}

			err := keys[0].PrivateKey.Decrypt([]byte("passphrase"))
			if err != nil {
				t.Errorf("prompt: error decrypting key: %s", err)
				return nil, errors.ErrKeyIncorrect
			}

			return nil, nil
		}

		md, err := ReadMessage(readerFromHex(test.messageHex), kring, prompt, nil)
		if err != nil {
			t.Errorf("#%d: error reading message: %s", i, err)
			return
		}

		if !md.IsSigned || md.SignedByKeyId != test.signedByKeyId || md.SignedBy == nil || !md.IsEncrypted || md.IsSymmetricallyEncrypted || len(md.EncryptedToKeyIds) == 0 || md.EncryptedToKeyIds[0] != test.encryptedToKeyId || md.MultiSig {
			t.Errorf("#%d: bad MessageDetails: %#v", i, md)
		}

		contents, err := ioutil.ReadAll(md.UnverifiedBody)
		if err != nil {
			t.Errorf("#%d: error reading UnverifiedBody: %s", i, err)
		}
		if string(contents) != expected {
			t.Errorf("#%d: bad UnverifiedBody got:%s want:%s", i, string(contents), expected)
		}

		if md.SignatureError != nil || md.Signature == nil {
			t.Errorf("#%d: failed to validate: %s", i, md.SignatureError)
		}
	}
}

func TestUnspecifiedRecipient(t *testing.T) {
	expected := "Recipient unspecified\n"
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))

	md, err := ReadMessage(readerFromHex(recipientUnspecifiedHex), kring, nil, nil)
	if err != nil {
		t.Errorf("error reading message: %s", err)
		return
	}

	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Errorf("error reading UnverifiedBody: %s", err)
	}
	if string(contents) != expected {
		t.Errorf("bad UnverifiedBody got:%s want:%s", string(contents), expected)
	}
}

func TestSymmetricallyEncrypted(t *testing.T) {
	firstTimeCalled := true

	prompt := func(keys []Key, symmetric bool) ([]byte, error) {
		if len(keys) != 0 {
			t.Errorf("prompt: len(keys) = %d (want 0)", len(keys))
		}

		if !symmetric {
			t.Errorf("symmetric is not set")
		}

		if firstTimeCalled {
			firstTimeCalled = false
			return []byte("wrongpassword"), nil
		}

		return []byte("password"), nil
	}

	md, err := ReadMessage(readerFromHex(symmetricallyEncryptedCompressedHex), nil, prompt, nil)
	if err != nil {
		t.Errorf("ReadMessage: %s", err)
		return
	}

	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Errorf("ReadAll: %s", err)
	}

	expectedCreationTime := uint32(1295992998)
	if md.LiteralData.Time != expectedCreationTime {
		t.Errorf("LiteralData.Time is %d, want %d", md.LiteralData.Time, expectedCreationTime)
	}

	const expected = "Symmetrically encrypted.\n"
	if string(contents) != expected {
		t.Errorf("contents got: %s want: %s", string(contents), expected)
	}
}

func testDetachedSignature(t *testing.T, kring KeyRing, signature io.Reader, sigInput, tag string, expectedSignerKeyId uint64) {
	signed := bytes.NewBufferString(sigInput)
	signer, err := CheckDetachedSignature(kring, signed, signature)
	if err != nil {
		t.Errorf("%s: signature error: %s", tag, err)
		return
	}
	if signer == nil {
		t.Errorf("%s: signer is nil", tag)
		return
	}
	if signer.PrimaryKey.KeyId != expectedSignerKeyId {
		t.Errorf("%s: wrong signer got:%x want:%x", tag, signer.PrimaryKey.KeyId, expectedSignerKeyId)
	}
}

func TestDetachedSignature(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))
	testDetachedSignature(t, kring, readerFromHex(detachedSignatureHex), signedInput, "binary", testKey1KeyId)
	testDetachedSignature(t, kring, readerFromHex(detachedSignatureTextHex), signedInput, "text", testKey1KeyId)
	testDetachedSignature(t, kring, readerFromHex(detachedSignatureV3TextHex), signedInput, "v3", testKey1KeyId)

	incorrectSignedInput := signedInput + "X"
	_, err := CheckDetachedSignature(kring, bytes.NewBufferString(incorrectSignedInput), readerFromHex(detachedSignatureHex))
	if err == nil {
		t.Fatal("CheckDetachedSignature returned without error for bad signature")
	}
	if err == errors.ErrUnknownIssuer {
		t.Fatal("CheckDetachedSignature returned ErrUnknownIssuer when the signer was known, but the signature invalid")
	}
}

func TestDetachedSignatureDSA(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(dsaTestKeyHex))
	testDetachedSignature(t, kring, readerFromHex(detachedSignatureDSAHex), signedInput, "binary", testKey3KeyId)
}

func TestMultipleSignaturePacketsDSA(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(dsaTestKeyHex))
	testDetachedSignature(t, kring, readerFromHex(missingHashFunctionHex+detachedSignatureDSAHex), signedInput, "binary", testKey3KeyId)
}

func testHashFunctionError(t *testing.T, signatureHex string) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))
	_, err := CheckDetachedSignature(kring, nil, readerFromHex(signatureHex))
	if err == nil {
		t.Fatal("Packet with bad hash type was correctly parsed")
	}
	unsupported, ok := err.(errors.UnsupportedError)
	if !ok {
		t.Fatalf("Unexpected class of error: %s", err)
	}
	if !strings.Contains(string(unsupported), "hash ") {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestUnknownHashFunction(t *testing.T) {
	// unknownHashFunctionHex contains a signature packet with hash
	// function type 153 (which isn't a real hash function id).
	testHashFunctionError(t, unknownHashFunctionHex)
}

func TestMissingHashFunction(t *testing.T) {
	// missingHashFunctionHex contains a signature packet that uses
	// RIPEMD160, which isn't compiled in.  Since that's the only signature
	// packet we don't find any suitable packets and end up with ErrUnknownIssuer
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))
	_, err := CheckDetachedSignature(kring, nil, readerFromHex(missingHashFunctionHex))
	if err == nil {
		t.Fatal("Packet with missing hash type was correctly parsed")
	}
	if err != errors.ErrUnknownIssuer {
		t.Fatalf("Unexpected class of error: %s", err)
	}
}

func TestReadingArmoredPrivateKey(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(armoredPrivateKeyBlock))
	if err != nil {
		t.Error(err)
	}
	if len(el) != 1 {
		t.Errorf("got %d entities, wanted 1\n", len(el))
	}
}

func rawToArmored(raw []byte, priv bool) (ret string, err error) {

	var writer io.WriteCloser
	var out bytes.Buffer
	var which string

	if priv {
		which = "PRIVATE"
	} else {
		which = "PUBLIC"
	}
	hdr := fmt.Sprintf("PGP %s KEY BLOCK", which)

	writer, err = armor.Encode(&out, hdr, nil)

	if err != nil {
		return
	}
	if _, err = writer.Write(raw); err != nil {
		return
	}
	writer.Close()
	ret = out.String()
	return
}

const detachedMsg = "Thou still unravish'd bride of quietness, Thou foster-child of silence and slow time,"

func trySigning(e *Entity) (string, error) {
	txt := bytes.NewBufferString(detachedMsg)
	var out bytes.Buffer
	err := ArmoredDetachSign(&out, e, txt, nil)
	return out.String(), err
}

func TestSigningSubkey(t *testing.T) {
	k := openPrivateKey(t, signingSubkey, signingSubkeyPassphrase, true, 2)
	_, err := trySigning(k)
	if err != nil {
		t.Fatal(err)
	}
}

func openPrivateKey(t *testing.T, armoredKey string, passphrase string, protected bool, nSubkeys int) *Entity {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(armoredKey))
	if err != nil {
		t.Error(err)
	}
	if len(el) != 1 {
		t.Fatalf("got %d entities, wanted 1\n", len(el))
	}
	k := el[0]
	if k.PrivateKey == nil {
		t.Fatalf("Got nil key, but wanted a private key")
	}
	if err := k.PrivateKey.Decrypt([]byte(passphrase)); err != nil {
		t.Fatalf("failed to decrypt key: %s", err)
	}
	if err := k.PrivateKey.Decrypt([]byte(passphrase + "X")); err != nil {
		t.Fatalf("failed to decrypt key with the wrong key (it shouldn't matter): %s", err)
	}

	decryptions := 0

	// Also decrypt all subkeys (with the same password)
	for i, subkey := range k.Subkeys {
		priv := subkey.PrivateKey
		if priv == nil {
			t.Fatalf("unexpected nil subkey @%d", i)
		}
		err := priv.Decrypt([]byte(passphrase + "X"))

		if protected && err == nil {
			t.Fatalf("expected subkey decryption to fail on %d with bad PW\n", i)
		} else if !protected && err != nil {
			t.Fatalf("Without passphrase-protection, decryption shouldn't fail")
		}
		if err := priv.Decrypt([]byte(passphrase)); err != nil {
			t.Fatalf("failed to decrypt subkey %d: %s\n", i, err)
		} else {
			decryptions++
		}
	}
	if decryptions != nSubkeys {
		t.Fatalf("expected %d decryptions; got %d", nSubkeys, decryptions)
	}
	return k
}

func testSerializePrivate(t *testing.T, keyString string, passphrase string, nSubkeys int) *Entity {

	key := openPrivateKey(t, keyString, passphrase, true, nSubkeys)

	var buf bytes.Buffer
	err := key.SerializePrivate(&buf, nil)
	if err != nil {
		t.Fatal(err)
	}

	armored, err := rawToArmored(buf.Bytes(), true)
	if err != nil {
		t.Fatal(err)
	}

	return openPrivateKey(t, armored, passphrase, false, nSubkeys)
}

func TestGnuS2KDummyEncryptionSubkey(t *testing.T) {
	key := testSerializePrivate(t, gnuDummyS2KPrivateKey, gnuDummyS2KPrivateKeyPassphrase, 1)
	_, err := trySigning(key)
	if err == nil {
		t.Fatal("Expected a signing failure, since we don't have a signing key")
	}
}

func TestGNUS2KDummySigningSubkey(t *testing.T) {
	key := testSerializePrivate(t, gnuDummyS2KPrivateKeyWithSigningSubkey, gnuDummyS2KPrivateKeyWithSigningSubkeyPassphrase, 2)
	_, err := trySigning(key)
	if err != nil {
		t.Fatal("Got a signing failure: %s\n", err)
	}
}

func TestReadingArmoredPublicKey(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(e2ePublicKey))
	if err != nil {
		t.Error(err)
	}
	if len(el) != 1 {
		t.Errorf("didn't get a valid entity")
	}
}

func TestNoArmoredData(t *testing.T) {
	_, err := ReadArmoredKeyRing(bytes.NewBufferString("foo"))
	if _, ok := err.(errors.InvalidArgumentError); !ok {
		t.Errorf("error was not an InvalidArgumentError: %s", err)
	}
}

func testReadMessageError(t *testing.T, messageHex string) {
	buf, err := hex.DecodeString(messageHex)
	if err != nil {
		t.Errorf("hex.DecodeString(): %v", err)
	}

	kr, err := ReadKeyRing(new(bytes.Buffer))
	if err != nil {
		t.Errorf("ReadKeyring(): %v", err)
	}

	_, err = ReadMessage(bytes.NewBuffer(buf), kr,
		func([]Key, bool) ([]byte, error) {
			return []byte("insecure"), nil
		}, nil)

	if err == nil {
		t.Errorf("ReadMessage(): Unexpected nil error")
	}
}

func TestIssue11503(t *testing.T) {
	testReadMessageError(t, "8c040402000aa430aa8228b9248b01fc899a91197130303030")
}

func TestIssue11504(t *testing.T) {
	testReadMessageError(t, "9303000130303030303030303030983002303030303030030000000130")
}

// TestSignatureV3Message tests the verification of V3 signature, generated
// with a modern V4-style key.  Some people have their clients set to generate
// V3 signatures, so it's useful to be able to verify them.
func TestSignatureV3Message(t *testing.T) {
	testSignedV3Message(t, signedMessageV3, keyV4forVerifyingSignedMessageV3)
}

func TestSignatureV4MessageJwb(t *testing.T) {
	testSignedV4Message(t, jwbSignedV3Message, jwbKey)
}

func TestSignatureV4MessageYield(t *testing.T) {
	testSignedV4Message(t, yieldSig, yieldKey)
}

func TestSignatureV4Spiros(t *testing.T) {
	testSignedV4Message(t, spirosSig, spirosKey)
}

func TestSignatureV4Silverbaq(t *testing.T) {
	testSignedV4Message(t, silverbaqSig, silverbaqKey)
}

func TestSignatureWithSubpacket33(t *testing.T) {
	testSignedV4Message(t, subpaacket33Sig, subpacket33Key)
}

func TestBadSignature(t *testing.T) {
	testSignedV4MessageWithError(t, badSig, keyForBagSig, true)
}

func TestBengtSignature(t *testing.T) {
	testSignedV4Message(t, bengtSig, bengtKey)
}

func testSignedV3Message(t *testing.T, armoredMsg string, armoredKey string) {
	sig, err := armor.Decode(strings.NewReader(armoredMsg))
	if err != nil {
		t.Error(err)
		return
	}
	key, err := ReadArmoredKeyRing(strings.NewReader(armoredKey))
	if err != nil {
		t.Error(err)
		return
	}
	md, err := ReadMessage(sig.Body, key, nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Error(err)
		return
	}

	// We'll see a sig error here after reading in the UnverifiedBody above,
	// if there was one to see.
	if err = md.SignatureError; err != nil {
		t.Error(err)
		return
	}

	if md.SignatureV3 == nil {
		t.Errorf("No available signature after checking signature")
		return
	}
	if md.Signature != nil {
		t.Errorf("Did not expect a signature V4 back")
		return
	}
	return
}

func testSignedV4Message(t *testing.T, armoredMsg string, armoredKey string) {
	testSignedV4MessageWithError(t, armoredMsg, armoredKey, false)
}

func testSignedV4MessageWithError(t *testing.T, armoredMsg string, armoredKey string, hasErr bool) {
	sig, err := armor.Decode(strings.NewReader(armoredMsg))
	if err != nil {
		t.Error(err)
		return
	}
	key, err := ReadArmoredKeyRing(strings.NewReader(armoredKey))
	if err != nil {
		t.Error(err)
		return
	}
	md, err := ReadMessage(sig.Body, key, nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Error(err)
		return
	}

	// We'll see a sig error here after reading in the UnverifiedBody above,
	// if there was one to see.
	err = md.SignatureError
	if !hasErr && err != nil {
		t.Error(err)
		return
	}
	if hasErr && err == nil {
		t.Error("expected a signature verification error")
		return
	}

	if md.Signature == nil {
		t.Errorf("Expected a V4 signature back")
		return
	}

	if md.SignatureV3 != nil {
		t.Errorf("Did not expect V3 signature back")
		return
	}
	return
}

func TestEdDSA(t *testing.T) {
	key, err := ReadArmoredKeyRing(strings.NewReader(eddsaPublicKey))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := armor.Decode(strings.NewReader(eddsaSignature))
	if err != nil {
		t.Fatal(err)
	}

	md, err := ReadMessage(sig.Body, key, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	literalData, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Fatal(err)
	}

	// We'll see a sig error here after reading in the UnverifiedBody above,
	// if there was one to see.
	if err = md.SignatureError; err != nil {
		t.Fatal(err)
	}

	if md.Signature == nil {
		t.Fatalf("No available signature after checking signature")
	}

	if string(literalData) != eddsaSignedMsg {
		t.Fatal("got wrong signed message")
	}
	return
}

func testSignWithRevokedSubkey(t *testing.T, privArmored, pubArmored, passphrase string) {
	priv := openPrivateKey(t, privArmored, passphrase, true, 3)
	els, err := ReadArmoredKeyRing(bytes.NewBufferString(pubArmored))
	if err != nil {
		t.Error(err)
	}
	if len(els) != 1 {
		t.Fatalf("got %d entities, wanted 1\n", len(els))
	}
	priv.CopySubkeyRevocations(els[0])
	sig, err := trySigning(priv)
	if err != nil {
		t.Fatal(err)
	}
	var ring EntityList
	ring = append(ring, priv)
	signer, issuer, err := checkArmoredDetachedSignature(ring, strings.NewReader(detachedMsg), strings.NewReader(sig))
	if err != nil {
		t.Fatal(err)
	}
	if issuer == nil {
		t.Fatal("expected non-nil issuer")
	}
	// Subkey[1] is revoked, so we better be using Subkey[2]
	if *issuer != signer.Subkeys[2].PublicKey.KeyId {
		t.Fatalf("Got wrong subkey: wanted %x, but got %x", signer.Subkeys[2].PublicKey.KeyId, *issuer)
	}

	// Now make sure that we can serialize and reimport and we'll get the same
	// results.  In this sense, we'll do better than GPG --export-secret-key,
	// since we'll actually export the revocation statement.
	var buf bytes.Buffer
	err = priv.SerializePrivate(&buf, &packet.Config{ReuseSignaturesOnSerialize: true})
	if err != nil {
		t.Fatal(err)
	}

	armored, err := rawToArmored(buf.Bytes(), true)
	if err != nil {
		t.Fatal(err)
	}
	priv2 := openPrivateKey(t, armored, "", false, 3)

	sig, err = trySigning(priv2)
	if err != nil {
		t.Fatal(err)
	}
	var ring2 EntityList
	ring2 = append(ring2, priv2)
	signer, issuer, err = checkArmoredDetachedSignature(ring2, strings.NewReader(detachedMsg), strings.NewReader(sig))
	if err != nil {
		t.Fatal(err)
	}
	if issuer == nil {
		t.Fatal("expected non-nil issuer")
	}
	// Subkey[1] is revoked, so we better be using Subkey[2]
	if *issuer != signer.Subkeys[2].PublicKey.KeyId {
		t.Fatalf("Got wrong subkey: wanted %x, but got %x", signer.Subkeys[2].PublicKey.KeyId, *issuer)
	}
}

func TestSignWithRevokedSubkeyOfflineMaster(t *testing.T) {
	testSignWithRevokedSubkey(t, keyWithRevokedSubkeysOfflineMasterPrivate, keyWithRevokedSubkeysOfflineMasterPublic, keyWithRevokedSubkeyPassphrase)
}

func TestSignWithRevokedSubkey(t *testing.T) {
	testSignWithRevokedSubkey(t, keyWithRevokedSubkeysPrivate, keyWithRevokedSubkeysPublic, keyWithRevokedSubkeyPassphrase)
}

func TestMultipleSigSubkey(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(matthiasuKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.Subkeys) != 1 {
		t.Fatal("Expected one subkey")
	}
	time1, _ := time.Parse("2006-01-02", "2017-03-21")
	if entity.Subkeys[0].Sig.KeyExpired(time1) {
		t.Fatal("Expected subkey not to be expired")
	}
}

func TestMessageEncryptionRoundtripWithPassphraseChange(t *testing.T) {
	el, err := ReadKeyRing(readerFromHex(testKeys1And2PrivateHex))
	if len(el) != 2 {
		t.Fatal("failed to load the keyring")
	}
	// This is the encryption key used in previous test TestSignedEncryptedMessage()
	// If that pass, we know the message can be decrypted by this private key
	keys := el.KeysById(0x2a67d68660df41c7, nil)
	if len(keys) != 1 {
		t.Fatalf("%d keys found with ID 0x2a67d68660df41c7", len(keys))
	}

	// Re-encrypt the Entity
	entity := keys[0].Entity
	oldPasswd := []byte("passphrase")
	newPasswd := []byte("123456")
	if entity.PrivateKey.Encrypted {
		if err = entity.PrivateKey.Decrypt(oldPasswd); err != nil {
			t.Fatalf("failed to decrypt primary private key: %s", err)
		}
		if err = entity.PrivateKey.Encrypt(newPasswd, nil); err != nil {
			t.Fatalf("failed to encrypt primary private key: %s", err)
		}
	}

	for _, sk := range entity.Subkeys {
		if !sk.PrivateKey.Encrypted {
			continue
		}
		if err = sk.PrivateKey.Decrypt(oldPasswd); err != nil {
			t.Fatalf("failed to decrypt subkey 0x%x: %s", sk.PublicKey.KeyId, err)
		}
		if err = sk.PrivateKey.Encrypt(newPasswd, nil); err != nil {
			t.Fatalf("failed to encrypt subkey 0x%x: %s", sk.PublicKey.KeyId, err)
		}
	}

	// Re-serialize the re-encrypted Entity
	keyBuf := bytes.NewBuffer(nil)
	if err = entity.SerializePrivate(keyBuf, nil); err != nil {
		t.Fatalf("failed to serialize primary key: %s", err)
	}

	var kring EntityList
	if kring, err = ReadKeyRing(keyBuf); err != nil {
		t.Fatalf("failed to load new key ring: %s", err)
	}
	prompt := func(keys []Key, symmetric bool) ([]byte, error) {
		if symmetric {
			t.Error("prompt: message was marked as symmetrically encrypted")
			return nil, errors.ErrKeyIncorrect
		}

		if len(keys) == 0 {
			t.Error("prompt: no key presented")
			return nil, errors.ErrKeyIncorrect
		}

		err := keys[0].PrivateKey.Decrypt(newPasswd)
		if err != nil {
			t.Errorf("prompt: private key decryption failed: %s", err)
			return nil, errors.ErrKeyIncorrect
		}

		return nil, nil
	}

	encryptedMsgReader := readerFromHex(signedEncryptedMessageHex)
	var msgD *MessageDetails
	if msgD, err = ReadMessage(encryptedMsgReader, kring, prompt, nil); err != nil {
		t.Fatalf("failed to read encrypted message: %s", err)
	}

	var content []byte
	if content, err = ioutil.ReadAll(msgD.UnverifiedBody); err != nil {
		t.Fatalf("failed to read body: %s", err)
	}
	expected := []byte("Signed and encrypted message\n")
	if !bytes.Equal(expected, content) {
		t.Fatalf("message mismatch")
	}
}

func TestIllformedArmors(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(noNewLinesKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read noNewLinesKey: %v", err)
	}

	el, err = ReadArmoredKeyRing(bytes.NewBufferString(noNewlinesKey2))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read noNewlinesKey2: %v", err)
	}

	el, err = ReadArmoredKeyRing(bytes.NewBufferString(noCRCKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read noCRCKey: %v", err)
	}

	el, err = ReadArmoredKeyRing(bytes.NewBufferString(spacesInsteadOfNewlinesKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read spacesInsteadOfNewlinesKey: %v", err)
	}
}

func TestSignedTextMessageNoSignature(t *testing.T) {
	kring, _ := ReadKeyRing(readerFromHex(testKeys1And2Hex))

	md, err := ReadMessage(readerFromHex(signedTextMessageMalformed), kring, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !md.IsSigned || md.SignedByKeyId != 0xa34d7e18c20c31bb || md.SignedBy == nil || md.IsEncrypted || md.IsSymmetricallyEncrypted || len(md.EncryptedToKeyIds) != 0 || md.IsSymmetricallyEncrypted || md.MultiSig {
		t.Errorf("bad MessageDetails: %#v", md)
	}

	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Errorf("error reading UnverifiedBody: %s", err)
	}
	if string(contents) != signedTextInput {
		t.Errorf("bad UnverifiedBody got:%s want:%s", string(contents), signedTextInput)
	}
	if md.SignatureError == nil || md.Signature != nil {
		t.Fatalf("Expected SignatureError, got nil")
	}
	t.Logf("SignatureError is: %s", md.SignatureError)
}

const testKey1KeyId = 0xA34D7E18C20C31BB
const testKey3KeyId = 0x338934250CCC0360

const signedInput = "Signed message\nline 2\nline 3\n"
const signedTextInput = "Signed message\r\nline 2\r\nline 3\r\n"

const recipientUnspecifiedHex = "848c0300000000000000000103ff62d4d578d03cf40c3da998dfe216c074fa6ddec5e31c197c9666ba292830d91d18716a80f699f9d897389a90e6d62d0238f5f07a5248073c0f24920e4bc4a30c2d17ee4e0cae7c3d4aaa4e8dced50e3010a80ee692175fa0385f62ecca4b56ee6e9980aa3ec51b61b077096ac9e800edaf161268593eedb6cc7027ff5cb32745d250010d407a6221ae22ef18469b444f2822478c4d190b24d36371a95cb40087cdd42d9399c3d06a53c0673349bfb607927f20d1e122bde1e2bf3aa6cae6edf489629bcaa0689539ae3b718914d88ededc3b"

const detachedSignatureHex = "889c04000102000605024d449cd1000a0910a34d7e18c20c31bb167603ff57718d09f28a519fdc7b5a68b6a3336da04df85e38c5cd5d5bd2092fa4629848a33d85b1729402a2aab39c3ac19f9d573f773cc62c264dc924c067a79dfd8a863ae06c7c8686120760749f5fd9b1e03a64d20a7df3446ddc8f0aeadeaeba7cbaee5c1e366d65b6a0c6cc749bcb912d2f15013f812795c2e29eb7f7b77f39ce77"

const detachedSignatureTextHex = "889c04010102000605024d449d21000a0910a34d7e18c20c31bbc8c60400a24fbef7342603a41cb1165767bd18985d015fb72fe05db42db36cfb2f1d455967f1e491194fbf6cf88146222b23bf6ffbd50d17598d976a0417d3192ff9cc0034fd00f287b02e90418bbefe609484b09231e4e7a5f3562e199bf39909ab5276c4d37382fe088f6b5c3426fc1052865da8b3ab158672d58b6264b10823dc4b39"

const detachedSignatureV3TextHex = "8900950305005255c25ca34d7e18c20c31bb0102bb3f04009f6589ef8a028d6e54f6eaf25432e590d31c3a41f4710897585e10c31e5e332c7f9f409af8512adceaff24d0da1474ab07aa7bce4f674610b010fccc5b579ae5eb00a127f272fb799f988ab8e4574c141da6dbfecfef7e6b2c478d9a3d2551ba741f260ee22bec762812f0053e05380bfdd55ad0f22d8cdf71b233fe51ae8a24"

const detachedSignatureDSAHex = "884604001102000605024d6c4eac000a0910338934250ccc0360f18d00a087d743d6405ed7b87755476629600b8b694a39e900a0abff8126f46faf1547c1743c37b21b4ea15b8f83"

const testKeys1And2Hex = "988d044d3c5c10010400b1d13382944bd5aba23a4312968b5095d14f947f600eb478e14a6fcb16b0e0cac764884909c020bc495cfcc39a935387c661507bdb236a0612fb582cac3af9b29cc2c8c70090616c41b662f4da4c1201e195472eb7f4ae1ccbcbf9940fe21d985e379a5563dde5b9a23d35f1cfaa5790da3b79db26f23695107bfaca8e7b5bcd0011010001b41054657374204b6579203120285253412988b804130102002205024d3c5c10021b03060b090807030206150802090a0b0416020301021e01021780000a0910a34d7e18c20c31bbb5b304009cc45fe610b641a2c146331be94dade0a396e73ca725e1b25c21708d9cab46ecca5ccebc23055879df8f99eea39b377962a400f2ebdc36a7c99c333d74aeba346315137c3ff9d0a09b0273299090343048afb8107cf94cbd1400e3026f0ccac7ecebbc4d78588eb3e478fe2754d3ca664bcf3eac96ca4a6b0c8d7df5102f60f6b0020003b88d044d3c5c10010400b201df61d67487301f11879d514f4248ade90c8f68c7af1284c161098de4c28c2850f1ec7b8e30f959793e571542ffc6532189409cb51c3d30dad78c4ad5165eda18b20d9826d8707d0f742e2ab492103a85bbd9ddf4f5720f6de7064feb0d39ee002219765bb07bcfb8b877f47abe270ddeda4f676108cecb6b9bb2ad484a4f0011010001889f04180102000905024d3c5c10021b0c000a0910a34d7e18c20c31bb1a03040085c8d62e16d05dc4e9dad64953c8a2eed8b6c12f92b1575eeaa6dcf7be9473dd5b24b37b6dffbb4e7c99ed1bd3cb11634be19b3e6e207bed7505c7ca111ccf47cb323bf1f8851eb6360e8034cbff8dd149993c959de89f8f77f38e7e98b8e3076323aa719328e2b408db5ec0d03936efd57422ba04f925cdc7b4c1af7590e40ab0020003988d044d3c5c33010400b488c3e5f83f4d561f317817538d9d0397981e9aef1321ca68ebfae1cf8b7d388e19f4b5a24a82e2fbbf1c6c26557a6c5845307a03d815756f564ac7325b02bc83e87d5480a8fae848f07cb891f2d51ce7df83dcafdc12324517c86d472cc0ee10d47a68fd1d9ae49a6c19bbd36d82af597a0d88cc9c49de9df4e696fc1f0b5d0011010001b42754657374204b6579203220285253412c20656e637279707465642070726976617465206b65792988b804130102002205024d3c5c33021b03060b090807030206150802090a0b0416020301021e01021780000a0910d4984f961e35246b98940400908a73b6a6169f700434f076c6c79015a49bee37130eaf23aaa3cfa9ce60bfe4acaa7bc95f1146ada5867e0079babb38804891f4f0b8ebca57a86b249dee786161a755b7a342e68ccf3f78ed6440a93a6626beb9a37aa66afcd4f888790cb4bb46d94a4ae3eb3d7d3e6b00f6bfec940303e89ec5b32a1eaaacce66497d539328b0020003b88d044d3c5c33010400a4e913f9442abcc7f1804ccab27d2f787ffa592077ca935a8bb23165bd8d57576acac647cc596b2c3f814518cc8c82953c7a4478f32e0cf645630a5ba38d9618ef2bc3add69d459ae3dece5cab778938d988239f8c5ae437807075e06c828019959c644ff05ef6a5a1dab72227c98e3a040b0cf219026640698d7a13d8538a570011010001889f04180102000905024d3c5c33021b0c000a0910d4984f961e35246b26c703ff7ee29ef53bc1ae1ead533c408fa136db508434e233d6e62be621e031e5940bbd4c08142aed0f82217e7c3e1ec8de574bc06ccf3c36633be41ad78a9eacd209f861cae7b064100758545cc9dd83db71806dc1cfd5fb9ae5c7474bba0c19c44034ae61bae5eca379383339dece94ff56ff7aa44a582f3e5c38f45763af577c0934b0020003"

const testKeys1And2PrivateHex = "9501d8044d3c5c10010400b1d13382944bd5aba23a4312968b5095d14f947f600eb478e14a6fcb16b0e0cac764884909c020bc495cfcc39a935387c661507bdb236a0612fb582cac3af9b29cc2c8c70090616c41b662f4da4c1201e195472eb7f4ae1ccbcbf9940fe21d985e379a5563dde5b9a23d35f1cfaa5790da3b79db26f23695107bfaca8e7b5bcd00110100010003ff4d91393b9a8e3430b14d6209df42f98dc927425b881f1209f319220841273a802a97c7bdb8b3a7740b3ab5866c4d1d308ad0d3a79bd1e883aacf1ac92dfe720285d10d08752a7efe3c609b1d00f17f2805b217be53999a7da7e493bfc3e9618fd17018991b8128aea70a05dbce30e4fbe626aa45775fa255dd9177aabf4df7cf0200c1ded12566e4bc2bb590455e5becfb2e2c9796482270a943343a7835de41080582c2be3caf5981aa838140e97afa40ad652a0b544f83eb1833b0957dce26e47b0200eacd6046741e9ce2ec5beb6fb5e6335457844fb09477f83b050a96be7da043e17f3a9523567ed40e7a521f818813a8b8a72209f1442844843ccc7eb9805442570200bdafe0438d97ac36e773c7162028d65844c4d463e2420aa2228c6e50dc2743c3d6c72d0d782a5173fe7be2169c8a9f4ef8a7cf3e37165e8c61b89c346cdc6c1799d2b41054657374204b6579203120285253412988b804130102002205024d3c5c10021b03060b090807030206150802090a0b0416020301021e01021780000a0910a34d7e18c20c31bbb5b304009cc45fe610b641a2c146331be94dade0a396e73ca725e1b25c21708d9cab46ecca5ccebc23055879df8f99eea39b377962a400f2ebdc36a7c99c333d74aeba346315137c3ff9d0a09b0273299090343048afb8107cf94cbd1400e3026f0ccac7ecebbc4d78588eb3e478fe2754d3ca664bcf3eac96ca4a6b0c8d7df5102f60f6b00200009d01d8044d3c5c10010400b201df61d67487301f11879d514f4248ade90c8f68c7af1284c161098de4c28c2850f1ec7b8e30f959793e571542ffc6532189409cb51c3d30dad78c4ad5165eda18b20d9826d8707d0f742e2ab492103a85bbd9ddf4f5720f6de7064feb0d39ee002219765bb07bcfb8b877f47abe270ddeda4f676108cecb6b9bb2ad484a4f00110100010003fd17a7490c22a79c59281fb7b20f5e6553ec0c1637ae382e8adaea295f50241037f8997cf42c1ce26417e015091451b15424b2c59eb8d4161b0975630408e394d3b00f88d4b4e18e2cc85e8251d4753a27c639c83f5ad4a571c4f19d7cd460b9b73c25ade730c99df09637bd173d8e3e981ac64432078263bb6dc30d3e974150dd0200d0ee05be3d4604d2146fb0457f31ba17c057560785aa804e8ca5530a7cd81d3440d0f4ba6851efcfd3954b7e68908fc0ba47f7ac37bf559c6c168b70d3a7c8cd0200da1c677c4bce06a068070f2b3733b0a714e88d62aa3f9a26c6f5216d48d5c2b5624144f3807c0df30be66b3268eeeca4df1fbded58faf49fc95dc3c35f134f8b01fd1396b6c0fc1b6c4f0eb8f5e44b8eace1e6073e20d0b8bc5385f86f1cf3f050f66af789f3ef1fc107b7f4421e19e0349c730c68f0a226981f4e889054fdb4dc149e8e889f04180102000905024d3c5c10021b0c000a0910a34d7e18c20c31bb1a03040085c8d62e16d05dc4e9dad64953c8a2eed8b6c12f92b1575eeaa6dcf7be9473dd5b24b37b6dffbb4e7c99ed1bd3cb11634be19b3e6e207bed7505c7ca111ccf47cb323bf1f8851eb6360e8034cbff8dd149993c959de89f8f77f38e7e98b8e3076323aa719328e2b408db5ec0d03936efd57422ba04f925cdc7b4c1af7590e40ab00200009501fe044d3c5c33010400b488c3e5f83f4d561f317817538d9d0397981e9aef1321ca68ebfae1cf8b7d388e19f4b5a24a82e2fbbf1c6c26557a6c5845307a03d815756f564ac7325b02bc83e87d5480a8fae848f07cb891f2d51ce7df83dcafdc12324517c86d472cc0ee10d47a68fd1d9ae49a6c19bbd36d82af597a0d88cc9c49de9df4e696fc1f0b5d0011010001fe030302e9030f3c783e14856063f16938530e148bc57a7aa3f3e4f90df9dceccdc779bc0835e1ad3d006e4a8d7b36d08b8e0de5a0d947254ecfbd22037e6572b426bcfdc517796b224b0036ff90bc574b5509bede85512f2eefb520fb4b02aa523ba739bff424a6fe81c5041f253f8d757e69a503d3563a104d0d49e9e890b9d0c26f96b55b743883b472caa7050c4acfd4a21f875bdf1258d88bd61224d303dc9df77f743137d51e6d5246b88c406780528fd9a3e15bab5452e5b93970d9dcc79f48b38651b9f15bfbcf6da452837e9cc70683d1bdca94507870f743e4ad902005812488dd342f836e72869afd00ce1850eea4cfa53ce10e3608e13d3c149394ee3cbd0e23d018fcbcb6e2ec5a1a22972d1d462ca05355d0d290dd2751e550d5efb38c6c89686344df64852bf4ff86638708f644e8ec6bd4af9b50d8541cb91891a431326ab2e332faa7ae86cfb6e0540aa63160c1e5cdd5a4add518b303fff0a20117c6bc77f7cfbaf36b04c865c6c2b42754657374204b6579203220285253412c20656e637279707465642070726976617465206b65792988b804130102002205024d3c5c33021b03060b090807030206150802090a0b0416020301021e01021780000a0910d4984f961e35246b98940400908a73b6a6169f700434f076c6c79015a49bee37130eaf23aaa3cfa9ce60bfe4acaa7bc95f1146ada5867e0079babb38804891f4f0b8ebca57a86b249dee786161a755b7a342e68ccf3f78ed6440a93a6626beb9a37aa66afcd4f888790cb4bb46d94a4ae3eb3d7d3e6b00f6bfec940303e89ec5b32a1eaaacce66497d539328b00200009d01fe044d3c5c33010400a4e913f9442abcc7f1804ccab27d2f787ffa592077ca935a8bb23165bd8d57576acac647cc596b2c3f814518cc8c82953c7a4478f32e0cf645630a5ba38d9618ef2bc3add69d459ae3dece5cab778938d988239f8c5ae437807075e06c828019959c644ff05ef6a5a1dab72227c98e3a040b0cf219026640698d7a13d8538a570011010001fe030302e9030f3c783e148560f936097339ae381d63116efcf802ff8b1c9360767db5219cc987375702a4123fd8657d3e22700f23f95020d1b261eda5257e9a72f9a918e8ef22dd5b3323ae03bbc1923dd224db988cadc16acc04b120a9f8b7e84da9716c53e0334d7b66586ddb9014df604b41be1e960dcfcbc96f4ed150a1a0dd070b9eb14276b9b6be413a769a75b519a53d3ecc0c220e85cd91ca354d57e7344517e64b43b6e29823cbd87eae26e2b2e78e6dedfbb76e3e9f77bcb844f9a8932eb3db2c3f9e44316e6f5d60e9e2a56e46b72abe6b06dc9a31cc63f10023d1f5e12d2a3ee93b675c96f504af0001220991c88db759e231b3320dcedf814dcf723fd9857e3d72d66a0f2af26950b915abdf56c1596f46a325bf17ad4810d3535fb02a259b247ac3dbd4cc3ecf9c51b6c07cebb009c1506fba0a89321ec8683e3fd009a6e551d50243e2d5092fefb3321083a4bad91320dc624bd6b5dddf93553e3d53924c05bfebec1fb4bd47e89a1a889f04180102000905024d3c5c33021b0c000a0910d4984f961e35246b26c703ff7ee29ef53bc1ae1ead533c408fa136db508434e233d6e62be621e031e5940bbd4c08142aed0f82217e7c3e1ec8de574bc06ccf3c36633be41ad78a9eacd209f861cae7b064100758545cc9dd83db71806dc1cfd5fb9ae5c7474bba0c19c44034ae61bae5eca379383339dece94ff56ff7aa44a582f3e5c38f45763af577c0934b0020000"

const dsaElGamalTestKeysHex = "9501e1044dfcb16a110400aa3e5c1a1f43dd28c2ffae8abf5cfce555ee874134d8ba0a0f7b868ce2214beddc74e5e1e21ded354a95d18acdaf69e5e342371a71fbb9093162e0c5f3427de413a7f2c157d83f5cd2f9d791256dc4f6f0e13f13c3302af27f2384075ab3021dff7a050e14854bbde0a1094174855fc02f0bae8e00a340d94a1f22b32e48485700a0cec672ac21258fb95f61de2ce1af74b2c4fa3e6703ff698edc9be22c02ae4d916e4fa223f819d46582c0516235848a77b577ea49018dcd5e9e15cff9dbb4663a1ae6dd7580fa40946d40c05f72814b0f88481207e6c0832c3bded4853ebba0a7e3bd8e8c66df33d5a537cd4acf946d1080e7a3dcea679cb2b11a72a33a2b6a9dc85f466ad2ddf4c3db6283fa645343286971e3dd700703fc0c4e290d45767f370831a90187e74e9972aae5bff488eeff7d620af0362bfb95c1a6c3413ab5d15a2e4139e5d07a54d72583914661ed6a87cce810be28a0aa8879a2dd39e52fb6fe800f4f181ac7e328f740cde3d09a05cecf9483e4cca4253e60d4429ffd679d9996a520012aad119878c941e3cf151459873bdfc2a9563472fe0303027a728f9feb3b864260a1babe83925ce794710cfd642ee4ae0e5b9d74cee49e9c67b6cd0ea5dfbb582132195a121356a1513e1bca73e5b80c58c7ccb4164453412f456c47616d616c2054657374204b65792031886204131102002205024dfcb16a021b03060b090807030206150802090a0b0416020301021e01021780000a091033af447ccd759b09fadd00a0b8fd6f5a790bad7e9f2dbb7632046dc4493588db009c087c6a9ba9f7f49fab221587a74788c00db4889ab00200009d0157044dfcb16a1004008dec3f9291205255ccff8c532318133a6840739dd68b03ba942676f9038612071447bf07d00d559c5c0875724ea16a4c774f80d8338b55fca691a0522e530e604215b467bbc9ccfd483a1da99d7bc2648b4318fdbd27766fc8bfad3fddb37c62b8ae7ccfe9577e9b8d1e77c1d417ed2c2ef02d52f4da11600d85d3229607943700030503ff506c94c87c8cab778e963b76cf63770f0a79bf48fb49d3b4e52234620fc9f7657f9f8d56c96a2b7c7826ae6b57ebb2221a3fe154b03b6637cea7e6d98e3e45d87cf8dc432f723d3d71f89c5192ac8d7290684d2c25ce55846a80c9a7823f6acd9bb29fa6cd71f20bc90eccfca20451d0c976e460e672b000df49466408d527affe0303027a728f9feb3b864260abd761730327bca2aaa4ea0525c175e92bf240682a0e83b226f97ecb2e935b62c9a133858ce31b271fa8eb41f6a1b3cd72a63025ce1a75ee4180dcc284884904181102000905024dfcb16a021b0c000a091033af447ccd759b09dd0b009e3c3e7296092c81bee5a19929462caaf2fff3ae26009e218c437a2340e7ea628149af1ec98ec091a43992b00200009501e1044dfcb1be1104009f61faa61aa43df75d128cbe53de528c4aec49ce9360c992e70c77072ad5623de0a3a6212771b66b39a30dad6781799e92608316900518ec01184a85d872365b7d2ba4bacfb5882ea3c2473d3750dc6178cc1cf82147fb58caa28b28e9f12f6d1efcb0534abed644156c91cca4ab78834268495160b2400bc422beb37d237c2300a0cac94911b6d493bda1e1fbc6feeca7cb7421d34b03fe22cec6ccb39675bb7b94a335c2b7be888fd3906a1125f33301d8aa6ec6ee6878f46f73961c8d57a3e9544d8ef2a2cbfd4d52da665b1266928cfe4cb347a58c412815f3b2d2369dec04b41ac9a71cc9547426d5ab941cccf3b18575637ccfb42df1a802df3cfe0a999f9e7109331170e3a221991bf868543960f8c816c28097e503fe319db10fb98049f3a57d7c80c420da66d56f3644371631fad3f0ff4040a19a4fedc2d07727a1b27576f75a4d28c47d8246f27071e12d7a8de62aad216ddbae6aa02efd6b8a3e2818cda48526549791ab277e447b3a36c57cefe9b592f5eab73959743fcc8e83cbefec03a329b55018b53eec196765ae40ef9e20521a603c551efe0303020950d53a146bf9c66034d00c23130cce95576a2ff78016ca471276e8227fb30b1ffbd92e61804fb0c3eff9e30b1a826ee8f3e4730b4d86273ca977b4164453412f456c47616d616c2054657374204b65792032886204131102002205024dfcb1be021b03060b090807030206150802090a0b0416020301021e01021780000a0910a86bf526325b21b22bd9009e34511620415c974750a20df5cb56b182f3b48e6600a0a9466cb1a1305a84953445f77d461593f1d42bc1b00200009d0157044dfcb1be1004009565a951da1ee87119d600c077198f1c1bceb0f7aa54552489298e41ff788fa8f0d43a69871f0f6f77ebdfb14a4260cf9fbeb65d5844b4272a1904dd95136d06c3da745dc46327dd44a0f16f60135914368c8039a34033862261806bb2c5ce1152e2840254697872c85441ccb7321431d75a747a4bfb1d2c66362b51ce76311700030503fc0ea76601c196768070b7365a200e6ddb09307f262d5f39eec467b5f5784e22abdf1aa49226f59ab37cb49969d8f5230ea65caf56015abda62604544ed526c5c522bf92bed178a078789f6c807b6d34885688024a5bed9e9f8c58d11d4b82487b44c5f470c5606806a0443b79cadb45e0f897a561a53f724e5349b9267c75ca17fe0303020950d53a146bf9c660bc5f4ce8f072465e2d2466434320c1e712272fafc20e342fe7608101580fa1a1a367e60486a7cd1246b7ef5586cf5e10b32762b710a30144f12dd17dd4884904181102000905024dfcb1be021b0c000a0910a86bf526325b21b2904c00a0b2b66b4b39ccffda1d10f3ea8d58f827e30a8b8e009f4255b2d8112a184e40cde43a34e8655ca7809370b0020000"

const signedMessageHex = "a3019bc0cbccc0c4b8d8b74ee2108fe16ec6d3ca490cbe362d3f8333d3f352531472538b8b13d353b97232f352158c20943157c71c16064626063656269052062e4e01987e9b6fccff4b7df3a34c534b23e679cbec3bc0f8f6e64dfb4b55fe3f8efa9ce110ddb5cd79faf1d753c51aecfa669f7e7aa043436596cccc3359cb7dd6bbe9ecaa69e5989d9e57209571edc0b2fa7f57b9b79a64ee6e99ce1371395fee92fec2796f7b15a77c386ff668ee27f6d38f0baa6c438b561657377bf6acff3c5947befd7bf4c196252f1d6e5c524d0300"

const signedTextMessageHex = "a3019bc0cbccc8c4b8d8b74ee2108fe16ec6d36a250cbece0c178233d3f352531472538b8b13d35379b97232f352158ca0b4312f57c71c1646462606365626906a062e4e019811591798ff99bf8afee860b0d8a8c2a85c3387e3bcf0bb3b17987f2bbcfab2aa526d930cbfd3d98757184df3995c9f3e7790e36e3e9779f06089d4c64e9e47dd6202cb6e9bc73c5d11bb59fbaf89d22d8dc7cf199ddf17af96e77c5f65f9bbed56f427bd8db7af37f6c9984bf9385efaf5f184f986fb3e6adb0ecfe35bbf92d16a7aa2a344fb0bc52fb7624f0200"

// Same message as signedTextMessageHex but "signature packet" is
// dropped from the end. So this message claims to be signed but it
// cannot be verified, so verification should fail.
const signedTextMessageMalformed = "900d03010201a34d7e18c20c31bb01cb2674004d4300d05369676e6564206d6573736167650d0a6c696e6520320d0a6c696e6520330d0a"

const signedEncryptedMessageHex = "848c032a67d68660df41c70103ff5789d0de26b6a50c985a02a13131ca829c413a35d0e6fa8d6842599252162808ac7439c72151c8c6183e76923fe3299301414d0c25a2f06a2257db3839e7df0ec964773f6e4c4ac7ff3b48c444237166dd46ba8ff443a5410dc670cb486672fdbe7c9dfafb75b4fea83af3a204fe2a7dfa86bd20122b4f3d2646cbeecb8f7be8d2c03b018bd210b1d3791e1aba74b0f1034e122ab72e760492c192383cf5e20b5628bd043272d63df9b923f147eb6091cd897553204832aba48fec54aa447547bb16305a1024713b90e77fd0065f1918271947549205af3c74891af22ee0b56cd29bfec6d6e351901cd4ab3ece7c486f1e32a792d4e474aed98ee84b3f591c7dff37b64e0ecd68fd036d517e412dcadf85840ce184ad7921ad446c4ee28db80447aea1ca8d4f574db4d4e37688158ddd19e14ee2eab4873d46947d65d14a23e788d912cf9a19624ca7352469b72a83866b7c23cb5ace3deab3c7018061b0ba0f39ed2befe27163e5083cf9b8271e3e3d52cc7ad6e2a3bd81d4c3d7022f8d"

const signedEncryptedMessage2Hex = "85010e03cf6a7abcd43e36731003fb057f5495b79db367e277cdbe4ab90d924ddee0c0381494112ff8c1238fb0184af35d1731573b01bc4c55ecacd2aafbe2003d36310487d1ecc9ac994f3fada7f9f7f5c3a64248ab7782906c82c6ff1303b69a84d9a9529c31ecafbcdb9ba87e05439897d87e8a2a3dec55e14df19bba7f7bd316291c002ae2efd24f83f9e3441203fc081c0c23dc3092a454ca8a082b27f631abf73aca341686982e8fbda7e0e7d863941d68f3de4a755c2964407f4b5e0477b3196b8c93d551dd23c8beef7d0f03fbb1b6066f78907faf4bf1677d8fcec72651124080e0b7feae6b476e72ab207d38d90b958759fdedfc3c6c35717c9dbfc979b3cfbbff0a76d24a5e57056bb88acbd2a901ef64bc6e4db02adc05b6250ff378de81dca18c1910ab257dff1b9771b85bb9bbe0a69f5989e6d1710a35e6dfcceb7d8fb5ccea8db3932b3d9ff3fe0d327597c68b3622aec8e3716c83a6c93f497543b459b58ba504ed6bcaa747d37d2ca746fe49ae0a6ce4a8b694234e941b5159ff8bd34b9023da2814076163b86f40eed7c9472f81b551452d5ab87004a373c0172ec87ea6ce42ccfa7dbdad66b745496c4873d8019e8c28d6b3"

const symmetricallyEncryptedCompressedHex = "8c0d04030302eb4a03808145d0d260c92f714339e13de5a79881216431925bf67ee2898ea61815f07894cd0703c50d0a76ef64d482196f47a8bc729af9b80bb6"

const dsaTestKeyHex = "9901a2044d6c49de110400cb5ce438cf9250907ac2ba5bf6547931270b89f7c4b53d9d09f4d0213a5ef2ec1f26806d3d259960f872a4a102ef1581ea3f6d6882d15134f21ef6a84de933cc34c47cc9106efe3bd84c6aec12e78523661e29bc1a61f0aab17fa58a627fd5fd33f5149153fbe8cd70edf3d963bc287ef875270ff14b5bfdd1bca4483793923b00a0fe46d76cb6e4cbdc568435cd5480af3266d610d303fe33ae8273f30a96d4d34f42fa28ce1112d425b2e3bf7ea553d526e2db6b9255e9dc7419045ce817214d1a0056dbc8d5289956a4b1b69f20f1105124096e6a438f41f2e2495923b0f34b70642607d45559595c7fe94d7fa85fc41bf7d68c1fd509ebeaa5f315f6059a446b9369c277597e4f474a9591535354c7e7f4fd98a08aa60400b130c24ff20bdfbf683313f5daebf1c9b34b3bdadfc77f2ddd72ee1fb17e56c473664bc21d66467655dd74b9005e3a2bacce446f1920cd7017231ae447b67036c9b431b8179deacd5120262d894c26bc015bffe3d827ba7087ad9b700d2ca1f6d16cc1786581e5dd065f293c31209300f9b0afcc3f7c08dd26d0a22d87580b4db41054657374204b65792033202844534129886204131102002205024d6c49de021b03060b090807030206150802090a0b0416020301021e01021780000a0910338934250ccc03607e0400a0bdb9193e8a6b96fc2dfc108ae848914b504481f100a09c4dc148cb693293a67af24dd40d2b13a9e36794"

const dsaTestKeyPrivateHex = "9501bb044d6c49de110400cb5ce438cf9250907ac2ba5bf6547931270b89f7c4b53d9d09f4d0213a5ef2ec1f26806d3d259960f872a4a102ef1581ea3f6d6882d15134f21ef6a84de933cc34c47cc9106efe3bd84c6aec12e78523661e29bc1a61f0aab17fa58a627fd5fd33f5149153fbe8cd70edf3d963bc287ef875270ff14b5bfdd1bca4483793923b00a0fe46d76cb6e4cbdc568435cd5480af3266d610d303fe33ae8273f30a96d4d34f42fa28ce1112d425b2e3bf7ea553d526e2db6b9255e9dc7419045ce817214d1a0056dbc8d5289956a4b1b69f20f1105124096e6a438f41f2e2495923b0f34b70642607d45559595c7fe94d7fa85fc41bf7d68c1fd509ebeaa5f315f6059a446b9369c277597e4f474a9591535354c7e7f4fd98a08aa60400b130c24ff20bdfbf683313f5daebf1c9b34b3bdadfc77f2ddd72ee1fb17e56c473664bc21d66467655dd74b9005e3a2bacce446f1920cd7017231ae447b67036c9b431b8179deacd5120262d894c26bc015bffe3d827ba7087ad9b700d2ca1f6d16cc1786581e5dd065f293c31209300f9b0afcc3f7c08dd26d0a22d87580b4d00009f592e0619d823953577d4503061706843317e4fee083db41054657374204b65792033202844534129886204131102002205024d6c49de021b03060b090807030206150802090a0b0416020301021e01021780000a0910338934250ccc03607e0400a0bdb9193e8a6b96fc2dfc108ae848914b504481f100a09c4dc148cb693293a67af24dd40d2b13a9e36794"

const armoredPrivateKeyBlock = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1.4.10 (GNU/Linux)

lQHYBE2rFNoBBADFwqWQIW/DSqcB4yCQqnAFTJ27qS5AnB46ccAdw3u4Greeu3Bp
idpoHdjULy7zSKlwR1EA873dO/k/e11Ml3dlAFUinWeejWaK2ugFP6JjiieSsrKn
vWNicdCS4HTWn0X4sjl0ZiAygw6GNhqEQ3cpLeL0g8E9hnYzJKQ0LWJa0QARAQAB
AAP/TB81EIo2VYNmTq0pK1ZXwUpxCrvAAIG3hwKjEzHcbQznsjNvPUihZ+NZQ6+X
0HCfPAdPkGDCLCb6NavcSW+iNnLTrdDnSI6+3BbIONqWWdRDYJhqZCkqmG6zqSfL
IdkJgCw94taUg5BWP/AAeQrhzjChvpMQTVKQL5mnuZbUCeMCAN5qrYMP2S9iKdnk
VANIFj7656ARKt/nf4CBzxcpHTyB8+d2CtPDKCmlJP6vL8t58Jmih+kHJMvC0dzn
gr5f5+sCAOOe5gt9e0am7AvQWhdbHVfJU0TQJx+m2OiCJAqGTB1nvtBLHdJnfdC9
TnXXQ6ZXibqLyBies/xeY2sCKL5qtTMCAKnX9+9d/5yQxRyrQUHt1NYhaXZnJbHx
q4ytu0eWz+5i68IYUSK69jJ1NWPM0T6SkqpB3KCAIv68VFm9PxqG1KmhSrQIVGVz
dCBLZXmIuAQTAQIAIgUCTasU2gIbAwYLCQgHAwIGFQgCCQoLBBYCAwECHgECF4AA
CgkQO9o98PRieSoLhgQAkLEZex02Qt7vGhZzMwuN0R22w3VwyYyjBx+fM3JFETy1
ut4xcLJoJfIaF5ZS38UplgakHG0FQ+b49i8dMij0aZmDqGxrew1m4kBfjXw9B/v+
eIqpODryb6cOSwyQFH0lQkXC040pjq9YqDsO5w0WYNXYKDnzRV0p4H1pweo2VDid
AdgETasU2gEEAN46UPeWRqKHvA99arOxee38fBt2CI08iiWyI8T3J6ivtFGixSqV
bRcPxYO/qLpVe5l84Nb3X71GfVXlc9hyv7CD6tcowL59hg1E/DC5ydI8K8iEpUmK
/UnHdIY5h8/kqgGxkY/T/hgp5fRQgW1ZoZxLajVlMRZ8W4tFtT0DeA+JABEBAAEA
A/0bE1jaaZKj6ndqcw86jd+QtD1SF+Cf21CWRNeLKnUds4FRRvclzTyUMuWPkUeX
TaNNsUOFqBsf6QQ2oHUBBK4VCHffHCW4ZEX2cd6umz7mpHW6XzN4DECEzOVksXtc
lUC1j4UB91DC/RNQqwX1IV2QLSwssVotPMPqhOi0ZLNY7wIA3n7DWKInxYZZ4K+6
rQ+POsz6brEoRHwr8x6XlHenq1Oki855pSa1yXIARoTrSJkBtn5oI+f8AzrnN0BN
oyeQAwIA/7E++3HDi5aweWrViiul9cd3rcsS0dEnksPhvS0ozCJiHsq/6GFmy7J8
QSHZPteedBnZyNp5jR+H7cIfVN3KgwH/Skq4PsuPhDq5TKK6i8Pc1WW8MA6DXTdU
nLkX7RGmMwjC0DBf7KWAlPjFaONAX3a8ndnz//fy1q7u2l9AZwrj1qa1iJ8EGAEC
AAkFAk2rFNoCGwwACgkQO9o98PRieSo2/QP/WTzr4ioINVsvN1akKuekmEMI3LAp
BfHwatufxxP1U+3Si/6YIk7kuPB9Hs+pRqCXzbvPRrI8NHZBmc8qIGthishdCYad
AHcVnXjtxrULkQFGbGvhKURLvS9WnzD/m1K2zzwxzkPTzT9/Yf06O6Mal5AdugPL
VrM0m72/jnpKo04=
=zNCn
-----END PGP PRIVATE KEY BLOCK-----`

const e2ePublicKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Charset: UTF-8

xv8AAABSBAAAAAATCCqGSM49AwEHAgME1LRoXSpOxtHXDUdmuvzchyg6005qIBJ4
sfaSxX7QgH9RV2ONUhC+WiayCNADq+UMzuR/vunSr4aQffXvuGnR383/AAAAFDxk
Z2lsQHlhaG9vLWluYy5jb20+wv8AAACGBBATCAA4/wAAAAWCVGvAG/8AAAACiwn/
AAAACZC2VkQCOjdvYf8AAAAFlQgJCgv/AAAAA5YBAv8AAAACngEAAE1BAP0X8veD
24IjmI5/C6ZAfVNXxgZZFhTAACFX75jUA3oD6AEAzoSwKf1aqH6oq62qhCN/pekX
+WAsVMBhNwzLpqtCRjLO/wAAAFYEAAAAABIIKoZIzj0DAQcCAwT50ain7vXiIRv8
B1DO3x3cE/aattZ5sHNixJzRCXi2vQIA5QmOxZ6b5jjUekNbdHG3SZi1a2Ak5mfX
fRxC/5VGAwEIB8L/AAAAZQQYEwgAGP8AAAAFglRrwBz/AAAACZC2VkQCOjdvYQAA
FJAA9isX3xtGyMLYwp2F3nXm7QEdY5bq5VUcD/RJlj792VwA/1wH0pCzVLl4Q9F9
ex7En5r7rHR5xwX82Msc+Rq9dSyO
=7MrZ
-----END PGP PUBLIC KEY BLOCK-----`

const dsaKeyWithSHA512 = `9901a2044f04b07f110400db244efecc7316553ee08d179972aab87bb1214de7692593fcf5b6feb1c80fba268722dd464748539b85b81d574cd2d7ad0ca2444de4d849b8756bad7768c486c83a824f9bba4af773d11742bdfb4ac3b89ef8cc9452d4aad31a37e4b630d33927bff68e879284a1672659b8b298222fc68f370f3e24dccacc4a862442b9438b00a0ea444a24088dc23e26df7daf8f43cba3bffc4fe703fe3d6cd7fdca199d54ed8ae501c30e3ec7871ea9cdd4cf63cfe6fc82281d70a5b8bb493f922cd99fba5f088935596af087c8d818d5ec4d0b9afa7f070b3d7c1dd32a84fca08d8280b4890c8da1dde334de8e3cad8450eed2a4a4fcc2db7b8e5528b869a74a7f0189e11ef097ef1253582348de072bb07a9fa8ab838e993cef0ee203ff49298723e2d1f549b00559f886cd417a41692ce58d0ac1307dc71d85a8af21b0cf6eaa14baf2922d3a70389bedf17cc514ba0febbd107675a372fe84b90162a9e88b14d4b1c6be855b96b33fb198c46f058568817780435b6936167ebb3724b680f32bf27382ada2e37a879b3d9de2abe0c3f399350afd1ad438883f4791e2e3b4184453412068617368207472756e636174696f6e207465737488620413110a002205024f04b07f021b03060b090807030206150802090a0b0416020301021e01021780000a0910ef20e0cefca131581318009e2bf3bf047a44d75a9bacd00161ee04d435522397009a03a60d51bd8a568c6c021c8d7cf1be8d990d6417b0020003`

const unknownHashFunctionHex = `8a00000040040001990006050253863c24000a09103b4fe6acc0b21f32ffff01010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101`

const missingHashFunctionHex = `8a00000040040001030006050253863c24000a09103b4fe6acc0b21f32ffff0101010101010101010101010101010101010101010101010101010101010101010101010101`

const campbellQuine = `a0b001000300fcffa0b001000d00f2ff000300fcffa0b001000d00f2ff8270a01c00000500faff8270a01c00000500faff000500faff001400ebff8270a01c00000500faff000500faff001400ebff428821c400001400ebff428821c400001400ebff428821c400001400ebff428821c400001400ebff428821c400000000ffff000000ffff000b00f4ff428821c400000000ffff000000ffff000b00f4ff0233214c40000100feff000233214c40000100feff0000`

const keyV4forVerifyingSignedMessageV3 = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

mI0EVfxoFQEEAMBIqmbDfYygcvP6Phr1wr1XI41IF7Qixqybs/foBF8qqblD9gIY
BKpXjnBOtbkcVOJ0nljd3/sQIfH4E0vQwK5/4YRQSI59eKOqd6Fx+fWQOLG+uu6z
tewpeCj9LLHvibx/Sc7VWRnrznia6ftrXxJ/wHMezSab3tnGC0YPVdGNABEBAAG0
JEdvY3J5cHRvIFRlc3QgS2V5IDx0aGVtYXhAZ21haWwuY29tPoi5BBMBCgAjBQJV
/GgVAhsDBwsJCAcDAgEGFQgCCQoLBBYCAwECHgECF4AACgkQeXnQmhdGW9PFVAP+
K7TU0qX5ArvIONIxh/WAweyOk884c5cE8f+3NOPOOCRGyVy0FId5A7MmD5GOQh4H
JseOZVEVCqlmngEvtHZb3U1VYtVGE5WZ+6rQhGsMcWP5qaT4soYwMBlSYxgYwQcx
YhN9qOr292f9j2Y//TTIJmZT4Oa+lMxhWdqTfX+qMgG4jQRV/GgVAQQArhFSiij1
b+hT3dnapbEU+23Z1yTu1DfF6zsxQ4XQWEV3eR8v+8mEDDNcz8oyyF56k6UQ3rXi
UMTIwRDg4V6SbZmaFbZYCOwp/EmXJ3rfhm7z7yzXj2OFN22luuqbyVhuL7LRdB0M
pxgmjXb4tTvfgKd26x34S+QqUJ7W6uprY4sAEQEAAYifBBgBCgAJBQJV/GgVAhsM
AAoJEHl50JoXRlvT7y8D/02ckx4OMkKBZo7viyrBw0MLG92i+DC2bs35PooHR6zz
786mitjOp5z2QWNLBvxC70S0qVfCIz8jKupO1J6rq6Z8CcbLF3qjm6h1omUBf8Nd
EfXKD2/2HV6zMKVknnKzIEzauh+eCKS2CeJUSSSryap/QLVAjRnckaES/OsEWhNB
=RZia
-----END PGP PUBLIC KEY BLOCK-----
`

const signedMessageV3 = `-----BEGIN PGP MESSAGE-----
Comment: GPGTools - https://gpgtools.org

owGbwMvMwMVYWXlhlrhb9GXG03JJDKF/MtxDMjKLFYAoUaEktbhEITe1uDgxPVWP
q5NhKjMrWAVcC9evD8z/bF/uWNjqtk/X3y5/38XGRQHm/57rrDRYuGnTw597Xqka
uM3137/hH3Os+Jf2dc0fXOITKwJvXJvecPVs0ta+Vg7ZO1MLn8w58Xx+6L58mbka
DGHyU9yTueZE8D+QF/Tz28Y78dqtF56R1VPn9Xw4uJqrWYdd7b3vIZ1V6R4Nh05d
iT57d/OhWwA=
=hG7R
-----END PGP MESSAGE-----
`

const gnuDummyS2KPrivateKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG/MacGPG2 v2.0.22 (Darwin)
Comment: GPGTools - https://gpgtools.org

lQCVBFNVKE4BBADjD9Xq+1wml4VS3hxkCuyhWp003ki7yN/ZAb5cUHyIzgY7BR9v
ydz7R2s5dkRksxqiD8qg/u/UwMGteREhA8ML8JXSZ5T/TMH8DJNB1HsoKlm2q/W4
/S04jy5X/+M9GvRi47gZyOmLsu57rXdJimrUf9r9qtKSPViWlzrq4cAE0wARAQAB
/gNlAkdOVQG0IFdpbGxpYW0gV29yZHN3b3J0aCA8d3dAb3guYWMudWs+iL4EEwEK
ACgFAlNVKE4CGwMFCRLMAwAGCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEJLY
KARjvfT1roEEAJ140DFf7DV0d51KMmwz8iwuU7OWOOMoOObdLOHox3soScrHvGqM
0dg7ZZUhQSIETQUDk2Fkcjpqizhs7sJinbWYcpiaEKv7PWYHLyIIH+RcYKv18hla
EFHaOoUdRfzZsNSwNznnlCSCJOwkVMa1eJGJrEElzoktqPeDsforPFKhnQH+BFNV
KE4BBACwsTltWOQUEjjKDXW28u7skuIT2jtGFc/bbzXcfg2bzTpoJlMNOBMdRDPD
TVccJhAYj8kX9WJDSj+gluMvt319lLrAXjaroZHvHFqJQDxlqyR3mCkITjL09UF/
wVy3sF7wek8KlJthYSiBZT496o1MOsj5k+E8Y/vOHQbvg9uK0wARAQAB/gMDAmEI
mZFRPn111gNki6npnVhXyDhv7FWJw/aLHkEISwmK4fDKOnx+Ueef64K5kZdUmnBC
r9HEAUZA8mKuhWnpDTCLYZwaucqMjD0KyVJiApyGl9QHU41LDyfobDWn/LabKb6t
8uz6qkGzg87fYz8XLDgLvolImbTbeqQa9wuBRK9XfRLVgWv7qemNeDCSdLFEDA6W
ENR+YjDJTZzZDlaH0yLMvudJO4lKnsS+5lhX69qeBJpfp+eMsPh/K8dCOi6mYuSP
SF2JI7hVpk9PurDO1ne20mLuqZvmuDHcddWM88FjXotytDtuHScaX94+vVLXQAKz
mROs4Z7GkNs2om03kWCqsGmAV1B0+bbmcxTH14/vwAFrYSJwcvHsaDhshcCoxJa8
pKxttlHlUYQ6YQZflIMnxvbZAIryDDK9kwut3GGStfoJXoi5jA8uh+WG+avn+iNI
k8lR0SSgo6n5/vyWS6l/ZBbF1JwX6oQ4ep7piKUEGAEKAA8FAlNVKE4CGwwFCRLM
AwAACgkQktgoBGO99PUaKAQAiK1zQQQIOVkqBa/E9Jx5UpCVF/fi0XsTfU2Y0Slg
FV7j9Bqe0obycJ2LFRNDndVReJQQj5vpwZ/B5dAoUqaMXmAD3DD+7ZY756u+g0rU
21Z4Nf+we9PfyA5+lxw+6PXNpYcxvU9wXf+t5vvTLrdnVAdR0hSxKWdOCgIS1VlQ
uxs=
=NolW
-----END PGP PRIVATE KEY BLOCK-----`

const gnuDummyS2KPrivateKeyPassphrase = "lucy"

const gnuDummyS2KPrivateKeyWithSigningSubkey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

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
CQJMlh5V7RMenq01TpLPvc8=
=tI6t
-----END PGP PRIVATE KEY BLOCK-----

`
const gnuDummyS2KPrivateKeyWithSigningSubkeyPassphrase = "urn"

const signingSubkey = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Version: GnuPG v1

lQO+BFZcVT8BCAC968125oFzhdiT2a+jdYM/ci4P/V2mrO4Wc45JswlE2lmrnn/X
1IyT/gFczvbr33bYvPsCazPxFVukk7fd8hLvozCCnarpeUY6PLRyiU6yX6Rp6E8m
5pAR0m6bRiuMYSSmaNwarpjpRdB1zusfsGlFF12V+ooRKZHUlUvwGJEJTpfFvErs
xiyaqVZJqql1mQkmYMBTPjWNA+7xgNGzyXKvdjPHNgzL2xx2eANEuynuM5C+daAi
p/vJrrC24Vv9BuSErGc0UAv42kLZQ/wupA0Mbv6hgSWPY8DkXOvdonrFlgewuR6J
SxDSjpEN9bFaQ3QRCNYK8+hylz4+WW6JtEy3ABEBAAH+AwMCmfRNAFbtf95g/yYR
MjwSrUckrkl81H+sZ1l8fxPQKeEwvrzBVko5k6vT+FRCOrzQcFZjcBbLKBB5098g
3V+nJmrPMhRq8HrFLs6yySj6RDRcmSuKsdI7W0iR2UFCYEJZNiihgIWcDv/SHr8U
OM+aKXaiCYD681Yow1En5b0cFWRS/h4E0na6SOQr9SKIn1IgYMHWrp7kl218rkl3
++doATzRJIARVHhEDFuZrF4VYY3P4eN/zvvuw7HOAyxnkbXdEkhYZtp7JoJq/F6N
SvrQ2wUgj8BFYcfXvPHl0jxqzxsTA6QcZrci+TUdL6iMPvuFyUKp2ZzP6TL+a2V2
iggz1IF5Jhj/qiWvS5zftfHsMp92oqeVHAntbQPXfRJAAzhDaI8DnBmaTnsU7uH9
eaemONtbhk0Ab07amiuO+IYf6mVU8uNbq4G3Zy70KoEBIuKwoKGoTq8LHmvMlSIF
sSyXVwphaPfO3bCBdJzSe7xb3AJi/Zl79vfYDu+5N+2qL+2Z0xf2AIo3JD1L3Ex9
Lm5PUEqohBjDRKP6bCCrggtBfCSN25u08Bidsl5Ldec5jwjMY9WqSKzkZe5NZAhZ
lppssQQTNerl5Eujz21UhmaJHxKQX2FuUF7sjq9sL7A2Lp/EYm8wvDgXV0BJbOZY
fgEtb9JBtfW21VyL5zjRESnKmuDuoveSOpLz+CBnKnqOPddRS8VDMFoYXB1afVJX
vfjbshlN1HRLdxSBw1Q918YXAZVxPbCT1lvHTtSB5seakgOgb8kQowkxUSSxu/D8
DydcQBc2USZOuoePssHUgTQI65STB1o0yS4sA19SriQ2I7erIdbElaWQ3OubMHIm
Yqe+wIR0tsKLcwnw0Cn70RNwDWv61jLstPTg1np0mLNe8ZV0jVCIh0Ftfx+ukjaz
yrQvU2lnbmluZyBTdWJrZXkgKFBXIGlzICdhYmNkJykgPHNpZ25pbmdAc3ViLmtl
eT6JATgEEwECACIFAlZcVT8CGwMGCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJ
EDE+Pwxw+p7819IH/2t3V0IuTttu9PmiOuKoL250biq7urScXRW+jO3S+I69tvZR
ubprMcW2xP9DMrz6oMcn7i6SESiXb3FHKH3FQVB+gCQ2CXeBlGW4FG3FI5qq1+Mg
lFbpRxr2G2FZOlbKYhEYjXD3xd03wlGLvcFvJhQdZFyl5475EGC92V3Dpb465uSA
KgimcBwSLqqLgPwCBVzQHPxPs7wc2vJcyexVIpvRMNt7iLNg6bw0cXC8fxhDk+F6
pQKJieFsGbWLlUYdOqHS6PLYXom3Mr5wdBbxmNX2MI8izxOAAa/AX91yhzm42Jhg
3KPtVQNvxHSZM0WuafTeo9MZRfLQk446EDP+7JCdA74EVlxVPwEIALALVFILo1rH
uZ0z4iEpfT5jSRfUzY73YpHjFTQKRL+Q8MVWNw9aHLYOeL1WtBevffiQ3zDWhG8q
Tx5h7/IiYH1HcUEx6Cd7K5+CnIqHAmDEOIKS6EXfRnTOBB4iuWm4Mt2mT0IFalOy
XNxGnZSC928MnoWpCQDkI5Pz0FsTOibS8t8YfDpd6+TWUkmnpJe08gkNquYk4YDo
bTcyu6UeLDeYhem9z5+YdPpFaCx5HLV9NLEBgnp2M8xXZDZh/vJjEloxCX1OFC3y
cps1ZJsoBBCelqLdduVY1N/olJo+h8FVD2CKW1Xz55fWaMAfThUNDYu9vFR7vMdX
tiivtNqZpvcAEQEAAf4DAwKZ9E0AVu1/3mCyKwygqIo2Gs+wYrKnOhNQB7tDbvW8
2K2HVtDk1u0HVhoCQ3869Z5lM9iWsmoYVh8fs9NAztEYW+1f47+bbdtnxJ2T44g6
knSko1j59o6GOoIvwqyMzBCBcwYCXmFJ5hL0K32laS3sKIfsQiylXzembrJkGBFv
BUEGWfZ2EEox1LjYplGqJN/dobbCPt2E6uS+cmlle92G2Jvoutfl1ogFDBelJzNV
XeEXZDv/fcNvWNAC/ZO8kr370DUoa2qlKlZAMT6SRgQ0JP2OVu+vlmb6l6jJZy2p
+nZ4+uISp2qvWQrIb2Oj5URG+vsbu0DPA8JPqsSWlhMrvmeBiQgtLrEDjpE7bjvY
lRrHagYwAdHIbxnfWE3UZIHVIqqj57GslkiuiPKEkWRQZLwhMToMOksyMgU9WobI
0I86U5v49mq6LN2G1RJOZDHc69F9mgraCYjMMBnA1Ogv5r5xaHYMRoRJabHARsFK
8iknkgQ2V5xgRpH+YXvPDHwe4awvBucHL4tHONyY+k1pzdnDgRFNhO8y+8XP+pG+
4KTILwFQ/2EqZt7xpR84Piy1cwjLz9z6uDmgXjqjJzVGefxn5U+9RfUWZzUri7a5
20GBhtpU07pBcBVml307PGuk8UOJfYMJUi7JwY7sI6HpAyxvw7eY4IV0CjZWNPVf
J6sgaaumzzuJlO5IMQB3REn7NyeBSNSQrEvL40AoeDKVSnEP1/SUmlJpklijE63X
cS7uxBDF88lyweyONClcYBJKumGH4JB0WUAnvM/wFm+x5GIkattbwrdUPPjfof1w
JER90c+qjE539NzMLdO4x4JfiQEsEZ21noB5i72kOmeX+s/HEJnc0q0zcdzDQMj/
JN33HNtzg2t3Z3uaCbOpp8wuri4QGp7Ris5bKngfiQEfBBgBAgAJBQJWXFU/AhsM
AAoJEDE+Pwxw+p78ZJoIAIqFO1v4GDJ3t9XylniCxQ7TfSIAIni5QlM5QHjLD0zG
0Js4HKYPTWqwZU43R/fb4CYsfEkRDHLjZNV8TjNAnsQONSuzsMBckIDwOGSP+wdR
YgULGRXsIuotK0qzZcrRitfSvHSCLjxaQ0gjfGns5xNzeZjrvLOf78PIV/4PzagY
lOiYzFLbfZ2oGWgZRhxo4NQPsUZLAUA2roRQIeguRRpTpQtW1Agqw7/qwEp+LnHE
p4csTYzBy59k5OZrZp3UV/47XKjbqgh8IC5kHXJJ/wzUGrPNc1ovR3yIxBwMVZr4
cxwJTbxVr/ZSA0i4qTvT4o85KM1HY/gmzlk13YTkH9idA74EVlxVagEIAK+tfSyr
9+h0LRgfp8/kaKX/LSoyhgULmqvY/6jceqtM3S2iehbqH/x0tKd0E9OVrjnIUo/D
S85/7wixppT56+ONU6uWcbqsCxClDHzF4JG9fE89Hb2t0vzREgGLYE4sAo5qYU+4
voYSutjsdZYRro0hMNwntyCx3wZvhhtHmkMg7aowSwf84lljOHNCv7LIDmYEz9xl
QODbeVNzwl8bXLe2og162VGXHJ5cRlKOMNOs4R10Rh0cweSPF0RDGdLxbOmOYnCi
tYN6AWOj5KdIf3slbOpmZpg6MaNGqtx2ErtUnos5/pziZJBgsuu4bzpeqExbMJ9w
3PDkcoIz1akryKUAEQEAAf4DAwL48mXB5mn4a2Dye08g7haozfkicHhwNRLeg3LO
QM9L4ZkTq9IdA7Hd97b6ewDygUQA5GxG0JjpZd0UNhYAKpWd2x678JvpPfJNdHhZ
dh9wo7EhW2HQi+A/qAzuHz58Znc4+vO9+3ECMvIdcaqZnQ2jDF3pooOOY9pOj7Hj
QPrNDeePGwbHpDgMPip7XdzWCQU3j9kohhhdgrAOKBI0wNh68HGPQ3E3KOzsEvLo
0f90L8DEFl8iTSFW4UqCVjfF4rWTIFKHMMTxut6Yivv2L8q66oV3gC3dKthd2kxV
IsBtJ9SmIjvdsTQ8yi67oHyfBMvzqPxdD0QJfBu8z+4LKxGOtrHoYRnX9MaSAJjE
47m9fhVlUeiaZXzAoI8J9D3NBoUJnFJ4zsJCUkCZY9gF4qZSWzuWathf2U9lSmDH
JlrxLIXChTGKYcjNOL42EOh+GQJjf/C5KVWSh9pfqMUFptuZ+k4A+xSDdnF8upoU
Odcm6fVobKXPouU8fLh7C5R9p+vYzJmFh9MP+2vd86CGxMDvB3l5GdacNY+1/ycA
gmDcqqdv3xB3n6+COEytOhIcrwF1cHA0nGw9sDeGX2Ly8ULhIld/axXoCXp14HTT
YIo7hijK0/FTUQg+J3HEvxfbl5vae4pPLp+x8zN9IHHx7SR4RKiYtZqqmuZAt3B0
WCNI18RO+rT3jNEsdY1vmwiKyHStwgb1dAYXSkBTNc8vFwIxFettpoHs6S9m+OQk
BCbc0ujOxCmduJDBznfw6b1ZAb8pQzVLpqDwPMAzgkLwajjs876as1/S9IU+P3js
kJzvEj52Glqs5X46LxdHEF/rKp3M2yOo/K5N8zDsp3xt3kBRd2Lx+9OsyBVoGuWn
XVHPqRp70gzo1WgUWVRI7V+XA62BflNDs6OnDmNjWH/ViQI+BBgBAgAJBQJWXFVq
AhsCASkJEDE+Pwxw+p78wF0gBBkBAgAGBQJWXFVqAAoJEBRr6IQvgxaLIcQH/2qn
zACX1+6obanMnYvWeF9dON+qfPGBN7NDtyhBDnsJuUL6WQGTGb3exFOFodQ+bCVV
pH7+uPENwpVbDd4um0Rkw43HejZa+IEREKBzh6IHtICIJ+GRcYb1bEKl0V3ezluz
sBhOvl23/A+mBDEqmWyfD0OMHejZDamKUVrLz/S8sP4Wp6m731AhxV3EjTjfzE4a
RxJiL7mcoDFzFg7hiCT5Tq6ZGFaZMW5690j3s0mu7lVj1aCjWKQAVFzeKKZFoZOo
Gjvd6xCdUmqwvqudypvkdbwZTHHibLVmgq7IJzTDaPQs73a0s5g5q5dVCWTw1zxc
6Y7qtqBrjDSJrOq2XRvxXQf/RQZIh/P9bAMGp8Ln6VOxfUWrhdAyiUrcbq7kuHwN
terflJi0KA7/hGoNNtK+FprMOqGQORfEbP0n8Q9NcE/ugE8/PG+Dttnbi7IUtBu9
iD5idEdZCllPr/1ekSIzxXIlBcrp92pd+SVDZ11cJR1tp+R+CyXah9VuBRVNZ5mI
rRXJmUbQHXkL/fCyDOkCFcrR+OG3j0bJvv2SQXkhbsbG4J/Q3hVXadZKqTSTNLWt
FbLYLwTpGXH2bBQyDkJJ/gI7iNUm6MtGPYrD2ZuB/XGyv/Q+KfNJk/Q9Dxb7eCOE
wxSLXhuDL3EPy4MVw8HE0TixCvq082aIbS8UAWOCnaqUyQ==
=3zTL
-----END PGP PRIVATE KEY BLOCK-----
`

const signingSubkeyPassphrase = "abcd"

const eddsaPublicKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mDMEVcdzEhYJKwYBBAHaRw8BAQdABLH577R+X2tGKoTX7GVYInAoCPaSpsaJqA52
nopSLsa0K0Vhcmx5IEFkb3B0ZXIgKFBXIGlzIGFiY2QpIDxlYXJseUBhZG9wdC5l
cj6IeQQTFggAIQUCVcdzEgIbAwULCQgHAgYVCAkKCwIEFgIDAQIeAQIXgAAKCRBY
ZCLvtzlOPSS/AQDVhDyt1Si33VqLEmtlKnLs/2Kvi9FeM7yKU3Faj5ki4AEAyaMO
3LKLyzMhYn7GavsS2wlP6hpuw8Vavjk2kWE7iwA=
=IE4q
-----END PGP PUBLIC KEY BLOCK-----
`

const eddsaSignature = `-----BEGIN PGP MESSAGE-----
Version: GnuPG v2

owGbwMvMwCEWkaL0frulny3jaeckhtDjM5g9UnNy8hVSE4tyKhUSU/ILSlKLivUU
PFKLUhUyixWK83NTFVxTXIIdFYpLCwryi0r0FEIyUhVKMjKLUvS4OuJYGMQ4GNhY
mUBGMXBxCsDMP7GA4X/4JlF9p1uHWr2yn/o+l1uRdcFn6xp7zq2/PzDZyqr0h+xk
+J9mYZEyTzxYwov3+41tk1POxp2d4xzP7qhw+vSpjus5sswA
=Eywk
-----END PGP MESSAGE-----
`

const eddsaSignedMsg = "Hello early adopters. Here is some EdDSA support. The third.\n"

const keyWithRevokedSubkeysOfflineMasterPublic = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

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
tmIuHu1WeYgLy8ZUju/jcJcb9XYI6rBP
=YFA2
-----END PGP PUBLIC KEY BLOCK-----
`

const keyWithRevokedSubkeysOfflineMasterPrivate = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

lQEVBFa5QKwBCADHqlsggv3b4EWV7WtHVeur84Vm2Xhb+htzcPPZFwpipqF7HV+l
hsYPsY0qEDarcsGYfoRcEh4j9xV3VwrUj7DNwf2s/ZXuni+9hoyAI6FdLn6ep9ju
q5z+R7okfqx70gi4wDQVDmpVT3MaYi7/fd3kqQjRUUIwBysPPXTfBFA8S8ARnp71
bBp+Xz+ESsxmyOjjmCea2H43N3x0d/qVSORo5f32U67z77Nn/ZXKwMqmJNE0+LtM
18icqWJQ3+R+9j3P01geidsHGCaPjW4Lb0io6g8pynbfA1ihlKasfwYDgMgt8TMA
QO/wum2ozq6NJF0PuJtakVn1izWagaHcGB9RABEBAAH+A2UCR05VAbQvUmV2b2tl
ZCBTdWJrZXkgKFBXIGlzICdhYmNkJykgPHJldm9rZWRAc3ViLmtleT6JATcEEwEK
ACEFAla5QKwCGwMFCwkIBwMFFQoJCAsFFgIDAQACHgECF4AACgkQIJquI8PlfWNc
2ggAhBNF6bY1JMOC3tCLhH1ijg8JLbq4JVW/tcSY+PFC8SpJN4btHSk20ZM4sgQ7
f8mnkUx8OYL2j9QelHEShLZkHsXfJWVGFtfbST07xIV/4bUnJEUp8Ph0b/lLrO/7
Xp4M7yxkQyMm9mhJ/Vb7QywAZh/fHTeXcjLkRuFT+v8gIbvbkAHhoUSVJLf4N1+l
X8WBQx/cc9QdtfsNgbB++6SBlTi7znd153SkVyyXGbwa6wgTBse1sAOwXsKr3U0y
HXHxIRdvPpXvHsWz9E5F1OzeNUtaZwvaRa9JI/zyP6fElMvVgakO5ra4wfpnt1ld
uN3NDOVX8RhYFnmNMMe8/GFe0J0DvgRWuUCsAQgAsaBga3WD/FGtlDOixnNobmGz
Hbzyn/pcqcni28thq7eKqs7ev2pGStCRqWlyPpdFnVo1sZ2HAJGiVmwSvyVlziCJ
xDFg+k32PyNi4ykPpQRomEGtoKnK2KS2ZXu6tKBDD0wAPcEdMLjPiTia/JFmHtPL
EjuwdoSzFDFntoMQ7khTchHhgf9YPFojB/aSGNXMCyq3PfY1ONZGHefYzlPvSfID
IHRFpafvngVy7eIHNyqVvU0Uz7LT6tFY6RvRsecbKWWT+sm+iUbVNRkGxZ98T5zy
RKwJXaHwrgcapGdYSk/O0wnQvFkln3jiw4WnMWXKr/NegW8c1hC8CeQn2w55fwAR
AQAB/gMDAgzoGW952mkM4qM5+ebuLarYn1KUnzL6ivVVJoo1xTNAn4ZGp8gJUTm4
Q9qi4VQo5yEJtDPxd+UWeL70dq0np3Fv9eYnC22IjTFtx84GkXxYD0mT0FJv3CNr
xISjN1i8YX57fF5TV9Spx1BhNlF93FRwaOIfqAa01VQTGzYpBSEyRgYksL1le7m3
YaEmF39uX5oIvCDRt8Gx0NRLhQ9OEIRZNo8jZPJYhh15fRYHV2HRx1G37EyzFPkG
VSp3HjHxiZLMntyys95CE2a4yFII8emyySiCMvHgQCzCgmGRCBEPBI+owhjuXx6I
URnua0IVCW+yKt5eEs1fzGuRIcYu2bF6VhQPW5gCvtlvMbvDcBxUrRakzBEIilEY
hSQY06klV95DTOpgv7UWAWcOmJO8SGky5lYuge5BcEUSH8JfMF3C2xgsE68rgsGG
gbjNurTHZzlzCN8K+mVYZ2rMP+/ZeNH44+9eMp3ynsE9hrbLDb0rfyXtxa6opIvn
rTJXytkuoOfKURnoqiAruIkyaKmOJkh8xut2eeXigar2AZI+cupVBNI8Q7e3aUJv
UahuYSBletbfdfFyseGjinAdZ+s0D6GflXYzF8D08kUnUCb9yDMWQiQ8y+fQdeYv
nHF1rgGY8720oGtdMEiAwTzzMwSpwRRtjdvZ2KRKLL7HzmXiVHyvh5RD/0co/l3F
cER/z1Ks9RXmAK8UxWJigjjuIVDCsWdQfUm69JY6cL+sryX/KZQPPbCrW2Nmkcil
1Q1A88v8hlUXMKZl0DAZ/Cpbb2fMMIKmJToyCBNtkVncJS9XxitQTF1zqYWnz7AF
+/ZL8khw71U/8hTcCP7mCv5WuPye7OhGRzcEttkWrM6E7B6EOF+65mDhwxInQ4Ep
sHWTT3698o3wI5mxLBo+MgO81yiJAR8EGAEKAAkFAla5QKwCGwwACgkQIJquI8Pl
fWPaaQf/XrjBHGNAbPMFpUs6GvpjLcD+KEQHF1lFZe6pO2BdLeTwTc7vbfqkkbeo
yJyJmjgm6CsjApmkRW9i6YCreC/Hb7nwXnO0BiXtIVm2SdCrthltLspcPY5dCu47
yUIzF2sf4lF98Ef5nU8srWL/0jb2DuNP2a7TPmnttuDncVnNq8kTwvC+SfKV7Qr3
GGMuVexl06UynVeTtX9fQGCqrcAUo0uP/3/g7VwV9lXFgs/kO+7t2f4WZZlim9eM
uAduO7BRWk9XNela8dHvl59mWjAcMI/ycNf+Kcu4yAUFcj7sxcDD/8FuiCQL3f3p
2E+DDy3HUhGeLU8ejzbhjC0Xr/U5+p0DvgRWuUJ4AQgAtg3kAKxBDSmDmLArNJKm
W5xogX7GtL/abzrMEOSouVXoiO30s+oK/XQ8QidyIaOfDtAxtkAQ6PG5+5wPOVo9
sgTSV0AmrnNzD3LYhgSasXC9hbONUsbDVH80UA2m3KHYdSKuvbvj8I0oXx0e/DFi
U5gkGskksWrtIhIWA/nZxD+hWzfJ7gcQF7zX54TPrgPGNkfjNeFOouKXdGUZXvl7
tUNcIRe7+DoSpzmHs0xW1At5FxjXVJwnqbuQDNIcPpFfMKrUoTHOLsPSifT4btVZ
m+F/wPqXdSdUt9GW61nH3as42BfV4zqnWoJhtQerkW4uqZ+O90vxtUP+OEu1wXRL
6wARAQAB/gMDAvpqiFLi6yhb4mWr2ofsIL23V3i7oCCnuHTe1c/up8T4TamSDsn/
tGXoZEDO0oMZe8oqIVR7396KBvlRx8bDDX/SYzOuXjl+b6o7lkhnA0VRrDkGmpm1
gzgvMs+JYvP2eiLVPji1yfNKunA5Wlu8CKiiHFPDnCgheYGgMpRs+QU1rneeTOeo
Sfk+aHCL6sHQ7O9riZvou6+33D/UcmwXH6QVVqstSnRhGuIkz9DdrGrZAiFqujHt
nvU0tKRrLjtfwM+UfzjeTakmqR0BWEuLAIf6BASlavH7Tn8cCg6njPSLCeDRm+Tl
vO7JTxh2KkDOhgM1+TjkUw3AmB32hXMfPuvxoZ+Hhl03xORSmTEdkCrw9IkLdLTB
oWo06JMKYHvMp3mtQXkecHUqPI46LhT2MguzBrxnDymYNX4yY6DMSYRyMCn1+A4w
tMZQdNbMygzbOHP6jle9jOwQFGTOpvP68slaBHMfjrOCluMxvudxhhw1qVuj3nhO
RQxtH32w9IrPjLnpzgxXo1vZ5dXvHRQbP1T8BCgHDdJtMWNn/VSetORUhIG8nsNj
cDHhXj3YBokvAyicjP8euziQh2WAHA7Bc5plPLBlGEDhMsIwTtpec0TN9A2PfqTV
L8C/xTa8uIwopiSWtGwtl/7MuyunIk7LBIhAlkQoOiLUe037sid8Fzn6f1ZLspNT
89f2icJVTgzb9Vrec5afR9L4gcOV3N/+COSgZZqsUmggPkJaLXxNa45qAPWzpV60
oUsDGqVTThSbCd/qaj+WfEKAaUOfeuN4wP+4PkhFv7soShsCYDe/46VukSqta9ma
rlc+JKgoZ2BVh6tBfQMHo1yaszVa34yvoH0JiP8MFSaQur9ljCv/wubex+djrm+w
zbsr+7HN+ey9mUMcV/ug5ntH6xMoBP2JAj4EGAEKAAkFAla5QngCGwIBKQkQIJqu
I8PlfWPAXSAEGQEKAAYFAla5QngACgkQWiVrsAiVPozJhwf/edwVPbyyI2EV7twE
C83AF1cEQ1Hpwsor079WWfoythLaX6hzInBOGT8UC5WdMXpKbiFjBi/0DqFCan0x
oJ1aysTvfAB8Hyq9y8FKc3gfFvibFzBvvLW0fCo1IkQllNQCu8hFv7e1tUvdQO/N
/2pcEncgLXzPAt3Iu/lbTyDH5B15wMQMH/6t+Z82qEh2q6x5j2EiBix2adeRaVF1
iDEpB0nW9GfSBeb6TPOap8l6FJGPYLqdDdd/S9q7O5hsnXvsr9BFT4rzqV8HzHQS
2SVOT60uIw8Vnk4iyYH5mVZ4i6iNferFSxfa2Ju32U/q3J5CHJhETt1lStDRsm8q
QXGApvASB/9vw/R13U1IFQKZi0SZ0LJBRbuXf+LEGe+15o00RoghB1FLzyZ3SHiK
OlnPdFtB4FpUHhE/qp7ehWLw27/5FF28PXJogIUdA5id3pa298bRCuvwUtJvjahS
aPIry53/Th2ZELWeXJ9nJYtzwtptvnCrr9rX4Bly+iopNfPdj9BVTOR3miC33bKE
8E0mKK5OrKtwp82viZKkmOeZmYZw2mOV5NmrtY5I3HQrsYRVoR9/9XUt7nCrRB93
e9rjHlB7837a0sCc60p4/+9y4lnqaHTV/IcmWgfvyb69F5Frpj3NfmZSY1HuBMDr
2qXGiMxMPqPwdaqiNTRwEeoWVZ1IBItUnQO+BFa5QqIBCADiZy6KgIfcdNSluaYO
h/w5HchCL6r+5FMKeX/BtttLl9l+0ysDZUZVMx5WMPjRLpBkRLFK9hDydfXkCBwA
vgtn4PNxRfETi4uIV2R7TBGh4Ld0Lw71oX1kZajB2EaKlQob+wmZ9vKypVebWurg
ulIRtLbWeBMqAol91Oa439lK4MrY/5L6Ia+uFDbpqkylhToIUxos0gVIUSW4nxVi
+AyhD8tVxrV0IghZmRucrXSFdCN4PhPWMV30eBiBirtjeCBsjE/x8U8gpa23JN/f
YKbEcKxtNOMgZmo5HyCiCunXov4xmt/j6cvkwAPo3lylUsBz3jm9BEk7lbe3Qliv
7HTLABEBAAH+AwMCRF0ld4vQ8q3iFCfmC5MZZmgJGKOXqjKajQazxTD4RDeXTE3M
Z0yV4Gqhsz2CixL73RTUspA58BIOv/hiA8Oze8vhsOzIjb91VUAsbyXkQa1MYKjo
fb9d0RESSB2QcW7ACXZpZBsBWVfl/cd/+rLEvZfgLMS+tKTEvNAA3zWG0fQ5bI7W
kS4Jsqq273d+2PHx+Prs+4wr/6AQw8dJoP6gbaMzjJQM7DWRIi+mXxThATTHzQxS
gOgxx/HauL2iHMsyuPOUl1lfZvVxhWJN9psPxYypyPWH0ZpeQCwDh2tecxIPPWTO
fdSxbIoK6rIR+EcaG/uu55n1mbJdDb1N2Zf87D0WgXZglPBfc0JvsxhgkLL9pTSe
RmJiREK7BlxLBZzdopBEbvaJL5MwOgL3fA3OwM8UU9qYmOge8OgxHp+Nd/nP3StS
l+h7z6VInGVTnf5sr0zhNxQD5N8Gyulf+1MneCjfK6/3s5rpLNOHneHybd0W0F2r
FkYbw4rSi6v2bMvHybG5EzQutYZuNoJ56DaXZDir0vR8aBb6JPSlNI1Kmx9OEce1
cP1jhd82DEmk7X3hRVYDwu2cw/6egbGv3+kxwMzktiEnri0fgbghsXrP/KiTSG+X
+h3Q4p+NQysYzOHc86YJO9FRD4FPkG1vkDwzUQvavAEWSLK/y1fV89Ky4XmQUT0E
K70oDa9VL6mfRqbQ4fMQX7f58OlJDTtPf2s11EcDfsD/WdMKUnNIiyPPjaEDj2+U
qcNkcewEk6D71ZvNSdRKzLCBFqVywqZOdWhBOZ4Mk9YxKiJBb2VfjfOMwReoBmrg
NgplL58qWC9QIuP32fxRMC0NN8E5Zdvz5S7NqhDwp6YURWFuHtFXzlrbgcYsR9SF
IGMeh0ULUS2dpOjboiHLvZCQ/KEQQMi5BokCPgQYAQoACQUCVrlCogIbAgEpCRAg
mq4jw+V9Y8BdIAQZAQoABgUCVrlCogAKCRDiQ1ltWGXWD46dB/4kH7APfFeEZc6F
DW2Au2U5a2zXJ84yetDcLrTTwQQyAUPqAZwSS9/ptFok9oeqMfkA9Ck7f2vbCZ/p
klse3Rw0pUI77n2ljvMwBOtjKhOi7z2/FDiknxFc1oR3mJEFKvHJoOmOzNn5q9Tq
xJVr7Sh1C+AAKz3K9BfWsW3+EhRqIfxAsdA2qvYCNE4nR5N5hK0wBkhIsQNx6rHp
tOGK6p0KEbQhjxtmO7YAzq1uAP2VgX0Q23+WchDBx8yJ1MmdWwtAgNrzKBmeHE15
o50FQH4adxJHv2T38U89gw9VdnoEnYBz4WaXfEbOsMQTq0303P74Ckvo8cozWdp/
ltnoR0yktXAH/2ugE11IKHYRkWsSG9zaFUsAKf17ja6ysgbXN6eVV/swv///ERDV
9KnmMBezbnqQ355wofT+TSfv6OQxWh8/0PtoDVRMRqUW+0mjSdrttqvWvvc0tDMj
h988QVJQ9nRcaVIUnkkC6JdWq1Z18LAYQia7HnDHv5mDL829qkPHZGnKy4ZFYnfU
hcEmNkTXuViQf5ahEHscZWpv1w1JUgSZC0f4184cm3DbZ9rufazniTJsYgpIF0ij
dYB5bLE6TJmYLlwnCRHxFZjy2urflKlNJRVdcMUUd6vt10yKWczA/MIvFxiR1ZRB
prm2Yi4e7VZ5iAvLxlSO7+Nwlxv1dgjqsE8=
=q2vt
-----END PGP PRIVATE KEY BLOCK-----`

const keyWithRevokedSubkeyPassphrase = `abcd`

const keyWithRevokedSubkeysPublic = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

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
tmIuHu1WeYgLy8ZUju/jcJcb9XYI6rBP
=YFA2
-----END PGP PUBLIC KEY BLOCK-----`

const keyWithRevokedSubkeysPrivate = `-----BEGIN PGP PRIVATE KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

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
=O4TR
-----END PGP PRIVATE KEY BLOCK-----`

const jwbSignedV3Message = `-----BEGIN PGP MESSAGE-----

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
=+5CA
-----END PGP MESSAGE-----`

const jwbKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
AAADxnB0UE8AAAAmAAAD6G5sTkwAAAAoAAAEDmVzRVMAAAAmAAAD6HRoVEgAAAAk
AAAENnRyVFIAAAAiAAAEWmZpRkkAAAAoAAAEfHBsUEwAAAAsAAAEpHJ1UlUAAAAi
AAAE0GFyRUcAAAAmAAAE8mVuVVMAAAAmAAAFGGRhREsAAAAuAAAFPgBWAWEAZQBv
AGIAZQBjAG4A/QAgAFIARwBCACAAcAByAG8AZgBpAGwARwBlAG4AZQByAGkBDQBr
AGkAIABSAEcAQgAgAHAAcgBvAGYAaQBsAFAAZQByAGYAaQBsACAAUgBHAEIAIABn
AGUAbgDoAHIAaQBjAFAAZQByAGYAaQBsACAAUgBHAEIAIABHAGUAbgDpAHIAaQBj
AG8EFwQwBDMEMAQ7BEwEPQQ4BDkAIAQ/BEAEPgREBDAEOQQ7ACAAUgBHAEIAUABy
AG8AZgBpAGwAIABnAOkAbgDpAHIAaQBxAHUAZQAgAFIAVgBCkBp1KAAgAFIARwBC
ACCCcl9pY8+P8ABQAHIAbwBmAGkAbABvACAAUgBHAEIAIABnAGUAbgBlAHIAaQBj
AG8ARwBlAG4AZQByAGkAcwBrACAAUgBHAEIALQBwAHIAbwBmAGkAbMd8vBgAIABS
AEcAQgAg1QS4XNMMx3wATwBiAGUAYwBuAP0AIABSAEcAQgAgAHAAcgBvAGYAaQBs
BeQF6AXVBeQF2QXcACAAUgBHAEIAIAXbBdwF3AXZAEEAbABsAGcAZQBtAGUAaQBu
AGUAcwAgAFIARwBCAC0AUAByAG8AZgBpAGwAwQBsAHQAYQBsAOEAbgBvAHMAIABS
AEcAQgAgAHAAcgBvAGYAaQBsZm6QGgAgAFIARwBCACBjz4/wZYdO9k4AgiwAIABS
AEcAQgAgMNcw7TDVMKEwpDDrAFAAcgBvAGYAaQBsACAAUgBHAEIAIABnAGUAbgBl
AHIAaQBjA5MDtQO9A7kDugPMACADwAPBA78DxgOvA7sAIABSAEcAQgBQAGUAcgBm
AGkAbAAgAFIARwBCACAAZwBlAG4A6QByAGkAYwBvAEEAbABnAGUAbQBlAGUAbgAg
AFIARwBCAC0AcAByAG8AZgBpAGUAbA5CDhsOIw5EDh8OJQ5MACAAUgBHAEIAIA4X
DjEOSA4nDkQOGwBHAGUAbgBlAGwAIABSAEcAQgAgAFAAcgBvAGYAaQBsAGkAWQBs
AGUAaQBuAGUAbgAgAFIARwBCAC0AcAByAG8AZgBpAGkAbABpAFUAbgBpAHcAZQBy
AHMAYQBsAG4AeQAgAHAAcgBvAGYAaQBsACAAUgBHAEIEHgQxBEkEOAQ5ACAEPwRA
BD4ERAQ4BDsETAAgAFIARwBCBkUGRAZBACAGKgY5BjEGSgZBACAAUgBHAEIAIAYn
BkQGOQYnBkUARwBlAG4AZQByAGkAYwAgAFIARwBCACAAUAByAG8AZgBpAGwAZQBH
AGUAbgBlAHIAZQBsACAAUgBHAEIALQBiAGUAcwBrAHIAaQB2AGUAbABzAGV0ZXh0
AAAAAENvcHlyaWdodCAyMDA3IEFwcGxlIEluYy4sIGFsbCByaWdodHMgcmVzZXJ2
ZWQuAFhZWiAAAAAAAADzUgABAAAAARbPWFlaIAAAAAAAAHRNAAA97gAAA9BYWVog
AAAAAAAAWnUAAKxzAAAXNFhZWiAAAAAAAAAoGgAAFZ8AALg2Y3VydgAAAAAAAAAB
Ac0AAHNmMzIAAAAAAAEMQgAABd7///MmAAAHkgAA/ZH///ui///9owAAA9wAAMBs
/+EAgEV4aWYAAE1NACoAAAAIAAUBEgADAAAAAQABAAABGgAFAAAAAQAAAEoBGwAF
AAAAAQAAAFIBKAADAAAAAQACAACHaQAEAAAAAQAAAFoAAAAAAAAASAAAAAEAAABI
AAAAAQACoAIABAAAAAEAAABIoAMABAAAAAEAAABIAAAAAP/bAEMAAgICAgIBAgIC
AgICAgMDBgQDAwMDBwUFBAYIBwgICAcICAkKDQsJCQwKCAgLDwsMDQ4ODg4JCxAR
Dw4RDQ4ODv/bAEMBAgICAwMDBgQEBg4JCAkODg4ODg4ODg4ODg4ODg4ODg4ODg4O
Dg4ODg4ODg4ODg4ODg4ODg4ODg4ODg4ODg4ODv/AABEIAEgASAMBIgACEQEDEQH/
xAAfAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBwgJCgv/xAC1EAACAQMDAgQDBQUE
BAAAAX0BAgMABBEFEiExQQYTUWEHInEUMoGRoQgjQrHBFVLR8CQzYnKCCQoWFxgZ
GiUmJygpKjQ1Njc4OTpDREVGR0hJSlNUVVZXWFlaY2RlZmdoaWpzdHV2d3h5eoOE
hYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX
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
n1G9ErbN5GmQOzo1On1W300V4/QD
=0eMh
-----END PGP PUBLIC KEY BLOCK-----
`

const yieldSig = `-----BEGIN PGP MESSAGE-----
Comment: https://keybase.io/download
Version: Keybase Go 1.0.15 (linux)

xA0DAAoB4RTVFXqRdDoBy+F0AOIAAAAA63siYm9keSI6eyJkZXZpY2UiOnsiY3Rp
bWUiOjAsImlkIjoiMDVhYjU4ODQ0NGY5ZTAzMWM5NzY5MmI5MGFiNWRkMTgiLCJr
aWQiOiIwMTIwMzlkYzI3ZjNjZGM4ZGIyZTYxMTFlZjhjMmM1MTJiMDQyZGEyZTM3
ZWRhNDEyMDI0MThjYjg2ZTJjODJlMWNiZDBhIiwibXRpbWUiOjAsIm5hbWUiOiJ2
YWlvc3V4Iiwic3RhdHVzIjoxLCJ0eXBlIjoiZGVza3RvcCJ9LCJrZXkiOnsiZWxk
ZXN0X2tpZCI6IjAxMDE2YTAyMzZiZmEyZGEyMGVjZTZhMzhkYmRkNzlkMDM0MTdi
ZGM3ZTVlY2U4MDgzNjViMTU5ZGYzY2YxYjBiN2ZmMGEiLCJmaW5nZXJwcmludCI6
IjBlNWEyZjExOGM3MmVhYzg3MDNiNzFhYmUxMTRkNTE1N2E5MTc0M2EiLCJob3N0
Ijoia2V5YmFzZS5pbyIsImtleV9pZCI6IkUxMTRENTE1N0E5MTc0M0EiLCJraWQi
OiIwMTAxNmEwMjM2YmZhMmRhMjBlY2U2YTM4ZGJkZDc5ZDAzNDE3YmRjN2U1ZWNl
ODA4MzY1YjE1OWRmM2NmMWIwYjdmZjBhIiwidWlkIjoiOGJiY2JhNDJkNDhkY2Zm
ZTZlYTZlZmEzZjU2NTZlMTkiLCJ1c2VybmFtZSI6InlpZWxkIn0sInNpYmtleSI6
eyJraWQiOiIwMTIwMzlkYzI3ZjNjZGM4ZGIyZTYxMTFlZjhjMmM1MTJiMDQyZGEy
ZTM3ZWRhNDEyMDI0MThjYjg2ZTJjODJlMWNiZDBhIiwicmV2ZXJzZV9zaWciOiJn
NlJpYjJSNWhxaGtaWFJoWTJobFpNT3BhR0Z6YUY5MGVYQmxDcU5yWlhuRUl3RWdP
ZHduODgzSTJ5NWhFZStNTEZFckJDMmk0MzdhUVNBa0dNdUc0c2d1SEwwS3AzQmhl
V3h2WVdURkJDOTdJbUp2WkhraU9uc2laR1YyYVdObElqcDdJbU4wYVcxbElqb3dM
Q0pwWkNJNklqQTFZV0kxT0RnME5EUm1PV1V3TXpGak9UYzJPVEppT1RCaFlqVmta
REU0SWl3aWEybGtJam9pTURFeU1ETTVaR015TjJZelkyUmpPR1JpTW1VMk1URXha
V1k0WXpKak5URXlZakEwTW1SaE1tVXpOMlZrWVRReE1qQXlOREU0WTJJNE5tVXlZ
emd5WlRGalltUXdZU0lzSW0xMGFXMWxJam93TENKdVlXMWxJam9pZG1GcGIzTjFl
Q0lzSW5OMFlYUjFjeUk2TVN3aWRIbHdaU0k2SW1SbGMydDBiM0FpZlN3aWEyVjVJ
anA3SW1Wc1pHVnpkRjlyYVdRaU9pSXdNVEF4Tm1Fd01qTTJZbVpoTW1SaE1qQmxZ
MlUyWVRNNFpHSmtaRGM1WkRBek5ERTNZbVJqTjJVMVpXTmxPREE0TXpZMVlqRTFP
V1JtTTJObU1XSXdZamRtWmpCaElpd2labWx1WjJWeWNISnBiblFpT2lJd1pUVmhN
bVl4TVRoak56SmxZV000TnpBellqY3hZV0psTVRFMFpEVXhOVGRoT1RFM05ETmhJ
aXdpYUc5emRDSTZJbXRsZVdKaGMyVXVhVzhpTENKclpYbGZhV1FpT2lKRk1URTBS
RFV4TlRkQk9URTNORE5CSWl3aWEybGtJam9pTURFd01UWmhNREl6Tm1KbVlUSmtZ
VEl3WldObE5tRXpPR1JpWkdRM09XUXdNelF4TjJKa1l6ZGxOV1ZqWlRnd09ETTJO
V0l4TlRsa1pqTmpaakZpTUdJM1ptWXdZU0lzSW5WcFpDSTZJamhpWW1OaVlUUXla
RFE0WkdObVptVTJaV0UyWldaaE0yWTFOalUyWlRFNUlpd2lkWE5sY201aGJXVWlP
aUo1YVdWc1pDSjlMQ0p6YVdKclpYa2lPbnNpYTJsa0lqb2lNREV5TURNNVpHTXlO
Mll6WTJSak9HUmlNbVUyTVRFeFpXWTRZekpqTlRFeVlqQTBNbVJoTW1Vek4yVmtZ
VFF4TWpBeU5ERTRZMkk0Tm1VeVl6Z3laVEZqWW1Rd1lTSXNJbkpsZG1WeWMyVmZj
MmxuSWpwdWRXeHNmU3dpZEhsd1pTSTZJbk5wWW10bGVTSXNJblpsY25OcGIyNGlP
akY5TENKamJHbGxiblFpT25zaWJtRnRaU0k2SW10bGVXSmhjMlV1YVc4Z1oyOGdZ
MnhwWlc1MElpd2lkbVZ5YzJsdmJpSTZJakV1TUM0eE5TSjlMQ0pqZEdsdFpTSTZN
VFExT1RnNU1UZzRPQ3dpWlhod2FYSmxYMmx1SWpvMU1EUTFOell3TURBc0ltMWxj
bXRzWlY5eWIyOTBJanA3SW1OMGFXMWxJam94TkRVNU9Ea3hPRFkwTENKb1lYTm9J
am9pWlRBME5qaGhPV1kyWTJNME5XWTVOVGMzWVRVM1pqaGlaVEF6WkRrMFpHUmha
R1F3WWpOa09EQTVNR0kxTVRZM04yRXpOalpsWXpBeFlUSTNZV1EyTUdWaU5USTNa
alJqTnpReVptVmpOVFUyTjJNd05qSXpOR1JsWldZek1HSTJOekZqWVRVd1pqTXpO
ek5oWVdVMk1qTTVPRGN4TXpBNFpXUTNNV0ptWlRRaUxDSnpaWEZ1YnlJNk5ESTVP
VGN5ZlN3aWNI6UpsZGlJNklqUTVaVEE1T1dVNE1UTmhNREl3TURSbU16QTBOakUz
T0RWak9EaGhZakU1TmpBNE1XVTROamhqTW1OaE9USTFZbVJsT0dVNU56RmxaVGxs
T1RKbFptTWlMQ0p6WlhGdWJ5STZOeXdpZEdGbklqb2ljMmxuYm1GMGRYSmxJbjJq
YzJsbnhFQSszd2lUeGpWaWtTdVBNb1JCcVlYclVzelhnMmpZbjBHaFBVRjBZNS8y
Y2REVGdPdkFwejlhNUVLWm84cHY1WDRJZmVhUWRNWGRQTFR2clpPNVpQNEVxSE5w
WjE5MGVYQmxJS04wWVdmTkFnS25kbVZ5YzJsdmJnRT0ifSwidHlwZSI6InNpYmtl
eSIsInZlcnNpb24iOjF9LCJjbGllbnQiOnsibmFtZSI6ImtleWJhc2UuaW8gZ28g
Y2xpZW50IiwidmVyc2lvbiI6IjEuMC4xNSJ9LCJjdGltZSI6MTQ1OTg5MTg4OCwi
ZXhwaXJlX2luIjo1MDQ1NzYwMDAsIm1lcmtsZV9yb290Ijp7ImN0aW1lIjoxNDU5
ODkxODY0LCJoYXNoIjoiZTA0NjhhOWY2Y2M0NWY5NTc3YTU3ZjhiZTAzZDk0ZGRh
ZGQwYjNkODA5MGI1MTY3N2EzNjZlYzAxYTI3YWQ2MGViNTI3ZjRjNzQy52ZlYzU1
NjdjMDYyMzRkZWVmMzBiNjcxY2E1MGYzMzczYWFlNjIzOTg3MTMwOGVkNzFiZmU0
Iiwic2Vxbm8iOjQyOTk3Mn0sInByZXYiOiI0OWUwOTllODEzYTAyMDA0ZjMwNDYx
Nzg1Yzg4YWIxOTYwODFlODY4YzJjYTkyNWJk5WU4ZTk3MWVlOWU5MmVmYyIsInNl
cW5vIjo3LCJ0YWci4zoic2lnbmF04nVyZSLgfQDCwVwEAAEKABAFAlcELrAJEOEU
1RV6kXQ6AADxrBAAVH7MdyaQLinPghHd1uscch7np1QxfJrjkbM5Fgq2pDzOJByn
Hoiy9VY0vomQ5Ja9WFX7Zfs7t9/S+Rj+0JHb/0UqlOeO5cBx6Lh7hXHLrNdyfx0J
3v8GkOmyMTatzqPS7h0jj8uK6RfG2AQNQ0MzIt95qaynhNzpdFDbC38smwU7TIMh
voxHXTaCFoDsprF+oeqAUQJMnJexNuhkIklA9RQWBUir651IeHSbCj6IfExQSosI
oWoESoE7i3E1Rmqi2QuKYmeLHK0iHQJ5BtjOL/q5Ug+N+EfYojRRfVyCdcvDdiiv
Wfoq2ih+Ieg8ZtKJ/ZSbJM7Vh38Y+2x6iRHbD9cfJkkvelNj6qZFYu7yJtppXIMT
jjh3SsRV4kFNNzv7GE5bdlkmeNlzgCTsqkbwrWDcTSfDXa895y9uk13kojAqdO7J
+s6Zxg4zEbSHmwaaT3hjHz8Bmb3E60pCZGEdlE7KJmgn2gTtSH/ticVZTB98/PAM
KoDVr+NAAB9cdmljQVC36lzrroHiQHdjrF9xos+1ezh50gCkty9z2m0q5WLtrbUA
GrS87oZ8Ckk/42aR6AmPzkfAmG8zOXclLnVzdfDnk5AXq8vzYaSy3k0lX6TYxdbe
m1E0sVJ/xI5uuOUY25D0MdAphxLpgZJTwyqxIhLQBr0cg4M4ino/clDN7ww=
=BnWr
-----END PGP MESSAGE-----`

const yieldKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

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
=clQL
-----END PGP PUBLIC KEY BLOCK-----
`

const spirosSig = `-----BEGIN PGP MESSAGE-----
Comment: GPGTools - https://gpgtools.org

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
=l8kM
-----END PGP MESSAGE-----`

const spirosKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Comment: GPGTools - https://gpgtools.org

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
fIVTBJ4aGN+oB8E=
=JZBP
-----END PGP PUBLIC KEY BLOCK-----`

const silverbaqKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v2

mQENBFUO4h8BCACnk68cbmsKoCSMdZ7Cyd7r8OXpCL8WfWuTvgQjNgPByki0CoKH
9RkNmngtRsF7A3C8MSxClqfMkY0oilSEjbQGiJhFxfFri/Y/0H10NHrhUHgRM0E9
onhv6s6mholgL3op542s/aZ7PnazUfhGcx2u6WVSKDSP1x2OUbkKOAVLvO+eyty3
v0YBVvsX9+JlT1DHo1eH6hbYgK/nbPVaTIjShIragLJ61SDcrerQO1xyWDFH5UUh
cAwFqz4ctESq8OsvH4m/VGrsQNofBsOiYG9CrzTJVkdi4vrLojUatb7e3KtprfA9
8TR8DD5KXQnga6Lw2B4yW1+8qbY9Ne28/IxlABEBAAG0EmZlYmVyLnNtQGdtYWls
LmNvbYkBHAQQAQIABgUCVQ7iHwAKCRCPyPX27j2XtNbAB/9spInm0EduRSSq7QmB
5Pw8UzCph+gjWEWVyuqSbM7kk9PpJUDhAnITGRz99xCSmP8Vz5JqEefGY9I71zbl
iktVctYM+xAsgUUTsTMaI44/YhCFYT1HaHhLaGOluPJjFTbnKqltCsLBDKALa1pO
RRNjrSP7CqPP24c7f8O7bRkWevws6Ovm3++rvANjzudIOfkHPtOTeJxhquIXy+6S
Tnc/0taNZofwzwYURfHUAj+94V43TFV+7okvGecg/riWrktPmROR0OPQnfQXwkGN
NZzGWYl9kzpk682rjrA+8BEM9OmAfxOSnTjvGQM/qI9wTudDTXoPWAtYNNTQh8SJ
P+o1
=HdMW
-----END PGP PUBLIC KEY BLOCK-----
`

const silverbaqSig = `
-----BEGIN PGP MESSAGE-----
Comment: https://keybase.io/download
Version: Keybase Go 1.0.16 (windows)

xA0DAAoBj8j19u49l7QBy+F0AOIAAAAA6XsiYm9keSI6eyJrZXkiOnsiZWxkZXN0
X2tpZCI6IjAxMjBiNDdhMzlkYzBlYmMzNzkyMmJhZmM2NmE4MGYwN2I5MDY0YjUw
YWZjZDAyMWEwYjM1ODdhNzVlY2RhMjI3MGNhMGEiLCJob3N0Ijoia2V5YmFzZS5p
byIsImtpZCI6IjAxMjBiNDdhMzlkYzBlYmMzNzkyMmJhZmM2NmE4MGYwN2I5MDY0
YjUwYWZjZDAyMWEwYjM1ODdhNzVlY2RhMjI3MGNhMGEiLCJ1aWQiOiJlNmFiOGI1
OWIwMjMzODUwMTE1ZjY4YjM3MzRjZDIxOSIsInVzZXJuYW1lIjoic2lsdmVyYmFx
In0sInNpYmtleSI6eyJmaW5nZXJwcmludCI6IjBmNDFiOGRiMmEwNzllZDYyYTEx
YzRkYjhmYzhmNWY2ZWUzZDk3YjQiLCJmdWxsX2hhc2giOiJkNTBiOTc0MjIxMGEw
OWZiZDdhYWE3M2U4NDc2ZjBhYWJkNjA0ZTMxZmUyYmFjMDE4MmZjNDMzMzIwMWM4
YTA3Iiwia2V5X2lkIjoiOEZDOEY1RjZFRTNEOTdCNCIsImtpZCI6IjAxMDEwNGNm
YThmZmVlYjk4NDgxYmI1OTA0NjA0ZGFmM2JkNmQ5ZTA3ODQyOGMyYWM5OWVmODE1
YzZmYzAwYTQ46DNmNzBhIiwicmV2ZXJzZV9zaWciOm51bGx9LCJ0eXBlIjoic2li
a2V5IiwidmVyc2lvbiI6MX0sImNsaWVudCI6eyJuYW1lIjoia2V5YmFzZS5pbyBn
byBjbGllbnQiLCJ2ZXJzaW9uIjoiMS4wLjE2In0sImN0aW1lIjoxNDY4MTU5OTI4
LCJleHBpcmVfaW4iOjUwNDU3NjAwMCwibWVya2xlX3Jvb3QiOnsiY3RpbWUiOjE0
NjgxNTk4NzIsImhhc2giOiIzMGE3NmJlMGQ1OWM4YjE5YWQxMmQ2MjM5OTQ0OTAx
MjI4ODQ4N2YxNjVmYzI1MWJjYmFhMmY0NmHnZTJhMjg2YTE2YzM5ZTA1YzY4MzAz
M2U1ZDcxZWMzMjVmOGMwOTUwMzBkMzY0OTA2NGJmMWJmNTY2MDhhZjM5NWVlZDRi
ZTAiLCJzZXFubyI6NTE2NTAxfSwicHJldiI6IjBjNTMwMDRkZTVhMDcxZGQyZmUw
NTU2MzcyNzk4NzLlY2I0ZjkxNzI3NDg4MTI2ZjAwOTQwODc1YjFhN2IzMGTkOCIs
InNlcW5vIjo1LCJ0YeNnIjoic2lnbuJhdHVy4WUi4H0AwsBcBAABCgAQBQJXgle4
CRCPyPX27j2XtAAA/eIIAAVGimAxSqx1Ei8UQOpWvfgEwpUbUjTuk/tHtOKxvaCl
sQiqdbhdUsZTB2DMTFphMn9a7qPmM0UAmteW/x6812aDZJP+zpoVgwJSd3Jz+1yL
ezls6HGQnwECFaPOX32P8/H1nsLIo59bm+i0XfBqgPedto+LL5jBTuHWy5EtjVGW
e1Tp7960C56eZqD+nOwz6IuLkcvvI92J8XZGteDB0BzIq20BQgOrPxkwFYP2qJg+
pzfKZWj1+3sNbHRgENV0wjUjACPJ1H+amJUSGYdSdONqzwwWUXjFSfUXpqCtKEv0
8UMaL3BkNH6WJrdIcHOyTfRCO1xHvT1Zu9ri/Gls8/U=
=aFCD
-----END PGP MESSAGE-----
`

const subpacket33Key = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQENBFhIScUBCADp3DNt8ENU2U4kuWzdIe6SXQQ9klLyOcm4MzRWJa4kvHjVvaG8
CFHeluTITvXk+HYS06a+w4h7uDj6bQRiQTu+byvj2NuvWjgfiVvkD2BLZ5gsgM7N
IKXlHm+mbTK2FdCCzM7cRiphlySPjL7lflpjOz+iMf2E6phLh6uTsD3js8sxu5Hk
9EJ9sUKDBgLpJf92wL0FaxADP6BLPqn1DzEAe/NE4O6nY6ITqLziNv2UBxdfACe9
06ZPetOJhb/HFHGMkpZNS13BZS60rvMPxtHmzGsTqxP9hdrVqQx/Yi1PrH3UTZRR
/dEty1hHCMbY/gWzDmOxUcV64BxhKztsFxvJABEBAAG0FlRhY28gVGVzdCA8dGFj
b0B0ZS5zdD6JAU4EEwEIADgWIQSagsdLdeNBDBwAM4z0x+j0rBHU8wUCWEhJxQIb
AwULCQgHAgYVCAkKCwIEFgIDAQIeAQIXgAAKCRD0x+j0rBHU8wuvCACWnqhXhMkc
ZxQQ5PebCWt2q1xAEYEbdHWglAnbbsBmJz8aKyfMx5BRVKl3/9efNbBvlM+e3oUo
J40uhTCeGdmmZteM4PUQPCCl+sGYZ4aZuTD8BDnUo79V2ymFMGvYmk8vw79TVjcT
7OGF7jQiJwYycqhOri2FT+jEPSYlLsn7G9PIa40DQcbKQBqSzzt9H5C8hPLQX2b1
OfL+TDv/wa8i6stx/SYn6a1drnFFtyoRTxFwdryPEkRTotS70+Cb4Mg+QH9WgL+D
AppZVBYzZNPWX1bGL5FC7Er/cVeluHEmnfpMURgvtWK87ZW53imbgYixnnbZ2peC
iLgltpS95wFJuQENBFhIScUBCAC5psaWkHr6sJTVEa66t7uWsivvrohoQO1luOjL
QVJcHpGHnSeP8ZFx1B1vseXokMeBDZ9yRowOusiBHve4Tov0bwgp70TCcNZjeB6q
NqtJK7//mPuXBR51JKSlaO9DHi5k3RQ8DjeWm+I5FbDOhfz3vorPcNmrtUZRHjTE
W6Vqaf+Lv6EyLlkD7QRvE/FYbAVQ/+Ht/0pvxeZbVW076YD7tVm+J7AyeMK5u6t9
S7F0K4l/dNpBtoiRToG19BwfsNbjjJh/UDjUGaEgD0gGyhzzrpM2LiwJTmv13yGt
oK8/Wc8hpbT1yp0vSVagJ0FJ50cbqAI8H5LOLHSRroZwpWVtABEBAAGJATYEGAEI
ACAWIQSagsdLdeNBDBwAM4z0x+j0rBHU8wUCWEhJxQIbDAAKCRD0x+j0rBHU8wnB
CACs3EwT0En//ItF/GJdjdrpZyZkcSkOiGFWY/TeTi5SGjAIinJcJWFYDibLKbWx
dbBirDm+ep93VCyGT8CbUjSL5f6zt67tJA4SM+djTW0luCTh8X5RgDntz36tVo6G
pyWSAatyMOmsJqTKYuksIhvkVaJQHLzFBer/5ltzJkTkVZJuuGxqNMP4VUcBXoJG
WIieaZJUW8cNgP3jAR4sa0dSQE+hWBRnsZlWVOdsO1BucoqHz/ytvposCWRPaf83
dDhhp617omZLdCQGFDov0uE0MvIkwQxQIAaQVRfB5ZGH67K3ebyrFxOu6xdoArAF
gdDIgUoBsCGhjerF6Qvcj1dx
=rJOL
-----END PGP PUBLIC KEY BLOCK-----
`

const subpaacket33Sig = `
-----BEGIN PGP MESSAGE-----

owEBUwGs/pANAwAIAfTH6PSsEdTzAcsMYgBYSEwKaGVsbG8KiQEzBAABCAAdFiEE
moLHS3XjQQwcADOM9Mfo9KwR1PMFAlhITAoACgkQ9Mfo9KwR1PNK2AgA4cTQNfnu
F6vfIfmLrnuNm+OFSffkgQmDQc48RV2ppA35r0dDJDi1lF/UBei+RZVNP2zntZ2z
glxivl4qM7qdqojI1HZxP9cT25GuQNWZI3M0LXsderQtv6z4M6q8wj5OjI6kNMN3
QoBLL+cVcqEy0ocW1+oQ4NGemiQ4TLnH3or83OVoUXHSbtM6jBmedqs2taReRx20
RYl5iEr6kQGHmqYt1K142QORPrjYyvGJl8k8cKDjRxOo65ufTB+iztG312cVTnnJ
/gCEVjvAARdrDuroooJwxqx5BFpi7Z5qb4osmxJwdib9lrt1r/fV7QwY1LrCn5eF
YcO0GnR3O1jT9w==
=YRAG
-----END PGP MESSAGE-----
`
const keyForBagSig = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v2.0.10
Comment: https://keybase.io/crypto

xm8EVV9BDxMFK4EEACIDAwRSzcsRfEMpwQ7RvwWHPxXf97lwjf8mqqCeTZXkntaK
LAYBBi4ZH2dPpVL2Nk6Bh0K7Zc9II0ksc0BL+z0fJ/3hIDhq9NgfusjqjiX8NYZ7
dbeT+gyDPp5gzXaxPqrF0vXNBW1heDMywpIEExMKABoFAlVfQQ8CGwEDCwkHAxUK
CAIeAQIXgAIZAQAKCRApXuhUvwxsQUGxAX9Y1nv+tbCxZGE4P22kbVQbi4BTyIkL
YgNKqeoOlNxuWeDdh0xaDsimsTpEKWtSagMBgK179Xs4gsE9jPBtHGuzYpbHbROS
hVCq9ssZjvId45D2UHFSjyAm8spEeFJzLARqEs5WBFVfQQ8SCCqGSM49AwEHAgME
ntY21MrNuVBiA28QJRFx/+nw8O6URDuF7P+a1Ou+c/mzeH8bH9NB+fozm+wt0+kU
sAQ1rrmAdK9oXfbxFHw+VgMBCgnChwQYEwoADwUCVV9BDwUJDwmcAAIbDAAKCRAp
XuhUvwxsQYFyAYCuu59mWoB68DUaSFY1YoIjXt10oKkJyqeaF/MDKCs4RZvnLkcM
Vfm4ANHRx8P4jTwBgKF4bBRJYKbGCh9Sdve3ivaihIjKueXbIkIwzlHKnDhH0ryF
Zy7AuYErrAMqZiEyG85SBFVfQQ8TCCqGSM49AwEHAgME2hMvKEfCqKxsX7+B70Lq
fOOQg3mAP5vNEE9fP/O4CHS3nG7DRv4Di1vZlHE7u2OHAfHSAFq3ir935z4C0d7B
csLAJwQYEwoADwUCVV9BDwUJDwmcAAIbIgBqCRApXuhUvwxsQV8gBBkTCgAGBQJV
X0EPAAoJEGcrs3re2HczDekBAJOT5B0avXmVYnPGkTwVZ3oeoXNwQEpYOm/kdPEx
BFVhAQCKGjWboi7vVftw1cU/IFp1uh7lYWEzDlB9yvUCyd7/INoFAYC9wV1R++yE
roIfq9hkzttJnvmuobSYaJHLMIQXnIADOBLHbHxrKInDtYqIToUoYsoBf2CMDj0r
Vwx1bGn7KprTMnu89hv/rIjnAcDYorsmqfECoXXZNBrVHS/DBpDROPPpag==
=wdBi
-----END PGP PUBLIC KEY BLOCK-----
`

const badSig = `
-----BEGIN PGP MESSAGE-----
Version: Keybase OpenPGP v2.0.10
Comment: https://keybase.io/crypto

yMQhAnicrVU7jONEAM0dHBIrIR06CShXaRCXZeOxPf5EOgknsZPdxInzcbIJ0S1j
e5xMsh5/87GPEy2cREWBkKBBSNAjIdEhPhUFHaK6ElEfBS1O9tA1VzLNaGbePL15
1nv++ZUXCkd35qVvs8d/bLkbv/7CrQvmpXL7QdHynbRYeVBc4cOErxwcJ5cr4hQr
RQYATnYEFvKYlzCwBQkInOjKAnIZmWcEDHnIsjYvYixLgsQyDHZsi8WOjCTedXhR
sBwGFU+KLqFzHAURoUlOK2FLhLzNWi4rMCKLGRlK0MWYlSHGEuQtl7EFmwf5xYUf
72/k4iwU41Pi53v54vIg7zn4/1n3+kAn8jYn2a5oY0vmRRcKHMvw0JFsBB1eAvIe
GOOIIg/naA/tOLb48KQYE+upqc9UuVBE2OWB6LqAs5BrQwe7ruxYEmsBKOXM2AEy
y7D5ieQyyIKOCCwRW5bA5Tj2oCrCGxzF+DIm85z27f2oqo2zzrHRMI51dTBQGuph
d0ZHOZD4tHLcunbwuBtguodt2FPmFDAzWvM9D9OkcrxIkiCulMvPvC7bURok/ozO
aKprE4USO5qeiQj0q5rubFxVG899zZ+Mm4tJbdzOjPUuqDbKczVQY0MRlpkRdaZr
0eM2m5zBR71d3G7UYb0xWqnxTrXUga/76nxORkG7kW76rLaO7KE9bDCe19BLTrm8
dpue22FFWh/P6NghUuxiUxRFRpeTDeezGSpTrXRxsVp4aid3sa+vtpuGMvA9LoW1
FmnqtaZeVVdaVRCqM4pb3vm05iM00NmmGvcutHq47p1tx17Y0nvDTUvpLzuTUrVV
WtWSkLQnV03W7wejvrNajbegm7uFhYCAPun3VMs3TdVXJwrSNWenQWF41V6I46Yz
Ij1+yq55U8mG422Uthy+LBulKc5SMKOmvM1EZKS8srswzUVD7ztVnMY1IA2Y7SIM
mRAqZlie1FDH9/lNuF3kjx3II00PalAdajM6OCcZM4kVfRrNO82RKnR7+rxPWsiO
fNJihr1+g27r9Hzlw244CPkaMDt4xDEmkuz6FgW5k5bOoiB25W5KRleqPpB7yxb2
LH9lLqnWaJyPESGskrZIdEYAqsqMdLGcnmf6vLHt6ECZ0RZJz5farpn2zGWWxPMz
hQyN1XYbmmQ7aWWjZjDvKaHb18nZWatjEhDVkJh2+2ikdM2J5OWvSNZubzNts/w2
6lUTq8PLqxrkYraj1cXh1KMOhUa2EUStzYYydUpiHfBJVo+wZCQrWZrR9jBKXSqk
67KRCAHVciHZRbldHRqTJeY4bLgl2DE2YakbBoFtRrQEE7oEcsk1/FLIcTO6TLFB
kwzW+5NEWDP3ZvSeCna5QYf0qJ36cxK1T3aSBvukPw34SXFznbJiBeSHdkL2PQB4
juUYgeHgSRHvAhLhS7JHMNfjpBjkMc5JMEZsXjiCyAhAcHnJFXjIActGDsQOlPN2
hJwgAgdhGeUfEHMQQibvMQ5AgEVx3wcxDi+vJYHDgvrFCpurRPODyDlFyTrCxYc/
3n+xcOeo8NKtm/vqLxy9fPu/H8JPb94ofCxnDz599/PLQj6+Pvrqw2/eWP3wW124
+/j9t+7yf43jG4Xvgn/okz/fe/L9l+88un/zM+x9cvX6R68+qvx967XTD34nX/wL
bL0JeQ==
=965F
-----END PGP MESSAGE-----`

const bengtKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Charset: UTF-8

xv8AAABSBFOPJKkTCCqGSM49AwEHAgMEyzUHBzhmkM0aV+7cUoLOD39xXR5J1ryY
YCSdFOcJOtO0piEBUX1Qr88Z8uUULADS93INugEUNqSxOGWuAu4WTM3/AAAAGDxi
ZW5ndC5sdWVlcnNAZ21haWwuY29tPsL/AAAAZgQQEwgAGP8AAAAFglOPJKn/AAAA
CZBcFMyn0JbJmAAAreQBAL/ipjkc+3WA3zaJQVImdGnlTL6ukN1CM6PPqWEvLFPA
AQCkJfkosi8b2MVGgAck3OZTxPuF6j414eJN51AvtpzXZc7/AAAAVgRTjySpEggq
hkjOPQMBBwIDBOKJkCiqTlPCPbnIKfFPUB3zoeGVwYLXar5bROtJRVO+8Ie2UEgt
XP0+aKQ/7EQWrrqFM9b9Xo/3YUSiQCixb10DAQgHwv8AAABmBBgTCAAY/wAAAAWC
U48kqf8AAAAJkFwUzKfQlsmYAAAXLgEAj18rfkwfFBkUhbZe4sU3QE+sgVpfCBVr
AfiqBnIZ5l0BALIoQEboLEbb3cj1SqTBfXytnnojPuRDES6M9qzulc9c
=WJdL
-----END PGP PUBLIC KEY BLOCK-----`

const bengtSig = `-----BEGIN PGP MESSAGE-----
Version: Keybase OpenPGP v2.0.62
Comment: https://keybase.io/crypto

yMP8AniczVa/i1xVFJ74ixgRhUjASnmdsIZz7u8b7PwLLAKCxvHec8/dfdndmcmb
t0mWsKWFlSAisRE7sRGElKZwiRBRIQErCwsLsTVIwELwvJl1sxsjSojgFMPMu+99
9zvf953z7rUnHx4dO/76M998fOP9ry4f+fr69a3RqxePTi41eVq2m1OXGtpoedIP
vyZpk5tTzTpv5zTnk+30+Qucm5XmPHfzdjqRJUfaGYzJkPWQfQ2UC2lwNsaAYLVG
qJRibHZWBpgBlDcKz/vxelvkeUDUKesSNDpdCsfiNRWHylauESjEnBSyMybp4mvE
AEX+aQ4lsOAnbw0koVTbySp3s64dmDfGUSRllGMIygfmkE0olSyhIUq+QJQ7YpAH
16bz/lCNzYLpeEHvHvc/YN5bC7iiGKrSgY3TAZWz2UbISmGI0WSA4cY5d3t+ZJ6s
9oOifZdofdBUqLc0H361ZbzvTlxpNpJILddoLYk+zSmUDZz3EPUhRAE6386pHUAX
hJABo3ZVARQfkg4p2BoCmZyK9xibfTsHrQ6L75PJnkpS3hBHLNoVG4DIZ58dh6SC
V9WSOyAmoDXkaw7RJZStCvtqREBPhiGTKVzYIPmi2eTiTEWXUXENoj2FyqBETOE+
W52NhZAo8dr/itmZlabjzWnP41k3ndYlP+rbzaUlKoKPoFYa2uo62TYVdME5iVAh
yUy0jJYsZ53EPIeaqnDWTLp6CE6MMQUDJgeYcspuiCnvgTsFNizBZx2fF/AQkkUw
KKFG0SChomQrArA2IWhyIZeaBD/5UE2IJDmQTGsn0lTGof/3ahkEXtSzmBprTOvj
kvo0Pjsf0rc/PvoLbd9z1/xt4hYY4357JgvCc96nfqAuK3M+N5kuL7ary5ZkiugM
B7EgEJPjDMk74eu09A8ZQme5oMcqXiWOQTR0jIFKllqtg9os0fb221k56MSgb4i4
74QibV321jmlcgBnPICqSpGIFNBWUTY6V71W8mDSPlsbrcy7WKFIQOiQEyhYC/A9
J3wl5whsCuKkZCYZk1E6P1ewphRvtMxSaxSSluzpHLMUX1KKSkf2i2n0L50o083U
TvZlf5HakzTdbBbS91OabgwzaDK/2wwJBd7TDzhgiM6ijPKJhDNR9Qw+R1sLuEAk
LUPJofesdDU2QTQqG4Ac0RaQcgL8oyEe8E5rVIAQtQf2SsUcsHpUFSJKio2yubKW
dcy1IivFnpM0vJUWsKpopS3cbcgSfM+QB+D2/Rsy7Vbv3xA82CGQnZVIyQgymUUe
TOJHiFWEwsDK2xxFtliMLMhAj9YYUbpkGWPyxsl/MeTMYp9xn9qNgfssbW9MUxmv
pfnag3HkP6piZ3hFLmrYe1MeOLkM4u2nLMhJQb6dZOPirO14PFiDVkYByOfO5Ezg
bTAxUyiEoSDrzDIl0acsxwEni6R0yS5Byhkc5FAigk5yUarAO5Vqef32aVUwpd5J
6rc6bnZ233hkdPzY6LFHHxrOY6Njjz/95yntlyuj35+rR2+9+93NsHXi7dtvfnjp
7O47H/x45bMX9E+/vvfRK7e+vH5k9P21Zz+//cPl7rcb3z7R37i6e/rm1eaL0yd2
P3n505+feukt+gPf/gF6
=oTR6
-----END PGP MESSAGE-----`

// Minimal repro for a key that had an encryption subkey which had two
// signatures, the first one was expired and the second one was valid.
// Correct behavior is to find the 'valid' signature even if it's not
// first. Related to https://github.com/keybase/keybase-issues/issues/2604
const matthiasuKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v2.0.66
Comment: https://keybase.io/crypto

xjMEWNDnWhYJKwYBBAHaRw8BAQdA8PVaNyG9OcHWwknOwFQ83rX/UioADnQCXEnb
a6zJU5nNGU1yIFRlc3QgPHRlc3RAa2V5YmFzZS5pbz7CdgQTFgoAHgUCWNDnWgIb
AwMLCQcDFQoIAh4BAheAAxYCAQIZAQAKCRBuzc94y89AjJuZAQBRT8F4KD+VxuBk
a5o4CiNpNbkDzUbq/H1o5Q/pbGPBjwEAwTvy+PN2lOCEbiGg1tEVdbVzR+xuPXAi
sppnFBlc/AvOOARY0OdaEgorBgEEAZdVAQUBAQdAqJuojJywPAdEYL0fGWsT8ZYk
vCeffralNVsbBIm+zWADAQoJwmcEGBYKAA8FAjo1aoAFCQ8JnAACGwgACgkQbs3P
eMvPQIx8PAEAeLkAm3PAty0++zEpC4RoppEjxdfYymxzvIYAcwlCs/YBAI94Qzpj
XA/h4QtRlXR4EfX/43opwPYnT/WzImlXzYYJwmcEGBYKAA8FAljQ51oFCQ8JnAAC
GwgACgkQbs3PeMvPQIwf4QEAFfAR5rdFl2bj2UqhW7S2UL7eb7sgdibqXU/a66hL
HMgBAPaACKEEt5+mQvLioH2wDPn2Wm2oPd+7XeuGB+ex8JwD
=vkDo
-----END PGP PUBLIC KEY BLOCK-----
`

// Keys with ill-formed armors - either without CRC or without proper
// newlines.

const noNewLinesKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

` +
	"mDMEWQoA2RYJKwYBBAHaRw8BAQdAQkGkdMrckk67dh9MKxeCa8PJsTnMJ+oDgCJ9" +
	"7gvM60G0E1JFQUQgVEVTVCBHTy1DUllQVE+IeQQTFggAIQUCWQoA2QIbAwULCQgH" +
	"AgYVCAkKCwIEFgIDAQIeAQIXgAAKCRBEkPOnE4+R/RiKAQDvIhM74qEtKSbTtu50" +
	"mMB49eTAMg/MogyFA8SUCbStPAEAxdxSKX1hFYFP4N8ML8BgLOJG4PXAdQ8wnfXD" +
	`vJxN/AQ=
=apgr
-----END PGP PUBLIC KEY BLOCK-----`

const spacesInsteadOfNewlinesKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

` +
	"mDMEWQoA2RYJKwYBBAHaRw8BAQdAQkGkdMrckk67dh9MKxeCa8PJsTnMJ+oDgCJ9 " +
	"7gvM60G0E1JFQUQgVEVTVCBHTy1DUllQVE+IeQQTFggAIQUCWQoA2QIbAwULCQgH " +
	"AgYVCAkKCwIEFgIDAQIeAQIXgAAKCRBEkPOnE4+R/RiKAQDvIhM74qEtKSbTtu50 " +
	"mMB49eTAMg/MogyFA8SUCbStPAEAxdxSKX1hFYFP4N8ML8BgLOJG4PXAdQ8wnfXD " +
	`vJxN/AQ=
=apgr
-----END PGP PUBLIC KEY BLOCK-----`

const noNewlinesKey2 = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1.4.14 (GNU/Linux)

` +
	"mQENBFM2jy4BCAC41SHJRtrpz0BNJmij7cnXAWow1Vm1XwNLXPGFutMgu+V2sW23" +
	"lSumQdtX9dBjl5+Cg4tKG9g1dVq6uwT3+YeNkm++V7v/atzZDdkDFbTW4C3KGEE+" +
	"lk30CWdBQbls+acpUdK7ESJxcvbR7/HCbOUy+0+58LEz24TzsHTvGCWbG2AQ0ZkK" +
	"UxB6vLfFcQSa60Z57mHZAZmg0N+96P7Kx7RAFivRFrrUpgWny76dbo3hsuEn0bzv" +
	"PcRlNNJKTfkvQz8naQbsZBiT/gOjwsy3W2bGjYB3YTRlq7QNT5em8jFuYFXB74Pl" +
	"QN7nkYHZfNuLAAE19kOxeeOisAtOqhpq0N0XABEBAAG0H1J5YW4gS29pcyA8cnlh" +
	"bi5rb2lzQGdtYWlsLmNvbT6JATgEEwECACIFAlM2jy4CGwMGCwkIBwMCBhUIAgkK" +
	"CwQWAgMBAh4BAheAAAoJEC294S54YdOEFlEH/3vPylm1ofhY23mrGR2C3mvSBkzd" +
	"Aq5lvivxSx/N55N7Y4ZdALe6TCQyQSwSVsTGw20fXuYvOCJ52jOPQIti9XjDXm6P" +
	"xTsEmDBl1M/BZXiXSqoOxMn0KdrqQs5HANIkksauNh1AJZDaPzB5MoO+hb5AVEE4" +
	"ufx+BvBWN8n3C+BKiKdhcb9r9FpaIcsHB6lmyJ+ajP/5Hbp/XQ4ZhndMuuroSEC9" +
	"rmjUB5paPNyMs5ZXBNJ1YU/6K2sLYdlQZEI1hN14BuzaEvhk92N30On38OEDTDYs" +
	"E+N47fJZhhgG6Hemi8icr5Y/XcZwENfo9gCLYftFt7nvHVRVcKaIpdSQRFi5AQ0E" +
	"UzaPLgEIAK2o0vmXmAaZUH9VjE43GkTIG7favZswBIsDKG1HFni1VMS8mBsbLD9+" +
	"YgtsBJdlbmyEmcB8xzgatXrTFRT1JAwDfSzKEtkx40TclM9QZ4D39OWpg5VRcpPD" +
	"4G4YripIKoIx2ICsAMewqVuaqXsjM7+Piya7wAFBqza55Wnp66RSimPSjzojhn0I" +
	"5JbVJ01KXDJ8jMBoyCxr8noNRt9sF+xhPd/4RcfG+B3G2rQ/07sflyCDWsEZHzc9" +
	"Rgnq4yu3/Jagf8kC8S1RcFSocGDaQtwnrJRaVoKo4708YB07hMVk66DETIQG3YUs" +
	"zr1F7Iot3wr5sW9YeyN47UqekJO/hDsAEQEAAYkBHwQYAQIACQUCUzaPLgIbDAAK" +
	"CRAtveEueGHThAh1CACBMdAg+Gu1MdGkcAZpSeUeRZ53RfYRVNYLFSaRi7elOSZv" +
	"f0RO+7+lbIZojYmCju8jp7J0e9htdo/10uQsgaMpAPSZVG6ZUydkUWOgVnsNNrxZ" +
	"rZMxH35AHU8x8VeZZzzgNxBWybvEJouBwBRUdqXC1yFxcNBEO+QMuLPlH+Sn8kLN" +
	"mB2j/tiFP/L7bIvyYTLPRa+4/8zwW6XWXpCSuDTp653K7PkYmgkU64ctbJFr91AC" +
	"U/lDgsMdjZlLGjvVwQNvh8uQcO2hJzA0MjvUTKmKg+qwe5nVYJ7Sm+n2RwTfmIOZ" +
	`i6OWaPnUwuDBWTY59LkEeAKnilLJfpPQCK7fANEe
=O610
-----END PGP PUBLIC KEY BLOCK-----`

const noCRCKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWQoA2RYJKwYBBAHaRw8BAQdAQkGkdMrckk67dh9MKxeCa8PJsTnMJ+oDgCJ9
7gvM60G0E1JFQUQgVEVTVCBHTy1DUllQVE+IeQQTFggAIQUCWQoA2QIbAwULCQgH
AgYVCAkKCwIEFgIDAQIeAQIXgAAKCRBEkPOnE4+R/RiKAQDvIhM74qEtKSbTtu50
mMB49eTAMg/MogyFA8SUCbStPAEAxdxSKX1hFYFP4N8ML8BgLOJG4PXAdQ8wnfXD
vJxN/AQ=
-----END PGP PUBLIC KEY BLOCK-----
`
