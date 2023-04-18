package xray

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_xray_sampling_rule", name="Sampling Rule")
// @Tags(identifierAttribute="arn")
func ResourceSamplingRule() *schema.Resource {
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
			"rule_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 9999),
			},
			"fixed_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"reservoir_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"service_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
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
			"url_path": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"version": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 32),
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSamplingRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	samplingRule := &xray.SamplingRule{
		RuleName:      aws.String(d.Get("rule_name").(string)),
		ResourceARN:   aws.String(d.Get("resource_arn").(string)),
		Priority:      aws.Int64(int64(d.Get("priority").(int))),
		FixedRate:     aws.Float64(d.Get("fixed_rate").(float64)),
		ReservoirSize: aws.Int64(int64(d.Get("reservoir_size").(int))),
		ServiceName:   aws.String(d.Get("service_name").(string)),
		ServiceType:   aws.String(d.Get("service_type").(string)),
		Host:          aws.String(d.Get("host").(string)),
		HTTPMethod:    aws.String(d.Get("http_method").(string)),
		URLPath:       aws.String(d.Get("url_path").(string)),
		Version:       aws.Int64(int64(d.Get("version").(int))),
	}

	if v, ok := d.GetOk("attributes"); ok {
		samplingRule.Attributes = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	params := &xray.CreateSamplingRuleInput{
		SamplingRule: samplingRule,
		Tags:         GetTagsIn(ctx),
	}

	out, err := conn.CreateSamplingRuleWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Sampling Rule: %s", err)
	}

	d.SetId(aws.StringValue(out.SamplingRuleRecord.SamplingRule.RuleName))

	return append(diags, resourceSamplingRuleRead(ctx, d, meta)...)
}

func resourceSamplingRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	samplingRule, err := GetSamplingRule(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Sampling Rule (%s): %s", d.Id(), err)
	}

	if samplingRule == nil {
		log.Printf("[WARN] XRay Sampling Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(samplingRule.RuleARN)
	d.Set("rule_name", samplingRule.RuleName)
	d.Set("resource_arn", samplingRule.ResourceARN)
	d.Set("priority", samplingRule.Priority)
	d.Set("fixed_rate", samplingRule.FixedRate)
	d.Set("reservoir_size", samplingRule.ReservoirSize)
	d.Set("service_name", samplingRule.ServiceName)
	d.Set("service_type", samplingRule.ServiceType)
	d.Set("host", samplingRule.Host)
	d.Set("http_method", samplingRule.HTTPMethod)
	d.Set("url_path", samplingRule.URLPath)
	d.Set("version", samplingRule.Version)
	d.Set("attributes", aws.StringValueMap(samplingRule.Attributes))
	d.Set("arn", arn)

	return diags
}

func resourceSamplingRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	if d.HasChanges("attributes", "priority", "fixed_rate", "reservoir_size", "service_name", "service_type",
		"host", "http_method", "url_path", "resource_arn") {
		samplingRuleUpdate := &xray.SamplingRuleUpdate{
			RuleName:      aws.String(d.Id()),
			Priority:      aws.Int64(int64(d.Get("priority").(int))),
			FixedRate:     aws.Float64(d.Get("fixed_rate").(float64)),
			ReservoirSize: aws.Int64(int64(d.Get("reservoir_size").(int))),
			ServiceName:   aws.String(d.Get("service_name").(string)),
			ServiceType:   aws.String(d.Get("service_type").(string)),
			Host:          aws.String(d.Get("host").(string)),
			HTTPMethod:    aws.String(d.Get("http_method").(string)),
			URLPath:       aws.String(d.Get("url_path").(string)),
			ResourceARN:   aws.String(d.Get("resource_arn").(string)),
		}

		if d.HasChange("attributes") {
			attributes := map[string]*string{}
			if v, ok := d.GetOk("attributes"); ok {
				if m, ok := v.(map[string]interface{}); ok {
					attributes = flex.ExpandStringMap(m)
				}
			}
			samplingRuleUpdate.Attributes = attributes
		}

		params := &xray.UpdateSamplingRuleInput{
			SamplingRuleUpdate: samplingRuleUpdate,
		}

		_, err := conn.UpdateSamplingRuleWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating XRay Sampling Rule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSamplingRuleRead(ctx, d, meta)...)
}

func resourceSamplingRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	log.Printf("[INFO] Deleting XRay Sampling Rule: %s", d.Id())
	_, err := conn.DeleteSamplingRuleWithContext(ctx, &xray.DeleteSamplingRuleInput{
		RuleName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, xray.ErrCodeInvalidRequestException, "Sampling rule does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting XRay Sampling Rule: %s", d.Id())
	}

	return diags
}

func GetSamplingRule(ctx context.Context, conn *xray.XRay, ruleName string) (*xray.SamplingRule, error) {
	params := &xray.GetSamplingRulesInput{}
	for {
		out, err := conn.GetSamplingRulesWithContext(ctx, params)
		if err != nil {
			return nil, err
		}
		for _, samplingRuleRecord := range out.SamplingRuleRecords {
			samplingRule := samplingRuleRecord.SamplingRule
			if aws.StringValue(samplingRule.RuleName) == ruleName {
				return samplingRule, nil
			}
		}
		if aws.StringValue(out.NextToken) == "" {
			break
		}
		params.NextToken = out.NextToken
	}
	return nil, nil
}
