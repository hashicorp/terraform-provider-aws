package iam

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVirtualMFADevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualMFADeviceCreate,
		Read:   resourceVirtualMFADeviceRead,
		Update: resourceVirtualMFADeviceUpdate,
		Delete: resourceVirtualMFADeviceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_32_string_seed": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "/",
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"qr_code_png": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"virtual_mfa_device_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`[\w+=,.@-]+`),
					"must consist of upper and lowercase alphanumeric characters with no spaces. You can also include any of the following characters: _+=,.@-",
				),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualMFADeviceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("virtual_mfa_device_name").(string)
	request := &iam.CreateVirtualMFADeviceInput{
		Path:                 aws.String(d.Get("path").(string)),
		VirtualMFADeviceName: aws.String(name),
	}

	if len(tags) > 0 {
		request.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateVirtualMFADevice(request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM Virtual MFA Device (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		output, err = conn.CreateVirtualMFADevice(request)
	}

	if err != nil {
		return fmt.Errorf("failed creating IAM Virtual MFA Device %s: %w", name, err)
	}

	vMfa := output.VirtualMFADevice
	d.SetId(aws.StringValue(vMfa.SerialNumber))

	d.Set("base_32_string_seed", string(vMfa.Base32StringSeed))
	d.Set("qr_code_png", string(vMfa.QRCodePNG))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 {
		err := virtualMFAUpdateTags(conn, d.Id(), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM Virtual MFA Device (%s): %s", d.Id(), err)
			return resourceVirtualMFADeviceRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for IAM Virtual MFA Device (%s): %w", d.Id(), err)
		}
	}

	return resourceVirtualMFADeviceRead(d, meta)
}

func resourceVirtualMFADeviceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindVirtualMFADevice(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Virtual MFA Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Virtual MFA Device (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.SerialNumber)

	// The call above returns empty tags
	tagsInput := &iam.ListMFADeviceTagsInput{
		SerialNumber: aws.String(d.Id()),
	}

	mfaTags, err := conn.ListMFADeviceTags(tagsInput)
	if err != nil {
		return fmt.Errorf("error listing IAM Virtual MFA Device Tags (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(mfaTags.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVirtualMFADeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	o, n := d.GetChange("tags_all")

	err := virtualMFAUpdateTags(conn, d.Id(), o, n)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed updating tags for IAM Virtual MFA Device (%s): %s", d.Id(), err)
		return resourceVirtualMFADeviceRead(d, meta)
	}

	if err != nil {
		return fmt.Errorf("failed updating tags for IAM Virtual MFA Device (%s): %w", d.Id(), err)
	}

	return resourceVirtualMFADeviceRead(d, meta)
}

func resourceVirtualMFADeviceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVirtualMFADevice(request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		return fmt.Errorf("Error deleting IAM Virtual MFA Device %s: %w", d.Id(), err)
	}
	return nil
}
