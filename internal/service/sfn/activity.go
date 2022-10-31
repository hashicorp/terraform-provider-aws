package sfn

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceActivity() *schema.Resource {
	return &schema.Resource{
		Create: resourceActivityCreate,
		Read:   resourceActivityRead,
		Update: resourceActivityUpdate,
		Delete: resourceActivityDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 80),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceActivityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SFNConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &sfn.CreateActivityInput{
		Name: aws.String(name),
		Tags: Tags(tags.IgnoreAWS()),
	}

	output, err := conn.CreateActivity(input)

	if err != nil {
		return fmt.Errorf("creating Step Function Activity (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ActivityArn))

	return resourceActivityRead(d, meta)
}

func resourceActivityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SFNConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindActivityByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Step Functions Activity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Step Functions Activity (%s): %w", d.Id(), err)
	}

	d.Set("creation_date", output.CreationDate.Format(time.RFC3339))
	d.Set("name", output.Name)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("listing tags for Step Functions Activity (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceActivityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SFNConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating Step Function Activity (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceActivityRead(d, meta)
}

func resourceActivityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SFNConn

	log.Printf("[DEBUG] Deleting Step Functions Activity: %s", d.Id())
	_, err := conn.DeleteActivity(&sfn.DeleteActivityInput{
		ActivityArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting Step Functions Activity (%s): %w", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(1*time.Minute, func() (interface{}, error) {
		return FindActivityByARN(conn, d.Id())
	})

	if err != nil {
		return fmt.Errorf("waiting for Step Functions Activity (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func FindActivityByARN(conn *sfn.SFN, arn string) (*sfn.DescribeActivityOutput, error) {
	input := &sfn.DescribeActivityInput{
		ActivityArn: aws.String(arn),
	}

	output, err := conn.DescribeActivity(input)

	if tfawserr.ErrCodeEquals(err, sfn.ErrCodeActivityDoesNotExist) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CreationDate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
