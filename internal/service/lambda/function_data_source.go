// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_lambda_function", name="Function")
// @Tags
func dataSourceFunction() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFunctionRead,

		Schema: map[string]*schema.Schema{
			"architectures": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code_sha256": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code_signing_config_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dead_letter_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrTargetARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnvironment: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variables": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ephemeral_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"file_system_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local_mount_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"handler": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logging_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_log_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"log_group": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_log_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"memory_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"qualified_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualified_invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"reserved_concurrent_executions": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrRole: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"runtime": {
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
				Deprecated: "This attribute is deprecated and will be removed in a future major version. Use `code_sha256` instead.",
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tracing_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMode: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv6_allowed_for_dual_stack": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	} else {
		// If no qualifier provided, set version to latest published version.
		latest, err := findLatestFunctionVersionByName(ctx, conn, functionName)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) latest version: %s", functionName, err)
		}

		// If no published version exists, AWS returns '$LATEST' for latestVersion
		if v := aws.ToString(latest.Version); v != FunctionVersionLatest {
			input.Qualifier = aws.String(v)
		}
	}

	output, err := findFunction(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s): %s", functionName, err)
	}

	function := output.Configuration
	functionARN := aws.ToString(function.FunctionArn)
	qualifierSuffix := fmt.Sprintf(":%s", aws.ToString(input.Qualifier))
	versionSuffix := fmt.Sprintf(":%s", aws.ToString(function.Version))
	qualifiedARN := functionARN
	if !strings.HasSuffix(functionARN, qualifierSuffix) && !strings.HasSuffix(functionARN, versionSuffix) {
		qualifiedARN = functionARN + versionSuffix
	}
	unqualifiedARN := strings.TrimSuffix(functionARN, qualifierSuffix)

	d.SetId(functionName)
	d.Set("architectures", function.Architectures)
	d.Set(names.AttrARN, unqualifiedARN)
	d.Set("code_sha256", function.CodeSha256)
	if function.DeadLetterConfig != nil && function.DeadLetterConfig.TargetArn != nil {
		if err := d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				names.AttrTargetARN: aws.ToString(function.DeadLetterConfig.TargetArn),
			},
		}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dead_letter_config: %s", err)
		}
	} else {
		d.Set("dead_letter_config", []interface{}{})
	}
	d.Set(names.AttrDescription, function.Description)
	if err := d.Set(names.AttrEnvironment, flattenEnvironment(function.Environment)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting environment: %s", err)
	}
	if err := d.Set("ephemeral_storage", flattenEphemeralStorage(function.EphemeralStorage)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_storage: %s", err)
	}
	if err := d.Set("file_system_config", flattenFileSystemConfigs(function.FileSystemConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting file_system_config: %s", err)
	}
	d.Set("handler", function.Handler)
	if output.Code != nil {
		d.Set("image_uri", output.Code.ImageUri)
	}
	d.Set("invoke_arn", invokeARN(meta.(*conns.AWSClient), unqualifiedARN))
	d.Set(names.AttrKMSKeyARN, function.KMSKeyArn)
	d.Set("last_modified", function.LastModified)
	if err := d.Set("layers", flattenLayers(function.Layers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting layers: %s", err)
	}
	if err := d.Set("logging_config", flattenLoggingConfig(function.LoggingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
	}
	d.Set("memory_size", function.MemorySize)
	d.Set("qualified_arn", qualifiedARN)
	d.Set("qualified_invoke_arn", invokeARN(meta.(*conns.AWSClient), qualifiedARN))
	if output.Concurrency != nil {
		d.Set("reserved_concurrent_executions", output.Concurrency.ReservedConcurrentExecutions)
	} else {
		d.Set("reserved_concurrent_executions", -1)
	}
	d.Set(names.AttrRole, function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("signing_job_arn", function.SigningJobArn)
	d.Set("signing_profile_version_arn", function.SigningProfileVersionArn)
	d.Set("source_code_hash", function.CodeSha256)
	d.Set("source_code_size", function.CodeSize)
	d.Set(names.AttrTimeout, function.Timeout)
	tracingConfigMode := awstypes.TracingModePassThrough
	if function.TracingConfig != nil {
		tracingConfigMode = function.TracingConfig.Mode
	}
	if err := d.Set("tracing_config", []interface{}{
		map[string]interface{}{
			names.AttrMode: string(tracingConfigMode),
		},
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tracing_config: %s", err)
	}
	d.Set(names.AttrVersion, function.Version)
	if err := d.Set(names.AttrVPCConfig, flattenVPCConfigResponse(function.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	setTagsOut(ctx, output.Tags)

	// See r/aws_lambda_function.
	if partition, region := meta.(*conns.AWSClient).Partition, meta.(*conns.AWSClient).Region; partition == names.StandardPartitionID && signerServiceIsAvailable(region) {
		var codeSigningConfigARN string

		if function.PackageType == awstypes.PackageTypeZip {
			output, err := conn.GetFunctionCodeSigningConfig(ctx, &lambda.GetFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) code signing config: %s", d.Id(), err)
			}

			if output != nil {
				codeSigningConfigARN = aws.ToString(output.CodeSigningConfigArn)
			}
		}

		d.Set("code_signing_config_arn", codeSigningConfigARN)
	}

	return diags
}
