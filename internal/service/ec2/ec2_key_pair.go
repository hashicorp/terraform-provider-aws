// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"bytes"
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/crypto/ssh"
)

// @SDKResource("aws_key_pair", name="Key Pair")
// @Tags(identifierAttribute="key_pair_id")
// @Testing(tagsTest=false)
func resourceKeyPair() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyPairCreate,
		ReadWithoutTimeout:   resourceKeyPairRead,
		UpdateWithoutTimeout: resourceKeyPairUpdate,
		DeleteWithoutTimeout: resourceKeyPairDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		MigrateState:  keyPairMigrateState,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255),
				ConflictsWith: []string{"key_name_prefix"},
			},
			"key_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255-sdkid.UniqueIDSuffixLength),
				ConflictsWith: []string{"key_name"},
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPublicKey: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v any) string {
					switch v := v.(type) {
					case string:
						return strings.TrimSpace(v)
					default:
						return ""
					}
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceKeyPairCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	keyName := create.Name(ctx, d.Get("key_name").(string), d.Get("key_name_prefix").(string))
	input := ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: []byte(d.Get(names.AttrPublicKey).(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeKeyPair),
	}

	output, err := conn.ImportKeyPair(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing EC2 Key Pair (%s): %s", keyName, err)
	}

	d.SetId(aws.ToString(output.KeyName))

	return append(diags, resourceKeyPairRead(ctx, d, meta)...)
}

func resourceKeyPairRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)

	keyPair, err := findKeyPairByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] EC2 Key Pair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Key Pair (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, keyPairARN(ctx, c, d.Id()))
	d.Set("fingerprint", keyPair.KeyFingerprint)
	d.Set("key_name", keyPair.KeyName)
	d.Set("key_name_prefix", create.NamePrefixFromName(aws.ToString(keyPair.KeyName)))
	d.Set("key_pair_id", keyPair.KeyPairId)
	d.Set("key_type", keyPair.KeyType)

	setTagsOut(ctx, keyPair.Tags)

	return diags
}

func resourceKeyPairUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceKeyPairRead(ctx, d, meta)...)
}

func resourceKeyPairDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Key Pair: %s", d.Id())
	input := ec2.DeleteKeyPairInput{
		KeyName: aws.String(d.Id()),
	}
	_, err := conn.DeleteKeyPair(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Key Pair (%s): %s", d.Id(), err)
	}

	return diags
}

// OpenSSHPublicKeysEqual returns whether or not two OpenSSH public key format strings represent the same key.
// Any key comment is ignored when comparing values.
func openSSHPublicKeysEqual(v1, v2 string) bool {
	key1, _, _, _, err := ssh.ParseAuthorizedKey([]byte(v1))

	if err != nil {
		return false
	}

	key2, _, _, _, err := ssh.ParseAuthorizedKey([]byte(v2))

	if err != nil {
		return false
	}

	return key1.Type() == key2.Type() && bytes.Equal(key1.Marshal(), key2.Marshal())
}
func keyPairARN(ctx context.Context, c *conns.AWSClient, keyName string) string {
	return c.RegionalARN(ctx, names.EC2, "key-pair/"+keyName)
}
