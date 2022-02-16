package pgpkeys

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

func TestFetchPubkeys(t *testing.T) {
	testset := []string{
		"keybase:jefferai",
		"keybase:hashicorp",
		"external:hbollon|https://github.com/hbollon.gpg",
		"external:hbollon|https://github.com/%s.gpg",
	}
	ret, err := FetchPubkeys(testset)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}

	fingerprints := map[string]string{}
	for username, key := range ret {
		data, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			t.Fatalf("error decoding key for user %s: %v", username, err)
		}
		entity, err := openpgp.ReadEntity(packet.NewReader(bytes.NewBuffer(data)))
		if err != nil {
			t.Fatalf("error parsing key for user %s: %v", username, err)
		}
		fingerprints[username] = hex.EncodeToString(entity.PrimaryKey.Fingerprint[:])
	}

	exp := map[string]string{
		"keybase:jefferai":  "0f801f518ec853daff611e836528efcac6caa3db",
		"keybase:hashicorp": "c874011f0ab405110d02105534365d9472d7468f",
		"external:hbollon":  "49ca0f0ba50de5e1fab6638b3b3614614b74b1d6",
	}

	if !reflect.DeepEqual(fingerprints, exp) {
		t.Fatalf("fingerprints do not match; expected \n%#v\ngot\n%#v\n", exp, fingerprints)
	}
}

func TestFetchKeybasePubkeys(t *testing.T) {
	testset := []string{"keybase:jefferai", "keybase:hashicorp"}
	ret, err := FetchPubkeys(testset)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}

	fingerprints := []string{}
	for _, user := range testset {
		data, err := base64.StdEncoding.DecodeString(ret[user])
		if err != nil {
			t.Fatalf("error decoding key for user %s: %v", user, err)
		}
		entity, err := openpgp.ReadEntity(packet.NewReader(bytes.NewBuffer(data)))
		if err != nil {
			t.Fatalf("error parsing key for user %s: %v", user, err)
		}
		fingerprints = append(fingerprints, hex.EncodeToString(entity.PrimaryKey.Fingerprint[:]))
	}

	exp := []string{
		"0f801f518ec853daff611e836528efcac6caa3db",
		"c874011f0ab405110d02105534365d9472d7468f",
	}

	if !reflect.DeepEqual(fingerprints, exp) {
		t.Fatalf("fingerprints do not match; expected \n%#v\ngot\n%#v\n", exp, fingerprints)
	}
}

func TestFetchExternalPubkeys(t *testing.T) {
	testset := []string{
		"external:hbollon|https://gitlab.com/%s.gpg",
		"external:hbollon|https://github.com/%s.gpg",
		"external:hbollon|https://github.com/hbollon.gpg",
	}
	ret, err := FetchPubkeys(testset)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}

	fingerprints := []string{}
	for _, user := range testset {
		userData := strings.Split(user, "|")
		data, err := base64.StdEncoding.DecodeString(ret[userData[0]])
		if err != nil {
			t.Fatalf("error decoding key for user %s: %v", user, err)
		}
		entity, err := openpgp.ReadEntity(packet.NewReader(bytes.NewBuffer(data)))
		if err != nil {
			t.Fatalf("error parsing key for user %s: %v", user, err)
		}
		fingerprints = append(fingerprints, hex.EncodeToString(entity.PrimaryKey.Fingerprint[:]))
	}

	exp := []string{
		"49ca0f0ba50de5e1fab6638b3b3614614b74b1d6",
		"49ca0f0ba50de5e1fab6638b3b3614614b74b1d6",
		"49ca0f0ba50de5e1fab6638b3b3614614b74b1d6",
	}

	if !reflect.DeepEqual(fingerprints, exp) {
		t.Fatalf("fingerprints do not match; expected \n%#v\ngot\n%#v\n", exp, fingerprints)
	}
}

func TestValidateFetchingURL(t *testing.T) {
	testset := []string{
		"",
		"https://github.com/%s.gpg",
		"https://gitlab.com/%s.gpg",
		"https://github.com/%s.gpg%d",
		"https://github.com/hbollon.gpg",
		"https://git hub.com/hbollon.gpg",
		"https://github.com/%s%s.gpg",
		"https://github.%q/%s.gpg",
		"http://github.com/%s.gpg",
		"github.com/%s.gpg",
	}

	exp := []error{
		fmt.Errorf("fetching url cannot be empty"),
		nil,
		nil,
		fmt.Errorf("fetching url cannot contain format specifiers others than %%s"),
		nil,
		fmt.Errorf("fetching url is not a valid URL"),
		fmt.Errorf("fetching url must contain exactly one %%s token"),
		fmt.Errorf("fetching url cannot contain format specifiers others than %%s"),
		fmt.Errorf("fetching url must be a valid URL in https:// format"),
		fmt.Errorf("fetching url must be a valid URL in https:// format"),
	}

	for i, url := range testset {
		err := validateFetchingURL(url)
		if err == nil && exp[i] != nil || err != nil && exp[i] == nil ||
			err != nil && exp[i] != nil && err.Error() != exp[i].Error() {
			t.Fatalf("validation result do not match; expected \n%#v\ngot\n%#v\n", exp[i], err)
		}
	}
}
