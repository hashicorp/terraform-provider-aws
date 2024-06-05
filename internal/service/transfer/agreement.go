// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_agreement", name="Agreement")
// @Tags(identifierAttribute="arn")
func resourceAgreement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAgreementCreate,
		ReadWithoutTimeout:   resourceAgreementRead,
		UpdateWithoutTimeout: resourceAgreementUpdate,
		DeleteWithoutTimeout: resourceAgreementDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"agreement_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_directory": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"local_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partner_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAgreementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID := d.Get("server_id").(string)
	input := &transfer.CreateAgreementInput{
		AccessRole:       aws.String(d.Get("access_role").(string)),
		BaseDirectory:    aws.String(d.Get("base_directory").(string)),
		LocalProfileId:   aws.String(d.Get("local_profile_id").(string)),
		PartnerProfileId: aws.String(d.Get("partner_profile_id").(string)),
		ServerId:         aws.String(serverID),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateAgreement(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Agreement: %s", err)
	}

	d.SetId(agreementCreateResourceID(serverID, aws.ToString(output.AgreementId)))

	return append(diags, resourceAgreementRead(ctx, d, meta)...)
}

func resourceAgreementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, agreementID, err := agreementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findAgreementByTwoPartKey(ctx, conn, serverID, agreementID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Agreement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Agreement (%s): %s", d.Id(), err)
	}

	d.Set("access_role", output.AccessRole)
	d.Set("agreement_id", output.AgreementId)
	d.Set(names.AttrARN, output.Arn)
	d.Set("base_directory", output.BaseDirectory)
	d.Set(names.AttrDescription, output.Description)
	d.Set("local_profile_id", output.LocalProfileId)
	d.Set("partner_profile_id", output.PartnerProfileId)
	d.Set("server_id", output.ServerId)
	d.Set(names.AttrStatus, output.Status)
	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceAgreementUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, agreementID, err := agreementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &transfer.UpdateAgreementInput{
			AgreementId: aws.String(agreementID),
			ServerId:    aws.String(serverID),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("base_directory") {
			input.BaseDirectory = aws.String(d.Get("base_directory").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("local_profile_id") {
			input.LocalProfileId = aws.String(d.Get("local_profile_id").(string))
		}

		if d.HasChange("partner_profile_id") {
			input.PartnerProfileId = aws.String(d.Get("partner_profile_id").(string))
		}

		_, err := conn.UpdateAgreement(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Agreement (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAgreementRead(ctx, d, meta)...)
}

func resourceAgreementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	serverID, agreementID, err := agreementParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Transfer Agreement: %s", d.Id())
	_, err = conn.DeleteAgreement(ctx, &transfer.DeleteAgreementInput{
		AgreementId: aws.String(agreementID),
		ServerId:    aws.String(serverID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Agreement (%s): %s", d.Id(), err)
	}

	return diags
}

const agreementResourceIDSeparator = "/"

func agreementCreateResourceID(serverID, agreementID string) string {
	parts := []string{serverID, agreementID}
	id := strings.Join(parts, agreementResourceIDSeparator)

	return id
}

func agreementParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, agreementResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVERID%[2]sAGREEMENTID", id, agreementResourceIDSeparator)
}

func findAgreementByTwoPartKey(ctx context.Context, conn *transfer.Client, serverID, agreementID string) (*awstypes.DescribedAgreement, error) {
	input := &transfer.DescribeAgreementInput{
		AgreementId: aws.String(agreementID),
		ServerId:    aws.String(serverID),
	}

	output, err := conn.DescribeAgreement(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Agreement == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Agreement, nil
}
