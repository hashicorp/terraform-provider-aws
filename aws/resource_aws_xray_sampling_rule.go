package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsXraySamplingRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsXraySamplingRuleCreate,
		Read:   resourceAwsXraySamplingRuleRead,
		Update: resourceAwsXraySamplingRuleUpdate,
		Delete: resourceAwsXraySamplingRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(1, 128),
				ConflictsWith: []string{"rule_arn"},
			},
			"rule_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"priority": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"fixed_rate": {
				Type:     schema.TypeFloat,
				Required: true,
			},
			"reservoir_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
				Type:     schema.TypeString,
				Required: true,
			},
			"http_method": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url_path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsXraySamplingRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn
	samplingRule := &xray.SamplingRule{
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

	if v, ok := d.GetOk("name"); ok {
		samplingRule.RuleName = aws.String(v.(string))
	} else {
		samplingRule.RuleARN = aws.String(d.Get("rule_arn").(string))
	}

	if v, ok := d.GetOk("attributes"); ok {
		samplingRule.Attributes = stringMapToPointers(v.(map[string]interface{}))
	}

	params := &xray.CreateSamplingRuleInput{
		SamplingRule: samplingRule,
	}

	out, err := conn.CreateSamplingRule(params)
	if err != nil {
		return err
	}

	d.SetId(*out.SamplingRuleRecord.SamplingRule.RuleName)

	return resourceAwsXraySamplingRuleRead(d, meta)
}

func resourceAwsXraySamplingRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn
	params := &xray.GetSamplingRulesInput{}

	for {
		out, err := conn.GetSamplingRules(params)
		log.Printf("[DEBUG] Retrieved Rules: %s", out.SamplingRuleRecords)
		if err != nil {
			d.SetId("")
			return err
		}
		for _, samplingRuleRecord := range out.SamplingRuleRecords {
			sampingRule := samplingRuleRecord.SamplingRule
			if aws.StringValue(sampingRule.RuleName) == d.Id() {
				d.Set("name", sampingRule.RuleName)
				d.Set("rule_arn", sampingRule.RuleARN)
				d.Set("resource_arn", sampingRule.ResourceARN)
				d.Set("priority", sampingRule.Priority)
				d.Set("fixed_rate", sampingRule.FixedRate)
				d.Set("reservoir_size", sampingRule.ReservoirSize)
				d.Set("service_name", sampingRule.ServiceName)
				d.Set("service_type", sampingRule.ServiceType)
				d.Set("host", sampingRule.Host)
				d.Set("http_method", sampingRule.HTTPMethod)
				d.Set("url_path", sampingRule.URLPath)
				d.Set("version", sampingRule.Version)
				d.Set("attributes", aws.StringValueMap(sampingRule.Attributes))
				break
			}
		}
		if out.NextToken == nil {
			break
		}
		params.NextToken = out.NextToken
	}
	return nil
}

func resourceAwsXraySamplingRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsXraySamplingRuleRead(d, meta)
}

func resourceAwsXraySamplingRuleDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
