package lambda

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	FunctionVersionLatest = "$LATEST"
	mutexKey              = `aws_lambda_function`
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFunctionCreate,
		ReadWithoutTimeout:   resourceFunctionRead,
		UpdateWithoutTimeout: resourceFunctionUpdate,
		DeleteWithoutTimeout: resourceFunctionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
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
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lambda.Architecture_Values(), false),
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
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/mnt/[a-zA-Z0-9-_.]+$`), "must start with '/mnt/'"),
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
				Type:     schema.TypeInt,
				Optional: true,
				Default:  128,
			},
			"package_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      lambda.PackageTypeZip,
				ValidateFunc: validation.StringInSlice(lambda.PackageType_Values(), false),
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
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"runtime": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(lambda.Runtime_Values(), false),
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
			"snap_start": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_on": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lambda.SnapStartApplyOn_Values(), true),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"tracing_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lambda.TracingMode_Values(), true),
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
				//     security_group_ids = []
				//     subnet_ids         = []
				//   }
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.Id() == "" || old == "1" || new == "0" {
						return false
					}

					if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids") {
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
	conn := meta.(*conns.AWSClient).LambdaConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	functionName := d.Get("function_name").(string)
	packageType := d.Get("package_type").(string)
	input := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{},
		Description:  aws.String(d.Get("description").(string)),
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int64(int64(d.Get("memory_size").(int))),
		PackageType:  aws.String(packageType),
		Publish:      aws.Bool(d.Get("publish").(bool)),
		Role:         aws.String(d.Get("role").(string)),
		Timeout:      aws.Int64(int64(d.Get("timeout").(int))),
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
		input.Architectures = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("code_signing_config_arn"); ok {
		input.CodeSigningConfigArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok && len(v.([]interface{})) > 0 {
		if v.([]interface{})[0] == nil {
			return sdkdiag.AppendErrorf(diags, "nil dead_letter_config supplied for function: %s", functionName)
		}

		input.DeadLetterConfig = &lambda.DeadLetterConfig{
			TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})["target_arn"].(string)),
		}
	}

	if v, ok := d.GetOk("environment"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
			input.Environment = &lambda.Environment{
				Variables: flex.ExpandStringMap(v),
			}
		}
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EphemeralStorage = &lambda.EphemeralStorage{
			Size: aws.Int64(int64(v.([]interface{})[0].(map[string]interface{})["size"].(int))),
		}
	}

	if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
		input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
	}

	if packageType == lambda.PackageTypeZip {
		input.Handler = aws.String(d.Get("handler").(string))
		input.Runtime = aws.String(d.Get("runtime").(string))
	}

	if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
		input.ImageConfig = expandImageConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KMSKeyArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layers"); ok && len(v.([]interface{})) > 0 {
		input.Layers = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("snap_start"); ok {
		input.SnapStart = expandSnapStart(v.([]interface{}))
	}

	if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TracingConfig = &lambda.TracingConfig{
			Mode: aws.String(v.([]interface{})[0].(map[string]interface{})["mode"].(string)),
		}
	}

	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input.VpcConfig = &lambda.VpcConfig{
			SecurityGroupIds: flex.ExpandStringSet(tfMap["security_group_ids"].(*schema.Set)),
			SubnetIds:        flex.ExpandStringSet(tfMap["subnet_ids"].(*schema.Set)),
		}
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := retryFunctionOp(ctx, func() (interface{}, error) {
		return conn.CreateFunctionWithContext(ctx, input)
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
		_, err := conn.PutFunctionConcurrencyWithContext(ctx, &lambda.PutFunctionConcurrencyInput{
			FunctionName:                 aws.String(d.Id()),
			ReservedConcurrentExecutions: aws.Int64(int64(v)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) concurrency: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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
	d.Set("architectures", aws.StringValueSlice(function.Architectures))
	functionARN := aws.StringValue(function.FunctionArn)
	d.Set("arn", functionARN)
	if function.DeadLetterConfig != nil && function.DeadLetterConfig.TargetArn != nil {
		if err := d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				"target_arn": aws.StringValue(function.DeadLetterConfig.TargetArn),
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
	if err := d.Set("snap_start", flattenSnapStart(function.SnapStart)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snap_start: %s", err)
	}
	d.Set("source_code_hash", function.CodeSha256)
	d.Set("source_code_size", function.CodeSize)
	d.Set("timeout", function.Timeout)
	tracingConfigMode := lambda.TracingModePassThrough
	if function.TracingConfig != nil {
		tracingConfigMode = aws.StringValue(function.TracingConfig.Mode)
	}
	if err := d.Set("tracing_config", []interface{}{
		map[string]interface{}{
			"mode": tracingConfigMode,
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

		qualifiedARN := aws.StringValue(latest.FunctionArn)
		d.Set("qualified_arn", qualifiedARN)
		d.Set("qualified_invoke_arn", functionInvokeARN(qualifiedARN, meta))
		d.Set("version", latest.Version)

		// Tagging operations are permitted on Lambda functions only.
		// Tags on aliases and versions are not supported.
		tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
		}
	}

	// Currently, this functionality is only enabled in AWS Commercial partition
	// and other partitions return ambiguous error codes (e.g. AccessDeniedException
	// in AWS GovCloud (US)) so we cannot just ignore the error as would typically.
	// Currently this functionality is not enabled in all Regions and returns ambiguous error codes
	// (e.g. AccessDeniedException), so we cannot just ignore the error as we would typically.
	if partition := meta.(*conns.AWSClient).Partition; partition == endpoints.AwsPartitionID && SignerServiceIsAvailable(meta.(*conns.AWSClient).Region) {
		var codeSigningConfigArn string

		// Code Signing is only supported on zip packaged lambda functions.
		if aws.StringValue(function.PackageType) == lambda.PackageTypeZip {
			output, err := conn.GetFunctionCodeSigningConfigWithContext(ctx, &lambda.GetFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Lambda Function (%s) code signing config: %s", d.Id(), err)
			}

			if output != nil {
				codeSigningConfigArn = aws.StringValue(output.CodeSigningConfigArn)
			}
		}

		d.Set("code_signing_config_arn", codeSigningConfigArn)
	}

	return diags
}

func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn()

	if d.HasChange("code_signing_config_arn") {
		if v, ok := d.GetOk("code_signing_config_arn"); ok {
			_, err := conn.PutFunctionCodeSigningConfigWithContext(ctx, &lambda.PutFunctionCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(v.(string)),
				FunctionName:         aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) code signing config: %s", d.Id(), err)
			}
		} else {
			_, err := conn.DeleteFunctionCodeSigningConfigWithContext(ctx, &lambda.DeleteFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s) code signing config: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lambda Function (%s) tags: %s", arn, err)
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

				input.DeadLetterConfig = &lambda.DeadLetterConfig{
					TargetArn: aws.String(v.([]interface{})[0].(map[string]interface{})["target_arn"].(string)),
				}
			} else {
				input.DeadLetterConfig = &lambda.DeadLetterConfig{
					TargetArn: aws.String(""),
				}
			}
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChanges("environment", "kms_key_arn") {
			input.Environment = &lambda.Environment{
				Variables: map[string]*string{},
			}

			if v, ok := d.GetOk("environment"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				if v, ok := v.([]interface{})[0].(map[string]interface{})["variables"].(map[string]interface{}); ok && len(v) > 0 {
					input.Environment = &lambda.Environment{
						Variables: flex.ExpandStringMap(v),
					}
				}
			}
		}

		if d.HasChange("ephemeral_storage") {
			if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EphemeralStorage = &lambda.EphemeralStorage{
					Size: aws.Int64(int64(v.([]interface{})[0].(map[string]interface{})["size"].(int))),
				}
			}
		}

		if d.HasChange("file_system_config") {
			if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
				input.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
			} else {
				input.FileSystemConfigs = []*lambda.FileSystemConfig{}
			}
		}

		if d.HasChange("handler") {
			input.Handler = aws.String(d.Get("handler").(string))
		}

		if d.HasChange("image_config") {
			if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
				input.ImageConfig = expandImageConfigs(v.([]interface{}))
			} else {
				input.ImageConfig = &lambda.ImageConfig{}
			}
		}

		if d.HasChange("kms_key_arn") {
			input.KMSKeyArn = aws.String(d.Get("kms_key_arn").(string))
		}

		if d.HasChange("layers") {
			input.Layers = flex.ExpandStringList(d.Get("layers").([]interface{}))
		}

		if d.HasChange("memory_size") {
			input.MemorySize = aws.Int64(int64(d.Get("memory_size").(int)))
		}

		if d.HasChange("role") {
			input.Role = aws.String(d.Get("role").(string))
		}

		if d.HasChange("runtime") {
			input.Runtime = aws.String(d.Get("runtime").(string))
		}

		if d.HasChange("snap_start") {
			input.SnapStart = expandSnapStart(d.Get("snap_start").([]interface{}))
		}

		if d.HasChange("timeout") {
			input.Timeout = aws.Int64(int64(d.Get("timeout").(int)))
		}

		if d.HasChange("tracing_config") {
			if v, ok := d.GetOk("tracing_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TracingConfig = &lambda.TracingConfig{
					Mode: aws.String(v.([]interface{})[0].(map[string]interface{})["mode"].(string)),
				}
			}
		}

		if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids") {
			if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				input.VpcConfig = &lambda.VpcConfig{
					SecurityGroupIds: flex.ExpandStringSet(tfMap["security_group_ids"].(*schema.Set)),
					SubnetIds:        flex.ExpandStringSet(tfMap["subnet_ids"].(*schema.Set)),
				}
			} else {
				input.VpcConfig = &lambda.VpcConfig{
					SecurityGroupIds: []*string{},
					SubnetIds:        []*string{},
				}
			}
		}

		_, err := retryFunctionOp(ctx, func() (interface{}, error) {
			return conn.UpdateFunctionConfigurationWithContext(ctx, input)
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
				input.Architectures = flex.ExpandStringList(v.([]interface{}))
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

		_, err := conn.UpdateFunctionCodeWithContext(ctx, input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "Error occurred while GetObject.") {
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
			_, err := conn.PutFunctionConcurrencyWithContext(ctx, &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String(d.Id()),
				ReservedConcurrentExecutions: aws.Int64(int64(v)),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Lambda Function (%s) concurrency: %s", d.Id(), err)
			}
		} else {
			_, err := conn.DeleteFunctionConcurrencyWithContext(ctx, &lambda.DeleteFunctionConcurrencyInput{
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

		outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.PublishVersionWithContext(ctx, input)
		}, lambda.ErrCodeResourceConflictException, "in progress")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing Lambda Function (%s) version: %s", d.Id(), err)
		}

		output := outputRaw.(*lambda.FunctionConfiguration)

		err = conn.WaitUntilFunctionUpdatedWithContext(ctx, &lambda.GetFunctionConfigurationInput{
			FunctionName: output.FunctionArn,
			Qualifier:    output.Version,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "publishing Lambda Function (%s) version: waiting for completion: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFunctionRead(ctx, d, meta)...)
}

func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn()

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())
	_, err := conn.DeleteFunctionWithContext(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Function (%s): %s", d.Id(), err)
	}

	if _, ok := d.GetOk("replace_security_groups_on_destroy"); ok {
		if err := replaceSecurityGroups(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return diags
}

func FindFunctionByName(ctx context.Context, conn *lambda.Lambda, name string) (*lambda.GetFunctionOutput, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	return findFunction(ctx, conn, input)
}

func findFunction(ctx context.Context, conn *lambda.Lambda, input *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	output, err := conn.GetFunctionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func findLatestFunctionVersionByName(ctx context.Context, conn *lambda.Lambda, name string) (*lambda.FunctionConfiguration, error) {
	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(name),
		MaxItems:     aws.Int64(10000),
	}
	var output *lambda.FunctionConfiguration

	err := conn.ListVersionsByFunctionPagesWithContext(ctx, input, func(page *lambda.ListVersionsByFunctionOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		// List is sorted from oldest to latest.
		if lastPage {
			output = page.Versions[len(page.Versions)-1]
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// replaceSecurityGroups will replace the security groups on orphaned lambda ENI's
//
// If the replacement_security_group_ids attribute is set, those values will be used as
// replacements. Otherwise, the default security group is used.
func replaceSecurityGroups(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	ec2Conn := meta.(*conns.AWSClient).EC2Conn()

	var sgIDs []string
	var vpcID string
	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		sgIDs = flex.ExpandStringValueSet(tfMap["security_group_ids"].(*schema.Set))
		vpcID = tfMap["vpc_id"].(string)
	} else { // empty VPC config, nothing to do
		return nil
	}

	if len(sgIDs) == 0 { // no security groups, nothing to do
		return nil
	}

	var replacmentSGIDs []*string
	if v, ok := d.GetOk("replacement_security_group_ids"); ok {
		replacmentSGIDs = flex.ExpandStringSet(v.(*schema.Set))
	} else {
		defaultSG, err := tfec2.FindSecurityGroupByNameAndVPCID(ctx, ec2Conn, "default", vpcID)
		if err != nil || defaultSG == nil {
			return fmt.Errorf("finding VPC (%s) default security group: %s", vpcID, err)
		}
		replacmentSGIDs = []*string{defaultSG.GroupId}
	}

	networkInterfaces, err := tfec2.FindLambdaNetworkInterfacesBySecurityGroupIDsAndFunctionName(ctx, ec2Conn, sgIDs, d.Id())
	if err != nil {
		return fmt.Errorf("finding Lambda Function (%s) network interfaces: %s", d.Id(), err)
	}

	for _, ni := range networkInterfaces {
		_, err := ec2Conn.ModifyNetworkInterfaceAttributeWithContext(ctx, &ec2.ModifyNetworkInterfaceAttributeInput{
			NetworkInterfaceId: ni.NetworkInterfaceId,
			Groups:             replacmentSGIDs,
		})
		if err != nil {
			return fmt.Errorf("modifying Lambda Function (%s) network interfaces: %s", d.Id(), err)
		}
	}

	return nil
}

func statusFunctionLastUpdateStatus(ctx context.Context, conn *lambda.Lambda, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFunctionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, aws.StringValue(output.Configuration.LastUpdateStatus), nil
	}
}

func statusFunctionState(ctx context.Context, conn *lambda.Lambda, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFunctionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, aws.StringValue(output.Configuration.State), nil
	}
}

func waitFunctionCreated(ctx context.Context, conn *lambda.Lambda, name string, timeout time.Duration) (*lambda.FunctionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.StatePending},
		Target:  []string{lambda.StateActive},
		Refresh: statusFunctionState(ctx, conn, name),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.StateReasonCode), aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitFunctionUpdated(ctx context.Context, conn *lambda.Lambda, functionName string, timeout time.Duration) (*lambda.FunctionConfiguration, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.LastUpdateStatusInProgress},
		Target:  []string{lambda.LastUpdateStatusSuccessful},
		Refresh: statusFunctionLastUpdateStatus(ctx, conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.LastUpdateStatusReasonCode), aws.StringValue(output.LastUpdateStatusReason)))

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
			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The role defined for the function cannot be assumed by Lambda") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The provided execution role does not have permissions") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "Lambda was unable to configure access to your environment variables because the KMS key is invalid for CreateGrant") {
				return true, err
			}

			if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceConflictException) {
				return true, err
			}

			return false, err
		})

	// Additional retries when throttled.
	if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
		output, err = tfresource.RetryWhen(ctx, functionExtraThrottlingTimeout,
			f,
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
					return true, err
				}

				return false, err
			})
	}

	return output, err
}

func checkHandlerRuntimeForZipFunction(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	packageType := d.Get("package_type")
	_, handlerOk := d.GetOk("handler")
	_, runtimeOk := d.GetOk("runtime")

	if packageType == lambda.PackageTypeZip && (!handlerOk || !runtimeOk) {
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

func flattenEnvironment(apiObject *lambda.EnvironmentResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Variables; v != nil {
		tfMap["variables"] = aws.StringValueMap(v)
	}

	return []interface{}{tfMap}
}

func flattenFileSystemConfigs(fscList []*lambda.FileSystemConfig) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(fscList))
	for _, fsc := range fscList {
		f := make(map[string]interface{})
		f["arn"] = aws.StringValue(fsc.Arn)
		f["local_mount_path"] = aws.StringValue(fsc.LocalMountPath)

		results = append(results, f)
	}
	return results
}

func expandFileSystemConfigs(fscMaps []interface{}) []*lambda.FileSystemConfig {
	fileSystemConfigs := make([]*lambda.FileSystemConfig, 0, len(fscMaps))
	for _, fsc := range fscMaps {
		fscMap := fsc.(map[string]interface{})
		fileSystemConfigs = append(fileSystemConfigs, &lambda.FileSystemConfig{
			Arn:            aws.String(fscMap["arn"].(string)),
			LocalMountPath: aws.String(fscMap["local_mount_path"].(string)),
		})
	}
	return fileSystemConfigs
}

func FlattenImageConfig(response *lambda.ImageConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if response == nil || response.Error != nil || response.ImageConfig == nil {
		return nil
	}

	settings["command"] = response.ImageConfig.Command
	settings["entry_point"] = response.ImageConfig.EntryPoint
	settings["working_directory"] = response.ImageConfig.WorkingDirectory

	return []map[string]interface{}{settings}
}

func expandImageConfigs(imageConfigMaps []interface{}) *lambda.ImageConfig {
	imageConfig := &lambda.ImageConfig{}
	// only one image_config block is allowed
	if len(imageConfigMaps) == 1 && imageConfigMaps[0] != nil {
		config := imageConfigMaps[0].(map[string]interface{})
		if len(config["entry_point"].([]interface{})) > 0 {
			imageConfig.EntryPoint = flex.ExpandStringList(config["entry_point"].([]interface{}))
		}
		if len(config["command"].([]interface{})) > 0 {
			imageConfig.Command = flex.ExpandStringList(config["command"].([]interface{}))
		}
		imageConfig.WorkingDirectory = aws.String(config["working_directory"].(string))
	}
	return imageConfig
}

func flattenEphemeralStorage(response *lambda.EphemeralStorage) []map[string]interface{} {
	if response == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["size"] = aws.Int64Value(response.Size)

	return []map[string]interface{}{m}
}

func expandSnapStart(tfList []interface{}) *lambda.SnapStart {
	snapStart := &lambda.SnapStart{ApplyOn: aws.String(lambda.SnapStartApplyOnNone)}
	if len(tfList) == 1 && tfList[0] != nil {
		item := tfList[0].(map[string]interface{})
		snapStart.ApplyOn = aws.String(item["apply_on"].(string))
	}
	return snapStart
}

func flattenSnapStart(apiObject *lambda.SnapStartResponse) []interface{} {
	if apiObject == nil || apiObject.ApplyOn == nil {
		return nil
	}
	if aws.StringValue(apiObject.ApplyOn) == lambda.SnapStartApplyOnNone {
		return nil
	}
	m := map[string]interface{}{
		"apply_on":            aws.StringValue(apiObject.ApplyOn),
		"optimization_status": aws.StringValue(apiObject.OptimizationStatus),
	}

	return []interface{}{m}
}
