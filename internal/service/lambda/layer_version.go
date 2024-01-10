// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	arn2 "github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const mutexLayerKey = `aws_lambda_layer_version`

// @SDKResource("aws_lambda_layer_version")
func ResourceLayerVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLayerVersionPublish,
		ReadWithoutTimeout:   resourceLayerVersionRead,
		DeleteWithoutTimeout: resourceLayerVersionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compatible_architectures": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lambda.Architecture_Values(), false),
				},
			},
			"compatible_runtimes": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 15,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lambda.Runtime_Values(), false),
				},
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filename": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"s3_bucket", "s3_key", "s3_object_version"},
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
			"s3_bucket": {
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
			"skip_destroy": {
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
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLayerVersionPublish(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	layerName := d.Get("layer_name").(string)
	filename, hasFilename := d.GetOk("filename")
	s3Bucket, bucketOk := d.GetOk("s3_bucket")
	s3Key, keyOk := d.GetOk("s3_key")
	s3ObjectVersion, versionOk := d.GetOk("s3_object_version")

	if !hasFilename && !bucketOk && !keyOk && !versionOk {
		return sdkdiag.AppendErrorf(diags, "filename or s3_* attributes must be set")
	}

	var layerContent *lambda.LayerVersionContentInput
	if hasFilename {
		conns.GlobalMutexKV.Lock(mutexLayerKey)
		defer conns.GlobalMutexKV.Unlock(mutexLayerKey)
		file, err := readFileContents(filename.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Unable to load %q: %s", filename.(string), err)
		}
		layerContent = &lambda.LayerVersionContentInput{
			ZipFile: file,
		}
	} else {
		if !bucketOk || !keyOk {
			return sdkdiag.AppendErrorf(diags, "s3_bucket and s3_key must all be set while using s3 code source")
		}
		layerContent = &lambda.LayerVersionContentInput{
			S3Bucket: aws.String(s3Bucket.(string)),
			S3Key:    aws.String(s3Key.(string)),
		}
		if versionOk {
			layerContent.S3ObjectVersion = aws.String(s3ObjectVersion.(string))
		}
	}

	params := &lambda.PublishLayerVersionInput{
		Content:     layerContent,
		Description: aws.String(d.Get("description").(string)),
		LayerName:   aws.String(layerName),
		LicenseInfo: aws.String(d.Get("license_info").(string)),
	}

	if v, ok := d.GetOk("compatible_runtimes"); ok && v.(*schema.Set).Len() > 0 {
		params.CompatibleRuntimes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("compatible_architectures"); ok && v.(*schema.Set).Len() > 0 {
		params.CompatibleArchitectures = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Publishing Lambda layer: %s", params)
	result, err := conn.PublishLayerVersionWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating lambda layer: %s", err)
	}

	d.SetId(aws.StringValue(result.LayerVersionArn))

	return append(diags, resourceLayerVersionRead(ctx, d, meta)...)
}

func resourceLayerVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	layerName, versionNumber, err := LayerVersionParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := FindLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Layer Version %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Layer Version (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.LayerVersionArn)
	d.Set("compatible_architectures", aws.StringValueSlice(output.CompatibleArchitectures))
	d.Set("compatible_runtimes", aws.StringValueSlice(output.CompatibleRuntimes))
	d.Set("created_date", output.CreatedDate)
	d.Set("description", output.Description)
	d.Set("layer_arn", output.LayerArn)
	d.Set("layer_name", layerName)
	d.Set("license_info", output.LicenseInfo)
	d.Set("signing_job_arn", output.Content.SigningJobArn)
	d.Set("signing_profile_version_arn", output.Content.SigningProfileVersionArn)
	d.Set("source_code_hash", output.Content.CodeSha256)
	d.Set("source_code_size", output.Content.CodeSize)
	d.Set("version", strconv.FormatInt(versionNumber, 10))

	return diags
}

func resourceLayerVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	if d.Get("skip_destroy").(bool) {
		log.Printf("[DEBUG] Retaining Lambda Layer Version %q", d.Id())
		return diags
	}

	layerName, versionNumber, err := LayerVersionParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteLayerVersionWithContext(ctx, &lambda.DeleteLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Layer Version (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLayerVersionByTwoPartKey(ctx context.Context, conn *lambda.Lambda, layerName string, versionNumber int64) (*lambda.GetLayerVersionOutput, error) {
	input := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	output, err := conn.GetLayerVersionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
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

func LayerVersionParseID(id string) (layerName string, version int64, err error) {
	arn, err := arn2.Parse(id)
	if err != nil {
		return
	}
	parts := strings.Split(arn.Resource, ":")
	if len(parts) != 3 || parts[0] != "layer" {
		err = fmt.Errorf("lambda_layer ID must be a valid Layer ARN")
		return
	}

	layerName = parts[1]
	version, err = strconv.ParseInt(parts[2], 10, 64)
	return
}
