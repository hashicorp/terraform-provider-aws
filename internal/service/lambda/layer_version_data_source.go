// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"compatible_architecture": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Architecture](),
				ConflictsWith:    []string{names.AttrVersion},
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
				ConflictsWith:    []string{names.AttrVersion},
			},
			"compatible_runtimes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created_date": {
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
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrVersion: {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"compatible_runtimes"},
			},
		},
	}
}

func dataSourceLayerVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	layerName := d.Get("layer_name").(string)
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

		output, err := conn.ListLayerVersions(ctx, input)

		if err == nil && len(output.LayerVersions) == 0 {
			err = tfresource.NewEmptyResultError(input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Lambda Layer Versions (%s): %s", layerName, err)
		}

		versionNumber = output.LayerVersions[0].Version
	}

	output, err := findLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Layer (%s) Version (%d): %s", layerName, versionNumber, err)
	}

	d.SetId(aws.ToString(output.LayerVersionArn))
	d.Set(names.AttrARN, output.LayerVersionArn)
	d.Set("compatible_architectures", output.CompatibleArchitectures)
	d.Set("compatible_runtimes", output.CompatibleRuntimes)
	d.Set("created_date", output.CreatedDate)
	d.Set(names.AttrDescription, output.Description)
	d.Set("layer_arn", output.LayerArn)
	d.Set("license_info", output.LicenseInfo)
	d.Set("signing_job_arn", output.Content.SigningJobArn)
	d.Set("signing_profile_version_arn", output.Content.SigningProfileVersionArn)
	d.Set("source_code_hash", output.Content.CodeSha256)
	d.Set("source_code_size", output.Content.CodeSize)
	d.Set(names.AttrVersion, output.Version)

	return diags
}
