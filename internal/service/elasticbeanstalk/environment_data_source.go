package elasticbeanstalk

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"application": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"solution_stack_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tier": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn

	// Get the name and description
	name := d.Get("name").(string)

	resp, err := conn.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentNames: []*string{aws.String(name)},
	})
	if err != nil {
		return fmt.Errorf("Error describing Environments (%s): %w", name, err)
	}

	if len(resp.Environments) > 1 || len(resp.Environments) < 1 {
		return fmt.Errorf("Error %d Environments matched, expected 1", len(resp.Environments))
	}

	app := resp.Environments[0]

	d.SetId(name)
	d.Set("arn", app.EnvironmentArn)
	d.Set("name", app.EnvironmentName)
	d.Set("application", app.ApplicationName)
	d.Set("version_label", app.VersionLabel)
	d.Set("solution_stack_name", app.SolutionStackName)
	d.Set("platform_arn", app.PlatformArn)
	d.Set("template_name", app.TemplateName)
	d.Set("description", app.Description)
	d.Set("endpoint_url", app.EndpointURL)
	d.Set("cname", app.CNAME)
	d.Set("resources", app.Resources)
	d.Set("tier", app.Tier)

	return nil
}
