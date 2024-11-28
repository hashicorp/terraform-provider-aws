// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_profile", name="Profile")
// @Tags(identifierAttribute="arn")
func resourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"as2_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"certificate_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProfileType](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := &transfer.CreateProfileInput{
		As2Id:       aws.String(d.Get("as2_id").(string)),
		ProfileType: awstypes.ProfileType(d.Get("profile_type").(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("certificate_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.CertificateIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Profile: %s", err)
	}

	d.SetId(aws.ToString(output.ProfileId))

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Profile (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("as2_id", output.As2Id)
	d.Set("certificate_ids", output.CertificateIds)
	d.Set("profile_id", output.ProfileId)
	d.Set("profile_type", output.ProfileType)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &transfer.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("certificate_ids") {
			input.CertificateIds = flex.ExpandStringValueSet(d.Get("certificate_ids").(*schema.Set))
		}

		_, err := conn.UpdateProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	log.Printf("[DEBUG] Deleting Transfer Profile: %s", d.Id())
	_, err := conn.DeleteProfile(ctx, &transfer.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findProfileByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedProfile, error) {
	input := &transfer.DescribeProfileInput{
		ProfileId: aws.String(id),
	}

	output, err := conn.DescribeProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Profile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Profile, nil
}
