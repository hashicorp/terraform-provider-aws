// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const imageVersionResourcePartCount = 2

// @SDKResource("aws_sagemaker_image_version", name="Image Version")
func resourceImageVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageVersionCreate,
		ReadWithoutTimeout:   resourceImageVersionRead,
		UpdateWithoutTimeout: resourceImageVersionUpdate,
		DeleteWithoutTimeout: resourceImageVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceImageVersionV0().CoreConfigSchema().ImpliedType(),
				Upgrade: imageVersionStateUpgradeV0,
				Version: 0,
			},
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aliases": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"base_image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"container_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"horovod": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"image_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.JobType](),
			},
			"ml_framework": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z]+ ?\d+\.\d+(\.\d+)?$`), ""),
			},
			"processor": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Processor](),
			},
			"programming_lang": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z]+ ?\d+\.\d+(\.\d+)?$`), ""),
			},
			"release_notes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"vendor_guidance": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VendorGuidance](),
			},
			names.AttrVersion: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceImageVersionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("image_name").(string)
	input := sagemaker.CreateImageVersionInput{
		ImageName:   aws.String(name),
		BaseImage:   aws.String(d.Get("base_image").(string)),
		ClientToken: aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk("job_type"); ok {
		input.JobType = awstypes.JobType(v.(string))
	}

	if v, ok := d.GetOk("processor"); ok {
		input.Processor = awstypes.Processor(v.(string))
	}

	if v, ok := d.GetOk("release_notes"); ok {
		input.ReleaseNotes = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vendor_guidance"); ok {
		input.VendorGuidance = awstypes.VendorGuidance(v.(string))
	}

	if v, ok := d.GetOk("horovod"); ok {
		input.Horovod = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ml_framework"); ok {
		input.MLFramework = aws.String(v.(string))
	}

	if v, ok := d.GetOk("programming_lang"); ok {
		input.ProgrammingLang = aws.String(v.(string))
	}

	if v, ok := d.GetOk("aliases"); ok && v.(*schema.Set).Len() > 0 {
		input.Aliases = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if _, err := conn.CreateImageVersion(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Image Version %s: %s", name, err)
	}

	out, err := waitImageVersionCreated(ctx, conn, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) to be created: %s", d.Id(), err)
	}

	parts := []string{name, strconv.Itoa(int(aws.ToInt32(out.Version)))}
	id, err := flex.FlattenResourceId(parts, imageVersionResourcePartCount, false)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Image Version %s: %s", name, err)
	}

	d.SetId(id)

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name, version, err := expandImageVersionResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Image Version (%s): %s", d.Id(), err)
	}
	d.Set("image_name", name) // to support import

	image, err := findImageVersionByTwoPartKey(ctx, conn, name, version)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Image Version (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, image.ImageVersionArn)
	d.Set("base_image", image.BaseImage)
	d.Set("image_arn", image.ImageArn)
	d.Set("container_image", image.ContainerImage)
	d.Set(names.AttrVersion, image.Version)
	d.Set("horovod", image.Horovod)
	d.Set("job_type", image.JobType)
	d.Set("processor", image.Processor)
	d.Set("release_notes", image.ReleaseNotes)
	d.Set("vendor_guidance", image.VendorGuidance)
	d.Set("ml_framework", image.MLFramework)
	d.Set("programming_lang", image.ProgrammingLang)

	// The DescribeImageVersion API response does not include aliases, so these must
	// be fetched separately using the ListAliases API
	aliases, err := findImageVersionAliasesByTwoPartKey(ctx, conn, name, version)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading aliases for SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	if err := d.Set("aliases", aliases); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting aliases: %s", err)
	}

	return diags
}

func resourceImageVersionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("image_name").(string)
	version := d.Get(names.AttrVersion).(int)

	input := sagemaker.UpdateImageVersionInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	if d.HasChange("horovod") {
		input.Horovod = aws.Bool(d.Get("horovod").(bool))
	}

	if d.HasChange("job_type") {
		input.JobType = awstypes.JobType(d.Get("job_type").(string))
	}

	if d.HasChange("processor") {
		input.Processor = awstypes.Processor(d.Get("processor").(string))
	}

	if d.HasChange("release_notes") {
		input.ReleaseNotes = aws.String(d.Get("release_notes").(string))
	}

	if d.HasChange("vendor_guidance") {
		input.VendorGuidance = awstypes.VendorGuidance(d.Get("vendor_guidance").(string))
	}

	if d.HasChange("ml_framework") {
		input.MLFramework = aws.String(d.Get("ml_framework").(string))
	}

	if d.HasChange("programming_lang") {
		input.ProgrammingLang = aws.String(d.Get("programming_lang").(string))
	}

	if d.HasChange("aliases") {
		// For UpdateImageVersion, we need to use AliasesToAdd and AliasesToDelete
		// instead of Aliases directly
		o, n := d.GetChange("aliases")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := flex.ExpandStringValueSet(ns.Difference(os)), flex.ExpandStringValueSet(os.Difference(ns))

		if len(add) > 0 {
			input.AliasesToAdd = add
		}
		if len(del) > 0 {
			input.AliasesToDelete = del
		}
	}

	if _, err := conn.UpdateImageVersion(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("image_name").(string)
	version := d.Get(names.AttrVersion).(int)

	input := sagemaker.DeleteImageVersionInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	if _, err := conn.DeleteImageVersion(ctx, &input); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	if _, err := waitImageVersionDeleted(ctx, conn, name, version); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

// findImageVersionByName is used immediately after creation to poll for status of
// the most recently created image version
//
// findImageVersionByTwoPartKey should be used for all subsequent operations once the
// version number is assigned.
func findImageVersionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersion(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// findImageVersionByTwoPartKey is used to fetch a specific version once a version number
// is assigned
func findImageVersionByTwoPartKey(ctx context.Context, conn *sagemaker.Client, name string, version int) (*sagemaker.DescribeImageVersionOutput, error) {
	input := sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	output, err := conn.DescribeImageVersion(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findImageVersionAliasesByTwoPartKey(ctx context.Context, conn *sagemaker.Client, name string, version int) ([]string, error) {
	input := sagemaker.ListAliasesInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	output, err := conn.ListAliases(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SageMakerImageVersionAliases, nil
}

// expandImageVersionResourceID wraps flex.ExpandResourceId and handles conversion
// of the version portion to an integer
func expandImageVersionResourceID(id string) (string, int, error) {
	parts, err := flex.ExpandResourceId(id, imageVersionResourcePartCount, false)
	if err != nil {
		return "", 0, err
	}

	name := parts[0]
	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return name, 0, err
	}

	return name, version, nil
}
