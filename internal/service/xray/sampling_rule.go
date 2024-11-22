// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_xray_sampling_rule", name="Sampling Rule")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray/types;types.SamplingRule")
func resourceSamplingRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSamplingRuleCreate,
		ReadWithoutTimeout:   resourceSamplingRuleRead,
		UpdateWithoutTimeout: resourceSamplingRuleUpdate,
		DeleteWithoutTimeout: resourceSamplingRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAttributes: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 32),
				},
			},
			"fixed_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 10),
			},
			names.AttrPriority: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 9999),
			},
			"reservoir_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			names.AttrServiceName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"service_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url_path": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			names.AttrVersion: {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func resourceSamplingRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	name := d.Get("rule_name").(string)
	samplingRule := &types.SamplingRule{
		FixedRate:     d.Get("fixed_rate").(float64),
		Host:          aws.String(d.Get("host").(string)),
		HTTPMethod:    aws.String(d.Get("http_method").(string)),
		Priority:      aws.Int32(int32(d.Get(names.AttrPriority).(int))),
		ReservoirSize: int32(d.Get("reservoir_size").(int)),
		ResourceARN:   aws.String(d.Get(names.AttrResourceARN).(string)),
		RuleName:      aws.String(name),
		ServiceName:   aws.String(d.Get(names.AttrServiceName).(string)),
		ServiceType:   aws.String(d.Get("service_type").(string)),
		URLPath:       aws.String(d.Get("url_path").(string)),
		Version:       aws.Int32(int32(d.Get(names.AttrVersion).(int))),
	}

	if v, ok := d.GetOk(names.AttrAttributes); ok && len(v.(map[string]interface{})) > 0 {
		samplingRule.Attributes = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	input := &xray.CreateSamplingRuleInput{
		SamplingRule: samplingRule,
		Tags:         getTagsIn(ctx),
	}

	output, err := conn.CreateSamplingRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Sampling Rule (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.SamplingRuleRecord.SamplingRule.RuleName))

	return append(diags, resourceSamplingRuleRead(ctx, d, meta)...)
}

func resourceSamplingRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	samplingRule, err := findSamplingRuleByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] XRay Sampling Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Sampling Rule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, samplingRule.RuleARN)
	d.Set(names.AttrAttributes, samplingRule.Attributes)
	d.Set("fixed_rate", samplingRule.FixedRate)
	d.Set("host", samplingRule.Host)
	d.Set("http_method", samplingRule.HTTPMethod)
	d.Set(names.AttrPriority, samplingRule.Priority)
	d.Set("reservoir_size", samplingRule.ReservoirSize)
	d.Set(names.AttrResourceARN, samplingRule.ResourceARN)
	d.Set("rule_name", samplingRule.RuleName)
	d.Set(names.AttrServiceName, samplingRule.ServiceName)
	d.Set("service_type", samplingRule.ServiceType)
	d.Set("url_path", samplingRule.URLPath)
	d.Set(names.AttrVersion, samplingRule.Version)

	return diags
}

func resourceSamplingRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		samplingRuleUpdate := &types.SamplingRuleUpdate{
			FixedRate:     aws.Float64(d.Get("fixed_rate").(float64)),
			Host:          aws.String(d.Get("host").(string)),
			HTTPMethod:    aws.String(d.Get("http_method").(string)),
			Priority:      aws.Int32(int32(d.Get(names.AttrPriority).(int))),
			ReservoirSize: aws.Int32(int32(d.Get("reservoir_size").(int))),
			ResourceARN:   aws.String(d.Get(names.AttrResourceARN).(string)),
			RuleName:      aws.String(d.Id()),
			ServiceName:   aws.String(d.Get(names.AttrServiceName).(string)),
			ServiceType:   aws.String(d.Get("service_type").(string)),
			URLPath:       aws.String(d.Get("url_path").(string)),
		}

		if d.HasChange(names.AttrAttributes) {
			samplingRuleUpdate.Attributes = flex.ExpandStringValueMap(d.Get(names.AttrAttributes).(map[string]interface{}))
		}

		input := &xray.UpdateSamplingRuleInput{
			SamplingRuleUpdate: samplingRuleUpdate,
		}

		_, err := conn.UpdateSamplingRule(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating XRay Sampling Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSamplingRuleRead(ctx, d, meta)...)
}

func resourceSamplingRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayClient(ctx)

	log.Printf("[INFO] Deleting XRay Sampling Rule: %s", d.Id())
	_, err := conn.DeleteSamplingRule(ctx, &xray.DeleteSamplingRuleInput{
		RuleName: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "Sampling rule does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting XRay Sampling Rule: %s", d.Id())
	}

	return diags
}

func findSamplingRuleByName(ctx context.Context, conn *xray.Client, name string) (*types.SamplingRule, error) {
	input := &xray.GetSamplingRulesInput{}

	pages := xray.NewGetSamplingRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.SamplingRuleRecords {
			if v := v.SamplingRule; v != nil && aws.ToString(v.RuleName) == name {
				return v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}
