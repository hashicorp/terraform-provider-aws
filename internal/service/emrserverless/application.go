package emrserverless

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceApplicationCreate,
		Read:   resourceApplicationRead,
		Update: resourceApplicationUpdate,
		Delete: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"release_label": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
		},
	}
}

func resourceApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &emrserverless.CreateApplicationInput{
		ClientToken:  aws.String(resource.UniqueId()),
		ReleaseLabel: aws.String(d.Get("release_label").(string)),
		Name:         aws.String(d.Get("name").(string)),
		Type:         aws.String(d.Get("type").(string)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	result, err := conn.CreateApplication(input)
	if err != nil {
		return fmt.Errorf("error creating EMR Serveless Application: %w", err)
	}

	d.SetId(aws.StringValue(result.ApplicationId))

	_, err = waitApplicationCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for EMR Serveless Application (%s) to create: %w", d.Id(), err)
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &emrserverless.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
			ClientToken:   aws.String(resource.UniqueId()),
		}

		_, err := conn.UpdateApplication(input)
		if err != nil {
			return fmt.Errorf("error updating EMR Serveless Application: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating EMR Serverless Application (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	application, err := FindApplicationByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Serverless Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EMR Serverless Application (%s): %w", d.Id(), err)
	}

	d.Set("arn", application.Arn)
	d.Set("name", application.Name)
	d.Set("type", strings.ToLower(aws.StringValue(application.Type)))
	d.Set("release_label", application.ReleaseLabel)

	tags := KeyValueTags(application.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn

	request := &emrserverless.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EMR Serverless Application: %s", d.Id())
	_, err := conn.DeleteApplication(request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, emrserverless.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting EMR Serverless Application (%s): %w", d.Id(), err)
	}

	return nil
}
