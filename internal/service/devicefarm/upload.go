package devicefarm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceUpload() *schema.Resource {
	return &schema.Resource{
		Create: resourceUploadCreate,
		Read:   resourceUploadRead,
		Update: resourceUploadUpdate,
		Delete: resourceUploadDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"project_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(devicefarm.UploadType_Values(), false),
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUploadCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.CreateUploadInput{
		Name:       aws.String(d.Get("name").(string)),
		ProjectArn: aws.String(d.Get("project_arn").(string)),
		Type:       aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	out, err := conn.CreateUpload(input)
	if err != nil {
		return fmt.Errorf("Error creating DeviceFarm Upload: %w", err)
	}

	arn := aws.StringValue(out.Upload.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Upload: %s", arn)
	d.SetId(arn)

	return resourceUploadRead(d, meta)
}

func resourceUploadRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	upload, err := FindUploadByArn(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Upload (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DeviceFarm Upload (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(upload.Arn)
	d.Set("name", upload.Name)
	d.Set("type", upload.Type)
	d.Set("content_type", upload.ContentType)
	d.Set("url", upload.Url)
	d.Set("category", upload.Category)
	d.Set("metadata", upload.Metadata)
	d.Set("arn", arn)

	projectArn, err := decodeProjectARN(arn, "upload", meta)
	if err != nil {
		return fmt.Errorf("error decoding project_arn (%s): %w", arn, err)
	}

	d.Set("project_arn", projectArn)

	return nil
}

func resourceUploadUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.UpdateUploadInput{
		Arn: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("content_type") {
		input.ContentType = aws.String(d.Get("content_type").(string))
	}

	log.Printf("[DEBUG] Updating DeviceFarm Upload: %s", d.Id())
	_, err := conn.UpdateUpload(input)
	if err != nil {
		return fmt.Errorf("Error Updating DeviceFarm Upload: %w", err)
	}

	return resourceUploadRead(d, meta)
}

func resourceUploadDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DeviceFarmConn

	input := &devicefarm.DeleteUploadInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Upload: %s", d.Id())
	_, err := conn.DeleteUpload(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting DeviceFarm Upload: %w", err)
	}

	return nil
}
