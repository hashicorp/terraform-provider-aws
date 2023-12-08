// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	FunctionVersionLatest = "$LATEST"
	mutexKey              = `aws_lambda_function`
	listVersionsMaxItems  = 10000
)

// @SDKResource("aws_lambda_function", name="Function")
// @Tags(identifierAttribute="arn")
func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("function_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"architectures": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.Architecture](),
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"code_signing_config_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"dead_letter_config": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environment": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				// Suppress diff if change is to an empty list
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "0" && new == "1" {
						_, n := d.GetChange("environment.0.variables")
						newn, ok := n.(map[string]interface{})
						if ok && len(newn) == 0 {
							return true
						}
					}
					return false
				},
			},
			"ephemeral_storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(512, 10240),
						},
					},
				},
			},
			"file_system_config": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				// Lambda file system supports 1 EFS file system per lambda function. This might increase in future.
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// EFS access point arn
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						// Local mount path inside a lambda function. Must start with "/mnt/".
						"local_mount_path": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`^/mnt/[0-9A-Za-z_.-]+$`), "must start with '/mnt/'"),
						},
					},
				},
			},
			"filename": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"filename", "image_uri", "s3_bucket"},
			},
			"function_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validFunctionName(),
			},
			"handler": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"image_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"entry_point": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"working_directory": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"image_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"filename", "image_uri", "s3_bucket"},
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layers": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"memory_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      128,
				ValidateFunc: validation.IntBetween(128, 10240),
			},
			"package_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          types.PackageTypeZip,
				ValidateDiagFunc: enum.Validate[types.PackageType](),
			},
			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"qualified_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualified_invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace_security_groups_on_destroy": {
				Deprecated: "AWS no longer supports this operation. This attribute now has " +
					"no effect and will be removed in a future major version.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"replacement_security_group_ids": {
				Deprecated: "AWS no longer supports this operation. This attribute now has " +
					"no effect and will be removed in a future major version.",
				Type:         schema.TypeSet,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				RequiredWith: []string{"replace_security_groups_on_destroy"},
			},
			"reserved_concurrent_executions": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.IntAtLeast(-1),
			},
			"role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"runtime": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.Runtime](),
			},
			"s3_bucket": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"filename", "image_uri", "s3_bucket"},
				RequiredWith: []string{"s3_key"},
			},
			"s3_key": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"s3_bucket"},
			},
			"s3_object_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filename", "image_uri"},
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
				Optional: true,
				Default:  false,
			},
			"snap_start": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_on": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.SnapStartApplyOn](),
						},
						"optimization_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"source_code_hash": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(1, 900),
			},
			"tracing_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.TracingMode](),
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv6_allowed_for_dual_stack": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},

				// Suppress diffs if the VPC configuration is provided, but empty
				// which is a valid Lambda function configuration. e.g.
				//   vpc_config {
				//     ipv6_allowed_for_dual_stack = false
				//     security_group_ids          = []
				//     subnet_ids                  = []
				//   }
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Id() == "" || old == "1" || new == "0" {
						return false
					}

					if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids", "vpc_config.0.ipv6_allowed_for_dual_stack") {
						return false
					}

					return true
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			checkHandlerRuntimeForZipFunction,
			updateComputedAttributesOnPublish,
			verify.SetTagsDiff,
		),
	}
}

const (
	functionExtraThrottlingTimeout = 9 * time.Minute
)

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	packageType := types.PackageType(d.Get("package_type").(string))
	input := &lambda.CreateFunctionInput{
		Code:         &types.FunctionCode{},
		Description:  aws.String(d.Get("description").(string)),
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int32(int32(d.Get("memory_size").(int))),
		PackageType:  packageType,
		Publish:      d.Get("publish").(bool),
		Role:         aws.String(d.Get("role").(string)),
		Tags:         getTagsIn(ctx),
		Timeout:      aws.Int32(int32(d.Get("timeout").(int))),
	}

	if v, ok := d.GetOk("filename"); ok {
		// Grab an exclusive lock so that we're only reading one function into memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364.
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		zipFile, err := readFileContents(v.(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ZIP file (%s): %s", v, err)
		}

		input.Code.ZipFile = zipFile
	} else if v, ok := d.GetOk("image_uri"); ok {
		input.Code.ImageUri = aws.String(v.(string))
	} else {
		input.Code.S3Bucket = aws.String(d.Get("s3_bucket").(string))
		input.Code.S3Key = aws.String(d.Get("s3_key").(string))
		if v, ok := d.GetOk("s3_object_version"); ok {
			input.Code.S3ObjectVersion = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("architectures"); ok && len(v.([]interface{})) > 0 {
		input.Architectures = expandArchitectures(v.([]interface{}))
	}

	if v, ok := d.GetOk("code_signing_config_arn"); ok {
		input.CodeSigningConfigArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 {
		if v.([]interface{})[0] == nil {
			return sdkdiag.AppendErrorf(diags, "nil dead_letter_config supplied for function: %s", functionName)
		}

		input.DeadLetterConfig = &types.DeadLetterConfig{
			TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})["target_arn"].(string)),
		}
	}

	if v, ok := d.GetOk("environment"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
			input.Environment = &types.Environment{
				Variables: flex.ExpandStringValueMap(v),
			}
		}
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EphemeralStorage = &types.EphemeralStorage{
			Size: aws.Int32(int32(v.([]interface{})[0].(map[string]interface{})["size"].(int))),
		}
	}

	if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
		input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
	}

	if packageType == types.PackageTypeZip {
		input.Handler = aws.String(d.Get("handler").(string))
		input.Runtime = types.Runtime(d.Get("runtime").(string))
	}

	if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
		input.ImageConfig = expandImageConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KMSKeyArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layers"); ok && len(v.([]interface{})) > 0 {
		input.Layers = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("snap_start"); ok {
		input.SnapStart = expandSnapStart(v.([]interface{}))
	}

	if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TracingConfig = &types.TracingConfig{
			Mode: types.TracingMode(v.([]interface{})[0].(map[string]interface{})["mode"].(string)),
		}
	}

	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input.VpcConfig = &types.VpcConfig{
			Ipv6AllowedForDualStack: aws.Bool(tfMap["ipv6_allowed_for_dual_stack"].(bool)),
			SecurityGroupIds:        flex.ExpandStringValueSet(tfMap["security_group_ids"].(*schema.Set)),
			SubnetIds:               flex.ExpandStringValueSet(tfMap["subnet_ids"].(*schema.Set)),
		}
	}

	_, err := retryFunctionOp(ctx, func() (interface{}, error) {
		return conn.CreateFunction(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function (%s): %s", functionName, err)
	}

	d.SetId(functionName)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindFunctionByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function (%s): waiting for completion: %s", d.Id(), err)
	}

	if _, err := waitFunctionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function (%s): waiting for completion: %s", d.Id(), err)
	}

	if v, ok := d.Get("reserved_concurrent_executions").(int); ok && v >= 0 {
		_, err := conn.PutFunctionConcurrency(ctx, &lambda.PutFunctionConcurrencyInput{
			FunctionName:                 aws.String(d.Id()),
			ReservedConcurrentExecutions: aws.Int32(int32(v)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) concurrency: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(d.Id()),
	}

	var hasQualifier bool
	if v, ok := d.GetOk("qualifier"); ok {
		hasQualifier = true
		input.Qualifier = aws.String(v.(string))
	}

	output, err := findFunction(ctx, conn, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s): %s", d.Id(), err)
	}

	function := output.Configuration
	d.Set("architectures", flattenArchitectures(function.Architectures))
	functionARN := aws.ToString(function.FunctionArn)
	d.Set("arn", functionARN)
	if function.DeadLetterConfig != nil && function.DeadLetterConfig.TargetArn != nil {
		if err := d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				"target_arn": aws.ToString(function.DeadLetterConfig.TargetArn),
			},
		}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dead_letter_config: %s", err)
		}
	} else {
		d.Set("dead_letter_config", []interface{}{})
	}
	d.Set("description", function.Description)
	if err := d.Set("environment", flattenEnvironment(function.Environment)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting environment: %s", err)
	}
	if err := d.Set("ephemeral_storage", flattenEphemeralStorage(function.EphemeralStorage)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_storage: %s", err)
	}
	if err := d.Set("file_system_config", flattenFileSystemConfigs(function.FileSystemConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting file_system_config: %s", err)
	}
	d.Set("handler", function.Handler)
	if err := d.Set("image_config", FlattenImageConfig(function.ImageConfigResponse)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_config: %s", err)
	}
	if output.Code != nil {
		d.Set("image_uri", output.Code.ImageUri)
	}
	d.Set("invoke_arn", functionInvokeARN(functionARN, meta))
	d.Set("kms_key_arn", function.KMSKeyArn)
	d.Set("last_modified", function.LastModified)
	if err := d.Set("layers", flattenLayers(function.Layers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting layers: %s", err)
	}
	d.Set("memory_size", function.MemorySize)
	d.Set("package_type", function.PackageType)
	if output.Concurrency != nil {
		d.Set("reserved_concurrent_executions", output.Concurrency.ReservedConcurrentExecutions)
	} else {
		d.Set("reserved_concurrent_executions", -1)
	}
	d.Set("role", function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("signing_job_arn", function.SigningJobArn)
	d.Set("signing_profile_version_arn", function.SigningProfileVersionArn)
	// Support in-place update of non-refreshable attribute.
	d.Set("skip_destroy", d.Get("skip_destroy"))
	if err := d.Set("snap_start", flattenSnapStart(function.SnapStart)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snap_start: %s", err)
	}
	d.Set("source_code_hash", function.CodeSha256)
	d.Set("source_code_size", function.CodeSize)
	d.Set("timeout", function.Timeout)
	tracingConfigMode := types.TracingModePassThrough
	if function.TracingConfig != nil {
		tracingConfigMode = function.TracingConfig.Mode
	}
	if err := d.Set("tracing_config", []interface{}{
		map[string]interface{}{
			"mode": string(tracingConfigMode),
		},
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tracing_config: %s", err)
	}
	if err := d.Set("vpc_config", flattenVPCConfigResponse(function.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	if hasQualifier {
		d.Set("qualified_arn", functionARN)
		d.Set("qualified_invoke_arn", functionInvokeARN(functionARN, meta))
		d.Set("version", function.Version)
	} else {
		latest, err := findLatestFunctionVersionByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) latest version: %s", d.Id(), err)
		}

		qualifiedARN := aws.ToString(latest.FunctionArn)
		d.Set("qualified_arn", qualifiedARN)
		d.Set("qualified_invoke_arn", functionInvokeARN(qualifiedARN, meta))
		d.Set("version", latest.Version)

		setTagsOut(ctx, output.Tags)
	}

	// Currently, this functionality is only enabled in AWS Commercial partition
	// and other partitions return ambiguous error codes (e.g. AccessDeniedException
	// in AWS GovCloud (US)) so we cannot just ignore the error as would typically.
	// Currently this functionality is not enabled in all Regions and returns ambiguous error codes
	// (e.g. AccessDeniedException), so we cannot just ignore the error as we would typically.
	if partition := meta.(*conns.AWSClient).Partition; partition == endpoints.AwsPartitionID && SignerServiceIsAvailable(meta.(*conns.AWSClient).Region) {
		var codeSigningConfigArn string

		// Code Signing is only supported on zip packaged lambda functions.
		if function.PackageType == types.PackageTypeZip {
			output, err := conn.GetFunctionCodeSigningConfig(ctx, &lambda.GetFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) code signing config: %s", d.Id(), err)
			}

			if output != nil {
				codeSigningConfigArn = aws.ToString(output.CodeSigningConfigArn)
			}
		}

		d.Set("code_signing_config_arn", codeSigningConfigArn)
	}

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	if d.HasChange("code_signing_config_arn") {
		if v, ok := d.GetOk("code_signing_config_arn"); ok {
			_, err := conn.PutFunctionCodeSigningConfig(ctx, &lambda.PutFunctionCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(v.(string)),
				FunctionName:         aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) code signing config: %s", d.Id(), err)
			}
		} else {
			_, err := conn.DeleteFunctionCodeSigningConfig(ctx, &lambda.DeleteFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s) code signing config: %s", d.Id(), err)
			}
		}
	}

	configUpdate := needsFunctionConfigUpdate(d)
	if configUpdate {
		input := &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(d.Id()),
		}

		if d.HasChange("dead_letter_config") {
			if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 {
				if v.([]interface{})[0] == nil {
					return sdkdiag.AppendErrorf(diags, "nil dead_letter_config supplied for function: %s", d.Id())
				}

				input.DeadLetterConfig = &types.DeadLetterConfig{
					TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})["target_arn"].(string)),
				}
			} else {
				input.DeadLetterConfig = &types.DeadLetterConfig{
					TargetArn: aws.String(""),
				}
			}
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChanges("environment", "kms_key_arn") {
			input.Environment = &types.Environment{
				Variables: map[string]string{},
			}

			if v, ok := d.GetOk("environment"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
					input.Environment = &types.Environment{
						Variables: flex.ExpandStringValueMap(v),
					}
				}
			}
		}

		if d.HasChange("ephemeral_storage") {
			if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EphemeralStorage = &types.EphemeralStorage{
					Size: aws.Int32(int32(v.([]interface{})[0].(map[string]interface{})["size"].(int))),
				}
			}
		}

		if d.HasChange("file_system_config") {
			if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
				input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
			} else {
				input.FileSystemConfigs = []types.FileSystemConfig{}
			}
		}

		if d.HasChange("handler") {
			input.Handler = aws.String(d.Get("handler").(string))
		}

		if d.HasChange("image_config") {
			if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
				input.ImageConfig = expandImageConfigs(v.([]interface{}))
			} else {
				input.ImageConfig = &types.ImageConfig{}
			}
		}

		if d.HasChange("kms_key_arn") {
			input.KMSKeyArn = aws.String(d.Get("kms_key_arn").(string))
		}

		if d.HasChange("layers") {
			input.Layers = flex.ExpandStringValueList(d.Get("layers").([]interface{}))
		}

		if d.HasChange("memory_size") {
			input.MemorySize = aws.Int32(int32(d.Get("memory_size").(int)))
		}

		if d.HasChange("role") {
			input.Role = aws.String(d.Get("role").(string))
		}

		if d.HasChange("runtime") {
			input.Runtime = types.Runtime(d.Get("runtime").(string))
		}

		if d.HasChange("snap_start") {
			input.SnapStart = expandSnapStart(d.Get("snap_start").([]interface{}))
		}

		if d.HasChange("timeout") {
			input.Timeout = aws.Int32(int32(d.Get("timeout").(int)))
		}

		if d.HasChange("tracing_config") {
			if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TracingConfig = &types.TracingConfig{
					Mode: types.TracingMode(v.([]interface{})[0].(map[string]interface{})["mode"].(string)),
				}
			}
		}

		if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids", "vpc_config.0.ipv6_allowed_for_dual_stack") {
			if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				input.VpcConfig = &types.VpcConfig{
					Ipv6AllowedForDualStack: aws.Bool(tfMap["ipv6_allowed_for_dual_stack"].(bool)),
					SecurityGroupIds:        flex.ExpandStringValueSet(tfMap["security_group_ids"].(*schema.Set)),
					SubnetIds:               flex.ExpandStringValueSet(tfMap["subnet_ids"].(*schema.Set)),
				}
			} else {
				input.VpcConfig = &types.VpcConfig{
					Ipv6AllowedForDualStack: aws.Bool(false),
					SecurityGroupIds:        []string{},
					SubnetIds:               []string{},
				}
			}
		}

		_, err := retryFunctionOp(ctx, func() (interface{}, error) {
			return conn.UpdateFunctionConfiguration(ctx, input)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lambda Function (%s) configuration: %s", d.Id(), err)
		}

		if _, err := waitFunctionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Lambda Function (%s) configuration update: %s", d.Id(), err)
		}
	}

	codeUpdate := needsFunctionCodeUpdate(d)
	if codeUpdate {
		input := &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(d.Id()),
		}

		if d.HasChange("architectures") {
			if v, ok := d.GetOk("architectures"); ok && len(v.([]interface{})) > 0 {
				input.Architectures = expandArchitectures(v.([]interface{}))
			}
		}

		if v, ok := d.GetOk("filename"); ok {
			// Grab an exclusive lock so that we're only reading one function into memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(mutexKey)
			defer conns.GlobalMutexKV.Unlock(mutexKey)

			zipFile, err := readFileContents(v.(string))

			if err != nil {
				// As filename isn't set in resourceFunctionRead(), don't ovewrite the last known good value.
				old, _ := d.GetChange("filename")
				d.Set("filename", old)

				return sdkdiag.AppendErrorf(diags, "reading ZIP file (%s): %s", v, err)
			}

			input.ZipFile = zipFile
		} else if v, ok := d.GetOk("image_uri"); ok {
			input.ImageUri = aws.String(v.(string))
		} else {
			input.S3Bucket = aws.String(d.Get("s3_bucket").(string))
			input.S3Key = aws.String(d.Get("s3_key").(string))
			if v, ok := d.GetOk("s3_object_version"); ok {
				input.S3ObjectVersion = aws.String(v.(string))
			}
		}

		_, err := conn.UpdateFunctionCode(ctx, input)

		if err != nil {
			var ipve *types.InvalidParameterValueException
			if errors.As(err, &ipve) && strings.Contains(ipve.ErrorMessage(), "Error occurred while GetObject.") {
				// As s3_bucket, s3_key and s3_object_version aren't set in resourceFunctionRead(), don't ovewrite the last known good values.
				for _, key := range []string{"s3_bucket", "s3_key", "s3_object_version"} {
					old, _ := d.GetChange(key)
					d.Set(key, old)
				}
			}

			return sdkdiag.AppendErrorf(diags, "updating Lambda Function (%s) code: %s", d.Id(), err)
		}

		if _, err := waitFunctionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lambda Function (%s) code: waiting for completion: %s", d.Id(), err)
		}
	}

	if d.HasChange("reserved_concurrent_executions") {
		if v, ok := d.Get("reserved_concurrent_executions").(int); ok && v >= 0 {
			_, err := conn.PutFunctionConcurrency(ctx, &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String(d.Id()),
				ReservedConcurrentExecutions: aws.Int32(int32(v)),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) concurrency: %s", d.Id(), err)
			}
		} else {
			_, err := conn.DeleteFunctionConcurrency(ctx, &lambda.DeleteFunctionConcurrencyInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s) concurrency: %s", d.Id(), err)
			}
		}
	}

	if d.Get("publish").(bool) && (codeUpdate || configUpdate || d.HasChange("publish")) {
		input := &lambda.PublishVersionInput{
			FunctionName: aws.String(d.Id()),
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.PublishVersion(ctx, input)
			},
			func(err error) (bool, error) {
				var rce *types.ResourceConflictException
				if errors.As(err, &rce) && strings.Contains(rce.ErrorMessage(), "in progress") {
					return true, err
				}
				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing Lambda Function (%s) version: %s", d.Id(), err)
		}

		output := outputRaw.(*lambda.PublishVersionOutput)

		err = lambda.NewFunctionUpdatedWaiter(conn).Wait(ctx, &lambda.GetFunctionConfigurationInput{
			FunctionName: output.FunctionArn,
			Qualifier:    output.Version,
		}, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing Lambda Function (%s) version: waiting for completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Lambda Function: %s", d.Id())
		return diags
	}

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterValueException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(d.Id()),
		})
	}, "because it is a replicated function")

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s): %s", d.Id(), err)
	}

	return diags
}

func FindFunctionByName(ctx context.Context, conn *lambda.Client, name string) (*lambda.GetFunctionOutput, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	return findFunction(ctx, conn, input)
}

func findFunction(ctx context.Context, conn *lambda.Client, input *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	output, err := conn.GetFunction(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Code == nil || output.Configuration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findLatestFunctionVersionByName(ctx context.Context, conn *lambda.Client, name string) (*types.FunctionConfiguration, error) {
	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(name),
		MaxItems:     aws.Int32(listVersionsMaxItems),
	}
	var output *types.FunctionConfiguration

	pages := lambda.NewListVersionsByFunctionPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return output, err
		}
		if len(page.Versions) > 0 && page.NextMarker == nil {
			// List is sorted from oldest to latest.
			output = &page.Versions[len(page.Versions)-1]
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusFunctionLastUpdateStatus(ctx context.Context, conn *lambda.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFunctionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, string(output.Configuration.LastUpdateStatus), nil
	}
}

func statusFunctionState(ctx context.Context, conn *lambda.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFunctionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, string(output.Configuration.State), nil
	}
}

func waitFunctionCreated(ctx context.Context, conn *lambda.Client, name string, timeout time.Duration) (*types.FunctionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.StatePending),
		Target:  enum.Slice(types.StateActive),
		Refresh: statusFunctionState(ctx, conn, name),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.StateReasonCode), aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitFunctionUpdated(ctx context.Context, conn *lambda.Client, functionName string, timeout time.Duration) (*types.FunctionConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LastUpdateStatusInProgress),
		Target:  enum.Slice(types.LastUpdateStatusSuccessful),
		Refresh: statusFunctionLastUpdateStatus(ctx, conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.LastUpdateStatusReasonCode), aws.ToString(output.LastUpdateStatusReason)))

		return output, err
	}

	return nil, err
}

// retryFunctionOp retries a Lambda Function Create or Update operation.
// It handles IAM eventual consistency and EC2 throttling.
func retryFunctionOp(ctx context.Context, f func() (interface{}, error)) (interface{}, error) { //nolint:unparam
	output, err := tfresource.RetryWhen(ctx, propagationTimeout,
		f,
		func(err error) (bool, error) {
			var ipve *types.InvalidParameterValueException
			if errors.As(err, &ipve) {
				msg := ipve.ErrorMessage()
				if strings.Contains(msg, "The role defined for the function cannot be assumed by Lambda") {
					return true, err
				}
				if strings.Contains(msg, "The provided execution role does not have permissions") {
					return true, err
				}
				if strings.Contains(msg, "throttled by EC2") {
					return true, err
				}
				if strings.Contains(msg, "Lambda was unable to configure access to your environment variables because the KMS key is invalid for CreateGrant") {
					return true, err
				}
			}

			var rce *types.ResourceConflictException
			if errors.As(err, &rce) {
				return true, err
			}

			return false, err
		},
	)

	// Additional retries when throttled.
	var ipve *types.InvalidParameterValueException
	if errors.As(err, &ipve) && strings.Contains(ipve.ErrorMessage(), "throttled by EC2") {
		output, err = tfresource.RetryWhen(ctx, functionExtraThrottlingTimeout,
			f,
			func(err error) (bool, error) {
				var ipve *types.InvalidParameterValueException
				if errors.As(err, &ipve) && strings.Contains(ipve.ErrorMessage(), "throttled by EC2") {
					return true, err
				}

				return false, err
			},
		)
	}

	return output, err
}

func checkHandlerRuntimeForZipFunction(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	packageType := d.Get("package_type").(string)
	_, handlerOk := d.GetOk("handler")
	_, runtimeOk := d.GetOk("runtime")

	if packageType == string(types.PackageTypeZip) && (!handlerOk || !runtimeOk) {
		return fmt.Errorf("handler and runtime must be set when PackageType is Zip")
	}
	return nil
}

func updateComputedAttributesOnPublish(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	configChanged := needsFunctionConfigUpdate(d)
	codeChanged := needsFunctionCodeUpdate(d)
	if codeChanged {
		d.SetNewComputed("last_modified")
	}

	publish := d.Get("publish").(bool)
	publishChanged := d.HasChange("publish")
	if publish && (configChanged || codeChanged || publishChanged) {
		d.SetNewComputed("version")
		d.SetNewComputed("qualified_arn")
		d.SetNewComputed("qualified_invoke_arn")
	}
	return nil
}

func needsFunctionCodeUpdate(d verify.ResourceDiffer) bool {
	return d.HasChange("filename") ||
		d.HasChange("source_code_hash") ||
		d.HasChange("s3_bucket") ||
		d.HasChange("s3_key") ||
		d.HasChange("s3_object_version") ||
		d.HasChange("image_uri") ||
		d.HasChange("architectures")
}

func needsFunctionConfigUpdate(d verify.ResourceDiffer) bool {
	return d.HasChange("description") ||
		d.HasChange("handler") ||
		d.HasChange("file_system_config") ||
		d.HasChange("image_config") ||
		d.HasChange("memory_size") ||
		d.HasChange("role") ||
		d.HasChange("timeout") ||
		d.HasChange("kms_key_arn") ||
		d.HasChange("layers") ||
		d.HasChange("dead_letter_config") ||
		d.HasChange("snap_start") ||
		d.HasChange("tracing_config") ||
		d.HasChange("vpc_config.0.ipv6_allowed_for_dual_stack") ||
		d.HasChange("vpc_config.0.security_group_ids") ||
		d.HasChange("vpc_config.0.subnet_ids") ||
		d.HasChange("runtime") ||
		d.HasChange("environment") ||
		d.HasChange("ephemeral_storage")
}

func readFileContents(v string) ([]byte, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, err
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return fileContent, nil
}

func functionInvokeARN(functionARN string, meta interface{}) string {
	return arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: "lambda",
		Resource:  fmt.Sprintf("path/2015-03-31/functions/%s/invocations", functionARN),
	}.String()
}

// SignerServiceIsAvailable returns whether the AWS Signer service is available in the specified AWS Region.
// The AWS SDK endpoints package does not support Signer.
// See https://docs.aws.amazon.com/general/latest/gr/signer.html#signer_lambda_region.
func SignerServiceIsAvailable(region string) bool {
	availableRegions := map[string]struct{}{
		endpoints.UsEast2RegionID:      {},
		endpoints.UsEast1RegionID:      {},
		endpoints.UsWest1RegionID:      {},
		endpoints.UsWest2RegionID:      {},
		endpoints.AfSouth1RegionID:     {},
		endpoints.ApEast1RegionID:      {},
		endpoints.ApSouth1RegionID:     {},
		endpoints.ApNortheast2RegionID: {},
		endpoints.ApSoutheast1RegionID: {},
		endpoints.ApSoutheast2RegionID: {},
		endpoints.ApNortheast1RegionID: {},
		endpoints.CaCentral1RegionID:   {},
		endpoints.EuCentral1RegionID:   {},
		endpoints.EuWest1RegionID:      {},
		endpoints.EuWest2RegionID:      {},
		endpoints.EuSouth1RegionID:     {},
		endpoints.EuWest3RegionID:      {},
		endpoints.EuNorth1RegionID:     {},
		endpoints.MeSouth1RegionID:     {},
		endpoints.SaEast1RegionID:      {},
	}
	_, ok := availableRegions[region]

	return ok
}

func flattenEnvironment(apiObject *types.EnvironmentResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Variables; v != nil {
		tfMap["variables"] = aws.StringMap(v)
	}

	return []interface{}{tfMap}
}

func flattenFileSystemConfigs(fscList []types.FileSystemConfig) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(fscList))
	for _, fsc := range fscList {
		f := make(map[string]interface{})
		f["arn"] = aws.ToString(fsc.Arn)
		f["local_mount_path"] = aws.ToString(fsc.LocalMountPath)

		results = append(results, f)
	}
	return results
}

func expandFileSystemConfigs(fscMaps []interface{}) []types.FileSystemConfig {
	fileSystemConfigs := make([]types.FileSystemConfig, 0, len(fscMaps))
	for _, fsc := range fscMaps {
		fscMap := fsc.(map[string]interface{})
		fileSystemConfigs = append(fileSystemConfigs, types.FileSystemConfig{
			Arn:            aws.String(fscMap["arn"].(string)),
			LocalMountPath: aws.String(fscMap["local_mount_path"].(string)),
		})
	}
	return fileSystemConfigs
}

func FlattenImageConfig(response *types.ImageConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if response == nil || response.Error != nil || response.ImageConfig == nil {
		return nil
	}

	settings["command"] = response.ImageConfig.Command
	settings["entry_point"] = response.ImageConfig.EntryPoint
	settings["working_directory"] = response.ImageConfig.WorkingDirectory

	return []map[string]interface{}{settings}
}

func expandImageConfigs(imageConfigMaps []interface{}) *types.ImageConfig {
	imageConfig := &types.ImageConfig{}
	// only one image_config block is allowed
	if len(imageConfigMaps) == 1 && imageConfigMaps[0] != nil {
		config := imageConfigMaps[0].(map[string]interface{})
		if len(config["entry_point"].([]interface{})) > 0 {
			imageConfig.EntryPoint = flex.ExpandStringValueList(config["entry_point"].([]interface{}))
		}
		if len(config["command"].([]interface{})) > 0 {
			imageConfig.Command = flex.ExpandStringValueList(config["command"].([]interface{}))
		}
		imageConfig.WorkingDirectory = aws.String(config["working_directory"].(string))
	}
	return imageConfig
}

func flattenEphemeralStorage(response *types.EphemeralStorage) []map[string]interface{} {
	if response == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["size"] = aws.ToInt32(response.Size)

	return []map[string]interface{}{m}
}

func expandSnapStart(tfList []interface{}) *types.SnapStart {
	snapStart := &types.SnapStart{ApplyOn: types.SnapStartApplyOnNone}
	if len(tfList) == 1 && tfList[0] != nil {
		item := tfList[0].(map[string]interface{})
		snapStart.ApplyOn = types.SnapStartApplyOn(item["apply_on"].(string))
	}
	return snapStart
}

func flattenSnapStart(apiObject *types.SnapStartResponse) []interface{} {
	if apiObject == nil {
		return nil
	}
	if apiObject.ApplyOn == types.SnapStartApplyOnNone {
		return nil
	}
	m := map[string]interface{}{
		"apply_on":            string(apiObject.ApplyOn),
		"optimization_status": string(apiObject.OptimizationStatus),
	}

	return []interface{}{m}
}

func expandArchitectures(tfList []interface{}) []types.Architecture {
	vs := make([]types.Architecture, 0, len(tfList))
	for _, v := range tfList {
		vs = append(vs, types.Architecture(v.(string)))
	}
	return vs
}

func flattenArchitectures(apiObject []types.Architecture) []interface{} {
	vs := make([]interface{}, 0, len(apiObject))
	for _, v := range apiObject {
		vs = append(vs, string(v))
	}
	return vs
}
