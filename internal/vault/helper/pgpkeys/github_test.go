package pgpkeys

import (
	"encoding/base64"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"reflect"
	"strings"
	"testing"
)

func TestFetchLatestGitHubPublicKey(t *testing.T) {
	t.Parallel()

	testuser := "github:chomatdam"
	ret, err := FetchLatestGitHubPublicKey(testuser)
	if err != nil {
		t.Fatalf("bad: %v", err)
	}

	var fingerprints []string
	decodedBytes, err := base64.StdEncoding.DecodeString(ret)
	if err != nil {
		t.Fatalf("error decoding key for user %s: %v", testuser, err)
	}
	decodedString := string(decodedBytes)

	entity, err := crypto.NewKeyFromReader(strings.NewReader(decodedString))
	if err != nil {
		t.Fatalf("error parsing key for user %s: %v", testuser, err)
	}

	current := entity.GetFingerprint()
	exp := "c4df834a8627a6dcc0f0d94f8481f3143070234d"

	if !reflect.DeepEqual(current, exp) {
		t.Fatalf("fingerprints do not match; expected \n%#v\ngot\n%#v\n", exp, fingerprints)
	}
}
