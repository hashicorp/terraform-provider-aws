// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_function", name="Function")
func dataSourceFunction() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFunctionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_value_store_associations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStage: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FunctionStage](),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	stage := awstypes.FunctionStage(d.Get(names.AttrStage).(string))
	outputDF, err := findFunctionByTwoPartKey(ctx, conn, name, stage)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Function (%s) %s stage: %s", name, stage, err)
	}

	d.SetId(aws.ToString(outputDF.FunctionSummary.Name))
	d.Set(names.AttrARN, outputDF.FunctionSummary.FunctionMetadata.FunctionARN)
	d.Set(names.AttrComment, outputDF.FunctionSummary.FunctionConfig.Comment)
	d.Set("etag", outputDF.ETag)
	if err := d.Set("key_value_store_associations", flattenKeyValueStoreAssociations(outputDF.FunctionSummary.FunctionConfig.KeyValueStoreAssociations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting key_value_store_associations: %s", err)
	}
	d.Set("last_modified_time", outputDF.FunctionSummary.FunctionMetadata.LastModifiedTime.Format(time.RFC3339))
	d.Set(names.AttrName, outputDF.FunctionSummary.Name)
	d.Set("runtime", outputDF.FunctionSummary.FunctionConfig.Runtime)
	d.Set(names.AttrStatus, outputDF.FunctionSummary.Status)

	outputGF, err := conn.GetFunction(ctx, &cloudfront.GetFunctionInput{
		Name:  aws.String(name),
		Stage: stage,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Function (%s) %s stage code: %s", name, stage, err)
	}

	d.Set("code", string(outputGF.FunctionCode))

	return diags
}
