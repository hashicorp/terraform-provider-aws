// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lambda

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lambda_layer_version", name="Layer Version")
func dataSourceLayerVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLayerVersionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code_sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compatible_architecture": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Architecture](),
				ConflictsWith:    []string{names.AttrVersion, "layer_version_arn"},
			},
			"compatible_architectures": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"compatible_runtime": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Runtime](),
				ConflictsWith:    []string{names.AttrVersion, "layer_version_arn"},
			},
			"compatible_runtimes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layer_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"layer_version_arn"},
			},
			"layer_version_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"layer_name", names.AttrVersion, "compatible_architecture", "compatible_runtime"},
			},
			"license_info": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_profile_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_hash": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "source_code_hash is deprecated. Use code_sha256 instead.",
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVersion: {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"compatible_runtimes", "layer_version_arn"},
			},
		},
	}
}

func dataSourceLayerVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	var output *lambda.GetLayerVersionOutput
	var err error

	if v, ok := d.GetOk("layer_version_arn"); ok {
		output, err = findLayerVersionByARN(ctx, conn, v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Lambda Layer Version (%s): %s", v.(string), err)
		}
	} else {
		layerName := d.Get("layer_name").(string)
		if layerName == "" {
			return sdkdiag.AppendErrorf(diags, "one of `layer_name` or `layer_version_arn` must be specified")
		}

		var versionNumber int64
		if v, ok := d.GetOk(names.AttrVersion); ok {
			versionNumber = int64(v.(int))
		} else {
			input := &lambda.ListLayerVersionsInput{
				LayerName: aws.String(layerName),
			}

			if v, ok := d.GetOk("compatible_architecture"); ok {
				input.CompatibleArchitecture = awstypes.Architecture(v.(string))
			}

			if v, ok := d.GetOk("compatible_runtime"); ok {
				input.CompatibleRuntime = awstypes.Runtime(v.(string))
			}

			listOutput, err := conn.ListLayerVersions(ctx, input)
			if err == nil && len(listOutput.LayerVersions) == 0 {
				err = tfresource.NewEmptyResultError()
			}
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing Lambda Layer Versions (%s): %s", layerName, err)
			}

			versionNumber = listOutput.LayerVersions[0].Version
		}

		output, err = findLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Lambda Layer (%s) Version (%d): %s", layerName, versionNumber, err)
		}
	}

	d.SetId(aws.ToString(output.LayerVersionArn))
	d.Set(names.AttrARN, output.LayerVersionArn)
	d.Set("code_sha256", output.Content.CodeSha256)
	d.Set("compatible_architectures", output.CompatibleArchitectures)
	d.Set("compatible_runtimes", output.CompatibleRuntimes)
	d.Set(names.AttrCreatedDate, output.CreatedDate)
	d.Set(names.AttrDescription, output.Description)
	d.Set("layer_arn", output.LayerArn)
	d.Set("layer_name", layerNameFromARN(aws.ToString(output.LayerArn)))
	d.Set("layer_version_arn", output.LayerVersionArn)
	d.Set("license_info", output.LicenseInfo)
	d.Set("signing_job_arn", output.Content.SigningJobArn)
	d.Set("signing_profile_version_arn", output.Content.SigningProfileVersionArn)
	d.Set("source_code_hash", output.Content.CodeSha256)
	d.Set("source_code_size", output.Content.CodeSize)
	d.Set(names.AttrVersion, output.Version)

	return diags
}

func findLayerVersionByARN(ctx context.Context, conn *lambda.Client, arn string) (*lambda.GetLayerVersionOutput, error) {
	layerARN, versionNumber, err := parseLayerVersionARN(arn)
	if err != nil {
		// ARN doesn't include version - try to find latest
		return findLatestLayerVersionByARN(ctx, conn, arn)
	}

	// AWS GetLayerVersion requires layer ARN (without version) and version number separately
	input := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerARN),
		VersionNumber: aws.Int64(versionNumber),
	}

	return findLayerVersion(ctx, conn, input)
}

func findLatestLayerVersionByARN(ctx context.Context, conn *lambda.Client, layerARN string) (*lambda.GetLayerVersionOutput, error) {
	input := &lambda.ListLayerVersionsInput{
		LayerName: aws.String(layerARN),
	}

	output, err := conn.ListLayerVersions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("unable to list layer versions (if this is a cross-account layer, the ARN must include the version number, e.g., :1, :2, etc.): %w", err)
	}

	if len(output.LayerVersions) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	versionNumber := output.LayerVersions[0].Version

	getInput := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerARN),
		VersionNumber: aws.Int64(versionNumber),
	}

	return findLayerVersion(ctx, conn, getInput)
}

func parseLayerVersionARN(arn string) (string, int64, error) {
	// ARN format: arn:aws:lambda:region:account:layer:name:version
	parts := strings.Split(arn, ":")
	if len(parts) < 7 {
		return "", 0, fmt.Errorf("invalid layer ARN format: %s", arn)
	}

	if len(parts) == 7 {
		// No version in ARN
		return "", 0, fmt.Errorf("no version in ARN")
	}

	versionStr := parts[7]
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid version number in ARN: %s", versionStr)
	}

	// Return the layer ARN without version (first 7 parts)
	layerARN := strings.Join(parts[:7], ":")
	return layerARN, version, nil
}

func layerNameFromARN(layerArn string) string {
	// Extract layer name from ARN: arn:aws:lambda:region:account:layer:name
	parts := strings.Split(layerArn, ":")
	if len(parts) >= 7 {
		return parts[6]
	}
	return ""
}
