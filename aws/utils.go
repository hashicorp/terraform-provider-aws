package aws

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	homedir "github.com/mitchellh/go-homedir"
)

// Base64Encode encodes data if the input isn't already encoded using base64.StdEncoding.EncodeToString.
// If the input is already base64 encoded, return the original input unchanged.
func base64Encode(data []byte) string {
	// Check whether the data is already Base64 encoded; don't double-encode
	if isBase64Encoded(data) {
		return string(data)
	}
	// data has not been encoded encode and return
	return base64.StdEncoding.EncodeToString(data)
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func looksLikeJsonString(s interface{}) bool {
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
}

func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func isResourceNotFoundError(err error) bool {
	_, ok := err.(*resource.NotFoundError)
	return ok
}

func isResourceTimeoutError(err error) bool {
	timeoutErr, ok := err.(*resource.TimeoutError)
	return ok && timeoutErr.LastError == nil
}

func appendUniqueString(slice []string, elem string) []string {
	for _, e := range slice {
		if e == elem {
			return slice
		}
	}
	return append(slice, elem)
}

// loadFileContent returns contents of a file in a given path
func loadFileContent(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}

func byteHashSum(content []byte) string {
	hash := sha1.Sum(content)

	return hex.EncodeToString(hash[:])
}

func stringHashSum(css string) string {
	v := []byte(css)
	hash := sha1.Sum(v)
	return hex.EncodeToString(hash[:])
}

func remoteFileContent(v string) ([]byte, error) {
	match, _ := regexp.MatchString(`^https?:\/\/`, v)

	if match {
		resp, err := http.Get(v)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		return ioutil.ReadAll(resp.Body)
	} else {
		return loadFileContent(v)
	}
}

func remoteFileHashSum(v string) string {
	content, err := remoteFileContent(v)

	if err != nil {
		return ""
	}

	return byteHashSum(content)
}
