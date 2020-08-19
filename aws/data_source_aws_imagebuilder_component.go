package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsImageBuilderComponent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsImageBuilderComponentRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"change_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"tags": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"semantic_version": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsImageBuilderComponentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	componentArn := d.Get("arn").(string)

	params := &imagebuilder.GetComponentInput{
		ComponentBuildVersionArn: aws.String(componentArn),
	}

	resp, err := conn.GetComponent(params)

	if err != nil {
		return fmt.Errorf("Error retrieving Component: %s", err)
	}

	d.SetId(*resp.Component.Arn)

	if err := d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(resp.Component.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("change_description", resp.Component.ChangeDescription)
	d.Set("data", resp.Component.Data)
	d.Set("date_created", resp.Component.DateCreated)
	d.Set("description", resp.Component.Description)
	d.Set("encrypted", resp.Component.Encrypted)
	d.Set("kms_key_id", resp.Component.KmsKeyId)
	d.Set("name", resp.Component.Name)
	d.Set("owner", resp.Component.Owner)
	d.Set("platform", resp.Component.Platform)
	d.Set("type", resp.Component.Type)
	d.Set("semantic_version", resp.Component.Version)

	return nil
}
