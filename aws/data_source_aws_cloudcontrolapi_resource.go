package aws

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudcontrol/finder"
)

func dataSourceAwsCloudControlApiResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsCloudControlApiResourceRead,

		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}`), "must be three alphanumeric sections separated by double colons (::)"),
			},
			"type_version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsCloudControlApiResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudcontrolapiconn

	identifier := d.Get("identifier").(string)
	resourceDescription, err := finder.ResourceByID(ctx, conn,
		identifier,
		d.Get("type_name").(string),
		d.Get("type_version_id").(string),
		d.Get("role_arn").(string),
	)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Cloud Control API Resource (%s): %w", identifier, err))
	}

	d.SetId(aws.StringValue(resourceDescription.Identifier))

	d.Set("properties", resourceDescription.Properties)

	return nil
}
