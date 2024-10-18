package pgpkeys

import (
	"encoding/base64"
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/hashicorp/go-cleanhttp"
	"io"
	"strings"
)

func FetchLatestGitHubPublicKey(pgpKey string) (string, error) {
	stringComponents := strings.Split(pgpKey, ":")
	if len(stringComponents) != 2 {
		return "", fmt.Errorf("invalid GPG key format for Github, received='%s', expected='github:$username'", pgpKey)
	}

	username := strings.TrimSpace(stringComponents[1])
	url := fmt.Sprintf("https://github.com/%s.gpg", username)
	resp, err := cleanhttp.DefaultClient().Get(url)

	if err != nil {
		return "", fmt.Errorf("retrieving Public Key from Github (%s): %w", pgpKey, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body containing ASCII GPG Public Key from GitHub")
	}

	bodyString := string(bodyBytes)
	key, err := crypto.NewKeyFromArmored(bodyString)
	if err != nil {
		return "", fmt.Errorf("error reading keyring: %w", err)
	}
	bytes, err := key.GetPublicKey()
	if err != nil {
		return "", fmt.Errorf("failed to serialize raw GPG Public Key from GitHub")
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
