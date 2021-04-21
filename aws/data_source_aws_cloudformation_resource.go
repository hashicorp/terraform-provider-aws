package aws

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsCloudFormationResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsCloudformationResourceRead,

		Schema: map[string]*schema.Schema{
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_model": {
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

func dataSourceAwsCloudformationResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.GetResourceInput{}

	if v, ok := d.GetOk("identifier"); ok {
		input.Identifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_name"); ok {
		input.TypeName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type_version_id"); ok {
		input.TypeVersionId = aws.String(v.(string))
	}

	output, err := conn.GetResourceWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Resource: %w", err))
	}

	if output == nil || output.ResourceDescription == nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Resource: empty response"))
	}

	d.SetId(aws.StringValue(output.ResourceDescription.Identifier))

	d.Set("resource_model", output.ResourceDescription.ResourceModel)

	return nil
}
