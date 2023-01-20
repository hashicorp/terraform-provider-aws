package lightsail

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/vault/helper/pgpkeys"
)

func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyPairCreate,
		ReadWithoutTimeout:   resourceKeyPairRead,
		DeleteWithoutTimeout: resourceKeyPairDelete,

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

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

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
		resp, err := conn.CreateKeyPairWithContext(ctx, &lightsail.CreateKeyPairInput{
			KeyPairName: aws.String(kName),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
		if resp.Operation == nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): no operation returned", kName)
		}
		if resp.KeyPair == nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): no key information returned", kName)
		}
		d.SetId(kName)

		// private_key and public_key are only available in the response from
		// CreateKey pair. Here we set the public_key, and encrypt the private_key
		// if a pgp_key is given, else we store the private_key in state
		d.Set("public_key", resp.PublicKeyBase64)

		// encrypt private key if pgp_key is given
		pgpKey, err := retrieveGPGKey(d.Get("pgp_key").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
		if pgpKey != "" {
			fingerprint, encrypted, err := encryptValue(pgpKey, aws.StringValue(resp.PrivateKeyBase64), "Lightsail Private Key")
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
			}

			d.Set("encrypted_fingerprint", fingerprint)
			d.Set("encrypted_private_key", encrypted)
		} else {
			d.Set("private_key", resp.PrivateKeyBase64)
		}

		op = resp.Operation
	} else {
		// importing key
		resp, err := conn.ImportKeyPairWithContext(ctx, &lightsail.ImportKeyPairInput{
			KeyPairName:     aws.String(kName),
			PublicKeyBase64: aws.String(pubKey),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
		d.SetId(kName)

		op = resp.Operation
	}

	err := waitOperation(ctx, conn, op.Id)

	if err != nil {
		// We don't return an error here because the Create call succeeded
		log.Printf("[ERR] Error waiting for KeyPair (%s) to become ready: %s", d.Id(), err)
	}

	return append(diags, resourceKeyPairRead(ctx, d, meta)...)
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

	resp, err := conn.GetKeyPairWithContext(ctx, &lightsail.GetKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			log.Printf("[WARN] Lightsail KeyPair (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Key Pair (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.KeyPair.Arn)
	d.Set("name", resp.KeyPair.Name)
	d.Set("fingerprint", resp.KeyPair.Fingerprint)

	return diags
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()
	resp, err := conn.DeleteKeyPairWithContext(ctx, &lightsail.DeleteKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Key Pair (%s): %s", d.Id(), err)
	}

	op := resp.Operation

	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Key Pair (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
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
			return "", fmt.Errorf("retrieving Public Key (%s): %w", pgpKey, err)
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
		return "", "", fmt.Errorf("encrypting %s: %w", description, err)
	}

	return fingerprints[0], base64.StdEncoding.EncodeToString(encryptedValue[0]), nil
}
