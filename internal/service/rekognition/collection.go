package rekognition

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceCollectionCreate,
		Read:   resourceCollectionRead,
		Update: resourceCollectionUpdate,
		Delete: resourceCollectionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"collection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCollectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RekognitionConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := rekognition.CreateCollectionInput{
		CollectionId: aws.String(d.Get("collection_id").(string)),
		Tags:         Tags(tags.IgnoreAWS()),
	}

	collection, err := conn.CreateCollection(&input)
	if err != nil {
		return fmt.Errorf("error creating rekognition collection: %s", err)
	}

	if collection == nil {
		return fmt.Errorf("error getting Rekognition Collection (%s): empty response", d.Id())
	}

	d.SetId(d.Get("collection_id").(string))

	return resourceCollectionRead(d, meta)
}

func resourceCollectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RekognitionConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	collection, err := conn.DescribeCollection(&rekognition.DescribeCollectionInput{
		CollectionId: aws.String(d.Id()),
	})

	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, rekognition.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Rekognition Collection (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading rekognition collection (%s): %s", d.Id(), err)
	}

	if collection == nil {
		return fmt.Errorf("error getting Rekognition Collection (%s): empty response", d.Id())
	}

	arn := aws.StringValue(collection.CollectionARN)
	d.Set("collection_id", d.Id())
	d.Set("collection_arn", arn)
	d.Set("face_count", collection.FaceCount)
	d.Set("face_model_version", collection.FaceModelVersion)

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCollectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RekognitionConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("collection_arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceCollectionRead(d, meta)
}

func resourceCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RekognitionConn()

	input := rekognition.DeleteCollectionInput{
		CollectionId: aws.String(d.Id()),
	}

	output, err := conn.DeleteCollection(&input)
	if err != nil {
		return fmt.Errorf("error deleting Rekognition Collection (%s): %s", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Rekognition Collection (%s): empty response", d.Id())
	}

	return nil
}
