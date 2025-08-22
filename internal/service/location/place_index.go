// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_location_place_index", name="Map")
// @Tags(identifierAttribute="index_arn")
func ResourcePlaceIndex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePlaceIndexCreate,
		ReadWithoutTimeout:   resourcePlaceIndexRead,
		UpdateWithoutTimeout: resourcePlaceIndexUpdate,
		DeleteWithoutTimeout: resourcePlaceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_source_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intended_use": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.IntendedUse("SingleUse"),
							ValidateDiagFunc: enum.Validate[awstypes.IntendedUse](),
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"index_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePlaceIndexCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	input := &location.CreatePlaceIndexInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSource = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_source_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DataSourceConfiguration = expandDataSourceConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("index_name"); ok {
		input.IndexName = aws.String(v.(string))
	}

	output, err := conn.CreatePlaceIndex(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating place index: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating place index: empty result")
	}

	d.SetId(aws.ToString(output.IndexName))

	return append(diags, resourcePlaceIndexRead(ctx, d, meta)...)
}

func resourcePlaceIndexRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	input := &location.DescribePlaceIndexInput{
		IndexName: aws.String(d.Id()),
	}

	output, err := conn.DescribePlaceIndex(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Location Service Place Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Place Index (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Place Index (%s): empty response", d.Id())
	}

	d.Set(names.AttrCreateTime, aws.ToTime(output.CreateTime).Format(time.RFC3339))
	d.Set("data_source", output.DataSource)

	if output.DataSourceConfiguration != nil {
		d.Set("data_source_configuration", []any{flattenDataSourceConfiguration(output.DataSourceConfiguration)})
	} else {
		d.Set("data_source_configuration", nil)
	}

	d.Set(names.AttrDescription, output.Description)
	d.Set("index_arn", output.IndexArn)
	d.Set("index_name", output.IndexName)

	setTagsOut(ctx, output.Tags)

	d.Set("update_time", aws.ToTime(output.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourcePlaceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	if d.HasChanges("data_source_configuration", names.AttrDescription) {
		input := &location.UpdatePlaceIndexInput{
			IndexName: aws.String(d.Id()),
			// Deprecated but still required by the API
			PricingPlan: awstypes.PricingPlan("RequestBasedUsage"),
		}

		if v, ok := d.GetOk("data_source_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.DataSourceConfiguration = expandDataSourceConfiguration(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdatePlaceIndex(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service Place Index (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlaceIndexRead(ctx, d, meta)...)
}

func resourcePlaceIndexDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationClient(ctx)

	input := &location.DeletePlaceIndexInput{
		IndexName: aws.String(d.Id()),
	}

	_, err := conn.DeletePlaceIndex(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Place Index (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataSourceConfiguration(tfMap map[string]any) *awstypes.DataSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataSourceConfiguration{}

	if v, ok := tfMap["intended_use"].(string); ok && v != "" {
		apiObject.IntendedUse = awstypes.IntendedUse(v)
	}

	return apiObject
}

func flattenDataSourceConfiguration(apiObject *awstypes.DataSourceConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"intended_use": string(apiObject.IntendedUse),
	}

	return tfMap
}
