package openpgp

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/keybase/go-crypto/openpgp/armor"
)

func TestMultisig(t *testing.T) {
	kring1, err := ReadArmoredKeyRing(bytes.NewBufferString(testKey1))
	if err != nil {
		t.Error(err)
		return
	}

	kring2, err := ReadArmoredKeyRing(bytes.NewBufferString(testKey2))
	if err != nil {
		t.Error(err)
		return
	}

	if len(kring1) != 1 && len(kring2) != 1 {
		t.Fatalf("Expected both keyrings to only have one key each: len 1: %d len 2: %d", len(kring1), len(kring2))
	}

	tryWithKey := func(keys EntityList) error {
		sig, err := armor.Decode(strings.NewReader(testSignature))
		if err != nil {
			return err
		}

		md, err := ReadMessage(sig.Body, keys, nil, nil)
		if err != nil {
			return err
		}

		if !md.MultiSig {
			return fmt.Errorf("Expected MultiSig to be true")
		}

		if md.SignedByKeyId == 0 {
			return fmt.Errorf("SignedByKeyId should never be zero, even if keyring doesn't contain good key")
		}

		if md.SignedBy == nil {
			return fmt.Errorf("Couldn't find key message was signed with or message wasn't signed (md is %+v)", md)
		}

		if md.SignedBy.PublicKey != keys[0].PrimaryKey {
			return fmt.Errorf("Message wasn't by expected key (SignedBy is %p)", md.SignedBy)
		}

		_, err = ioutil.ReadAll(md.UnverifiedBody)
		if err != nil {
			return err
		}

		if md.SignatureError != nil {
			return fmt.Errorf("md.SignatureError: %s", md.SignatureError)
		}

		if md.Signature == nil && md.SignatureV3 == nil {
			return fmt.Errorf("Signature is nil after reading (md is %+v)", md)
		}

		return nil
	}

	badkey, err := ReadArmoredKeyRing(bytes.NewBufferString(matthiasuKey))
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Trying keyring 1")
	if err := tryWithKey(kring1); err != nil {
		t.Error(err)
	}

	t.Logf("Trying keyring 2")
	if err := tryWithKey(kring2); err != nil {
		t.Error(err)
	}

	t.Logf("Trying entirely unrelated key from different test file")
	err = tryWithKey(badkey)
	if err == nil {
		t.Error("Expected an error but got nil when trying unrelated key.")
	}
	t.Logf("When trying with bad key, error was: %s", err)
}

func TestMultisigMalformed(t *testing.T) {
	keys, err := ReadArmoredKeyRing(bytes.NewBufferString(testKey2))
	if err != nil {
		t.Fatal(err)
		return
	}

	sig, err := armor.Decode(strings.NewReader(testSignatureMalformed))
	if err != nil {
		t.Fatalf("Got error during decode: %v", err)
	}

	md, err := ReadMessage(sig.Body, keys, nil, nil)
	if err != nil {
		t.Fatalf("Got error during ReadMessage: %v", err)
	}

	if !md.MultiSig {
		t.Fatal("Expected MultiSig to be true")
	}

	_, err = ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Fatalf("Got error during ReadAll: %v", err)
	}

	if md.SignatureError == nil || md.Signature != nil {
		t.Fatalf("Expected SignatureError, got nil")
	}

	t.Logf("SignatureError is: %v", md.SignatureError)
}

const testKey1 = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWfcV/BYJKwYBBAHaRw8BAQdAYhAuI4LPgxnu8MDU/XJpSlfCFPelz58v5QpU
9R9MtFa0ClRlc3QgS2V5IDGIeQQTFggAIQUCWfcV/AIbAwULCQgHAgYVCAkKCwIE
FgIDAQIeAQIXgAAKCRCc9gMcYqzhOWz5AP91QOM2xFiy5FZ+suqpP5zbygMNe/PJ
wunDkjryQRaWqgEAjalogSO20NeTEbDBWiglggMvJrTFXMqdsZ5bdUkoSQi4OARZ
9xX8EgorBgEEAZdVAQUBAQdAIXy/6CNrP/Tq6uDnTu9Vra8Qc05uGY18gUqou9/0
m1wDAQgHiGEEGBYIAAkFAln3FfwCGwwACgkQnPYDHGKs4TkTYwEA9JWX8sASAY6u
NSuMuq3f3fKwYVR3kB0hYRd7ffic+aABALBDGedfGTfjKLWAqd+NFO4fKlQJjg0Y
+EnrmcTzas4G
=JK7o
-----END PGP PUBLIC KEY BLOCK-----
`

const testKey2 = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWfcWjxYJKwYBBAHaRw8BAQdAxsmWcp2FwAiRHylbOrDnoKKUBAa1wgQlE1mJ
fNj4EFS0ClRlc3QgS2V5IDKIeQQTFggAIQUCWfcWjwIbAwULCQgHAgYVCAkKCwIE
FgIDAQIeAQIXgAAKCRCIHHukx/1YbM3dAQDNiAF2ZqxrDvxv5chMeazuvsu9o5J8
mtpPludqpWKsvAD6AsH0fhDeIwKVBk1uigw3ut7VKyyNSSNezy3RczengQy4OARZ
9xaPEgorBgEEAZdVAQUBAQdAF1hhJLcRj77GF+lc9gEVziFZ1yJW/8LYSMZ0AAo9
kkgDAQgHiGEEGBYIAAkFAln3Fo8CGwwACgkQiBx7pMf9WGxlJgD/RriX0jfA3Hjl
pSCtbGRJGm6LZgYEn9XHzfmZ+ZTG9bsA/AxYjMrv4I3Ft4x6ogrqzvxmcga3zgGc
QjcG/YKbNUQJ
=/+FB
-----END PGP PUBLIC KEY BLOCK-----
`

const testSignature = `-----BEGIN PGP MESSAGE-----

owGbwMvMwCHWIVO95PjfiByGCWDunG/MMklrHloynhZLYoj8bmpdgQq4OuJYGMQ4
GNhYmUCyDFycAjAtVeyMDAsPbJoqV7fr73PmkwouPLvPX02Oeye648Tq72GOz+68
lTdiZLigI+QhMS++YiL/9uLLKyZmlnX3iZ89rLYhTKZe2b5JHIv5MBeCzP9zUvnk
xI4vmaJe3MJav1zjZZc67ZSJYVy9+Q33resy8X0M/4Mmhrs7zvyssnTjiTOTPzus
uXp/8adNSepq6kncd9iiU5kB
=QVa3
-----END PGP MESSAGE-----
`

// This signature payload ends prematurely, it claims to be signed by
// both testKey1 (9CF6031C62ACE139) and testKey2 (881C7BA4C7FD586C)
// but it only contains `signature packet` for 9CF6031C62ACE139. If we
// test it with 881C7BA4C7FD586C, it should fail to read.
const testSignatureMalformed = `-----BEGIN PGP MESSAGE-----

kA0DAAgWiBx7pMf9WGwAkA0DAAgWnPYDHGKs4TkByxZiAFn3NTt4eHh4eHh4eHh4
eHh4eHgKiF4EABYIAAYFAln3NTsACgkQnPYDHGKs4Tl6BwEAocCylR5+uv3nA8kg
RAy7z9VjXu4VuMir91ZB5tztHzIBANAsEkgYnl94kQ+3c9OokWl2i44XzcMmsFYc
fyM/ghcK
=R0Rj
-----END PGP ARMORED FILE-----
`
