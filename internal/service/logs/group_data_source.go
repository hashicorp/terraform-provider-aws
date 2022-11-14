package logs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"retention_in_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	logGroup, err := FindLogGroupByName(conn, name)

	if err != nil {
		return fmt.Errorf("reading CloudWatch Logs Log Group (%s): %w", name, err)
	}

	d.SetId(name)
	d.Set("arn", TrimLogGroupARNWildcardSuffix(aws.StringValue(logGroup.Arn)))
	d.Set("creation_time", logGroup.CreationTime)
	d.Set("kms_key_id", logGroup.KmsKeyId)
	d.Set("retention_in_days", logGroup.RetentionInDays)

	tags, err := ListTags(conn, name)

	if err != nil {
		return fmt.Errorf("listing tags for CloudWatch Logs Log Group (%s): %w", name, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
