// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/iam/types;types.VirtualMFADevice", importIgnore="base_32_string_seed;qr_code_png")
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
			names.AttrARN: {
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
			names.AttrPath: {
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
			names.AttrUserName: {
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	name := d.Get("virtual_mfa_device_name").(string)
	input := &iam.CreateVirtualMFADeviceInput{
		Path:                 aws.String(d.Get(names.AttrPath).(string)),
		Tags:                 getTagsIn(ctx),
		VirtualMFADeviceName: aws.String(name),
	}

	output, err := conn.CreateVirtualMFADevice(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateVirtualMFADevice(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Virtual MFA Device (%s): %s", name, err)
	}

	vMFA := output.VirtualMFADevice
	d.SetId(aws.ToString(vMFA.SerialNumber))

	// Base32StringSeed and QRCodePNG must be read here, because they are not available via ListVirtualMFADevices
	d.Set("base_32_string_seed", string(vMFA.Base32StringSeed))
	d.Set("qr_code_png", string(vMFA.QRCodePNG))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := virtualMFADeviceCreateTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	vMFA, err := findVirtualMFADeviceBySerialNumber(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Virtual MFA Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, vMFA.SerialNumber)

	path, name, err := parseVirtualMFADeviceARN(aws.ToString(vMFA.SerialNumber))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrPath, path)
	d.Set("virtual_mfa_device_name", name)

	if v := vMFA.EnableDate; v != nil {
		d.Set("enable_date", aws.ToTime(v).Format(time.RFC3339))
	}

	if u := vMFA.User; u != nil {
		d.Set(names.AttrUserName, u.UserName)
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
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if v := d.Get(names.AttrUserName); v != "" {
		_, err := conn.DeactivateMFADevice(ctx, &iam.DeactivateMFADeviceInput{
			UserName:     aws.String(v.(string)),
			SerialNumber: aws.String(d.Id()),
		})
		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return diags
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deactivating IAM Virtual MFA Device (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting IAM Virtual MFA Device: %s", d.Id())
	_, err := conn.DeleteVirtualMFADevice(ctx, &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	return diags
}

func findVirtualMFADeviceBySerialNumber(ctx context.Context, conn *iam.Client, serialNumber string) (*awstypes.VirtualMFADevice, error) {
	input := &iam.ListVirtualMFADevicesInput{}
	var output awstypes.VirtualMFADevice

	pages := iam.NewListVirtualMFADevicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.VirtualMFADevices {
			if !reflect.ValueOf(v).IsZero() && aws.ToString(v.SerialNumber) == serialNumber {
				output = v
				break
			}
		}
	}

	if reflect.ValueOf(output).IsZero() {
		return nil, &retry.NotFoundError{}
	}

	return &output, nil
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

func virtualMFADeviceTags(ctx context.Context, conn *iam.Client, identifier string) ([]awstypes.Tag, error) {
	output, err := conn.ListMFADeviceTags(ctx, &iam.ListMFADeviceTagsInput{
		SerialNumber: aws.String(identifier),
	})
	if err != nil {
		return nil, err
	}

	return output.Tags, nil
}
