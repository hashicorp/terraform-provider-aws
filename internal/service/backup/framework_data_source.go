package backup

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceFramework() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFrameworkRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_parameter": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"scope": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_resource_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"compliance_resource_types": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"tags": tftags.TagsSchemaComputed(),
								},
							},
						},
					},
				},
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceFrameworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	resp, err := conn.DescribeFramework(&backup.DescribeFrameworkInput{
		FrameworkName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("Error getting Backup Framework: %w", err)
	}

	d.SetId(aws.StringValue(resp.FrameworkName))

	d.Set("arn", resp.FrameworkArn)
	d.Set("deployment_status", resp.DeploymentStatus)
	d.Set("description", resp.FrameworkDescription)
	d.Set("name", resp.FrameworkName)
	d.Set("status", resp.FrameworkStatus)

	if err := d.Set("creation_time", resp.CreationTime.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting creation_time: %s", err)
	}

	if err := d.Set("control", flattenFrameworkControls(resp.FrameworkControls)); err != nil {
		return fmt.Errorf("error setting control: %w", err)
	}

	tags, err := ListTags(conn, aws.StringValue(resp.FrameworkArn))

	if err != nil {
		return fmt.Errorf("error listing tags for Backup Framework (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
