// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/vault/helper/pgpkeys"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResKeyPair = "KeyPair"
)

// @SDKResource("aws_lightsail_key_pair", name=KeyPair)
// @Tags(identifierAttribute="id", resourceType="KeyPair")
func ResourceKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyPairCreate,
		ReadWithoutTimeout:   resourceKeyPairRead,
		UpdateWithoutTimeout: resourceKeyPairUpdate,
		DeleteWithoutTimeout: resourceKeyPairDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted_private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
			},
			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrPrivateKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	kName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	var pubKey string
	var op *types.Operation
	if pubKeyInterface, ok := d.GetOk(names.AttrPublicKey); ok {
		pubKey = pubKeyInterface.(string)
	}

	if pubKey == "" {
		// creating new key
		resp, err := conn.CreateKeyPair(ctx, &lightsail.CreateKeyPairInput{
			KeyPairName: aws.String(kName),
			Tags:        getTagsIn(ctx),
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
		d.Set(names.AttrPublicKey, resp.PublicKeyBase64)

		// encrypt private key if pgp_key is given
		pgpKey, err := retrieveGPGKey(d.Get("pgp_key").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
		if pgpKey != "" {
			fingerprint, encrypted, err := encryptValue(pgpKey, aws.ToString(resp.PrivateKeyBase64), "Lightsail Private Key")
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
			}

			d.Set("encrypted_fingerprint", fingerprint)
			d.Set("encrypted_private_key", encrypted)
		} else {
			d.Set(names.AttrPrivateKey, resp.PrivateKeyBase64)
		}

		op = resp.Operation
	} else {
		// importing key
		resp, err := conn.ImportKeyPair(ctx, &lightsail.ImportKeyPairInput{
			KeyPairName:     aws.String(kName),
			PublicKeyBase64: aws.String(pubKey),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
		d.SetId(kName)

		op = resp.Operation

		if err := createTags(ctx, conn, kName, getTagsIn(ctx)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Lightsail Key Pair (%s): %s", kName, err)
		}
	}

	diag := expandOperations(ctx, conn, []types.Operation{*op}, "CreateKeyPair", ResKeyPair, kName)

	if diag != nil {
		return diag
	}

	return append(diags, resourceKeyPairRead(ctx, d, meta)...)
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	resp, err := conn.GetKeyPair(ctx, &lightsail.GetKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if err != nil {
		if IsANotFoundError(err) {
			log.Printf("[WARN] Lightsail KeyPair (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Key Pair (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.KeyPair.Arn)
	d.Set("fingerprint", resp.KeyPair.Fingerprint)
	d.Set(names.AttrName, resp.KeyPair.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(resp.KeyPair.Name)))

	setTagsOut(ctx, resp.KeyPair.Tags)

	return diags
}

func resourceKeyPairUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceKeyPairRead(ctx, d, meta)
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	resp, err := conn.DeleteKeyPair(ctx, &lightsail.DeleteKeyPairInput{
		KeyPairName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Key Pair (%s): %s", d.Id(), err)
	}

	diag := expandOperations(ctx, conn, []types.Operation{*resp.Operation}, "DeleteKeyPair", ResKeyPair, d.Id())

	if diag != nil {
		return diag
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

	return fingerprints[0], itypes.Base64Encode(encryptedValue[0]), nil
}

func FindKeyPairById(ctx context.Context, conn *lightsail.Client, id string) (*types.KeyPair, error) {
	in := &lightsail.GetKeyPairInput{KeyPairName: aws.String(id)}
	out, err := conn.GetKeyPair(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.KeyPair == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.KeyPair, nil
}
