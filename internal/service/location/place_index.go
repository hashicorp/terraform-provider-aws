// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      locationservice.IntendedUseSingleUse,
							ValidateFunc: validation.StringInSlice(locationservice.IntendedUse_Values(), false),
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
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePlaceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.CreatePlaceIndexInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("data_source"); ok {
		input.DataSource = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_source_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DataSourceConfiguration = expandDataSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("index_name"); ok {
		input.IndexName = aws.String(v.(string))
	}

	output, err := conn.CreatePlaceIndexWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating place index: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating place index: empty result")
	}

	d.SetId(aws.StringValue(output.IndexName))

	return append(diags, resourcePlaceIndexRead(ctx, d, meta)...)
}

func resourcePlaceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribePlaceIndexInput{
		IndexName: aws.String(d.Id()),
	}

	output, err := conn.DescribePlaceIndexWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
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

	d.Set(names.AttrCreateTime, aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("data_source", output.DataSource)

	if output.DataSourceConfiguration != nil {
		d.Set("data_source_configuration", []interface{}{flattenDataSourceConfiguration(output.DataSourceConfiguration)})
	} else {
		d.Set("data_source_configuration", nil)
	}

	d.Set(names.AttrDescription, output.Description)
	d.Set("index_arn", output.IndexArn)
	d.Set("index_name", output.IndexName)

	setTagsOut(ctx, output.Tags)

	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourcePlaceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	if d.HasChanges("data_source_configuration", names.AttrDescription) {
		input := &locationservice.UpdatePlaceIndexInput{
			IndexName: aws.String(d.Id()),
			// Deprecated but still required by the API
			PricingPlan: aws.String(locationservice.PricingPlanRequestBasedUsage),
		}

		if v, ok := d.GetOk("data_source_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DataSourceConfiguration = expandDataSourceConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdatePlaceIndexWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service Place Index (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePlaceIndexRead(ctx, d, meta)...)
}

func resourcePlaceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DeletePlaceIndexInput{
		IndexName: aws.String(d.Id()),
	}

	_, err := conn.DeletePlaceIndexWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Place Index (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataSourceConfiguration(tfMap map[string]interface{}) *locationservice.DataSourceConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &locationservice.DataSourceConfiguration{}

	if v, ok := tfMap["intended_use"].(string); ok && v != "" {
		apiObject.IntendedUse = aws.String(v)
	}

	return apiObject
}

func flattenDataSourceConfiguration(apiObject *locationservice.DataSourceConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IntendedUse; v != nil {
		tfMap["intended_use"] = aws.StringValue(v)
	}

	return tfMap
}
