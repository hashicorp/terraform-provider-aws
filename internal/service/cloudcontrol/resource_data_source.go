package cloudcontrol

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceResourceRead,

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

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudControlClient()

	identifier := d.Get("identifier").(string)
	typeName := d.Get("type_name").(string)
	resourceDescription, err := FindResource(ctx, conn,
		identifier,
		typeName,
		d.Get("type_version_id").(string),
		d.Get("role_arn").(string),
	)

	if err != nil {
		return diag.Errorf("reading Cloud Control API (%s) Resource (%s): %s", typeName, identifier, err)
	}

	d.SetId(aws.ToString(resourceDescription.Identifier))

	d.Set("properties", resourceDescription.Properties)

	return nil
}
