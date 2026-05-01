// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package b2bi

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/b2bi"
	awstypes "github.com/aws/aws-sdk-go-v2/service/b2bi/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_b2bi_profile", name="Profile")
// @Tags(identifierAttribute="profile_arn")
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
			"business_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},
			"email": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(5, 254),
			},
			"logging": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Logging](),
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},
			"phone": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(7, 22),
			},
			"profile_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &b2bi.CreateProfileInput{
		BusinessName: aws.String(d.Get("business_name").(string)),
		Logging:      awstypes.Logging(d.Get("logging").(string)),
		Name:         aws.String(name),
		Phone:        aws.String(d.Get("phone").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("email"); ok {
		input.Email = aws.String(v.(string))
	}

	output, err := conn.CreateProfile(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating B2BI Profile (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ProfileId))

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	output, err := findProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] B2BI Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading B2BI Profile (%s): %s", d.Id(), err)
	}

	d.Set("business_name", output.BusinessName)
	d.Set("email", output.Email)
	d.Set("logging", output.Logging)
	d.Set("log_group_name", output.LogGroupName)
	d.Set(names.AttrName, output.Name)
	d.Set("phone", output.Phone)
	d.Set("profile_arn", output.ProfileArn)
	d.Set("profile_id", output.ProfileId)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &b2bi.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("business_name") {
			input.BusinessName = aws.String(d.Get("business_name").(string))
		}

		if d.HasChange("email") {
			input.Email = aws.String(d.Get("email").(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange("phone") {
			input.Phone = aws.String(d.Get("phone").(string))
		}

		_, err := conn.UpdateProfile(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating B2BI Profile (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	log.Printf("[DEBUG] Deleting B2BI Profile: %s", d.Id())
	_, err := conn.DeleteProfile(ctx, &b2bi.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting B2BI Profile (%s): %s", d.Id(), err)
	}

	return diags
}

func findProfileByID(ctx context.Context, conn *b2bi.Client, id string) (*b2bi.GetProfileOutput, error) {
	input := &b2bi.GetProfileInput{
		ProfileId: aws.String(id),
	}

	output, err := conn.GetProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
