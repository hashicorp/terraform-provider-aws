// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
	input := &sagemaker.CreateImageVersionInput{
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

	if v, ok := d.GetOk("aliases"); ok {
		aliases := v.(*schema.Set).List()
		input.Aliases = make([]string, len(aliases))
		for i, alias := range aliases {
			input.Aliases[i] = alias.(string)
		}
	}

	_, err := conn.CreateImageVersion(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Image Version %s: %s", name, err)
	}

	// Get the version from the API response
	output, err := conn.DescribeImageVersion(ctx, &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing SageMaker AI Image Version %s after creation: %s", name, err)
	}

	// Set the ID to be a combination of name and version
	versionNumber := aws.ToInt32(output.Version)
	id := fmt.Sprintf("%s:%d", name, versionNumber)
	d.SetId(id)

	// Wait for the image version to be created
	if _, err := waitImageVersionCreated(ctx, conn, id); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) to be created: %s", id, err)
	}

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	id := d.Id()
	var image *sagemaker.DescribeImageVersionOutput
	var err error

	// Check if the ID contains a version (has a colon)
	if strings.Contains(id, ":") {
		// New format - use the new function
		image, err = findImageVersionByNameAndVersion(ctx, conn, id)
	} else {
		// Legacy format - just the name
		image, err = findImageVersionByName(ctx, conn, id)

		// If successful, update the ID to the new format
		if err == nil && image != nil {
			newID := fmt.Sprintf("%s:%d", id, aws.ToInt32(image.Version))
			d.SetId(newID)
			id = newID
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Image Version (%s); removing from state", id)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Image Version (%s): %s", id, err)
	}

	// Parse the ID to get the name
	parts := strings.Split(id, ":")
	name := parts[0]

	d.Set(names.AttrARN, image.ImageVersionArn)
	d.Set("base_image", image.BaseImage)
	d.Set("image_arn", image.ImageArn)
	d.Set("container_image", image.ContainerImage)
	d.Set(names.AttrVersion, image.Version)
	d.Set("image_name", name)
	d.Set("horovod", image.Horovod)
	d.Set("job_type", image.JobType)
	d.Set("processor", image.Processor)
	d.Set("release_notes", image.ReleaseNotes)
	d.Set("vendor_guidance", image.VendorGuidance)
	d.Set("ml_framework", image.MLFramework)
	d.Set("programming_lang", image.ProgrammingLang)

	// The AWS SDK doesn't have an Aliases field in DescribeImageVersionOutput
	// We need to fetch aliases separately using ListAliases API
	idParts := strings.Split(id, ":")
	imageName := idParts[0]
	versionStr := idParts[1]
	versionNum, err := strconv.Atoi(versionStr)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid version number in resource ID: %s", d.Id())
	}

	aliasesInput := &sagemaker.ListAliasesInput{
		ImageName: aws.String(imageName),
		Version:   aws.Int32(int32(versionNum)),
	}

	aliasesOutput, err := conn.ListAliases(ctx, aliasesInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing aliases for SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	if err := d.Set("aliases", aliasesOutput.SageMakerImageVersionAliases); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting aliases: %s", err)
	}

	return diags
}

func resourceImageVersionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	// Parse the ID to get name and version
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return sdkdiag.AppendErrorf(diags, "invalid resource ID format: %s", d.Id())
	}

	name := parts[0]
	versionStr := parts[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid version number in resource ID: %s", d.Id())
	}

	input := &sagemaker.UpdateImageVersionInput{
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
		oldAliasesSet, newAliasesSet := d.GetChange("aliases")
		oldAliases := oldAliasesSet.(*schema.Set).List()
		newAliases := newAliasesSet.(*schema.Set).List()

		// Find aliases to add (in new but not in old)
		var aliasesToAdd []string
		for _, newAlias := range newAliases {
			found := false
			for _, oldAlias := range oldAliases {
				if newAlias.(string) == oldAlias.(string) {
					found = true
					break
				}
			}
			if !found {
				aliasesToAdd = append(aliasesToAdd, newAlias.(string))
			}
		}

		// Find aliases to delete (in old but not in new)
		var aliasesToDelete []string
		for _, oldAlias := range oldAliases {
			found := false
			for _, newAlias := range newAliases {
				if oldAlias.(string) == newAlias.(string) {
					found = true
					break
				}
			}
			if !found {
				aliasesToDelete = append(aliasesToDelete, oldAlias.(string))
			}
		}

		if len(aliasesToAdd) > 0 {
			input.AliasesToAdd = aliasesToAdd
		}

		if len(aliasesToDelete) > 0 {
			input.AliasesToDelete = aliasesToDelete
		}
	}

	if _, err := conn.UpdateImageVersion(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	return append(diags, resourceImageVersionRead(ctx, d, meta)...)
}

func resourceImageVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	// Parse the ID to get name and version
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return sdkdiag.AppendErrorf(diags, "invalid resource ID format: %s", d.Id())
	}

	name := parts[0]
	versionStr := parts[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "invalid version number in resource ID: %s", d.Id())
	}

	input := &sagemaker.DeleteImageVersionInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	if _, err := conn.DeleteImageVersion(ctx, input); err != nil {
		if errs.IsAErrorMessageContains[*awstypes.ResourceNotFound](err, "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Image Version (%s): %s", d.Id(), err)
	}

	if _, err := waitImageVersionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Image Version (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findImageVersionByNameAndVersion(ctx context.Context, conn *sagemaker.Client, id string) (*sagemaker.DescribeImageVersionOutput, error) {
	// Parse the ID to get name and version
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid resource ID format: %s", id)
	}

	name := parts[0]
	versionStr := parts[1]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version number in resource ID: %s", id)
	}

	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
		Version:   aws.Int32(int32(version)),
	}

	output, err := conn.DescribeImageVersion(ctx, input)

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

// Keep this for backward compatibility
func findImageVersionByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersion(ctx, input)

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

// FindImageVersionByNameAndVersion finds a SageMaker Image Version by name and version
func FindImageVersionByNameAndVersion(ctx context.Context, conn *sagemaker.Client, id string) (*sagemaker.DescribeImageVersionOutput, error) {
	return findImageVersionByNameAndVersion(ctx, conn, id)
}
