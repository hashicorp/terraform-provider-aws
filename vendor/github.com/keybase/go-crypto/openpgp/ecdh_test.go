package openpgp

import (
	"bytes"
	"github.com/keybase/go-crypto/openpgp/armor"
	"github.com/keybase/go-crypto/openpgp/packet"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

const privKey = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lNIEV+7iWhMFK4EEACIDAwSxj2UpOqGEAUZQc43HIoE2htc9+5nOePeDHqJi5czo
ecYAS5liyPFAJ3NIAqRh7UJ6pfgoz/mjvgH2fn6YLWv15hKWuSNxa0+RbWvU3lTi
nB/aBgkqTIQkJlkH/IP2Om/+BwMC1yIjA1zY/tDrYXLaotDLdkh17bvPJ5mlEqTK
dpwqgYJvg6z0V9scUM3axHBvOzFgapRP3yEgRdE9T1bS/Uq5CQRWe/AV3kVjNQta
D737y28b4XHQskb9yk2VRrZTvAO0Jk1heCBFQ0MgMzg0IDx0aGVtYXgrZWNjLTM4
NEBnbWFpbC5jb20+iJkEExMJACEFAlfu4loCGwMFCwkIBwIGFQgJCgsCBBYCAwEC
HgECF4AACgkQ5ynPBjIBakqqUQF/dAiY/YEKrGxfEiXlM0PkIPX7l2usSNsTCg4K
GY6nZfDOsqlotBqGKDHOAT3Og83TAX9SG7qG7vQvjrkR2VjnG5J9tXY8+ZotC2bW
yJlOjLm47Is58ehoWbIxOORpaBzoo+Gc1gRX7uJaEgUrgQQAIgMDBLaJ+2BLw65B
8ApW5hZ1AbiPCrXfG+ADBg3mdmKK419qxN4gFdh96+HoRlyqmnK743zLYaEYs8mF
S3cIiDQYCTJ3VeyTXEqk8vWCXBsXXzOtwtcg45+b1qNTOBjiOad39wMBCQn+BwMC
G8W4E8jtuznrjyFq9Gr1klnNQuh19pfecveH5oVle/tVCZzfLYctq5vxbIiHduzT
Pnf42FXz4iyVn0zb3bjgfDsMRcHQpWnhidBNQiw351Z4qIQCvocZIgN8Vv/iVBiI
gQQYEwkACQUCV+7iWgIbDAAKCRDnKc8GMgFqSlzYAX9XQX/GynVsIWV9ju7Gvs+z
GIxAujUbJBdFBLivrGNoM/+4OEufal/lbL0uqWIWXsIBgL/Fr+IDnXs+nyC9CXoE
siHEVKaKfz1oilbbrAma6U/in5NXDVtuYMbxxGPbIRiREg==
=aFgY
-----END PGP PRIVATE KEY BLOCK-----`

const pubKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mG8EV+7iWhMFK4EEACIDAwSxj2UpOqGEAUZQc43HIoE2htc9+5nOePeDHqJi5czo
ecYAS5liyPFAJ3NIAqRh7UJ6pfgoz/mjvgH2fn6YLWv15hKWuSNxa0+RbWvU3lTi
nB/aBgkqTIQkJlkH/IP2Om+0Jk1heCBFQ0MgMzg0IDx0aGVtYXgrZWNjLTM4NEBn
bWFpbC5jb20+iJkEExMJACEFAlfu4loCGwMFCwkIBwIGFQgJCgsCBBYCAwECHgEC
F4AACgkQ5ynPBjIBakqqUQF/dAiY/YEKrGxfEiXlM0PkIPX7l2usSNsTCg4KGY6n
ZfDOsqlotBqGKDHOAT3Og83TAX9SG7qG7vQvjrkR2VjnG5J9tXY8+ZotC2bWyJlO
jLm47Is58ehoWbIxOORpaBzoo+G4cwRX7uJaEgUrgQQAIgMDBLaJ+2BLw65B8ApW
5hZ1AbiPCrXfG+ADBg3mdmKK419qxN4gFdh96+HoRlyqmnK743zLYaEYs8mFS3cI
iDQYCTJ3VeyTXEqk8vWCXBsXXzOtwtcg45+b1qNTOBjiOad39wMBCQmIgQQYEwkA
CQUCV+7iWgIbDAAKCRDnKc8GMgFqSlzYAX9XQX/GynVsIWV9ju7Gvs+zGIxAujUb
JBdFBLivrGNoM/+4OEufal/lbL0uqWIWXsIBgL/Fr+IDnXs+nyC9CXoEsiHEVKaK
fz1oilbbrAma6U/in5NXDVtuYMbxxGPbIRiREg==
=tfMD
-----END PGP PUBLIC KEY BLOCK-----`

const gpgEncryption = `-----BEGIN PGP MESSAGE-----

hJ4DzvP5Ex/d4TYSAwMEpjdfscEAp/NyYwViM2H6dPUr8vLA1fJ8pLefQi9u8pRU
JYnAzt3rf1NflTv/bHGuLxXvM+g8DvqT9yMHbTszmM40ghDgbfESCRT2w0SY6dnZ
1IadR8JH4lQEnG76EnJZMA1wq5TFcQ7/F8V+rJlpfBJ09PTFOZIq4eWG3Ql3ciLW
UNc5HhvHycU8U7ZohrXQs9JIAZ/QiU0irj8G2yAoOMGi/XVz3qyz4ZwtxhTHfMlI
NfBc9h72rI/hIjOdSM8ClO2ijOShevljVrd8YOxnTeJgVwtwFd3S9IA1
=KFaW
-----END PGP MESSAGE-----`

const passphrase = `abcd`

const decryption = "test message\n"

func openAndDecryptKey(t *testing.T, key string, passphrase string) EntityList {
	entities, err := ReadArmoredKeyRing(strings.NewReader(key))
	if err != nil {
		t.Fatalf("error opening keys: %v", err)
	}
	if len(entities) != 1 {
		t.Fatal("expected only 1 key")
	}
	k := entities[0]
	unlocker := func(k *packet.PrivateKey) {
		if !k.Encrypted {
			t.Fatal("expected a locked key")
		}
		err := k.Decrypt([]byte(passphrase))
		if err != nil {
			t.Fatalf("failed to unlock key: %s", err)
		}
	}
	unlocker(k.PrivateKey)
	for _, subkey := range k.Subkeys {
		unlocker(subkey.PrivateKey)
	}
	return entities
}

func TestECDHDecryption(t *testing.T) {
	keys := openAndDecryptKey(t, privKey, passphrase)
	b, err := armor.Decode(strings.NewReader(gpgEncryption))
	if err != nil {
		t.Fatal(err)
	}
	source := b.Body
	md, err := ReadMessage(source, keys, nil, nil)
	if err != nil {
		t.Fatalf("failed to read msg: %s", err)
	}
	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Errorf("error reading UnverifiedBody: %s", err)
	}
	if string(contents) != decryption {
		t.Errorf("bad UnverifiedBody got:\"%s\" want:\"%s\"", string(contents), decryption)
	}
}

func ecdhRoundtrip(t *testing.T, privKey string) {
	entities, err := ReadArmoredKeyRing(strings.NewReader(privKey))
	if err != nil {
		t.Fatalf("error opening keys: %v", err)
	}
	if len(entities) != 1 {
		t.Fatal("expected only 1 key")
	}
	if !entities[0].Subkeys[0].PublicKey.PubKeyAlgo.CanEncrypt() {
		t.Fatal("key cannot encrypt")
	}
	buf := new(bytes.Buffer)
	armored, err := armor.Encode(buf, "PGP MESSAGE", nil)
	writer, err := Encrypt(armored, entities[:1], nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to Encrypt: %s", err)
	}
	msgstr := "Hello Elliptic Curve Cryptography World."
	io.Copy(writer, bytes.NewBufferString(msgstr))
	writer.Close()
	armored.Close()

	block, err := armor.Decode(bytes.NewBuffer(buf.Bytes()))
	md, err := ReadMessage(block.Body, entities, nil, nil)
	if err != nil {
		t.Fatalf("Failed to decrypt: %s", err)
	}

	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if string(contents) != msgstr {
		t.Errorf("bad UnverifiedBody got:\"%s\" want:\"%s\"", string(contents), msgstr)
	}
}

func TestECDHRoundTrip(t *testing.T) {
	ecdhRoundtrip(t, privKey384)
	ecdhRoundtrip(t, privKey521)
}

func TestECDHRoundTripCv25519(t *testing.T) {
	ecdhRoundtrip(t, privKeyCv25519)
}

func TestInvalid(t *testing.T) {
	testDecrypt := func(priv_key, payload string) {
		entities, err := ReadArmoredKeyRing(strings.NewReader(priv_key))
		block, err := armor.Decode(strings.NewReader(payload))
		_, err = ReadMessage(block.Body, entities, nil, nil)
		if err == nil {
			t.Fatalf("Should fail with error.")
		}
	}

	testDecrypt(privKey384, payloadInvalidPadding)
	testDecrypt(privKey384, payloadInvalidPadding2)
	testDecrypt(privKey384, payloadInvalidKDFParams)

	testDecrypt(privKeyCv25519, bad25519_1)
	testDecrypt(privKeyCv25519, bad25519_2)
}

// TODO(zapu) - the effort in being compatible with GnuPG here has
// been pushed back a bit. When using elliptic.Unmarshal, we are
// strict on the byte level, not bit level, so anything longer on
// shorter than curve's expected byte-length will be rejected.

/*
func TestLongCoords(t *testing.T) {
	entities, err := ReadArmoredKeyRing(strings.NewReader(privKey521))
	block, err := armor.Decode(strings.NewReader(payload521longMPIs))
	md, err := ReadMessage(block.Body, entities, nil, nil)
	if err != nil {
		t.Fatalf("Failed to ReadMessage.")
	}

	expected := "purpleschala"
	contents, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		t.Fatalf("Failed to ReadAll")
	}

	if string(contents) != expected {
		t.Errorf("bad UnverifiedBody got:\"%s\" want:\"%s\"", string(contents), expected)
	}
}
*/

func TestImports(t *testing.T) {
	entities, err := ReadArmoredKeyRing(strings.NewReader(pub25519kbpgp))
	if err != nil {
		t.Fatalf("error opening keys: %v", err)
	}
	if len(entities) != 1 {
		t.Fatal("expected only 1 key")
	}
	if !entities[0].Subkeys[0].PublicKey.PubKeyAlgo.CanEncrypt() {
		t.Fatal("key cannot encrypt")
	}
}

func TestInvalidEddsaSignatureImport(t *testing.T) {
	// ECDH Cv25519 implementation should also have a working EdDSA
	// implementation for key signatures. This test tries to import
	// badly signed keys.

	// This one has a bad self-signature. Reading should fail.
	_, err := ReadArmoredKeyRing(strings.NewReader(pub25519invalidSig))
	if err == nil {
		t.Fatalf("reading invalid key should fail with error")
	}

	// Second one has properly self-signed primary key, but a subkey
	// with a bad signature.
	entities, err := ReadArmoredKeyRing(strings.NewReader(pub25519invalidSig2))
	if len(entities) != 1 {
		t.Fatal("Expected one key")
	}
	if len(entities[0].Subkeys) != 0 {
		t.Fatal("Expected 0 subkeys")
	}
	if len(entities[0].BadSubkeys) != 1 {
		t.Fatal("Expected 1 bad subkey")
	}
}

func TestECDHBitLengths(t *testing.T) {
	readAndGetBitLength := func(armored string) (uint16) {
		entities, err := ReadArmoredKeyRing(strings.NewReader(armored))
		if err  != nil {
			t.Fatalf("Error in ReadArmoredKeyRing: %v", err)
		}
		bitLen, err := entities[0].PrimaryKey.BitLength()
		if err != nil {
			t.Fatalf("Error in BitLength(): %v", err)
		}
		return bitLen
	}

	if bLen := readAndGetBitLength(privKey521); bLen != 521 {
		t.Fatalf("Got BitLength %d, expected 521", bLen)
	}
	if bLen := readAndGetBitLength(privKey384); bLen != 384 {
		t.Fatalf("Got BitLength %d, expected 384", bLen)
	}
	if bLen := readAndGetBitLength(privKeyCv25519); bLen != 256 {
		t.Fatalf("Got BitLength %d, expected 256", bLen)
	}
}

const privKey521 = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lNkEWAJ/HhMFK4EEACMEIwQBX1achVr3ad6/1AYQM0Xpb0yOch0Va2+d1WjAi/TU
lVMYFq3Sv1HRgwz87iaEGv2lViKTZ2Zbqh68ndyBoAY9CpQAzHrEnFozvQBQxSHe
JaWxdiJIF3ZtLRrxMm+SBSKcQge2TwXmFr/coEKU3uS6PNHz9/1qKvOflbLwgiP6
PWt01HYAAgUeo/x+60pfXvYBT/YwzYtEpMgY3ahEM64gNzCSwbggGdCK02H53Rir
hQc4NHL/N/dYachvcGllNP2yi5ygNeSjYiDxtCRuaXN0IGtleSB0ZXN0ZXIgPG0r
dGVzdGluZ0B6YXB1Lm5ldD6IvQQTEwoAIQUCWAJ/HgIbAwULCQgHAgYVCAkKCwIE
FgIDAQIeAQIXgAAKCRDwOj0LJyvhfR5OAgkBhXIMxYkE8EuBDPjtHG7DliwBt+Ht
++KWGHxWqkAFWQitjGK33JANOyuMjMr8ealisUsbRO4io51vsOa6BVrvQVsCCQEn
VHpmetF7urR2j+V/Qr3SmT01sj0opToya52YoM1eS7+bSJRtPYyz4GomHSbMe76m
zxqcXBu7xS1moh/HQP4gW5zeBFgCfx4SBSuBBAAjBCMEAP1NEe5jGggGOhGr99OX
zwvBPLbcsyIf7cpqDi1IAHCxcnoYzVIoBJEjkdyHpuTQAvjddSF+SNGk48O4z+Ev
tmlAAI1ChPg4ZLEk1fLqq/mxsyc3HT5Ny6cKYMeW3cfCAVLlmcLYPMt5ELCOBWj+
Iy6fp22eVsMaL2S2teDJ+ZsN2abeAwEKCQACCQFm8eXql6OnFxTUQ1ODtW0ub4MM
BNz1lcGW5PV06vXOwxKEcS1H3HK/ALqD3c7F+mQOAiWnmCXpNRgqKEfd1Rsz1SHH
iKMEGBMKAAkFAlgCfx4CGwwACgkQ8Do9Cycr4X3dmAIHUn62iaxtsJ3/FlSZhXxy
d8fW4Z3NhFlCLVL6p4NijQUJQPZMcDyh9fPvSdLE1CvBMtzow2qvEVUWiunus7nl
mPwCBioXoB7rOhvEz59qnTLAjPLMOw9ib+IEjthSzrGJpfQVn1n/izJbfeG7Ghg+
FAvmbYconl4Q0uWVJFs6Ys23JuUn
=IypP
-----END PGP PRIVATE KEY BLOCK-----`

const privKey384 = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lKQEWAU34RMFK4EEACIDAwT2oNRSt6wQ2XR+yLsL5uLtmI+PGgaCpx84z1YSVRJg
0/v1/7OnDpEmGv68GXajeZ5K4pDq31HNBrDEe5NwvLXCANOnSn4YQrYpCYVzJCgi
yK67knfVPRTHuqfbrFrvi80AAYCjD4ZyVQ6aoSsWpD1KQYD9UFRrcpgwj8L3Rnmc
1bt2KeRBRYLLWCLhzvWpKGrh/k4YS7QgTmlzdCAzODQgVGVzdGVyIDxtKzM4NEB6
YXB1Lm5ldD6ImQQTEwkAIQUCWAU34QIbAwULCQgHAgYVCAkKCwIEFgIDAQIeAQIX
gAAKCRC8jTWktJdxUg5CAX909W2xrO7CZa/M85D2yDY3r8vnjUaPN+aZbcYvYXeY
QxsPdyvHQP+50OBz4KYW5coBgMA1TM7/nwL+y8vibwTuWlDtwQF2YY8MsBSeAl0Z
6sZzAnwVTtHBJFZuCnHx1RxjF5yoBFgFN+ESBSuBBAAiAwMEwPrC34Behc3CLLfm
z8nLWPM2RCNb+n+TweDf2vFYwSo96PQvecbB/KI7uthPtxVoecyLhRPR53FVp9Oc
Ihfssx2uqVNsJD+0rG9RP7dFwhYjh2a45k0o0qpxXGXOPdusAwEJCQABgMA6AgN4
XTfgtBJRRQ4GkTY1vFivGvwzq2qkansqEI6DLqqiOSqgtkhmgmnYFYOD5BRwiIEE
GBMJAAkFAlgFN+ECGwwACgkQvI01pLSXcVJjyAF/TsbOlZHUQ/swWUUXHK+/ZGEv
Sy+NVV6aThtefPRKdzFoOtVqm2zb0JiGoNTF0BbsAYCN3Z6tWZebr9Zv6I1H1w8U
tWNM382gD7IUjkEz7BIWMbHkn4m8KsAxsax7VhmiTuo=
=tJxL
-----END PGP PRIVATE KEY BLOCK-----`

const privKeyCv25519 = `-----BEGIN PGP PRIVATE KEY BLOCK-----

lFgEV/bL8xYJKwYBBAHaRw8BAQdA/tN2DTMq9IDsDjE+d0jdrQv4nUh15IwhEuK6
98RzHTAAAQC6gzVTi8V5Eis0pBg8g0iW0hp++dPczDXGg+Kc1jkzEw+ItB1NaWNo
YWwgWiAoMjU1MTkpIDxtQHphcHUubmV0Poh5BBMWCAAhBQJX9svzAhsDBQsJCAcC
BhUICQoLAgQWAgMBAh4BAheAAAoJEOctHO20d21/oE8A/jeDMoqnVrart8PlBBOh
U7POysui1CFQb4bYokaURPNzAQCE7gJ0oD2pOlU6zgia1+6JPfAnUL8rQ4PFsZ7b
5gT2BpxdBFf2y/MSCisGAQQBl1UBBQEBB0DR3in/BpS2e2jyZL1lX+DrUzJwXeQs
6CTF+o83Jt32UgMBCAcAAP9tMK39aEcGzSUmICdAqybiurbh1anP453af1dwgiRb
8BF8iGEEGBYIAAkFAlf2y/MCGwwACgkQ5y0c7bR3bX9JPgEA7WiFuFuTI4L0e8mV
3UeahfoOyLOY71uHDNdfB66DCa0BANK4aMk+j7bpJoWFNpkWq9JnhpfXV9L2dh3R
kKuKwZkI
=RN+z
-----END PGP PRIVATE KEY BLOCK-----`

// Payload encrypted for privKey384 with invalid padding - only last
// byte of padding is of correct value, rest is 0. Otherwise the key
// is correct, just the padding is invalid and indicates tampering.

// TODO: It looks like gpg2 only verifies that last byte. So this
// payload will actually go through gpg2 without problems.
const payloadInvalidPadding = `-----BEGIN PGP MESSAGE-----

wY4D4iMKJN6I/VgSAwME6YQzhZdxHePAro1u7Xj7m+gjcnh2DUV1mliU5WbixLNB
DJYXK2a654hwGsd7UKOPkjKTzMkCYSq3W8T1fGedBF/AH95v59ixx1btzg5istEt
UtpBDikx6VHbHuYnPb/gIJ5pg2CDg7w88mXmrUtsIQGer+k6NvVHuLirpcx/Nd/G
0uAB5OU8ItIF2PZeki5gVZsDcaThyrHg3eDL4bzt4NPilu8k/uD25eRpHMNilOsY
AskqQEsGWIT0x6WvAMEtaHCK4j+PZhka4LHiymo5UeBf4A7g4OTWGz7P8OM4dPR7
av0MqAvw4gT0mPzhw+8A
=rmux
-----END PGP MESSAGE-----`

// Invalid padding again, this time all padding bytes are garbage
// gpg2 complains with:
// gpg: public key decryption failed: Wrong secret key used
// gpg: decryption failed: No secret key
// when given this payload.
const payloadInvalidPadding2 = `-----BEGIN PGP MESSAGE-----

wY4D4iMKJN6I/VgSAwMEJCTy2PmsaOf5/IrR/x9+rQSdGcV4lX1G32abha8mI/Iy
HEbylH4I5xMPMcLEE/IEf0OG8cmoa6Cku/O9gpM2cDPKFgzvntQGzTV0pacaCM7Z
AVFkBdGmU8Zbjn2HSFm6IEP9ZT3zPzkdyKHfBTS4w5xWVTxtRGBAtl/ZIBmogUie
0uAB5MTm3+pjTFMnZOXnZVPCf9Thx6Lga+Ct4Q2w4PniSKGmPuCA5dXokZcRsw0v
C6xEvJlx+45z5xz2KSeDxOYt1mDqeWnw4CfieHiSUODJ4FTgVeTHYEqmaC2FhiIu
rIfsm3rl4rGKDxbhv14A
=jxQh
-----END PGP MESSAGE-----`

// Invalid KDF params will result in invalid key for AES unwrapping.
// Encrypted message for with 384 pub key. gpg2 fails with:
// gpg: public key decryption failed: Checksum error
// gpg: decryption failed: No secret key
const payloadInvalidKDFParams = `-----BEGIN PGP MESSAGE-----

wY4D4iMKJN6I/VgSAwME+T6U0UCx5h03E/TwuwdxdJADIyhmkyXBrUyu8iVXp5Ny
ue9wfjFH72iqvHq/QeOOYI73xU4TESpFRUOjPD3aPXFwxYAaPRu4qMDpvnK18tSc
HIBPjrhz1ZBe5Ek541+3IH+80o+hpOSUKdVR0DbxQ1qpY79S7VXcDCtJXfjvp1WT
0uAB5LQ14PWJ0EYhgeZxmGJt0uXhq07gOuBc4XNZ4HjiI4195eA95Wn7vkkqm5Fq
9YlMKfRkjXO/S1INtnEcyX3xfeHGdsMj4DvifXny2eDX4NXgW+TU+b3hvaY9u6QS
7eE4lCHI4inRXO7hh6EA
=0LQy
-----END PGP MESSAGE-----`

// "Invalid data"
const bad25519_1 = `-----BEGIN PGP MESSAGE-----

wW4DR1BH23/8iIwSAgMEu1ESwaqbgUmf6B/em+mGRi1oLk2YBhc9NI/S9VVGCwgA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACDZVt9FSTCqIve2XHyIqhSg
/bGvE+XpY0jPs86Mbxg1F9LgAeS+FpD/U5Kw/LBX6z9FidwQ4YLo4KPgyuHEGeDt
4jPB0TLgQuVDMKvhpHEvBVgKfejBPeRZ41OCtoL0r0f6W1m93YyPHOAY4va8EZTg
5+DW4N7kjum4qJta2wDrpNATn/BbFeIx9CgZ4S0rAA==
=Nxt6
-----END PGP MESSAGE-----`

// Curve25519 ECDH encryption with invalid encoding: packet size is
// wrong, so the parser ends up reading new openpgp packet too early.
// gpg fails with "gpg: [don't know]: invalid packet (ctb=51)".
const bad25519_2 = `-----BEGIN PGP MESSAGE-----

wW4DR1BH23/8iIwSAQhA9DDlK1QCvLLFHSWSRui8HTF+PVfpeWbYDjrtNuHtGHcg
XmBEOyBvK2feO/ckDh4HVPLsH6VlusXuwPZUS5cBEwPS4AHk7+kAQ7hb/PFa3nkq
k3la1OGaV+DP4MPhO8bgdOI0lf2X4F/lWNr5QenN4SCH2vCkNjq8qmu4L8psfUAx
SoATPSoAgr7ghuJBU8c94MDgmOC+5FHKHDYHn1BNf1O6wnR9bcvijpFbi+E+GwA=
=A5sH
-----END PGP MESSAGE-----`

// Test importing public key generated by KBPGP.
const pub25519kbpgp = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v2.0.58
Comment: https://keybase.io/crypto

xjMEWIIDOhYJKwYBBAHaRw8BAQdAcU8B6JZJA0F0W3JGq9xV+SNqscreP0PJPfV3
cQSvlBTNCE1yIFJvYm90wnYEExYKAB4FAliCAzoCGwMDCwkHAxUKCAIeAQIXgAMW
AgECGQEACgkQWVPW4aTVZYmOhAEAggy1zPXuK0M1+922RWXgMGH4ycNx2i6TnpM+
mzb0JocBAAuZCxR5uRl5mZ4slgPb8j+t5YFWG0W1d/R0nEoGnNoDzjgEWIIDOhIK
KwYBBAGXVQEFAQEHQEuUO+lr0wNBoGSz0kKAU1RLkKkUtlg/XV9RFdFt/jtfAwEK
CcJnBBgWCgAPBQJYggM6BQkPCZwAAhsIAAoJEFlT1uGk1WWJAaMBAMtIXiRT6FJz
tDEPDrdvwJiywmyD0KLYw0V/wCQE0NHNAQD7U26ZoAWzDdS553hg2nDgWmYRNjud
eub6eQsZRC2NBA==
=7Jd3
-----END PGP PUBLIC KEY BLOCK-----

`

// A Cv25519 public key (actually a EDDSA + ECDH cv25519 subkey), that
// has invalid signatures.
const pub25519invalidSig = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v2.0.58
Comment: https://keybase.io/crypto

xjMEWIIJxxYJKwYBBAHaRw8BAQdA3EM36GIaEMjfQMjK8k55bmnqs+lEYk0jO2d3
fhNI5BXNCE1yIFJvYm90wnYEExYKAB4FAliCCccCGwMDCwkHAxUKCAIeAQIXgAMW
AgECGQEACgkQygcU+wA7kcewngEAKRNGmPMAmMM1zkDOtpVobg+n0GDcv3IETwtS
4xrbSwYBAOdQd7rryN/xL6NJ4QNppg2683QSXRamWwbpLtGHKNnxzjgEWIIJxxIK
KwYBBAGXVQEFAQEHQK6fmdOpuAVI0poVERTGVMDisJudpSFtZYgte0aCsZJyAwEK
CcJnBBgWCgAPBQJYggnHBQkPCZwAAhsIAAoJEMoHFPsAO5HHXnMBADj/EUUxtH8z
tDJh2vYUVQSv5YFxHoN2E7MHwPA6uV8JAQBFy8Z5WUxzdt9qcZ3HwcT1d65Gi/yj
fRLRJGB1wtYLOw==
=ApbL
-----END PGP PUBLIC KEY BLOCK-----

`

const pub25519invalidSig2 = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: Keybase OpenPGP v2.0.58
Comment: https://keybase.io/crypto

xjMEWIIytRYJKwYBBAHaRw8BAQdACBMcQVp8V+hiGLrBpgWdNYbawz7u5mxWPaM9
MmBI8NzNCE1yIFJvYm90wnYEExYKAB4FAliCMrUCGwMDCwkHAxUKCAIeAQIXgAMW
AgECGQEACgkQJGkQb7i3juNBigEABtxludlQgXYCcW38gwEAFYQWRgAwwaVukd+G
6/0LsvsBAL/NxbyJaQWi/q20UtUL4Xd6Bf3AiOX2tHaZAmiOTaEKzjgEWIIytRIK
KwYBBAGXVQEFAQEHQAjf/2ojzFgzBwtwRXpNnylp0toLMnus7tZamb+zryF0AwEK
CcJnBBgWCgAPBQJYgjK1BQkPCZwAAhsIAAoJECRpEG+4t47jzMsBANA8ekJ2dT5C
Ah7uZZrR8ZdSpyxffRFB4XUJrpQfhn2/AQCB5CNbch2Wry/8X/E44/SIjGtYIW9+
/begQfmUIvxAAw==
=fF1S
-----END PGP PUBLIC KEY BLOCK-----

`

// This payload has "unusual" formatting of encryption key coordinates
// MPIs. The big numbers are encoded to be longer, padded with a lot
// of zeros (notice the As in the base64 representation below). The
// total MPI size (read in readPointMPI) is 1459 bits for this buffer,
// where normally it would be around 1064: 521 bits per coordinate +
// bits needed to encode "0x4" header. Decoding should be flexible
// here, because depending on the implementation, it may round the
// total size to nearest byte or save exact number of bits (2 * 521 +
// 3 bits for 0x4).
const payload521longMPIs = `-----BEGIN PGP MESSAGE-----
Version: Keybase OpenPGP v2.0.58
Comment: https://keybase.io/crypto

wcA0A6pi0WoSlxTXEgWzBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA/ZTuFZa3
go5bAv8SLZd5vTQzjQiqiXfaQUX3dQu+zytgEeiugIshlJ7JykTPGhdQFVSiKQYe
a4RKpgbn8SkNhyIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAaR1kL+5vnf9E0wv
vGDKg1jo0uEGYVodxTwk8QuqbxWiCT/jOH9kybjECSPlkDkbUIOezOfaTlwde0Wo
XUNWZ/WrMOJHerpQvMvPNjEwMSGnsIZPh8/Hafj7j8OauMG5EWgCrzsxv1mgsRXP
QkNog/5dM9JIAcT8RpDaFecdhRag6ZPuRKmNuhiFtR7o0spcqX2UkJ3FPB7UydX3
ch9PkTNL1BVD++JqYQE9eaIqlCTAsHwCgO6pQkQqPUvB
=y7cW
-----END PGP MESSAGE-----

`
