// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_origin_access_identity", name="Origin Access Identity")
func resourceOriginAccessIdentity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOriginAccessIdentityCreate,
		ReadWithoutTimeout:   resourceOriginAccessIdentityRead,
		UpdateWithoutTimeout: resourceOriginAccessIdentityUpdate,
		DeleteWithoutTimeout: resourceOriginAccessIdentityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_access_identity_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_canonical_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOriginAccessIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.CreateCloudFrontOriginAccessIdentityInput{
		CloudFrontOriginAccessIdentityConfig: expandCloudFrontOriginAccessIdentityConfig(d),
	}

	output, err := conn.CreateCloudFrontOriginAccessIdentity(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Origin Access Identity: %s", err)
	}

	d.SetId(aws.ToString(output.CloudFrontOriginAccessIdentity.Id))

	return append(diags, resourceOriginAccessIdentityRead(ctx, d, meta)...)
}

func resourceOriginAccessIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findOriginAccessIdentityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Origin Access Identity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Origin Access Identity (%s): %s", d.Id(), err)
	}

	apiObject := output.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig
	d.Set("caller_reference", apiObject.CallerReference)
	d.Set("cloudfront_access_identity_path", "origin-access-identity/cloudfront/"+d.Id())
	d.Set(names.AttrComment, apiObject.Comment)
	d.Set("etag", output.ETag)
	d.Set("iam_arn", originAccessIdentityARN(meta.(*conns.AWSClient), d.Id()))
	d.Set("s3_canonical_user_id", output.CloudFrontOriginAccessIdentity.S3CanonicalUserId)

	return diags
}

func resourceOriginAccessIdentityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.UpdateCloudFrontOriginAccessIdentityInput{
		CloudFrontOriginAccessIdentityConfig: expandCloudFrontOriginAccessIdentityConfig(d),
		Id:                                   aws.String(d.Id()),
		IfMatch:                              aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateCloudFrontOriginAccessIdentity(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Origin Access Identity (%s): %s", d.Id(), err)
	}

	return append(diags, resourceOriginAccessIdentityRead(ctx, d, meta)...)
}

func resourceOriginAccessIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	_, err := conn.DeleteCloudFrontOriginAccessIdentity(ctx, &cloudfront.DeleteCloudFrontOriginAccessIdentityInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if errs.IsA[*awstypes.NoSuchCloudFrontOriginAccessIdentity](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Origin Access Identity (%s): %s", d.Id(), err)
	}

	return diags
}

func findOriginAccessIdentityByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetCloudFrontOriginAccessIdentityOutput, error) {
	input := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
		Id: aws.String(id),
	}

	output, err := conn.GetCloudFrontOriginAccessIdentity(ctx, input)

	if errs.IsA[*awstypes.NoSuchCloudFrontOriginAccessIdentity](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CloudFrontOriginAccessIdentity == nil || output.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandCloudFrontOriginAccessIdentityConfig(d *schema.ResourceData) *awstypes.CloudFrontOriginAccessIdentityConfig { // nosemgrep:ci.cloudfront-in-func-name
	apiObject := &awstypes.CloudFrontOriginAccessIdentityConfig{
		Comment: aws.String(d.Get(names.AttrComment).(string)),
	}

	// This sets CallerReference if it's still pending computation (ie: new resource)
	if v, ok := d.GetOk("caller_reference"); !ok {
		apiObject.CallerReference = aws.String(id.UniqueId())
	} else {
		apiObject.CallerReference = aws.String(v.(string))
	}

	return apiObject
}

func originAccessIdentityARN(c *conns.AWSClient, originAccessControlID string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "iam",
		AccountID: "cloudfront",
		Resource:  "user/CloudFront Origin Access Identity " + originAccessControlID,
	}.String()
}
