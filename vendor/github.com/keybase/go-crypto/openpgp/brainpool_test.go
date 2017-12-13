package openpgp

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/keybase/go-crypto/openpgp/armor"
	"github.com/keybase/go-crypto/openpgp/packet"
)

const msg = `Hello World!`

func signWithKeyFile(t *testing.T, name, password string) {
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	es, err := ReadArmoredKeyRing(f)
	if err != nil {
		t.Fatal(err)
	}

	// Cycle through all entities we find a signing key.
	for _, e := range es {
		if err = e.PrivateKey.Decrypt([]byte(password)); err != nil {
			continue
		}

		buf := new(bytes.Buffer)
		if err = DetachSign(buf, e, strings.NewReader(msg), nil); err == nil {
			p, err := packet.Read(buf)
			if err != nil {
				t.Fatal(err)
			}
			sig, ok := p.(*packet.Signature)
			if !ok {
				t.Fatal("couldn't parse signature from buffer")
			}
			signed := sig.Hash.New()
			signed.Write([]byte(msg))
			if err := e.PrimaryKey.VerifySignature(signed, sig); err != nil {
				t.Fatal(err)
			}

			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}

func verifySig(t *testing.T, keyFile, sigFile string) {
	var f *os.File
	var err error
	var b *armor.Block
	var p packet.Packet

	if f, err = os.Open(keyFile); err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if b, err = armor.Decode(f); err != nil {
		t.Fatal(err)
	}
	if p, err = packet.Read(b.Body); err != nil {
		t.Fatal(err)
	}

	priv, ok := p.(*packet.PrivateKey)
	if !ok {
		t.Fatal("couldn't parse private key")
	}

	if f, err = os.Open(sigFile); err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if b, err = armor.Decode(f); err != nil {
		t.Fatal(err)
	}
	if p, err = packet.Read(b.Body); err != nil {
		t.Fatal(err)
	}

	sig, ok := p.(*packet.Signature)
	if !ok {
		t.Fatal("couldn't parse signature")
	}

	signed := sig.Hash.New()
	signed.Write([]byte(msg))
	if err := priv.PublicKey.VerifySignature(signed, sig); err != nil {
		t.Fatal(err)
	}
}

func TestParseP256r1(t *testing.T) {
	signWithKeyFile(t, "testdata/brainpoolP256r1.pgp", "256")
	verifySig(t, "testdata/brainpoolP256r1.pgp", "testdata/brainpoolP256r1.asc")
}

func TestParseP384r1(t *testing.T) {
	signWithKeyFile(t, "testdata/brainpoolP384r1.pgp", "384")
	verifySig(t, "testdata/brainpoolP384r1.pgp", "testdata/brainpoolP384r1.asc")
}

func TestParseP512r1(t *testing.T) {
	signWithKeyFile(t, "testdata/brainpoolP512r1.pgp", "512")
	verifySig(t, "testdata/brainpoolP512r1.pgp", "testdata/brainpoolP512r1.asc")
}
