// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_origin_access_control", name="Origin Access Control")
func resourceOriginAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOriginAccessControlCreate,
		ReadWithoutTimeout:   resourceOriginAccessControlRead,
		UpdateWithoutTimeout: resourceOriginAccessControlUpdate,
		DeleteWithoutTimeout: resourceOriginAccessControlDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Managed by Terraform",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"origin_access_control_origin_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OriginAccessControlOriginTypes](),
			},
			"signing_behavior": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OriginAccessControlSigningBehaviors](),
			},
			"signing_protocol": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OriginAccessControlSigningProtocols](),
			},
		},
	}
}

func resourceOriginAccessControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloudfront.CreateOriginAccessControlInput{
		OriginAccessControlConfig: &awstypes.OriginAccessControlConfig{
			Description:                   aws.String(d.Get(names.AttrDescription).(string)),
			Name:                          aws.String(name),
			OriginAccessControlOriginType: awstypes.OriginAccessControlOriginTypes(d.Get("origin_access_control_origin_type").(string)),
			SigningBehavior:               awstypes.OriginAccessControlSigningBehaviors(d.Get("signing_behavior").(string)),
			SigningProtocol:               awstypes.OriginAccessControlSigningProtocols(d.Get("signing_protocol").(string)),
		},
	}

	output, err := conn.CreateOriginAccessControl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Origin Access Control (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.OriginAccessControl.Id))

	return append(diags, resourceOriginAccessControlRead(ctx, d, meta)...)
}

func resourceOriginAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findOriginAccessControlByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Origin Access Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Access Control (%s): %s", d.Id(), err)
	}

	config := output.OriginAccessControl.OriginAccessControlConfig
	d.Set(names.AttrDescription, config.Description)
	d.Set("etag", output.ETag)
	d.Set(names.AttrName, config.Name)
	d.Set("origin_access_control_origin_type", config.OriginAccessControlOriginType)
	d.Set("signing_behavior", config.SigningBehavior)
	d.Set("signing_protocol", config.SigningProtocol)

	return diags
}

func resourceOriginAccessControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.UpdateOriginAccessControlInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
		OriginAccessControlConfig: &awstypes.OriginAccessControlConfig{
			Description:                   aws.String(d.Get(names.AttrDescription).(string)),
			Name:                          aws.String(d.Get(names.AttrName).(string)),
			OriginAccessControlOriginType: awstypes.OriginAccessControlOriginTypes(d.Get("origin_access_control_origin_type").(string)),
			SigningBehavior:               awstypes.OriginAccessControlSigningBehaviors(d.Get("signing_behavior").(string)),
			SigningProtocol:               awstypes.OriginAccessControlSigningProtocols(d.Get("signing_protocol").(string)),
		},
	}

	_, err := conn.UpdateOriginAccessControl(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Origin Access Control (%s): %s", d.Id(), err)
	}

	return append(diags, resourceOriginAccessControlRead(ctx, d, meta)...)
}

func resourceOriginAccessControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[INFO] Deleting CloudFront Origin Access Control: %s", d.Id())
	_, err := conn.DeleteOriginAccessControl(ctx, &cloudfront.DeleteOriginAccessControlInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchOriginAccessControl](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Origin Access Control (%s): %s", d.Id(), err)
	}

	return diags
}

func findOriginAccessControlByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetOriginAccessControlOutput, error) {
	input := &cloudfront.GetOriginAccessControlInput{
		Id: aws.String(id),
	}

	output, err := conn.GetOriginAccessControl(ctx, input)

	if errs.IsA[*awstypes.NoSuchOriginAccessControl](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OriginAccessControl == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
