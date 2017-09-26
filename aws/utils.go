package aws

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
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

func hashSum(value interface{}) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value.(string))))
}

func hashPassword(value interface{}) string {
	return hashSum(value)
}

// Manages password hashing in state file based on hash_password config value.
// Should be called from CRUD functions. Switches on method to accomplish different
// tasks.
func managePasswordHash(d *schema.ResourceData, method string) (bool, *string) {
	var requestUpdate bool
	var masterUserPassword *string
	o_passwd, n_passwd := d.GetChange("password")
	n_passwdHash := hashPassword(n_passwd)
	hash_flag := d.Get("hash_password").(bool)
	switch method {
	case "create":
		if hash_flag {
			d.Set("password", n_passwdHash)
		}
	case "update":
		values := []string{o_passwd.(string), n_passwd.(string), n_passwdHash}
		encounter := map[string]bool{}
		var match bool
		for value := range values {
			if encounter[values[value]] {
				match = true
			} else {
				encounter[values[value]] = true
			}
		}
		if !match {
			masterUserPassword = aws.String(n_passwd.(string))
			requestUpdate = true
		}
		if hash_flag {
			d.Set("password", n_passwdHash)
		}
	}
	return requestUpdate, masterUserPassword
}
