// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_profile", name="Profile")
// @Tags(identifierAttribute="profile_id")
func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(transfer.ProfileType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	input := &transfer.CreateProfileInput{
		As2Id:       aws.String(d.Get("as2_id").(string)),
		ProfileType: aws.String(d.Get("profile_type").(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("certificate_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.CertificateIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Profile: %s", err)
	}

	d.SetId(aws.StringValue(output.ProfileId))

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	output, err := FindProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Profile (%s): %s", d.Id(), err)
	}

	d.Set("as2_id", output.As2Id)
	d.Set("certificate_ids", aws.StringValueSlice(output.CertificateIds))
	d.Set("profile_id", output.ProfileId)
	d.Set("profile_type", output.ProfileType)
	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &transfer.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("certificate_ids") {
			input.CertificateIds = flex.ExpandStringSet(d.Get("certificate_ids").(*schema.Set))
		}

		_, err := conn.UpdateProfileWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	log.Printf("[DEBUG] Deleting Transfer Profile: %s", d.Id())
	_, err := conn.DeleteProfileWithContext(ctx, &transfer.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Profile (%s): %s", d.Id(), err)
	}

	return diags
}
