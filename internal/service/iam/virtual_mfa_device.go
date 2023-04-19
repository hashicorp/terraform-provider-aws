package iam

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_virtual_mfa_device", name="Virtual MFA Device")
// @Tags
func ResourceVirtualMFADevice() *schema.Resource {
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

func resourceVirtualMFADeviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	name := d.Get("virtual_mfa_device_name").(string)
	tags := GetTagsIn(ctx)
	request := &iam.CreateVirtualMFADeviceInput{
		Path:                 aws.String(d.Get("path").(string)),
		Tags:                 tags,
		VirtualMFADeviceName: aws.String(name),
	}

	output, err := conn.CreateVirtualMFADeviceWithContext(ctx, request)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if request.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating IAM Virtual MFA Device (%s) with tags: %s. Trying create without tags.", name, err)
		request.Tags = nil

		output, err = conn.CreateVirtualMFADeviceWithContext(ctx, request)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Virtual MFA Device %s: %s", name, err)
	}

	vMfa := output.VirtualMFADevice
	d.SetId(aws.StringValue(vMfa.SerialNumber))

	d.Set("base_32_string_seed", string(vMfa.Base32StringSeed))
	d.Set("qr_code_png", string(vMfa.QRCodePNG))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if request.Tags == nil && len(tags) > 0 {
		err := virtualMFAUpdateTags(ctx, conn, d.Id(), nil, KeyValueTags(ctx, tags))

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.ErrorISOUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for IAM Virtual MFA Device (%s): %s", d.Id(), err)
			return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "adding tags after create for IAM Virtual MFA Device (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
}

func resourceVirtualMFADeviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	output, err := FindVirtualMFADevice(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Virtual MFA Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.SerialNumber)

	// The call above returns empty tags
	mfaTags, err := conn.ListMFADeviceTagsWithContext(ctx, &iam.ListMFADeviceTagsInput{
		SerialNumber: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing IAM Virtual MFA Device (%s) tags: %s", d.Id(), err)
	}

	SetTagsOut(ctx, mfaTags.Tags)

	return diags
}

func resourceVirtualMFADeviceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	o, n := d.GetChange("tags_all")

	err := virtualMFAUpdateTags(ctx, conn, d.Id(), o, n)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed updating tags for IAM Virtual MFA Device (%s): %s", d.Id(), err)
		return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating tags for IAM Virtual MFA Device (%s): %s", d.Id(), err)
	}

	return append(diags, resourceVirtualMFADeviceRead(ctx, d, meta)...)
}

func resourceVirtualMFADeviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMConn()

	request := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVirtualMFADeviceWithContext(ctx, request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting IAM Virtual MFA Device %s: %s", d.Id(), err)
	}
	return diags
}
