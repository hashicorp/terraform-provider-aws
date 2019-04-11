package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/hashicorp/terraform/helper/schema"
)

func saveTagsKinesisVideoStream(conn *kinesisvideo.KinesisVideo, d *schema.ResourceData, arn string) error {
	resp, err := conn.ListTagsForStream(&kinesisvideo.ListTagsForStreamInput{
		StreamARN: aws.String(arn),
	})
	if err != nil {
		return err
	}

	if err := d.Set("tags", tagsToMapGeneric(resp.Tags)); err != nil {
		return err
	}

	return nil
}

func setTagsKinesisVideoStream(conn *kinesisvideo.KinesisVideo, d *schema.ResourceData, arn string) error {
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		o := oraw.(map[string]interface{})
		n := nraw.(map[string]interface{})
		create, remove := diffTagsGeneric(o, n)
		if len(remove) > 0 {
			log.Printf("[DEBUG] Removing tags: %#v", remove)
			keys := make([]*string, 0, len(remove))
			for k := range remove {
				keys = append(keys, aws.String(k))
			}

			_, err := conn.UntagStream(&kinesisvideo.UntagStreamInput{
				StreamARN:  aws.String(arn),
				TagKeyList: keys,
			})
			if err != nil {
				return err
			}
		}
		if len(create) > 0 {
			log.Printf("[DEBUG] Creating tags: %#v", create)
			_, err := conn.TagStream(&kinesisvideo.TagStreamInput{
				StreamARN: aws.String(arn),
				Tags:      create,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
