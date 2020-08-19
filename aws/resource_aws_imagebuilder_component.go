package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
)

func resourceAwsImageBuilderComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderComponentCreate,
		Read:   resourceAwsImageBuilderComponentRead,
		Update: resourceAwsImageBuilderComponentUpdate,
		Delete: resourceAwsImageBuilderComponentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"change_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"data": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(1, 16000),
				ConflictsWith: []string{"uri"},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 126),
			},
			"platform": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Windows", "Linux"}, false),
			},
			"semantic_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"tags": tagsSchema(),
			"uri": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"data"},
			},
		},
	}
}

func resourceAwsImageBuilderComponentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateComponentInput{
		ClientToken:     aws.String(resource.UniqueId()),
		Name:            aws.String(d.Get("name").(string)),
		Platform:        aws.String(d.Get("platform").(string)),
		SemanticVersion: aws.String(d.Get("semantic_version").(string)),
	}

	if v, ok := d.GetOk("change_description"); ok {
		input.SetChangeDescription(v.(string))
	}
	data, dataok := d.GetOk("data")
	if dataok {
		input.SetData(data.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		input.SetKmsKeyId(v.(string))
	}
	uri, uriok := d.GetOk("uri")
	if uriok {
		input.SetUri(uri.(string))
	}
	if !uriok && !dataok {
		return errors.New("one of data or uri must be set")
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags())
	}

	log.Printf("[DEBUG] Creating Component: %#v", input)

	resp, err := conn.CreateComponent(input)
	if err != nil {
		return fmt.Errorf("error creating Component: %s", err)
	}

	d.SetId(aws.StringValue(resp.ComponentBuildVersionArn))

	return resourceAwsImageBuilderComponentRead(d, meta)
}

func resourceAwsImageBuilderComponentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetComponent(&imagebuilder.GetComponentInput{
		ComponentBuildVersionArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Component (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Component (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.Component.Arn)
	d.Set("name", resp.Component.Name)
	d.Set("semantic_version", resp.Component.Version)
	d.Set("change_description", resp.Component.ChangeDescription)
	d.Set("data", resp.Component.Data)
	d.Set("description", resp.Component.Description)
	d.Set("kms_key_id", resp.Component.KmsKeyId)
	d.Set("platform", resp.Component.Platform)

	tags, err := keyvaluetags.ImagebuilderListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Component (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsImageBuilderComponentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	// tags are the only thing we can actually change, everything else is ForceNew
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Component (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderComponentRead(d, meta)
}

func resourceAwsImageBuilderComponentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	_, err := conn.DeleteComponent(&imagebuilder.DeleteComponentInput{
		ComponentBuildVersionArn: aws.String(d.Id()),
	})

	if isAWSErr(err, imagebuilder.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Component (%s): %s", d.Id(), err)
	}

	return nil
}
