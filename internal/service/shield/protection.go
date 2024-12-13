// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
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

// @SDKResource("aws_shield_protection", name="Protection")
// @Tags(identifierAttribute="arn")
func resourceProtection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProtectionCreate,
		UpdateWithoutTimeout: resourceProtectionUpdate,
		ReadWithoutTimeout:   resourceProtectionRead,
		DeleteWithoutTimeout: resourceProtectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProtectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &shield.CreateProtectionInput{
		Name:        aws.String(name),
		ResourceArn: aws.String(d.Get(names.AttrResourceARN).(string)),
		Tags:        getTagsIn(ctx),
	}

	output, err := conn.CreateProtection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Shield Protection (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ProtectionId))

	return append(diags, resourceProtectionRead(ctx, d, meta)...)
}

func resourceProtectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	protection, err := findProtectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Shield Protection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Shield Protection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, protection.ProtectionArn)
	d.Set(names.AttrName, protection.Name)
	d.Set(names.AttrResourceARN, protection.ResourceArn)

	return diags
}

func resourceProtectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceProtectionRead(ctx, d, meta)...)
}

func resourceProtectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	log.Printf("[DEBUG] Deleting Shield Protection: %s", d.Id())
	_, err := conn.DeleteProtection(ctx, &shield.DeleteProtectionInput{
		ProtectionId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Shield Protection (%s): %s", d.Id(), err)
	}
	return diags
}

func findProtectionByID(ctx context.Context, conn *shield.Client, id string) (*types.Protection, error) {
	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(id),
	}

	return findProtection(ctx, conn, input)
}

func findProtectionByResourceARN(ctx context.Context, conn *shield.Client, arn string) (*types.Protection, error) {
	input := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(arn),
	}

	return findProtection(ctx, conn, input)
}

func findProtection(ctx context.Context, conn *shield.Client, input *shield.DescribeProtectionInput) (*types.Protection, error) {
	output, err := conn.DescribeProtection(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Protection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Protection, nil
}
