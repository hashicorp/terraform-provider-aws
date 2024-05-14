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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_function", name="Function")
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_value_store_associations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"live_stage_etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"runtime": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FunctionRuntime](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	functionName := d.Get(names.AttrName).(string)
	input := &cloudfront.CreateFunctionInput{
		FunctionCode: []byte(d.Get("code").(string)),
		FunctionConfig: &awstypes.FunctionConfig{
			Comment: aws.String(d.Get(names.AttrComment).(string)),
			Runtime: awstypes.FunctionRuntime(d.Get("runtime").(string)),
		},
		Name: aws.String(functionName),
	}

	if v, ok := d.GetOk("key_value_store_associations"); ok {
		input.FunctionConfig.KeyValueStoreAssociations = expandKeyValueStoreAssociations(v.(*schema.Set).List())
	}

	output, err := conn.CreateFunction(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Function (%s): %s", functionName, err)
	}

	d.SetId(aws.ToString(output.FunctionSummary.Name))

	if d.Get("publish").(bool) {
		input := &cloudfront.PublishFunctionInput{
			Name:    aws.String(d.Id()),
			IfMatch: output.ETag,
		}

		_, err := conn.PublishFunction(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing CloudFront Function (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	outputDF, err := findFunctionByTwoPartKey(ctx, conn, d.Id(), awstypes.FunctionStageDevelopment)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Function (%s) DEVELOPMENT stage: %s", d.Id(), err)
	}

	d.Set(names.AttrARN, outputDF.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set(names.AttrComment, outputDF.FunctionSummary.FunctionConfig.Comment)
	d.Set("etag", outputDF.ETag)
	if err := d.Set("key_value_store_associations", flattenKeyValueStoreAssociations(outputDF.FunctionSummary.FunctionConfig.KeyValueStoreAssociations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting key_value_store_associations: %s", err)
	}
	d.Set(names.AttrName, outputDF.FunctionSummary.Name)
	d.Set("runtime", outputDF.FunctionSummary.FunctionConfig.Runtime)
	d.Set(names.AttrStatus, outputDF.FunctionSummary.Status)

	outputGF, err := conn.GetFunction(ctx, &cloudfront.GetFunctionInput{
		Name:  aws.String(d.Id()),
		Stage: awstypes.FunctionStageDevelopment,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Function (%s) DEVELOPMENT stage code: %s", d.Id(), err)
	}

	d.Set("code", string(outputGF.FunctionCode))

	outputDF, err = findFunctionByTwoPartKey(ctx, conn, d.Id(), awstypes.FunctionStageLive)

	if tfresource.NotFound(err) {
		d.Set("live_stage_etag", "")
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Function (%s) LIVE stage: %s", d.Id(), err)
	} else {
		d.Set("live_stage_etag", outputDF.ETag)
	}

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	etag := d.Get("etag").(string)

	if d.HasChanges("code", names.AttrComment, "key_value_store_associations", "runtime") {
		input := &cloudfront.UpdateFunctionInput{
			FunctionCode: []byte(d.Get("code").(string)),
			FunctionConfig: &awstypes.FunctionConfig{
				Comment: aws.String(d.Get(names.AttrComment).(string)),
				Runtime: awstypes.FunctionRuntime(d.Get("runtime").(string)),
			},
			IfMatch: aws.String(etag),
			Name:    aws.String(d.Id()),
		}

		if v, ok := d.GetOk("key_value_store_associations"); ok {
			input.FunctionConfig.KeyValueStoreAssociations = expandKeyValueStoreAssociations(v.(*schema.Set).List())
		}

		output, err := conn.UpdateFunction(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFront Function (%s): %s", d.Id(), err)
		}

		etag = aws.ToString(output.ETag)
	}

	if d.Get("publish").(bool) {
		input := &cloudfront.PublishFunctionInput{
			IfMatch: aws.String(etag),
			Name:    aws.String(d.Id()),
		}

		_, err := conn.PublishFunction(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing CloudFront Function (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[INFO] Deleting CloudFront Function: %s", d.Id())
	_, err := conn.DeleteFunction(ctx, &cloudfront.DeleteFunctionInput{
		IfMatch: aws.String(d.Get("etag").(string)),
		Name:    aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchFunctionExists](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Function (%s): %s", d.Id(), err)
	}

	return diags
}

func findFunctionByTwoPartKey(ctx context.Context, conn *cloudfront.Client, name string, stage awstypes.FunctionStage) (*cloudfront.DescribeFunctionOutput, error) {
	input := &cloudfront.DescribeFunctionInput{
		Name:  aws.String(name),
		Stage: stage,
	}

	output, err := conn.DescribeFunction(ctx, input)

	if errs.IsA[*awstypes.NoSuchFunctionExists](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FunctionSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandKeyValueStoreAssociations(tfList []interface{}) *awstypes.KeyValueStoreAssociations {
	if len(tfList) == 0 {
		return nil
	}

	items := tfslices.ApplyToAll(tfList, func(v interface{}) awstypes.KeyValueStoreAssociation {
		return awstypes.KeyValueStoreAssociation{
			KeyValueStoreARN: aws.String(v.(string)),
		}
	})

	return &awstypes.KeyValueStoreAssociations{
		Items:    items,
		Quantity: aws.Int32(int32(len(items))),
	}
}

func flattenKeyValueStoreAssociations(input *awstypes.KeyValueStoreAssociations) []string {
	if input == nil {
		return nil
	}

	return tfslices.ApplyToAll(input.Items, func(v awstypes.KeyValueStoreAssociation) string {
		return aws.ToString(v.KeyValueStoreARN)
	})
}
