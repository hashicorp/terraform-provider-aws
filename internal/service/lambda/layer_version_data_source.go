// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_lambda_layer_version")
func DataSourceLayerVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLayerVersionRead,

		Schema: map[string]*schema.Schema{
			"layer_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"compatible_runtimes"},
			},
			"compatible_runtime": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(lambda.Runtime_Values(), false),
				ConflictsWith: []string{"version"},
			},
			"compatible_runtimes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_info": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
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
			"signing_profile_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_job_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compatible_architecture": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.StringInSlice(lambda.Architecture_Values(), false),
				ConflictsWith: []string{"version"},
			},
			"compatible_architectures": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceLayerVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)
	layerName := d.Get("layer_name").(string)

	var version int64

	if v, ok := d.GetOk("version"); ok {
		version = int64(v.(int))
	} else {
		listInput := &lambda.ListLayerVersionsInput{
			LayerName: aws.String(layerName),
		}
		if v, ok := d.GetOk("compatible_runtime"); ok {
			listInput.CompatibleRuntime = aws.String(v.(string))
		}

		if v, ok := d.GetOk("compatible_architecture"); ok {
			listInput.CompatibleArchitecture = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Looking up latest version for lambda layer %s", layerName)
		listOutput, err := conn.ListLayerVersionsWithContext(ctx, listInput)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Lambda Layer Versions (%s): %s", layerName, err)
		}

		if len(listOutput.LayerVersions) == 0 {
			return sdkdiag.AppendErrorf(diags, "listing Lambda Layer Versions (%s): empty response", layerName)
		}

		version = aws.Int64Value(listOutput.LayerVersions[0].Version)
	}

	input := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(version),
	}

	log.Printf("[DEBUG] Getting Lambda Layer Version: %s, version %d", layerName, version)
	output, err := conn.GetLayerVersionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Lambda Layer Version (%s, version %d): %s", layerName, version, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Lambda Layer Version (%s, version %d): empty response", layerName, version)
	}

	if err := d.Set("version", output.Version); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer version: %s", err)
	}
	if err := d.Set("compatible_runtimes", flex.FlattenStringList(output.CompatibleRuntimes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer compatible runtimes: %s", err)
	}
	if err := d.Set("compatible_architectures", flex.FlattenStringList(output.CompatibleArchitectures)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer compatible architectures: %s", err)
	}
	if err := d.Set("description", output.Description); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer description: %s", err)
	}
	if err := d.Set("license_info", output.LicenseInfo); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer license info: %s", err)
	}
	if err := d.Set("arn", output.LayerVersionArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer version arn: %s", err)
	}
	if err := d.Set("layer_arn", output.LayerArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer arn: %s", err)
	}
	if err := d.Set("created_date", output.CreatedDate); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer created date: %s", err)
	}
	if err := d.Set("source_code_hash", output.Content.CodeSha256); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer source code hash: %s", err)
	}
	if err := d.Set("source_code_size", output.Content.CodeSize); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer source code size: %s", err)
	}
	if err := d.Set("signing_profile_version_arn", output.Content.SigningProfileVersionArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer signing profile arn: %s", err)
	}
	if err := d.Set("signing_job_arn", output.Content.SigningJobArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda layer signing job arn: %s", err)
	}

	d.SetId(aws.StringValue(output.LayerVersionArn))

	return diags
}
