// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

const mutexLayerKey = `aws_lambda_layer_version`

// @SDKResource("aws_lambda_layer_version", name="Layer Version")
func resourceLayerVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLayerVersionCreate,
		ReadWithoutTimeout:   resourceLayerVersionRead,
		DeleteWithoutTimeout: resourceLayerVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code_sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compatible_architectures": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Architecture](),
				},
			},
			"compatible_runtimes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 15,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Runtime](),
				},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrS3Bucket, "s3_key", "s3_object_version"},
			},
			"layer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"license_info": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			names.AttrS3Bucket: {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filename"},
			},
			"s3_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filename"},
			},
			"s3_object_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"filename"},
			},
			"signing_job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_profile_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"source_code_hash": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLayerVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName := d.Get("layer_name").(string)
	filename, hasFilename := d.GetOk("filename")
	s3Bucket, bucketOk := d.GetOk(names.AttrS3Bucket)
	s3Key, keyOk := d.GetOk("s3_key")
	s3ObjectVersion, versionOk := d.GetOk("s3_object_version")

	if !hasFilename && !bucketOk && !keyOk && !versionOk {
		return sdkdiag.AppendErrorf(diags, "filename or s3_* attributes must be set")
	}

	var layerContent *awstypes.LayerVersionContentInput
	if hasFilename {
		conns.GlobalMutexKV.Lock(mutexLayerKey)
		defer conns.GlobalMutexKV.Unlock(mutexLayerKey)

		file, err := readFileContents(filename.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ZIP file (%s): %s", filename, err)
		}

		layerContent = &awstypes.LayerVersionContentInput{
			ZipFile: file,
		}
	} else {
		if !bucketOk || !keyOk {
			return sdkdiag.AppendErrorf(diags, "s3_bucket and s3_key must all be set while using s3 code source")
		}
		layerContent = &awstypes.LayerVersionContentInput{
			S3Bucket: aws.String(s3Bucket.(string)),
			S3Key:    aws.String(s3Key.(string)),
		}
		if versionOk {
			layerContent.S3ObjectVersion = aws.String(s3ObjectVersion.(string))
		}
	}

	input := &lambda.PublishLayerVersionInput{
		Content:     layerContent,
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		LayerName:   aws.String(layerName),
		LicenseInfo: aws.String(d.Get("license_info").(string)),
	}

	if v, ok := d.GetOk("compatible_architectures"); ok && v.(*schema.Set).Len() > 0 {
		input.CompatibleArchitectures = flex.ExpandStringyValueSet[awstypes.Architecture](v.(*schema.Set))
	}

	if v, ok := d.GetOk("compatible_runtimes"); ok && v.(*schema.Set).Len() > 0 {
		input.CompatibleRuntimes = flex.ExpandStringyValueSet[awstypes.Runtime](v.(*schema.Set))
	}

	output, err := conn.PublishLayerVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "publishing Lambda Layer (%s) Version: %s", layerName, err)
	}

	d.SetId(aws.ToString(output.LayerVersionArn))

	return append(diags, resourceLayerVersionRead(ctx, d, meta)...)
}

func resourceLayerVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName, versionNumber, err := layerVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Layer Version %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Layer Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.LayerVersionArn)
	d.Set("code_sha256", output.Content.CodeSha256)
	d.Set("compatible_architectures", output.CompatibleArchitectures)
	d.Set("compatible_runtimes", output.CompatibleRuntimes)
	d.Set(names.AttrCreatedDate, output.CreatedDate)
	d.Set(names.AttrDescription, output.Description)
	d.Set("layer_arn", output.LayerArn)
	d.Set("layer_name", layerName)
	d.Set("license_info", output.LicenseInfo)
	d.Set("signing_job_arn", output.Content.SigningJobArn)
	d.Set("signing_profile_version_arn", output.Content.SigningProfileVersionArn)
	d.Set("source_code_hash", d.Get("source_code_hash"))
	d.Set("source_code_size", output.Content.CodeSize)
	d.Set(names.AttrVersion, strconv.FormatInt(versionNumber, 10))

	return diags
}

func resourceLayerVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	if d.Get(names.AttrSkipDestroy).(bool) {
		log.Printf("[DEBUG] Retaining Lambda Layer Version %q", d.Id())
		return diags
	}

	layerName, versionNumber, err := layerVersionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Lambda Layer Version: %s", d.Id())
	_, err = conn.DeleteLayerVersion(ctx, &lambda.DeleteLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Layer Version (%s): %s", d.Id(), err)
	}

	return diags
}

func layerVersionParseResourceID(id string) (layerName string, version int64, err error) {
	v, err := arn.Parse(id)
	if err != nil {
		return
	}
	parts := strings.Split(v.Resource, ":")
	if len(parts) != 3 || parts[0] != "layer" {
		err = fmt.Errorf("lambda_layer ID must be a valid Layer ARN")
		return
	}

	layerName = parts[1]
	version, err = strconv.ParseInt(parts[2], 10, 64)
	return
}

func findLayerVersionByTwoPartKey(ctx context.Context, conn *lambda.Client, layerName string, versionNumber int64) (*lambda.GetLayerVersionOutput, error) {
	input := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	return findLayerVersion(ctx, conn, input)
}

func findLayerVersion(ctx context.Context, conn *lambda.Client, input *lambda.GetLayerVersionInput) (*lambda.GetLayerVersionOutput, error) {
	output, err := conn.GetLayerVersion(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
