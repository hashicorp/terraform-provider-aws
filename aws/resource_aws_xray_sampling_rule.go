package aws

import (
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
				Required:      true,
				ValidateFunc:  validation.StringLenBetween(1, 128),
				ConflictsWith: []string{"rule_arn"},
			},
			"rule_arn": {
				Type:          schema.TypeString,
				Required:      true,
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
	return nil
}

func resourceAwsXraySamplingRuleRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsXraySamplingRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsXraySamplingRuleRead(d, meta)
}

func resourceAwsXraySamplingRuleDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
