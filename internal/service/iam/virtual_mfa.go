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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceVirtualMfaDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualMfaDeviceCreate,
		Read:   resourceVirtualMfaDeviceRead,
		// Update: resourceVirtualMfaDeviceUpdate,
		Delete: resourceVirtualMfaDeviceDelete,
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
	}
}

func resourceVirtualMfaDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	name := d.Get("virtual_mfa_device_name").(string)
	request := &iam.CreateVirtualMFADeviceInput{
		Path:                 aws.String(d.Get("path").(string)),
		VirtualMFADeviceName: aws.String(name),
	}

	createResp, err := conn.CreateVirtualMFADevice(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM Virtual MFA Device %s: %w", name, err)
	}
	d.SetId(aws.StringValue(createResp.VirtualMFADevice.SerialNumber))

	return resourceVirtualMfaDeviceRead(d, meta)
}

func resourceVirtualMfaDeviceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	output, err := FindVirtualMfaDevice(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Virtual MFA Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Virtual MFA Device (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.SerialNumber)
	d.Set("base_32_string_seed", string(output.Base32StringSeed))
	d.Set("qr_code_png", string(output.QRCodePNG))

	return nil
}

// func resourceVirtualMfaDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
// 	if d.HasChanges("name", "path") {
// 		conn := meta.(*conns.AWSClient).IAMConn
// 		on, nn := d.GetChange("name")
// 		_, np := d.GetChange("path")

// 		request := &iam.UpdateVirtualMFADeviceInput{
// 			VirtualMfaDeviceName:    aws.String(on.(string)),
// 			NewVirtualMfaDeviceName: aws.String(nn.(string)),
// 			NewPath:                 aws.String(np.(string)),
// 		}
// 		_, err := conn.UpdateVirtualMfaDevice(request)
// 		if err != nil {
// 			return fmt.Errorf("Error updating IAM VirtualMfaDevice %s: %s", d.Id(), err)
// 		}
// 		d.SetId(nn.(string))
// 		return resourceVirtualMfaDeviceRead(d, meta)
// 	}
// 	return nil
// }

func resourceVirtualMfaDeviceDelete(d *schema.ResourceData, meta interface{}) error {
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
