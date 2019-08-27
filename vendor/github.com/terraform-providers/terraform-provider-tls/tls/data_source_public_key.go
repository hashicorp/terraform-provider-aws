package tls

import (
	"encoding/pem"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourcePublicKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePublicKeyRead,
		Schema: map[string]*schema.Schema{
			"private_key_pem": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "PEM formatted string to use as the private key",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the algorithm to use to generate the private key",
			},
			"public_key_pem": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key_openssh": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key_fingerprint_md5": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePublicKeyRead(d *schema.ResourceData, meta interface{}) error {
	// Read private key
	bytes := []byte("")
	if v, ok := d.GetOk("private_key_pem"); ok {
		bytes = []byte(v.(string))
	} else {
		return fmt.Errorf("invalid private key %#v", v)
	}
	// decode PEM encoding to ANS.1 PKCS1 DER
	keyPemBlock, _ := pem.Decode(bytes)

	if keyPemBlock == nil || (keyPemBlock.Type != "RSA PRIVATE KEY" && keyPemBlock.Type != "EC PRIVATE KEY") {
		typ := "unknown"

		if keyPemBlock != nil {
			typ = keyPemBlock.Type
		}

		return fmt.Errorf("failed to decode PEM block containing private key of type %#v", typ)
	}

	keyAlgo := ""
	switch keyPemBlock.Type {
	case "RSA PRIVATE KEY":
		keyAlgo = "RSA"
	case "EC PRIVATE KEY":
		keyAlgo = "ECDSA"
	}
	d.Set("algorithm", keyAlgo)
	// Converts a private key from its ASN.1 PKCS#1 DER encoded form
	key, err := parsePrivateKey(d, "private_key_pem", "algorithm")
	if err != nil {
		return fmt.Errorf("error converting key to algo: %s - %s", keyAlgo, err)
	}

	return readPublicKey(d, key)
}
