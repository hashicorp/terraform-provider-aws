package aws

import (
	// "fmt"
	// "log"
	// "sort"
	// "time"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsSsoPermissionSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsoPermissionSetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]+$`), "must match [\\w+=,.@-]"),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"session_duration": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"relay_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"inline_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"managed_policies": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
				},
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsSsoPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}
