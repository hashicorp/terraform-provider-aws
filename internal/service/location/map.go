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

// @SDKResource("aws_location_map", name="Map")
// @Tags(identifierAttribute="map_arn")
func ResourceMap() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMapCreate,
		ReadWithoutTimeout:   resourceMapRead,
		UpdateWithoutTimeout: resourceMapUpdate,
		DeleteWithoutTimeout: resourceMapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"style": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 100),
						},
					},
				},
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"map_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.CreateMapInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Configuration = expandConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("map_name"); ok {
		input.MapName = aws.String(v.(string))
	}

	output, err := conn.CreateMapWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating map: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating map: empty result")
	}

	d.SetId(aws.StringValue(output.MapName))

	return append(diags, resourceMapRead(ctx, d, meta)...)
}

func resourceMapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DescribeMapInput{
		MapName: aws.String(d.Id()),
	}

	output, err := conn.DescribeMapWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Location Service Map (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Location Service Map (%s): empty response", d.Id())
	}

	if output.Configuration != nil {
		d.Set(names.AttrConfiguration, []interface{}{flattenConfiguration(output.Configuration)})
	} else {
		d.Set(names.AttrConfiguration, nil)
	}

	d.Set(names.AttrCreateTime, aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, output.Description)
	d.Set("map_arn", output.MapArn)
	d.Set("map_name", output.MapName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceMapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &locationservice.UpdateMapInput{
			MapName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateMapWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Location Service Map (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMapRead(ctx, d, meta)...)
}

func resourceMapDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	input := &locationservice.DeleteMapInput{
		MapName: aws.String(d.Id()),
	}

	_, err := conn.DeleteMapWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Map (%s): %s", d.Id(), err)
	}

	return diags
}

func expandConfiguration(tfMap map[string]interface{}) *locationservice.MapConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &locationservice.MapConfiguration{}

	if v, ok := tfMap["style"].(string); ok && v != "" {
		apiObject.Style = aws.String(v)
	}

	return apiObject
}

func flattenConfiguration(apiObject *locationservice.MapConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Style; v != nil {
		tfMap["style"] = aws.StringValue(v)
	}

	return tfMap
}
