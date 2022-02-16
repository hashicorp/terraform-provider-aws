package pgpkeys

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-provider-aws/internal/vault/sdk/helper/jsonutil"
)

const (
	kbPrefix       = "keybase:"
	externalPrefix = "external:"
	externalRegex  = `^external:[^ |]+\|[^ |]+$`
)

func FetchPubkeys(input []string) (map[string]string, error) {
	client := cleanhttp.DefaultClient()
	if client == nil {
		return nil, fmt.Errorf("unable to create an http client")
	}

	if len(input) == 0 {
		return nil, nil
	}

	kbUsernames := make([]string, 0, len(input))
	extUsernames := make([]string, 0, len(input))
	extReg, _ := regexp.Compile(externalRegex)
	for _, v := range input {
		if strings.HasPrefix(v, kbPrefix) {
			kbUsernames = append(kbUsernames, strings.TrimSuffix(strings.TrimPrefix(v, kbPrefix), "\n"))
		} else if match := extReg.MatchString(v); match {
			extUser := strings.TrimPrefix(v, externalPrefix)
			if err := validateFetchingURL(strings.Split(extUser, "|")[1]); err != nil {
				return nil, err
			}
			extUsernames = append(extUsernames, strings.TrimSuffix(strings.TrimPrefix(extUser, externalPrefix), "\n"))
		} else {
			return nil, fmt.Errorf("unrecognized format for key ID %q", v)
		}
	}

	ret := make(map[string]string, len(kbUsernames)+len(extUsernames))
	if len(kbUsernames) != 0 {
		retKb, err := FetchKeybasePubkeys(client, kbUsernames)
		if err != nil {
			return nil, err
		}
		for k, v := range retKb {
			ret[k] = v
		}
	}

	if len(extUsernames) != 0 {
		retExt, err := FetchPubkeysFromURL(client, extUsernames)
		if err != nil {
			return nil, err
		}
		for k, v := range retExt {
			ret[k] = v
		}
	}

	return ret, nil
}

// FetchKeybasePubkeys fetches public keys from Keybase given a set of
// usernames, which are derived from correctly formatted input entries. It
// doesn't use their client code due to both the API and the fact that it is
// considered alpha and probably best not to rely on it.  The keys are returned
// as base64-encoded strings.
func FetchKeybasePubkeys(client *http.Client, usernames []string) (map[string]string, error) {
	ret := make(map[string]string, len(usernames))
	url := fmt.Sprintf("https://keybase.io/_/api/1.0/user/lookup.json?usernames=%s&fields=public_keys", strings.Join(usernames, ","))
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type PublicKeys struct {
		Primary struct {
			Bundle string
		}
	}

	type LThem struct {
		PublicKeys `json:"public_keys"`
	}

	type KbResp struct {
		Status struct {
			Name string
		}
		Them []LThem
	}

	out := &KbResp{
		Them: []LThem{},
	}

	if err := jsonutil.DecodeJSONFromReader(resp.Body, out); err != nil {
		return nil, err
	}

	if out.Status.Name != "OK" {
		return nil, fmt.Errorf("got non-OK response: %q", out.Status.Name)
	}

	missingNames := make([]string, 0, len(usernames))
	var keyReader *bytes.Reader
	serializedEntity := bytes.NewBuffer(nil)
	for i, themVal := range out.Them {
		if themVal.Primary.Bundle == "" {
			missingNames = append(missingNames, usernames[i])
			continue
		}
		keyReader = bytes.NewReader([]byte(themVal.Primary.Bundle))
		entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
		if err != nil {
			return nil, err
		}
		if len(entityList) != 1 {
			return nil, fmt.Errorf("primary key could not be parsed for user %q", usernames[i])
		}
		if entityList[0] == nil {
			return nil, fmt.Errorf("primary key was nil for user %q", usernames[i])
		}

		serializedEntity.Reset()
		err = entityList[0].Serialize(serializedEntity)
		if err != nil {
			return nil, fmt.Errorf("error serializing entity for user %q: %w", usernames[i], err)
		}

		// The API returns values in the same ordering requested, so this should properly match
		ret[kbPrefix+usernames[i]] = base64.StdEncoding.EncodeToString(serializedEntity.Bytes())
	}

	if len(missingNames) > 0 {
		return nil, fmt.Errorf("unable to fetch keys for user(s) %q from keybase", strings.Join(missingNames, ","))
	}

	return ret, nil
}

// FetchPubkeysFromURL fetches public keys given a set of pairs composed
// by usernames and URLs, which are derived from correctly formatted input entries.
// The user/url pairs are in the format of "username|url"
// The url must be a valid url and must be reachable. It can be a raw url or a
// formatable url which will use the associated username to fetch the public key.
// Provided url must be in https and return a gpg key in ascii armored format.
// The keys are returned as base64-encoded strings.
func FetchPubkeysFromURL(client *http.Client, usernames []string) (map[string]string, error) {
	ret := make(map[string]string, len(usernames))
	missingNames := make([]string, 0, len(usernames))
	for i, user := range usernames {
		// Split the username and url
		userData := strings.Split(user, "|")
		var urlUser string
		// Format the url if it's a formatable url using username
		if strings.Contains(userData[1], "%s") {
			urlUser = fmt.Sprintf(userData[1], userData[0])
		} else {
			urlUser = userData[1]
		}

		resp, err := client.Get(urlUser)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read the response body which is a raw gpg key in ascii armored format
		publicKey, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var keyReader *bytes.Reader
		serializedEntity := bytes.NewBuffer(nil)
		if string(publicKey) == "" {
			missingNames = append(missingNames, userData[0])
			continue
		}
		keyReader = bytes.NewReader(publicKey)
		entityList, err := openpgp.ReadArmoredKeyRing(keyReader)
		if err != nil {
			return nil, err
		}
		if len(entityList) != 1 {
			return nil, fmt.Errorf("primary key could not be parsed for user %q", usernames[i])
		}
		if entityList[0] == nil {
			return nil, fmt.Errorf("primary key was nil for user %q", usernames[i])
		}

		serializedEntity.Reset()
		err = entityList[0].Serialize(serializedEntity)
		if err != nil {
			return nil, fmt.Errorf("error serializing entity for user %q: %w", usernames[i], err)
		}

		ret[externalPrefix+userData[0]] = base64.StdEncoding.EncodeToString(serializedEntity.Bytes())
	}

	if len(missingNames) > 0 {
		return nil, fmt.Errorf("unable to fetch keys for user(s) %q from external api", strings.Join(missingNames, ","))
	}

	return ret, nil
}

// validateFetchingURL check for fetching url's validity
// It returns an error if the url is not valid
// A valid url must not be empty, starts with https://,
// strictly contains one "%s" token and no other format specifiers
func validateFetchingURL(rawUrl string) error {
	forbiddenTokens := []string{
		"%d", "%f", "%p", "%b",
		"%x", "%c", "%q", "%v",
	}

	// Check for empty url
	if rawUrl == "" {
		return fmt.Errorf("fetching url cannot be empty")
	}

	// Check for https:// scheme
	if !strings.HasPrefix(rawUrl, "https://") {
		return fmt.Errorf("fetching url must be a valid URL in https:// format")
	}

	// Check for forbidden tokens
	for _, token := range forbiddenTokens {
		if strings.Contains(rawUrl, token) {
			return fmt.Errorf("fetching url cannot contain format specifiers others than %%s")
		}
	}

	// Check for "%s" token count
	if strings.Count(rawUrl, "%s") == 1 {
		// Format URL for the next check
		rawUrl = fmt.Sprintf(rawUrl, "test")
	} else if strings.Count(rawUrl, "%s") > 1 {
		return fmt.Errorf("fetching url must contain exactly one %%s token")
	}

	// Check for formatable URL validity
	if u, err := url.Parse(rawUrl); err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("fetching url is not a valid URL")
	}

	return nil
}
