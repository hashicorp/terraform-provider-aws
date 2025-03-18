// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_macsec_key_association", name="MACSec Key Association")
func resourceMacSecKeyAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacSecKeyAssociatioCreate,
		ReadWithoutTimeout:   resourceMacSecKeyAssociationRead,
		DeleteWithoutTimeout: resourceMacSecKeyAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cak": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"ckn"},
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Fa-f]{64}$`), "Must be 64-character hex code string"),
			},
			"ckn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				AtLeastOneOf: []string{"ckn", "secret_arn"},
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9A-Fa-f]{64}$`), "Must be 64-character hex code string"),
			},
			names.AttrConnectionID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"ckn", "secret_arn"},
				ValidateFunc: verify.ValidARN,
			},
			"start_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacSecKeyAssociatioCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	connectionID := d.Get(names.AttrConnectionID).(string)
	input := &directconnect.AssociateMacSecKeyInput{
		ConnectionId: aws.String(connectionID),
	}

	if v, ok := d.GetOk("ckn"); ok {
		input.Cak = aws.String(d.Get("cak").(string))
		input.Ckn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("secret_arn"); ok {
		input.SecretARN = aws.String(v.(string))
	}

	output, err := conn.AssociateMacSecKey(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MACSec Key Association with Direct Connect Connection (%s): %s", connectionID, err)
	}

	var secretARN string
	for _, key := range output.MacSecKeys {
		secretARN = aws.ToString(key.SecretARN)
	}

	d.SetId(macSecKeyAssociationCreateResourceID(secretARN, connectionID))

	return append(diags, resourceMacSecKeyAssociationRead(ctx, d, meta)...)
}

func resourceMacSecKeyAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	secretARN, connectionID, err := macSecKeyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	key, err := findMacSecKeyByTwoPartKey(ctx, conn, connectionID, secretARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect MACSec Key Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect MACSec Key Association (%s): %s", d.Id(), err)
	}

	d.Set("ckn", key.Ckn)
	d.Set(names.AttrConnectionID, connectionID)
	d.Set("secret_arn", key.SecretARN)
	d.Set("start_on", key.StartOn)
	d.Set(names.AttrState, key.State)

	return diags
}

func resourceMacSecKeyAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	secretARN, connectionID, err := macSecKeyAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Direct Connect MACSec Key Association: %s", d.Id())
	input := directconnect.DisassociateMacSecKeyInput{
		ConnectionId: aws.String(connectionID),
		SecretARN:    aws.String(secretARN),
	}
	_, err = conn.DisassociateMacSecKey(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MACSec Key Association (%s): %s", d.Id(), err)
	}

	return diags
}

const macSecKeyAssociationResourceIDSeparator = "_"

func macSecKeyAssociationCreateResourceID(secretARN, connectionID string) string {
	parts := []string{secretARN, connectionID}
	id := strings.Join(parts, macSecKeyAssociationResourceIDSeparator)

	return id
}

func macSecKeyAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, macSecKeyAssociationResourceIDSeparator, 2)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected secretArn%[2]sconnectionId", id, macSecKeyAssociationResourceIDSeparator)
}

func findMacSecKeyByTwoPartKey(ctx context.Context, conn *directconnect.Client, connectionID, secretARN string) (*awstypes.MacSecKey, error) {
	output, err := findConnectionByID(ctx, conn, connectionID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.MacSecKeys, func(v awstypes.MacSecKey) bool {
		return aws.ToString(v.SecretARN) == secretARN
	}))
}
