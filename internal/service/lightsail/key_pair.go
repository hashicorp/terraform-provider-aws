package lightsail

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/vault/helper/pgpkeys"
)

func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyPairCreate,
		Read:   resourceKeyPairRead,
		Delete: resourceKeyPairDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},

			// optional fields
			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			// additional info returned from the API
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// fields returned from CreateKey
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// encrypted fields if pgp_key is given
			"encrypted_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKeyPairCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	var kName string
	if v, ok := d.GetOk("name"); ok {
		kName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		kName = resource.PrefixedUniqueId(v.(string))
	} else {
		kName = resource.UniqueId()
	}

	var pubKey string
	var op *lightsail.Operation
	if pubKeyInterface, ok := d.GetOk("public_key"); ok {
		pubKey = pubKeyInterface.(string)
	}

	if pubKey == "" {
		// creating new key
		resp, err := conn.CreateKeyPair(&lightsail.CreateKeyPairInput{
			KeyPairName: aws.String(kName),
		})
		if err != nil {
			return err
		}
		if resp.Operation == nil {
			return fmt.Errorf("No operation found for CreateKeyPair response")
		}
		if resp.KeyPair == nil {
			return fmt.Errorf("No KeyPair information found for CreateKeyPair response")
		}
		d.SetId(kName)

		// private_key and public_key are only available in the response from
		// CreateKey pair. Here we set the public_key, and encrypt the private_key
		// if a pgp_key is given, else we store the private_key in state
		d.Set("public_key", resp.PublicKeyBase64)

		// encrypt private key if pgp_key is given
		pgpKey, err := retrieveGPGKey(d.Get("pgp_key").(string))
		if err != nil {
			return err
		}
		if pgpKey != "" {
			fingerprint, encrypted, err := encryptValue(pgpKey, *resp.PrivateKeyBase64, "Lightsail Private Key")
			if err != nil {
				return err
			}

			d.Set("encrypted_fingerprint", fingerprint)
			d.Set("encrypted_private_key", encrypted)
		} else {
			d.Set("private_key", resp.PrivateKeyBase64)
		}

		op = resp.Operation
	} else {
		// importing key
		resp, err := conn.ImportKeyPair(&lightsail.ImportKeyPairInput{
			KeyPairName:     aws.String(kName),
			PublicKeyBase64: aws.String(pubKey),
		})

		if err != nil {
			log.Printf("[ERR] Error importing key: %s", err)
			return err
		}
		d.SetId(kName)

		op = resp.Operation
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Started"},
		Target:     []string{"Completed", "Succeeded"},
		Refresh:    resourceOperationRefreshFunc(op.Id, meta),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		// We don't return an error here because the Create call succeeded
		log.Printf("[ERR] Error waiting for KeyPair (%s) to become ready: %s", d.Id(), err)
	}

	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	resp, err := conn.GetKeyPair(&lightsail.GetKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		log.Printf("[WARN] Error getting KeyPair (%s): %s", d.Id(), err)
		// check for known not found error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail KeyPair (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("arn", resp.KeyPair.Arn)
	d.Set("name", resp.KeyPair.Name)
	d.Set("fingerprint", resp.KeyPair.Fingerprint)

	return nil
}

func resourceKeyPairDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	resp, err := conn.DeleteKeyPair(&lightsail.DeleteKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	op := resp.Operation
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Started"},
		Target:     []string{"Completed", "Succeeded"},
		Refresh:    resourceOperationRefreshFunc(op.Id, meta),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for KeyPair (%s) to become destroyed: %s",
			d.Id(), err)
	}

	return nil
}

// retrieveGPGKey returns the PGP key specified as the pgpKey parameter, or queries
// the public key from the keybase service if the parameter is a keybase username
// prefixed with the phrase "keybase:"
func retrieveGPGKey(pgpKey string) (string, error) {
	const keybasePrefix = "keybase:"

	encryptionKey := pgpKey
	if strings.HasPrefix(pgpKey, keybasePrefix) {
		publicKeys, err := pgpkeys.FetchKeybasePubkeys([]string{pgpKey})
		if err != nil {
			return "", fmt.Errorf("Error retrieving Public Key for %s: %w", pgpKey, err)
		}
		encryptionKey = publicKeys[pgpKey]
	}

	return encryptionKey, nil
}

// encryptValue encrypts the given value with the given encryption key. Description
// should be set such that errors return a meaningful user-facing response.
func encryptValue(encryptionKey, value, description string) (string, string, error) {
	fingerprints, encryptedValue, err :=
		pgpkeys.EncryptShares([][]byte{[]byte(value)}, []string{encryptionKey})
	if err != nil {
		return "", "", fmt.Errorf("Error encrypting %s: %w", description, err)
	}

	return fingerprints[0], base64.StdEncoding.EncodeToString(encryptedValue[0]), nil
}
