package rekognition

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceCollection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCollectionRead,

		Schema: map[string]*schema.Schema{
			"collection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"collection_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"face_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"face_model_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCollectionRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("collection_id").(string))
	conn := meta.(*conns.AWSClient).RekognitionConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &rekognition.DescribeCollectionInput{
		CollectionId: aws.String(d.Id()),
	}

	output, err := conn.DescribeCollection(input)
	if err != nil {
		return fmt.Errorf("error getting Rekognition Collection (%s): %w", d.Id(), err)
	}
	if output == nil {
		return fmt.Errorf("error getting Rekognition Collection (%s): empty response", d.Id())
	}
	d.Set("collection_id", d.Id())
	d.Set("collection_arn", output.CollectionARN)
	d.Set("face_count", output.FaceCount)
	d.Set("face_model_version", output.FaceModelVersion)
	tags, err := ListTags(conn, d.Get("collection_arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Rekognition Collection (%s): %w", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
