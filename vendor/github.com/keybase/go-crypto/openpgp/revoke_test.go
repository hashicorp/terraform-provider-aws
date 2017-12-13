package openpgp

import (
	"bytes"
	"testing"

	"github.com/keybase/go-crypto/openpgp/packet"
)

func TestRevokedKey(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(revokedKey1))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if revLen := len(entity.Revocations); revLen != 1 {
		t.Fatalf("Expected to see 1 revocation, got %v", revLen)
	}
	if urevLen := len(entity.UnverifiedRevocations); urevLen != 0 {
		t.Fatalf("Expected to see 0 unverified revocations, got %v", urevLen)
	}
	iden, ok := entity.Identities["Alice"]
	if !ok {
		t.Fatal("Expected to find \"Alice\" identity.")
	}
	if iden.SelfSignature == nil {
		t.Fatal("Identity.SelfSignature is nil.")
	}
	if idenLen := len(iden.Signatures); idenLen != 0 {
		t.Fatalf("Expected to see 0 Identity.Signatures, got %v", idenLen)
	}
}

func TestKeyRevocation2(t *testing.T) {
	kring, err := ReadKeyRing(readerFromHex(revokedKeyHex))
	if err != nil || len(kring) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	key := kring[0]
	if len(key.Revocations) != 1 {
		t.Fatalf("Expected to see one revocation.")
	}
	revocation := kring[0].Revocations[0]
	if *revocation.IssuerKeyId != key.PrimaryKey.KeyId {
		t.Fatalf("Expected IsserKeyId to be %x, got %x", key.PrimaryKey.KeyId, *revocation.IssuerKeyId)
	}
	if revocation.RevocationReason == nil {
		t.Fatal("Expected revocation reason not to be nil.")
	}
}

func TestRevokedIdentityKey(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(revokedIdentityKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.Identities) != 2 {
		t.Fatal("Expected two identities")
	}
	if id, ok := entity.Identities["This One WIll be rev0ked"]; ok && id.Revocation != nil {
		t.Fatalf("Unexpected valid identity (%v)", entity.Identities)
	}
	if id, ok := entity.Identities["Hello AA"]; ok && id.Revocation == nil {
		t.Fatalf("Unexpected bad identity (%v)", entity.Identities)
	}
}

func TestDesignatedRevoker(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(designatedRevokedKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.Revocations) != 0 || len(entity.UnverifiedRevocations) != 1 {
		t.Fatal("Expected unverified revocation")
	}
	rev := entity.UnverifiedRevocations[0]
	if issuer := *rev.IssuerKeyId; issuer != 0x9AD4C1F7C4EE24FE {
		t.Fatalf("Unexpected revocation issuer: %x", issuer)
	}

	// Designated revocation should not affect KeysByIdUsage searching.
	id := uint64(4595481070173372547)
	keys := el.KeysById(id, nil)
	if len(keys) != 1 {
		t.Errorf("Expected KeysById to find revoked key %X, but got %d matches", id, len(keys))
	}
	keys = el.KeysByIdUsage(id, nil, 0)
	if len(keys) != 1 {
		t.Errorf("Expected KeysByIdUsage to revoked key %X, but got %d matches", id, len(keys))
	}
}

func TestDesignatedRevoker2(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(designatedRevokedKey2))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.Revocations) != 0 || len(entity.UnverifiedRevocations) != 1 {
		t.Fatal("Expected unverified revocation")
	}
	rev := entity.UnverifiedRevocations[0]
	if issuer := *rev.IssuerKeyId; issuer != 0x9086605E0B5C4673 {
		t.Fatalf("Unexpected revocation issuer: %x", issuer)
	}

	// Try a couple of "invalid" FindVerifiedDesignatedRevoke calls,
	// with keysets that should not verify revocation.
	sig, key := FindVerifiedDesignatedRevoke(el, entity)
	if sig != nil || key != nil {
		t.Fatal("FindVerifiedDesignatedRevoke verified revocation when given invalid keyset")
	}

	var emptyList EntityList
	sig, key = FindVerifiedDesignatedRevoke(emptyList, entity)
	if sig != nil || key != nil {
		t.Fatal("FindVerifiedDesignatedRevoke verified revocation when given empty keyset")
	}

	revokerList, err := ReadArmoredKeyRing(bytes.NewBufferString(designatedRevoker1))
	if err != nil || len(revokerList) != 1 {
		t.Fatalf("Failed to read revoker's key: %v", err)
	}

	sig, key = FindVerifiedDesignatedRevoke(revokerList, entity)
	if sig == nil || key == nil {
		t.Fatal("FindVerifiedDesignatedRevoke returned nil")
	}
	if sig != entity.UnverifiedRevocations[0] || key.PublicKey != revokerList[0].PrimaryKey {
		t.Fatal("FindVerifiedDesignatedRevoke did not find proper sig and/or key.")
	}
}

func TestNoopFindDesignated(t *testing.T) {
	// Test calling FindVerifiedDesignatedRevoke on key that does not
	// have any UnverifiedRevocations.
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(revokedIdentityKey))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	sig, key := FindVerifiedDesignatedRevoke(el, el[0])
	if sig != nil || key != nil {
		t.Fatal("FindVerifiedDesignatedRevoke should return nil, nil")
	}
}

func TestDesignatedBadSig(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(designatedRevokedKey2))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.UnverifiedRevocations) != 1 {
		t.Fatal("Expected one unverified revocation.")
	}
	// Break UnverifiedRevocation signature, it should not pass
	// verification in FindVerifiedDesignatedRevoke anymore.
	entity.UnverifiedRevocations[0].EdDSASigR = packet.FromBytes([]byte{0x01, 0x02, 0x03})

	revokerList, err := ReadArmoredKeyRing(bytes.NewBufferString(designatedRevoker1))
	sig, key := FindVerifiedDesignatedRevoke(revokerList, entity)
	if sig != nil || key != nil {
		t.Fatal("FindVerifiedDesignatedRevoke did not fail verification")
	}
}

func TestMisplacedRevocation(t *testing.T) {
	el, err := ReadArmoredKeyRing(bytes.NewBufferString(keyMisplacedRevocation))
	if err != nil || len(el) != 1 {
		t.Fatalf("Failed to read key: %v", err)
	}
	entity := el[0]
	if len(entity.Revocations) != 1 {
		t.Fatal("Expected revocation")
	}
	iden, ok := entity.Identities["Alice"]
	if !ok {
		t.Fatal("Expected to find \"Alice\" identity.")
	}
	if iden.SelfSignature == nil {
		t.Fatal("Identity.SelfSignature is nil.")
	}
	if idenLen := len(iden.Signatures); idenLen != 0 {
		t.Fatalf("Expected to see 0 Identity.Signatures, got %v", idenLen)
	}
}

// Self-revoked key
const revokedKey1 = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQFCBFj3GqwRAwC922rw75mP/WuF/wdZOcAPVfqukqGd5S5x7ajUGi77sXqqhAnr
j+XsneekldcHqlJuti7IHxMcbOZQN0rYinpk6ODfB3J1ShcHTC2IpWsngzt+tL6V
zSIXbR5rLUGg2RMAoPMi18hqBq8xQQDG2rEWCRybRvnvAv0axMy37OAeye6Ky8m2
0l1vDFeNO7/OH9eO5oNEwNuVG/shjZkGTD/YuB8huPvcyMR3xxs6Qmjn0XRfUWxt
xPvfctP9HS7MPeDqa/DsMZ5hh7B1eiwmk2cj5E6ZOFk2G8sC/jtcA3wVF7eHsJvA
CL14MLeQ9g+04CT7VhvPt2f3X3GF7XQ/2pgBfnzDi26VU9ND75NBmwVulbJw8QG7
JOpMi3FeHhsWtbQGcZg3Vcw8IamnqhEaFJ9Nb/hV4rKm0IXfgohJBCARAgAJBQJY
9xqvAh0AAAoJEDl/NacbGDDEyDYAn3QKeWn52B9lHes3pNlRqFS4/VlvAJ9DP+Kf
Ec8PxRr9qYH8KpacyYWua7QFQWxpY2WIYQQTEQIAIQUCWPcarAIbAwULCQgHAgYV
CAkKCwIEFgIDAQIeAQIXgAAKCRA5fzWnGxgwxB00AJ4inWM/H4FuFxd8A2TmmN1J
nb/W7ACgozlKd8s90o72ccJq4zxLLOC/ik25AQ0EWPcarBAEAMNfbgy0zfpDz6zi
kU+9ysCnQPaAQjNrFCu3JnJ29TGTRjGq95NOYgaU3/guAf8d1QSBAPzC+c+o/TWQ
2+y6qKJnZbsvFzVjBiJW6zpFDyWvupfATzKE3rsWYeyCwdPfwHTejWGXeoJKkSAy
em+0wm2VI6CKRsrf88UCwD9wk7VrAAMFBAC1+2hcC1TcJuZwwhDd3xllXgrMHGyG
I92RmaTjttJgOvlN5Pyz6q5HgB5EFkzbW3YCGm/YY+KTXKWUp9u2Eh9cc8R9Pm7c
HzJlEINC+VMe/+Nzd15ceySNGNIUW6D9OTtzMmgrkXCvRnZ0DDsnexVOM4pI6Up4
afCdmQfHhocmZ4hJBBgRAgAJBQJY9xqsAhsMAAoJEDl/NacbGDDEsCgAn2RJ+SJB
i7W/Rh1FjTXpL+d7zPqzAJ0Vzhg3SkrLt8/VGRRSJRUMpb4bPw==
=w/2P
-----END PGP PUBLIC KEY BLOCK-----
`

// Public key that has two identities, one of which is revoked.
const revokedIdentityKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWOIZOBYJKwYBBAHaRw8BAQdAOw15aNPr+v1ACWdSwaKmT+vAfpZJu2aiX/ED
NR70fYm0GFRoaXMgT25lIFdJbGwgYmUgcmV2MGtlZIh5BBMWCAAhBQJY4hlKAhsD
BQsJCAcCBhUICQoLAgQWAgMBAh4BAheAAAoJEIUbNJhCKy361LQBAPH+mCf0r0z9
SZXw4B8fJ+jCl//0ato6Nk8bsedA2MyjAP4tx/h9XHjmANhKpue9YCyUFdV2NSKs
TIJ/EpNwz1QjArQISGVsbG8gQUGIYQQwFggACQUCWOIZcgIdIAAKCRCFGzSYQist
+nW9AQCaXyyTOmUw9gaw0SsS27NLtsYcu/affY4KLYQRW2ZjlgD9GLR5IKYtlX21
n/8Gw7KAuHaIQLK+wcbXnFabzM7TYA2IeQQTFggAIQUCWOIZOAIbAwULCQgHAgYV
CAkKCwIEFgIDAQIeAQIXgAAKCRCFGzSYQist+lGFAP9EFlJ0BCgOe6ART8xk93f3
fF+wOdMzdQ+6hni8wqW3OQEAq3VufchOPYJSL4fA+Oq7uEw5Z5Q9tBViES2Br7+I
1Au4OARY4hk4EgorBgEEAZdVAQUBAQdAAfA2+lbpmA1YXqHefB8gShHq201PsJmA
AQ2EB67c/XcDAQgHiGEEGBYIAAkFAljiGTgCGwwACgkQhRs0mEIrLfqOYwD/TaDI
Y81Z5IXtMVSMjg7sgNI93W9+xY5u0fHH5KThko4BAM7utt+MrMl67IrSLj0HLtVt
iO3AEa577DoHC0fseUgG
=uJYe
-----END PGP PUBLIC KEY BLOCK-----
`

// Key that has a designated revoker direct sig and also a designated
// revocation signature.
const designatedRevokedKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWN6JhRYJKwYBBAHaRw8BAQdA6NMRLTcnG9zXYIlH8aTxXttm6Ibnd+JcdnZR
7ZaarAOIYQQgFggACQUCWN6J+wIdAwAKCRCa1MH3xO4k/kqzAQCJRWV9XtLuBALs
pLfqb3V8+dumX9dNZhzrJejoOyNwIwEAzjpTdaSApbvfdon0ndf05UB+hkR2Sal5
bDXHANjltAiIeQQfFggAIQUCWN6J0RcMgBbsLs6ylR7EOEBNML2a1MH3xO4k/gIH
AAAKCRA/xm2vd7dAgxE6AP45XxRMDBG4MSvyqZw3zQ3XT0DzZyDfwmh4bNd2FZJg
lgD9ErTgyWuxVo4c/k/W6vowu6tV0rhMjH9MfwxmzY20igu0B1Jldm9rZWWIeQQT
FggAIQUCWN6JhQIbAwULCQgHAgYVCAkKCwIEFgIDAQIeAQIXgAAKCRA/xm2vd7dA
g0wmAPwOALfHBhKEiMTxCtAJ4ynJLiVXYmb+AdxLb6Q+ISmNuAEAt6uDcdM9pfX8
BjB78WoVjkxwRZpIMM3tcjz6VcR15w+4OARY3omFEgorBgEEAZdVAQUBAQdApcyK
X+duQaFIZV882qD8PZd3b9qS/ZN1EJSBOkJNiWQDAQgHiGEEGBYIAAkFAljeiYUC
GwwACgkQP8Ztr3e3QIO2KAD+NUOcZekVrfgx7STVdx2N9/zaK8cZSVgp2dWJ4DKE
1PsA+gM9O4+vwInhP8xGtH816FXJtGiw/mAyxCUeRTgi8KEH
=qbn3
-----END PGP PUBLIC KEY BLOCK-----
`

// Key that has been designated by designatedRevoker1 and contains a
// verifiable revocation signature.
const designatedRevokedKey2 = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWPY5DhYJKwYBBAHaRw8BAQdATJ1ECHK+nn/iRBTSJ+tGAVn9TtlOzAQeSNIh
FCbqkmSIYQQgFggACQUCWPY52gIdAgAKCRCQhmBeC1xGcwULAQDH4ohXPkNND4Ez
LRyXPNhCSC7IW8bfHqLWj0VH/cXBFwD/ci+R1C/pNXKzawLDw2k2Kqd1gn5Gd16C
RAU/0Q4MWAqIeQQfFggAIQUCWPY5TxcMgBbJcZz6AbUchwVji8mQhmBeC1xGcwIH
AAAKCRAYEqe7+/Ynv5hkAP0YaIHYyP55EVqiM/8JZJYK/A8x273QpfttY7KG8op0
cAD+J0nz4RnGJfhrfZGa1EwFNlQ6uyF8/BAJeat42x6w5gW0CEpvaG4gRG9liHkE
ExYIACEFAlj2OQ4CGwMFCwkIBwIGFQgJCgsCBBYCAwECHgECF4AACgkQGBKnu/v2
J7+B3QEAlnd3pLw0X8ccY/J7q0lvsZqhjg5JUCHE/VhHv9ff804BAN+9pttBx91G
AK/J0xl/dFxg4nAb+MrJabMlFJBfU2cKuDgEWPY5DhIKKwYBBAGXVQEFAQEHQNIf
z8EWK30QHiLVcO0yNlXRKpsygbQR9TnCzySnZlV/AwEIB4hhBBgWCAAJBQJY9jkO
AhsMAAoJEBgSp7v79ie/rccA/2JVMMi0lCB+pgNXtsy+VsGQN1Wn93hMtp96jTH6
ZXu5AP9gPV6r//WSuvfLl0yO4agWaa+lersoYwyovTEkqe0UAQ==
=hUOq
-----END PGP PUBLIC KEY BLOCK-----
`

// Revoker key that signed revocation of designatedRevokedKey2.
const designatedRevoker1 = `-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEWPY5HxYJKwYBBAHaRw8BAQdAS7VZfelXtQ13zj/1vC9w6KijlYF5Q0wknInU
7vXikhe0DEphY2sgUmV2b2tlcoh5BBMWCAAhBQJY9jkfAhsDBQsJCAcCBhUICQoL
AgQWAgMBAh4BAheAAAoJEJCGYF4LXEZzF+AA/3yM9sepkr7FXXOWd+fx+R4/0iMZ
HE4ykX7nhRsXE72BAQDRt/5NrJg5jdGgaE9ho9aXEv854Dx1FJxBxiQomKLmArg4
BFj2OR8SCisGAQQBl1UBBQEBB0A3KqdTAoZN2mMJfwvKwbC8Ibv7cDjHL+2zGm+R
/ur3PAMBCAeIYQQYFggACQUCWPY5HwIbDAAKCRCQhmBeC1xGcyDJAQDG9QqWpV4c
Sm3K1NCp/0bIlRI/aFycA65lhHNoIZgPZwEApkjPInTzm1ZyVl4zgZxFltLgPbnU
J25shXYSVsIQJQ0=
=wIyY
-----END PGP PUBLIC KEY BLOCK-----
`

// In this bundle, key revocation packet appears after identities.
// gpg2 does not mark that key as revoked, we are more flexible in
// uid/subkey parsing so we happen to mark that key as revoked.
const keyMisplacedRevocation = `-----BEGIN PGP PUBLIC KEY BLOCK-----

xsCCBFj3GqwRAwC922rw75mP/WuF/wdZOcAPVfqukqGd5S5x7ajUGi77sXqqhAnr
j+XsneekldcHqlJuti7IHxMcbOZQN0rYinpk6ODfB3J1ShcHTC2IpWsngzt+tL6V
zSIXbR5rLUGg2RMAoPMi18hqBq8xQQDG2rEWCRybRvnvAv0axMy37OAeye6Ky8m2
0l1vDFeNO7/OH9eO5oNEwNuVG/shjZkGTD/YuB8huPvcyMR3xxs6Qmjn0XRfUWxt
xPvfctP9HS7MPeDqa/DsMZ5hh7B1eiwmk2cj5E6ZOFk2G8sC/jtcA3wVF7eHsJvA
CL14MLeQ9g+04CT7VhvPt2f3X3GF7XQ/2pgBfnzDi26VU9ND75NBmwVulbJw8QG7
JOpMi3FeHhsWtbQGcZg3Vcw8IamnqhEaFJ9Nb/hV4rKm0IXfgs0FQWxpY2XCYQQT
EQIAIQUCWPcarAIbAwULCQgHAgYVCAkKCwIEFgIDAQIeAQIXgAAKCRA5fzWnGxgw
xB00AJ4inWM/H4FuFxd8A2TmmN1Jnb/W7ACgozlKd8s90o72ccJq4zxLLOC/ik3C
SQQgEQIACQUCWPcarwIdAAAKCRA5fzWnGxgwxMg2AJ90Cnlp+dgfZR3rN6TZUahU
uP1ZbwCfQz/inxHPD8Ua/amB/CqWnMmFrmvOwE0EWPcarBAEAMNfbgy0zfpDz6zi
kU+9ysCnQPaAQjNrFCu3JnJ29TGTRjGq95NOYgaU3/guAf8d1QSBAPzC+c+o/TWQ
2+y6qKJnZbsvFzVjBiJW6zpFDyWvupfATzKE3rsWYeyCwdPfwHTejWGXeoJKkSAy
em+0wm2VI6CKRsrf88UCwD9wk7VrAAMFBAC1+2hcC1TcJuZwwhDd3xllXgrMHGyG
I92RmaTjttJgOvlN5Pyz6q5HgB5EFkzbW3YCGm/YY+KTXKWUp9u2Eh9cc8R9Pm7c
HzJlEINC+VMe/+Nzd15ceySNGNIUW6D9OTtzMmgrkXCvRnZ0DDsnexVOM4pI6Up4
afCdmQfHhocmZ8JJBBgRAgAJBQJY9xqsAhsMAAoJEDl/NacbGDDEsCgAn2RJ+SJB
i7W/Rh1FjTXpL+d7zPqzAJ0Vzhg3SkrLt8/VGRRSJRUMpb4bPw==
=riYc
-----END PGP PUBLIC KEY BLOCK-----
`
