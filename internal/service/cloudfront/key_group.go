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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_key_group", name="Key Group")
func resourceKeyGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyGroupCreate,
		ReadWithoutTimeout:   resourceKeyGroupRead,
		UpdateWithoutTimeout: resourceKeyGroupUpdate,
		DeleteWithoutTimeout: resourceKeyGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"items": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceKeyGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	apiObject := &awstypes.KeyGroupConfig{
		Items: flex.ExpandStringValueSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	input := &cloudfront.CreateKeyGroupInput{
		KeyGroupConfig: apiObject,
	}

	output, err := conn.CreateKeyGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Key Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.KeyGroup.Id))

	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findKeyGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Key Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Key Group (%s): %s", d.Id(), err)
	}

	keyGroupConfig := output.KeyGroup.KeyGroupConfig
	d.Set(names.AttrComment, keyGroupConfig.Comment)
	d.Set("etag", output.ETag)
	d.Set("items", keyGroupConfig.Items)
	d.Set(names.AttrName, keyGroupConfig.Name)

	return diags
}

func resourceKeyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	apiObject := &awstypes.KeyGroupConfig{
		Items: flex.ExpandStringValueSet(d.Get("items").(*schema.Set)),
		Name:  aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrComment); ok {
		apiObject.Comment = aws.String(v.(string))
	}

	input := &cloudfront.UpdateKeyGroupInput{
		Id:             aws.String(d.Id()),
		IfMatch:        aws.String(d.Get("etag").(string)),
		KeyGroupConfig: apiObject,
	}

	_, err := conn.UpdateKeyGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceKeyGroupRead(ctx, d, meta)...)
}

func resourceKeyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := &cloudfront.DeleteKeyGroupInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteKeyGroup(ctx, input)

	if errs.IsA[*awstypes.NoSuchResource](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Key Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findKeyGroupByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetKeyGroupOutput, error) {
	input := &cloudfront.GetKeyGroupInput{
		Id: aws.String(id),
	}

	output, err := conn.GetKeyGroup(ctx, input)

	if errs.IsA[*awstypes.NoSuchResource](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.KeyGroup == nil || output.KeyGroup.KeyGroupConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
