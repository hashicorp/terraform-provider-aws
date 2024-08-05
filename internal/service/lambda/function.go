// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambda;lambda.GetFunctionOutput")
// @Testing(importIgnore="filename;last_modified;publish")
func resourceFunction() *schema.Resource {
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
					ValidateDiagFunc: enum.Validate[awstypes.Architecture](),
				},
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
						names.AttrTargetARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEnvironment: {
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
						names.AttrSize: {
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
						names.AttrARN: {
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
				ExactlyOneOf: []string{"filename", "image_uri", names.AttrS3Bucket},
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
				ExactlyOneOf: []string{"filename", "image_uri", names.AttrS3Bucket},
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
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
			"logging_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_log_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: enum.Validate[awstypes.ApplicationLogLevel](),
							DiffSuppressFunc: suppressLoggingConfigUnspecifiedLogLevels,
						},
						"log_format": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogFormat](),
						},
						"log_group": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validLogGroupName(),
						},
						"system_log_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							ValidateDiagFunc: enum.Validate[awstypes.SystemLogLevel](),
							DiffSuppressFunc: suppressLoggingConfigUnspecifiedLogLevels,
						},
					},
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
				Default:          awstypes.PackageTypeZip,
				ValidateDiagFunc: enum.Validate[awstypes.PackageType](),
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
				Type:     schema.TypeBool,
				Optional: true,
			},
			"replacement_security_group_ids": {
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
			names.AttrRole: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"runtime": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.Runtime](),
			},
			names.AttrS3Bucket: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"filename", "image_uri", names.AttrS3Bucket},
				RequiredWith: []string{"s3_key"},
			},
			"s3_key": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{names.AttrS3Bucket},
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
			names.AttrSkipDestroy: {
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
							ValidateDiagFunc: enum.Validate[awstypes.SnapStartApplyOn](),
						},
						"optimization_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"source_code_hash": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
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
						names.AttrMode: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TracingMode](),
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
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv6_allowed_for_dual_stack": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
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

func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	functionName := d.Get("function_name").(string)
	packageType := awstypes.PackageType(d.Get("package_type").(string))
	input := &lambda.CreateFunctionInput{
		Code:         &awstypes.FunctionCode{},
		Description:  aws.String(d.Get(names.AttrDescription).(string)),
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int32(int32(d.Get("memory_size").(int))),
		PackageType:  packageType,
		Publish:      d.Get("publish").(bool),
		Role:         aws.String(d.Get(names.AttrRole).(string)),
		Tags:         getTagsIn(ctx),
		Timeout:      aws.Int32(int32(d.Get(names.AttrTimeout).(int))),
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
		input.Code.S3Bucket = aws.String(d.Get(names.AttrS3Bucket).(string))
		input.Code.S3Key = aws.String(d.Get("s3_key").(string))
		if v, ok := d.GetOk("s3_object_version"); ok {
			input.Code.S3ObjectVersion = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("architectures"); ok && len(v.([]interface{})) > 0 {
		input.Architectures = flex.ExpandStringyValueList[awstypes.Architecture](v.([]interface{}))
	}

	if v, ok := d.GetOk("code_signing_config_arn"); ok {
		input.CodeSigningConfigArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 {
		if v.([]interface{})[0] == nil {
			return sdkdiag.AppendErrorf(diags, "nil dead_letter_config supplied for function: %s", functionName)
		}

		input.DeadLetterConfig = &awstypes.DeadLetterConfig{
			TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})[names.AttrTargetARN].(string)),
		}
	}

	if v, ok := d.GetOk(names.AttrEnvironment); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
			input.Environment = &awstypes.Environment{
				Variables: flex.ExpandStringValueMap(v),
			}
		}
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EphemeralStorage = &awstypes.EphemeralStorage{
			Size: aws.Int32(int32(v.([]interface{})[0].(map[string]interface{})[names.AttrSize].(int))),
		}
	}

	if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
		input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
	}

	if packageType == awstypes.PackageTypeZip {
		input.Handler = aws.String(d.Get("handler").(string))
		input.Runtime = awstypes.Runtime(d.Get("runtime").(string))
	}

	if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
		input.ImageConfig = expandImageConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 {
		input.LoggingConfig = expandLoggingConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KMSKeyArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layers"); ok && len(v.([]interface{})) > 0 {
		input.Layers = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("snap_start"); ok {
		input.SnapStart = expandSnapStart(v.([]interface{}))
	}

	if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TracingConfig = &awstypes.TracingConfig{
			Mode: awstypes.TracingMode(v.([]interface{})[0].(map[string]interface{})[names.AttrMode].(string)),
		}
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input.VpcConfig = &awstypes.VpcConfig{
			Ipv6AllowedForDualStack: aws.Bool(tfMap["ipv6_allowed_for_dual_stack"].(bool)),
			SecurityGroupIds:        flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set)),
			SubnetIds:               flex.ExpandStringValueSet(tfMap[names.AttrSubnetIDs].(*schema.Set)),
		}
	}

	_, err := retryFunctionOp(ctx, func() (*lambda.CreateFunctionOutput, error) {
		return conn.CreateFunction(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Function (%s): %s", functionName, err)
	}

	d.SetId(functionName)

	_, err = tfresource.RetryWhenNotFound(ctx, lambdaPropagationTimeout, func() (interface{}, error) {
		return findFunctionByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "awiting for Lambda Function (%s) create: %s", d.Id(), err)
	}

	if _, err := waitFunctionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "awiting for Lambda Function (%s) create: %s", d.Id(), err)
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
	d.Set("architectures", function.Architectures)
	functionARN := aws.ToString(function.FunctionArn)
	d.Set(names.AttrARN, functionARN)
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
	if err := d.Set("image_config", flattenImageConfig(function.ImageConfigResponse)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting image_config: %s", err)
	}
	if output.Code != nil {
		d.Set("image_uri", output.Code.ImageUri)
	}
	d.Set("invoke_arn", invokeARN(meta.(*conns.AWSClient), functionARN))
	d.Set(names.AttrKMSKeyARN, function.KMSKeyArn)
	d.Set("last_modified", function.LastModified)
	if err := d.Set("layers", flattenLayers(function.Layers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting layers: %s", err)
	}
	if err := d.Set("logging_config", flattenLoggingConfig(function.LoggingConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
	}
	d.Set("memory_size", function.MemorySize)
	d.Set("package_type", function.PackageType)
	if output.Concurrency != nil {
		d.Set("reserved_concurrent_executions", output.Concurrency.ReservedConcurrentExecutions)
	} else {
		d.Set("reserved_concurrent_executions", -1)
	}
	d.Set(names.AttrRole, function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("signing_job_arn", function.SigningJobArn)
	d.Set("signing_profile_version_arn", function.SigningProfileVersionArn)
	// Support in-place update of non-refreshable attribute.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))
	if err := d.Set("snap_start", flattenSnapStart(function.SnapStart)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snap_start: %s", err)
	}
	d.Set("source_code_hash", d.Get("source_code_hash"))
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
	if err := d.Set(names.AttrVPCConfig, flattenVPCConfigResponse(function.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	if hasQualifier {
		d.Set("qualified_arn", functionARN)
		d.Set("qualified_invoke_arn", invokeARN(meta.(*conns.AWSClient), functionARN))
		d.Set(names.AttrVersion, function.Version)
	} else {
		latest, err := findLatestFunctionVersionByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) latest version: %s", d.Id(), err)
		}

		qualifiedARN := aws.ToString(latest.FunctionArn)
		d.Set("qualified_arn", qualifiedARN)
		d.Set("qualified_invoke_arn", invokeARN(meta.(*conns.AWSClient), qualifiedARN))
		d.Set(names.AttrVersion, latest.Version)

		setTagsOut(ctx, output.Tags)
	}

	// Currently, this functionality is only enabled in AWS Commercial partition
	// and other partitions return ambiguous error codes (e.g. AccessDeniedException
	// in AWS GovCloud (US)) so we cannot just ignore the error as would typically.
	// Currently this functionality is not enabled in all Regions and returns ambiguous error codes
	// (e.g. AccessDeniedException), so we cannot just ignore the error as we would typically.
	if partition, region := meta.(*conns.AWSClient).Partition, meta.(*conns.AWSClient).Region; partition == names.StandardPartitionID && signerServiceIsAvailable(region) {
		var codeSigningConfigARN string

		// Code Signing is only supported on zip packaged lambda functions.
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

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)

	if d.HasChange("code_signing_config_arn") {
		if v, ok := d.GetOk("code_signing_config_arn"); ok {
			input := &lambda.PutFunctionCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(v.(string)),
				FunctionName:         aws.String(d.Id()),
			}

			_, err := conn.PutFunctionCodeSigningConfig(ctx, input)

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

				input.DeadLetterConfig = &awstypes.DeadLetterConfig{
					TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})[names.AttrTargetARN].(string)),
				}
			} else {
				input.DeadLetterConfig = &awstypes.DeadLetterConfig{
					TargetArn: aws.String(""),
				}
			}
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChanges(names.AttrEnvironment, names.AttrKMSKeyARN) {
			input.Environment = &awstypes.Environment{
				Variables: map[string]string{},
			}

			if v, ok := d.GetOk(names.AttrEnvironment); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
					input.Environment = &awstypes.Environment{
						Variables: flex.ExpandStringValueMap(v),
					}
				}
			}
		}

		if d.HasChange("ephemeral_storage") {
			if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EphemeralStorage = &awstypes.EphemeralStorage{
					Size: aws.Int32(int32(v.([]interface{})[0].(map[string]interface{})[names.AttrSize].(int))),
				}
			}
		}

		if d.HasChange("file_system_config") {
			if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
				input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
			} else {
				input.FileSystemConfigs = []awstypes.FileSystemConfig{}
			}
		}

		if d.HasChange("handler") {
			input.Handler = aws.String(d.Get("handler").(string))
		}

		if d.HasChange("image_config") {
			if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
				input.ImageConfig = expandImageConfigs(v.([]interface{}))
			} else {
				input.ImageConfig = &awstypes.ImageConfig{}
			}
		}

		if d.HasChange(names.AttrKMSKeyARN) {
			input.KMSKeyArn = aws.String(d.Get(names.AttrKMSKeyARN).(string))
		}

		if d.HasChange("layers") {
			input.Layers = flex.ExpandStringValueList(d.Get("layers").([]interface{}))
		}

		if d.HasChange("logging_config") {
			input.LoggingConfig = expandLoggingConfig(d.Get("logging_config").([]interface{}))
		}

		if d.HasChange("memory_size") {
			input.MemorySize = aws.Int32(int32(d.Get("memory_size").(int)))
		}

		if d.HasChange(names.AttrRole) {
			input.Role = aws.String(d.Get(names.AttrRole).(string))
		}

		if d.HasChange("runtime") {
			input.Runtime = awstypes.Runtime(d.Get("runtime").(string))
		}

		if d.HasChange("snap_start") {
			input.SnapStart = expandSnapStart(d.Get("snap_start").([]interface{}))
		}

		if d.HasChange(names.AttrTimeout) {
			input.Timeout = aws.Int32(int32(d.Get(names.AttrTimeout).(int)))
		}

		if d.HasChange("tracing_config") {
			if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TracingConfig = &awstypes.TracingConfig{
					Mode: awstypes.TracingMode(v.([]interface{})[0].(map[string]interface{})[names.AttrMode].(string)),
				}
			}
		}

		if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids", "vpc_config.0.ipv6_allowed_for_dual_stack") {
			if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				input.VpcConfig = &awstypes.VpcConfig{
					Ipv6AllowedForDualStack: aws.Bool(tfMap["ipv6_allowed_for_dual_stack"].(bool)),
					SecurityGroupIds:        flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set)),
					SubnetIds:               flex.ExpandStringValueSet(tfMap[names.AttrSubnetIDs].(*schema.Set)),
				}
			} else {
				input.VpcConfig = &awstypes.VpcConfig{
					Ipv6AllowedForDualStack: aws.Bool(false),
					SecurityGroupIds:        []string{},
					SubnetIds:               []string{},
				}
			}
		}

		_, err := retryFunctionOp(ctx, func() (*lambda.UpdateFunctionConfigurationOutput, error) {
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
				input.Architectures = flex.ExpandStringyValueList[awstypes.Architecture](v.([]interface{}))
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
			input.S3Bucket = aws.String(d.Get(names.AttrS3Bucket).(string))
			input.S3Key = aws.String(d.Get("s3_key").(string))
			if v, ok := d.GetOk("s3_object_version"); ok {
				input.S3ObjectVersion = aws.String(v.(string))
			}
		}

		_, err := conn.UpdateFunctionCode(ctx, input)

		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Error occurred while GetObject.") {
				// As s3_bucket, s3_key and s3_object_version aren't set in resourceFunctionRead(), don't ovewrite the last known good values.
				for _, key := range []string{names.AttrS3Bucket, "s3_key", "s3_object_version"} {
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

		outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ResourceConflictException](ctx, lambdaPropagationTimeout, func() (interface{}, error) {
			return conn.PublishVersion(ctx, input)
		}, "in progress")

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

	if v, ok := d.GetOk(names.AttrSkipDestroy); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Lambda Function: %s", d.Id())
		return diags
	}

	if _, ok := d.GetOk("replace_security_groups_on_destroy"); ok {
		if err := replaceSecurityGroupsOnDestroy(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterValueException](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
			FunctionName: aws.String(d.Id()),
		})
	}, "because it is a replicated function")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s): %s", d.Id(), err)
	}

	return diags
}

func findFunctionByName(ctx context.Context, conn *lambda.Client, name string) (*lambda.GetFunctionOutput, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	return findFunction(ctx, conn, input)
}

func findFunction(ctx context.Context, conn *lambda.Client, input *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	output, err := conn.GetFunction(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findLatestFunctionVersionByName(ctx context.Context, conn *lambda.Client, name string) (*awstypes.FunctionConfiguration, error) {
	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(name),
		MaxItems:     aws.Int32(listVersionsMaxItems),
	}
	var output *awstypes.FunctionConfiguration

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

// replaceSecurityGroupsOnDestroy sets the VPC configuration security groups
// prior to resource destruction
//
// This function is called when the replace_security_groups_on_destroy
// argument is set. If the replacement_security_group_ids attribute is set,
// those values will be used as replacements. Otherwise, the default
// security group is used.
//
// Configuring this option can decrease destroy times for the security
// groups included in the VPC configuration block during normal operation
// by freeing them from association with ENI's left behind after destruction
// of the function.
func replaceSecurityGroupsOnDestroy(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaClient(ctx)
	ec2Conn := meta.(*conns.AWSClient).EC2Client(ctx)

	var sgIDs []string
	var vpcID string
	if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		sgIDs = flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set))
		vpcID = tfMap[names.AttrVPCID].(string)
	} else { // empty VPC config, nothing to do
		return nil
	}

	if len(sgIDs) == 0 { // no security groups, nothing to do
		return nil
	}

	var replacementSGIDs []string
	if v, ok := d.GetOk("replacement_security_group_ids"); ok {
		replacementSGIDs = flex.ExpandStringValueSet(v.(*schema.Set))
	} else {
		defaultSG, err := tfec2.FindSecurityGroupByNameAndVPCID(ctx, ec2Conn, "default", vpcID)
		if err != nil || defaultSG == nil {
			return fmt.Errorf("finding VPC (%s) default security group: %s", vpcID, err)
		}
		replacementSGIDs = []string{aws.ToString(defaultSG.GroupId)}
	}

	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(d.Id()),
		VpcConfig: &awstypes.VpcConfig{
			SecurityGroupIds: replacementSGIDs,
		},
	}

	if _, err := retryFunctionOp(ctx, func() (*lambda.UpdateFunctionConfigurationOutput, error) {
		return conn.UpdateFunctionConfiguration(ctx, input)
	}); err != nil {
		return fmt.Errorf("updating Lambda Function (%s) configuration: %s", d.Id(), err)
	}

	if _, err := waitFunctionUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for Lambda Function (%s) configuration update: %s", d.Id(), err)
	}

	return nil
}

func statusFunctionLastUpdateStatus(ctx context.Context, conn *lambda.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFunctionByName(ctx, conn, name)

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
		output, err := findFunctionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, string(output.Configuration.State), nil
	}
}

func waitFunctionCreated(ctx context.Context, conn *lambda.Client, name string, timeout time.Duration) (*awstypes.FunctionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatePending),
		Target:  enum.Slice(awstypes.StateActive),
		Refresh: statusFunctionState(ctx, conn, name),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.StateReasonCode), aws.ToString(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitFunctionUpdated(ctx context.Context, conn *lambda.Client, functionName string, timeout time.Duration) (*awstypes.FunctionConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LastUpdateStatusInProgress),
		Target:  enum.Slice(awstypes.LastUpdateStatusSuccessful),
		Refresh: statusFunctionLastUpdateStatus(ctx, conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.LastUpdateStatusReasonCode), aws.ToString(output.LastUpdateStatusReason)))

		return output, err
	}

	return nil, err
}

// retryFunctionOp retries a Lambda Function Create or Update operation.
// It handles IAM eventual consistency and EC2 throttling.
type functionCU interface {
	lambda.CreateFunctionOutput | lambda.UpdateFunctionConfigurationOutput
}

func retryFunctionOp[T functionCU](ctx context.Context, f func() (*T, error)) (*T, error) {
	output, err := tfresource.RetryWhen(ctx, lambdaPropagationTimeout,
		func() (interface{}, error) {
			return f()
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "The role defined for the function cannot be assumed by Lambda") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "The provided execution role does not have permissions") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "throttled by EC2") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Lambda was unable to configure access to your environment variables because the KMS key is invalid for CreateGrant") {
				return true, err
			}

			if errs.IsA[*awstypes.ResourceConflictException](err) {
				return true, err
			}

			return false, err
		},
	)

	// Additional retries when throttled.
	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "throttled by EC2") {
		const (
			functionExtraThrottlingTimeout = 9 * time.Minute
		)
		output, err = tfresource.RetryWhen(ctx, functionExtraThrottlingTimeout,
			func() (interface{}, error) {
				return f()
			},
			func(err error) (bool, error) {
				if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "throttled by EC2") {
					return true, err
				}

				return false, err
			},
		)
	}

	if err != nil {
		return nil, err
	}

	return output.(*T), err
}

func checkHandlerRuntimeForZipFunction(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	packageType := d.Get("package_type").(string)
	_, handlerOk := d.GetOk("handler")
	_, runtimeOk := d.GetOk("runtime")

	if packageType == string(awstypes.PackageTypeZip) && (!handlerOk || !runtimeOk) {
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
		d.SetNewComputed(names.AttrVersion)
		d.SetNewComputed("qualified_arn")
		d.SetNewComputed("qualified_invoke_arn")
	}
	return nil
}

func needsFunctionCodeUpdate(d sdkv2.ResourceDiffer) bool {
	return d.HasChange("filename") ||
		d.HasChange("source_code_hash") ||
		d.HasChange(names.AttrS3Bucket) ||
		d.HasChange("s3_key") ||
		d.HasChange("s3_object_version") ||
		d.HasChange("image_uri") ||
		d.HasChange("architectures")
}

func needsFunctionConfigUpdate(d sdkv2.ResourceDiffer) bool {
	return d.HasChange(names.AttrDescription) ||
		d.HasChange("handler") ||
		d.HasChange("file_system_config") ||
		d.HasChange("image_config") ||
		d.HasChange("logging_config") ||
		d.HasChange("memory_size") ||
		d.HasChange(names.AttrRole) ||
		d.HasChange(names.AttrTimeout) ||
		d.HasChange(names.AttrKMSKeyARN) ||
		d.HasChange("layers") ||
		d.HasChange("dead_letter_config") ||
		d.HasChange("snap_start") ||
		d.HasChange("tracing_config") ||
		d.HasChange("vpc_config.0.ipv6_allowed_for_dual_stack") ||
		d.HasChange("vpc_config.0.security_group_ids") ||
		d.HasChange("vpc_config.0.subnet_ids") ||
		d.HasChange("runtime") ||
		d.HasChange(names.AttrEnvironment) ||
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

// See https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-custom-integrations.html.
func invokeARN(c *conns.AWSClient, functionOrAliasARN string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "apigateway",
		Region:    c.Region,
		AccountID: "lambda",
		Resource:  fmt.Sprintf("path/2015-03-31/functions/%s/invocations", functionOrAliasARN),
	}.String()
}

// SignerServiceIsAvailable returns whether the AWS Signer service is available in the specified AWS Region.
// The AWS SDK endpoints package does not support Signer.
// See https://docs.aws.amazon.com/general/latest/gr/signer.html#signer_lambda_region.
func signerServiceIsAvailable(region string) bool {
	availableRegions := map[string]struct{}{
		names.USEast1RegionID:      {},
		names.USEast2RegionID:      {},
		names.USWest1RegionID:      {},
		names.USWest2RegionID:      {},
		names.AFSouth1RegionID:     {},
		names.APEast1RegionID:      {},
		names.APSouth1RegionID:     {},
		names.APNortheast2RegionID: {},
		names.APSoutheast1RegionID: {},
		names.APSoutheast2RegionID: {},
		names.APNortheast1RegionID: {},
		names.CACentral1RegionID:   {},
		names.EUCentral1RegionID:   {},
		names.EUWest1RegionID:      {},
		names.EUWest2RegionID:      {},
		names.EUSouth1RegionID:     {},
		names.EUWest3RegionID:      {},
		names.EUNorth1RegionID:     {},
		names.MESouth1RegionID:     {},
		names.SAEast1RegionID:      {},
	}
	_, ok := availableRegions[region]

	return ok
}

func flattenEnvironment(apiObject *awstypes.EnvironmentResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Variables; v != nil {
		tfMap["variables"] = aws.StringMap(v)
	}

	return []interface{}{tfMap}
}

func flattenFileSystemConfigs(apiObjects []awstypes.FileSystemConfig) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
		tfMap["local_mount_path"] = aws.ToString(apiObject.LocalMountPath)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandFileSystemConfigs(tfList []interface{}) []awstypes.FileSystemConfig {
	apiObjects := make([]awstypes.FileSystemConfig, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObjects = append(apiObjects, awstypes.FileSystemConfig{
			Arn:            aws.String(tfMap[names.AttrARN].(string)),
			LocalMountPath: aws.String(tfMap["local_mount_path"].(string)),
		})
	}

	return apiObjects
}

func flattenImageConfig(apiObject *awstypes.ImageConfigResponse) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject == nil || apiObject.Error != nil || apiObject.ImageConfig == nil {
		return nil
	}

	tfMap["command"] = apiObject.ImageConfig.Command
	tfMap["entry_point"] = apiObject.ImageConfig.EntryPoint
	tfMap["working_directory"] = apiObject.ImageConfig.WorkingDirectory

	return []interface{}{tfMap}
}

func expandImageConfigs(tfList []interface{}) *awstypes.ImageConfig {
	apiObject := &awstypes.ImageConfig{}

	// only one image_config block is allowed
	if len(tfList) == 1 && tfList[0] != nil {
		tfMap := tfList[0].(map[string]interface{})

		if len(tfMap["entry_point"].([]interface{})) > 0 {
			apiObject.EntryPoint = flex.ExpandStringValueList(tfMap["entry_point"].([]interface{}))
		}

		if len(tfMap["command"].([]interface{})) > 0 {
			apiObject.Command = flex.ExpandStringValueList(tfMap["command"].([]interface{}))
		}

		apiObject.WorkingDirectory = aws.String(tfMap["working_directory"].(string))
	}

	return apiObject
}

func expandLoggingConfig(tfList []interface{}) *awstypes.LoggingConfig {
	apiObject := &awstypes.LoggingConfig{}

	if len(tfList) == 1 && tfList[0] != nil {
		tfMap := tfList[0].(map[string]interface{})

		if v := tfMap["application_log_level"].(string); len(v) > 0 {
			apiObject.ApplicationLogLevel = awstypes.ApplicationLogLevel(v)
		}

		if v := tfMap["log_format"].(string); len(v) > 0 {
			apiObject.LogFormat = awstypes.LogFormat(v)
		}

		if v := tfMap["log_group"].(string); len(v) > 0 {
			apiObject.LogGroup = aws.String(v)
		}

		if v := tfMap["system_log_level"].(string); len(v) > 0 {
			apiObject.SystemLogLevel = awstypes.SystemLogLevel(v)
		}
	}

	return apiObject
}

func flattenLoggingConfig(apiObject *awstypes.LoggingConfig) []map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"application_log_level": apiObject.ApplicationLogLevel,
		"log_format":            apiObject.LogFormat,
		"log_group":             aws.ToString(apiObject.LogGroup),
		"system_log_level":      apiObject.SystemLogLevel,
	}

	return []map[string]interface{}{tfMap}
}

// Suppress diff if log levels have not been specified, unless log_format has changed
func suppressLoggingConfigUnspecifiedLogLevels(k, old, new string, d *schema.ResourceData) bool {
	if d.HasChanges("logging_config.0.log_format") {
		return false
	}
	if old != "" && new == "" {
		return true
	}
	return false
}

func flattenEphemeralStorage(apiObject *awstypes.EphemeralStorage) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap[names.AttrSize] = aws.ToInt32(apiObject.Size)

	return []interface{}{tfMap}
}

func expandSnapStart(tfList []interface{}) *awstypes.SnapStart {
	apiObject := &awstypes.SnapStart{
		ApplyOn: awstypes.SnapStartApplyOnNone,
	}

	if len(tfList) == 1 && tfList[0] != nil {
		tfMap := tfList[0].(map[string]interface{})
		apiObject.ApplyOn = awstypes.SnapStartApplyOn(tfMap["apply_on"].(string))
	}

	return apiObject
}

func flattenSnapStart(apiObject *awstypes.SnapStartResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	if apiObject.ApplyOn == awstypes.SnapStartApplyOnNone {
		return nil
	}

	tfMap := map[string]interface{}{
		"apply_on":            apiObject.ApplyOn,
		"optimization_status": apiObject.OptimizationStatus,
	}

	return []interface{}{tfMap}
}
