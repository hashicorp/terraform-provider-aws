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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
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
