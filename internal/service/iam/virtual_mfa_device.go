// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_virtual_mfa_device", name="Virtual MFA Device")
// @Tags(identifierAttribute="id", resourceType="VirtualMFADevice")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/iam.VirtualMFADevice", importIgnore="base_32_string_seed;qr_code_png")
func resourceVirtualMFADevice() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualMFADeviceCreate,
		ReadWithoutTimeout:   resourceVirtualMFADeviceRead,
		UpdateWithoutTimeout: resourceVirtualMFADeviceUpdate,
		DeleteWithoutTimeout: resourceVirtualMFADeviceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			"enable_date": {
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtual_mfa_device_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[\w+=,.@-]+`),
					"must consist of upper and lowercase alphanumeric characters with no spaces. You can also include any of the following characters: _+=,.@-",
				),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualMFADeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	name := d.Get("virtual_mfa_device_name").(string)
	input := &iam.CreateVirtualMFADeviceInput{
		Path:                 aws.String(d.Get("path").(string)),
		Tags:                 getTagsIn(ctx),
		VirtualMFADeviceName: aws.String(name),
	}

	output, err := conn.CreateVirtualMFADeviceWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateVirtualMFADeviceWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Virtual MFA Device (%s): %s", name, err)
	}

	vMFA := output.VirtualMFADevice
	d.SetId(aws.StringValue(vMFA.SerialNumber))

	// Base32StringSeed and QRCodePNG must be read here, because they are not available via ListVirtualMFADevices
	d.Set("base_32_string_seed", string(vMFA.Base32StringSeed))
	d.Set("qr_code_png", string(vMFA.QRCodePNG))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := virtualMFADeviceCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting IAM Virtual MFA Device (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
}

func resourceVirtualMFADeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	vMFA, err := findVirtualMFADeviceBySerialNumber(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Virtual MFA Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	d.Set("arn", vMFA.SerialNumber)

	path, name, err := parseVirtualMFADeviceARN(aws.StringValue(vMFA.SerialNumber))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	d.Set("path", path)
	d.Set("virtual_mfa_device_name", name)

	if v := vMFA.EnableDate; v != nil {
		d.Set("enable_date", aws.TimeValue(v).Format(time.RFC3339))
	}

	if u := vMFA.User; u != nil {
		d.Set("user_name", u.UserName)
	}

	// The call above returns empty tags.
	tags, err := virtualMFADeviceTags(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing IAM Virtual MFA Device (%s) tags: %s", d.Id(), err)
	}

	setTagsOut(ctx, tags)

	return diags
}

func resourceVirtualMFADeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
}

func resourceVirtualMFADeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn(ctx)

	if v := d.Get("user_name"); v != "" {
		_, err := conn.DeactivateMFADeviceWithContext(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(v.(string)),
			SerialNumber: aws.String(d.Id()),
		})
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IAM Virtual MFA Device (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting IAM Virtual MFA Device: %s", d.Id())
	_, err := conn.DeleteVirtualMFADeviceWithContext(ctx, &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	return diags
}

func findVirtualMFADeviceBySerialNumber(ctx context.Context, conn *iam.IAM, serialNumber string) (*iam.VirtualMFADevice, error) {
	input := &iam.ListVirtualMFADevicesInput{}
	var output *iam.VirtualMFADevice

	err := conn.ListVirtualMFADevicesPagesWithContext(ctx, input, func(page *iam.ListVirtualMFADevicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VirtualMFADevices {
			if v != nil && aws.StringValue(v.SerialNumber) == serialNumber {
				output = v
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func parseVirtualMFADeviceARN(s string) (path, name string, err error) {
	arn, err := arn.Parse(s)
	if err != nil {
		return "", "", err
	}

	re := regexache.MustCompile(`^mfa(/|/[\x{0021}-\x{007E}]+/)([0-9A-Za-z_+=,.@-]+)$`)
	matches := re.FindStringSubmatch(arn.Resource)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("IAM Virtual MFA Device ARN: invalid resource section (%s)", arn.Resource)
	}

	return matches[1], matches[2], nil
}

func virtualMFADeviceTags(ctx context.Context, conn *iam.IAM, identifier string) ([]*iam.Tag, error) {
	output, err := conn.ListMFADeviceTagsWithContext(ctx, &iam.ListMFADeviceTagsInput{
		SerialNumber: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
