package lambda

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	homedir "github.com/mitchellh/go-homedir"
)

const keyMutex = `aws_lambda_function`

const FunctionVersionLatest = "$LATEST"

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceFunctionCreate,
		Read:   resourceFunctionRead,
		Update: resourceFunctionUpdate,
		Delete: resourceFunctionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
	functionPutConcurrencyTimeout  = 1 * time.Minute
	functionExtraThrottlingTimeout = 9 * time.Minute
)

func resourceFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	functionName := d.Get("function_name").(string)
	reservedConcurrentExecutions := d.Get("reserved_concurrent_executions").(int)
	iamRole := d.Get("role").(string)

	log.Printf("[DEBUG] Creating Lambda Function %s with role %s", functionName, iamRole)

	filename, hasFilename := d.GetOk("filename")
	s3Bucket, bucketOk := d.GetOk("s3_bucket")
	s3Key, keyOk := d.GetOk("s3_key")
	s3ObjectVersion, versionOk := d.GetOk("s3_object_version")
	imageUri, hasImageUri := d.GetOk("image_uri")

	if !hasFilename && !bucketOk && !keyOk && !versionOk && !hasImageUri {
		return errors.New("filename, s3_* or image_uri attributes must be set")
	}

	var functionCode *lambda.FunctionCode
	if hasFilename {
		// Grab an exclusive lock so that we're only reading one function into
		// memory at a time.
		// See https://github.com/hashicorp/terraform/issues/9364
		conns.GlobalMutexKV.Lock(keyMutex)
		defer conns.GlobalMutexKV.Unlock(keyMutex)
		file, err := loadFileContent(filename.(string))
		if err != nil {
			return fmt.Errorf("unable to load %q: %w", filename.(string), err)
		}
		functionCode = &lambda.FunctionCode{
			ZipFile: file,
		}
	} else if hasImageUri {
		functionCode = &lambda.FunctionCode{
			ImageUri: aws.String(imageUri.(string)),
		}
	} else {
		if !bucketOk || !keyOk {
			return errors.New("s3_bucket and s3_key must all be set while using S3 code source")
		}
		functionCode = &lambda.FunctionCode{
			S3Bucket: aws.String(s3Bucket.(string)),
			S3Key:    aws.String(s3Key.(string)),
		}
		if versionOk {
			functionCode.S3ObjectVersion = aws.String(s3ObjectVersion.(string))
		}
	}

	packageType := d.Get("package_type").(string)
	params := &lambda.CreateFunctionInput{
		Code:         functionCode,
		Description:  aws.String(d.Get("description").(string)),
		FunctionName: aws.String(functionName),
		MemorySize:   aws.Int64(int64(d.Get("memory_size").(int))),
		Role:         aws.String(iamRole),
		Timeout:      aws.Int64(int64(d.Get("timeout").(int))),
		Publish:      aws.Bool(d.Get("publish").(bool)),
		PackageType:  aws.String(packageType),
	}

	if v, ok := d.GetOk("architectures"); ok && len(v.([]interface{})) > 0 {
		params.Architectures = flex.ExpandStringList(v.([]interface{}))
	}

	if packageType == lambda.PackageTypeZip {
		params.Handler = aws.String(d.Get("handler").(string))
		params.Runtime = aws.String(d.Get("runtime").(string))
	}

	if v, ok := d.GetOk("code_signing_config_arn"); ok {
		params.CodeSigningConfigArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layers"); ok && len(v.([]interface{})) > 0 {
		params.Layers = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("dead_letter_config"); ok {
		dlcMaps := v.([]interface{})
		if len(dlcMaps) == 1 { // Schema guarantees either 0 or 1
			// Prevent panic on nil dead_letter_config. See GH-14961
			if dlcMaps[0] == nil {
				return fmt.Errorf("nil dead_letter_config supplied for function: %s", functionName)
			}
			dlcMap := dlcMaps[0].(map[string]interface{})
			params.DeadLetterConfig = &lambda.DeadLetterConfig{
				TargetArn: aws.String(dlcMap["target_arn"].(string)),
			}
		}
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 {
		ephemeralStorage := v.([]interface{})
		configMap := ephemeralStorage[0].(map[string]interface{})
		params.EphemeralStorage = &lambda.EphemeralStorage{
			Size: aws.Int64(int64(configMap["size"].(int))),
		}
	}

	if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
		params.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
		params.ImageConfig = expandImageConfigs(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})

		params.VpcConfig = &lambda.VpcConfig{
			SecurityGroupIds: flex.ExpandStringSet(config["security_group_ids"].(*schema.Set)),
			SubnetIds:        flex.ExpandStringSet(config["subnet_ids"].(*schema.Set)),
		}
	}

	if v, ok := d.GetOk("snap_start"); ok {
		params.SnapStart = expandSnapStart(v.([]interface{}))
	}

	if v, ok := d.GetOk("tracing_config"); ok {
		tracingConfig := v.([]interface{})
		tracing := tracingConfig[0].(map[string]interface{})
		params.TracingConfig = &lambda.TracingConfig{
			Mode: aws.String(tracing["mode"].(string)),
		}
	}

	if v, ok := d.GetOk("environment"); ok {
		environments := v.([]interface{})
		environment, ok := environments[0].(map[string]interface{})
		if !ok {
			return errors.New("At least one field is expected inside environment")
		}

		if environmentVariables, ok := environment["variables"]; ok {
			variables := flex.ExpandStringValueMap(environmentVariables.(map[string]interface{}))

			params.Environment = &lambda.Environment{
				Variables: aws.StringMap(variables),
			}
		}
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		params.KMSKeyArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	err := resource.Retry(propagationTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
		_, err := conn.CreateFunction(params)

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The role defined for the function cannot be assumed by Lambda") {
			log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The provided execution role does not have permissions") {
			log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
			log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
			return resource.RetryableError(err)
		}

		if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "Lambda was unable to configure access to your environment variables because the KMS key is invalid for CreateGrant") {
			log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
			return resource.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceConflictException) {
			log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateFunction(params)
	}

	if err != nil {
		if !tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
			return fmt.Errorf("error creating Lambda Function (1): %w", err)
		}

		err := resource.Retry(functionExtraThrottlingTimeout, func() *resource.RetryError {
			_, err := conn.CreateFunction(params)

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
				log.Printf("[DEBUG] Received %s, retrying CreateFunction", err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.CreateFunction(params)
		}

		if err != nil {
			return fmt.Errorf("error creating Lambda Function (2): %w", err)
		}
	}

	d.SetId(d.Get("function_name").(string))

	if _, err := waitFunctionCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for Lambda Function (%s) create: %w", d.Id(), err)
	}

	if reservedConcurrentExecutions >= 0 {
		log.Printf("[DEBUG] Setting Concurrency to %d for the Lambda Function %s", reservedConcurrentExecutions, functionName)

		concurrencyParams := &lambda.PutFunctionConcurrencyInput{
			FunctionName:                 aws.String(functionName),
			ReservedConcurrentExecutions: aws.Int64(int64(reservedConcurrentExecutions)),
		}

		err := resource.Retry(functionPutConcurrencyTimeout, func() *resource.RetryError {
			_, err := conn.PutFunctionConcurrency(concurrencyParams)

			if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.PutFunctionConcurrency(concurrencyParams)
		}

		if err != nil {
			return fmt.Errorf("error setting Lambda Function (%s) concurrency: %w", functionName, err)
		}
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionRead(d *schema.ResourceData, meta interface{}) error {
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

	output, err := findFunction(conn, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Lambda Function %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Lambda Function (%s): %w", d.Id(), err)
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
			return fmt.Errorf("setting dead_letter_config: %w", err)
		}
	} else {
		d.Set("dead_letter_config", []interface{}{})
	}
	d.Set("description", function.Description)
	if err := d.Set("environment", flattenEnvironment(function.Environment)); err != nil {
		return fmt.Errorf("setting environment: %w", err)
	}
	if err := d.Set("ephemeral_storage", flattenEphemeralStorage(function.EphemeralStorage)); err != nil {
		return fmt.Errorf("setting ephemeral_storage: %w", err)
	}
	if err := d.Set("file_system_config", flattenFileSystemConfigs(function.FileSystemConfigs)); err != nil {
		return fmt.Errorf("setting file_system_config: %w", err)
	}
	d.Set("handler", function.Handler)
	if err := d.Set("image_config", FlattenImageConfig(function.ImageConfigResponse)); err != nil {
		return fmt.Errorf("setting image_config: %w", err)
	}
	if output.Code != nil {
		d.Set("image_uri", output.Code.ImageUri)
	}
	d.Set("invoke_arn", functionInvokeARN(functionARN, meta))
	d.Set("kms_key_arn", function.KMSKeyArn)
	d.Set("last_modified", function.LastModified)
	if err := d.Set("layers", flattenLayers(function.Layers)); err != nil {
		return fmt.Errorf("setting layers: %w", err)
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
		return fmt.Errorf("setting snap_start: %w", err)
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
		return fmt.Errorf("setting tracing_config: %s", err)
	}
	if err := d.Set("vpc_config", flattenVPCConfigResponse(function.VpcConfig)); err != nil {
		return fmt.Errorf("setting vpc_config: %w", err)
	}

	if hasQualifier {
		d.Set("qualified_arn", functionARN)
		d.Set("qualified_invoke_arn", functionInvokeARN(functionARN, meta))
		d.Set("version", function.Version)
	} else {
		latest, err := findLatestFunctionVersionByName(conn, d.Id())

		if err != nil {
			return fmt.Errorf("reading Lambda Function (%s) latest version: %w", d.Id(), err)
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
			return fmt.Errorf("setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("setting tags_all: %w", err)
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
			output, err := conn.GetFunctionCodeSigningConfig(&lambda.GetFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			})

			if err != nil {
				return fmt.Errorf("reading Lambda Function (%s) code signing config: %w", d.Id(), err)
			}

			if output != nil {
				codeSigningConfigArn = aws.StringValue(output.CodeSigningConfigArn)
			}
		}

		d.Set("code_signing_config_arn", codeSigningConfigArn)
	}

	return nil
}

func resourceFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn()

	// If Code Signing Config is updated, calls PutFunctionCodeSigningConfig
	// If removed, calls DeleteFunctionCodeSigningConfig
	if d.HasChange("code_signing_config_arn") {
		if v, ok := d.GetOk("code_signing_config_arn"); ok {
			configUpdateInput := &lambda.PutFunctionCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(v.(string)),
				FunctionName:         aws.String(d.Id()),
			}

			_, err := conn.PutFunctionCodeSigningConfig(configUpdateInput)

			if err != nil {
				return fmt.Errorf("error updating code signing config arn (Function: %s): %w", d.Id(), err)
			}
		} else {
			configDeleteInput := &lambda.DeleteFunctionCodeSigningConfigInput{
				FunctionName: aws.String(d.Id()),
			}

			_, err := conn.DeleteFunctionCodeSigningConfig(configDeleteInput)

			if err != nil {
				return fmt.Errorf("error deleting code signing config arn (Function: %s): %w", d.Id(), err)
			}
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Lambda Function (%s) tags: %w", arn, err)
		}
	}

	configReq := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		configReq.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("ephemeral_storage") {
		ephemeralStorage := d.Get("ephemeral_storage").([]interface{})
		if len(ephemeralStorage) == 1 {
			configMap := ephemeralStorage[0].(map[string]interface{})
			configReq.EphemeralStorage = &lambda.EphemeralStorage{
				Size: aws.Int64(int64(configMap["size"].(int))),
			}
		}
	}
	if d.HasChange("handler") {
		configReq.Handler = aws.String(d.Get("handler").(string))
	}
	if d.HasChange("file_system_config") {
		configReq.FileSystemConfigs = make([]*lambda.FileSystemConfig, 0)
		if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
			configReq.FileSystemConfigs = expandFileSystemConfigs(v.([]interface{}))
		}
	}
	if d.HasChange("image_config") {
		configReq.ImageConfig = &lambda.ImageConfig{}
		if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
			configReq.ImageConfig = expandImageConfigs(v.([]interface{}))
		}
	}
	if d.HasChange("memory_size") {
		configReq.MemorySize = aws.Int64(int64(d.Get("memory_size").(int)))
	}
	if d.HasChange("role") {
		configReq.Role = aws.String(d.Get("role").(string))
	}
	if d.HasChange("timeout") {
		configReq.Timeout = aws.Int64(int64(d.Get("timeout").(int)))
	}
	if d.HasChange("kms_key_arn") {
		configReq.KMSKeyArn = aws.String(d.Get("kms_key_arn").(string))
	}
	if d.HasChange("layers") {
		layers := d.Get("layers").([]interface{})
		configReq.Layers = flex.ExpandStringList(layers)
	}
	if d.HasChange("dead_letter_config") {
		dlcMaps := d.Get("dead_letter_config").([]interface{})
		configReq.DeadLetterConfig = &lambda.DeadLetterConfig{
			TargetArn: aws.String(""),
		}
		if len(dlcMaps) == 1 { // Schema guarantees either 0 or 1
			dlcMap := dlcMaps[0].(map[string]interface{})
			configReq.DeadLetterConfig.TargetArn = aws.String(dlcMap["target_arn"].(string))
		}
	}
	if d.HasChange("snap_start") {
		snapStart := d.Get("snap_start").([]interface{})
		configReq.SnapStart = expandSnapStart(snapStart)
	}
	if d.HasChange("tracing_config") {
		tracingConfig := d.Get("tracing_config").([]interface{})
		if len(tracingConfig) == 1 { // Schema guarantees either 0 or 1
			config := tracingConfig[0].(map[string]interface{})
			configReq.TracingConfig = &lambda.TracingConfig{
				Mode: aws.String(config["mode"].(string)),
			}
		}
	}
	if d.HasChanges("vpc_config.0.security_group_ids", "vpc_config.0.subnet_ids") {
		configReq.VpcConfig = &lambda.VpcConfig{
			SecurityGroupIds: []*string{},
			SubnetIds:        []*string{},
		}
		if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 {
			vpcConfig := v.([]interface{})[0].(map[string]interface{})
			configReq.VpcConfig.SecurityGroupIds = flex.ExpandStringSet(vpcConfig["security_group_ids"].(*schema.Set))
			configReq.VpcConfig.SubnetIds = flex.ExpandStringSet(vpcConfig["subnet_ids"].(*schema.Set))
		}
	}

	if d.HasChange("runtime") {
		configReq.Runtime = aws.String(d.Get("runtime").(string))
	}

	if d.HasChanges("environment", "kms_key_arn") {
		if v, ok := d.GetOk("environment"); ok {
			environments := v.([]interface{})
			environment, ok := environments[0].(map[string]interface{})
			if !ok {
				return errors.New("At least one field is expected inside environment")
			}

			if environmentVariables, ok := environment["variables"]; ok {
				variables := flex.ExpandStringValueMap(environmentVariables.(map[string]interface{}))

				configReq.Environment = &lambda.Environment{
					Variables: aws.StringMap(variables),
				}
			}
		} else {
			configReq.Environment = &lambda.Environment{
				Variables: aws.StringMap(map[string]string{}),
			}
		}
	}
	configUpdate := hasConfigChanges(d)
	if configUpdate {
		log.Printf("[DEBUG] Send Update Lambda Function Configuration request: %#v", configReq)

		err := resource.Retry(propagationTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
			_, err := conn.UpdateFunctionConfiguration(configReq)

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The role defined for the function cannot be assumed by Lambda") {
				log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "The provided execution role does not have permissions") {
				log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
				log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "Lambda was unable to configure access to your environment variables because the KMS key is invalid for CreateGrant") {
				log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
				return resource.RetryableError(err)
			}

			if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceConflictException) {
				log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateFunctionConfiguration(configReq)
		}

		if err != nil {
			if !tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
				return fmt.Errorf("error modifying Lambda Function (%s) configuration : %w", d.Id(), err)
			}

			// Allow more time for EC2 throttling
			err := resource.Retry(functionExtraThrottlingTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
				_, err = conn.UpdateFunctionConfiguration(configReq)

				if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
					log.Printf("[DEBUG] Received %s, retrying UpdateFunctionConfiguration", err)
					return resource.RetryableError(err)
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.UpdateFunctionConfiguration(configReq)
			}

			if err != nil {
				return fmt.Errorf("error modifying Lambda Function Configuration %s: %w", d.Id(), err)
			}
		}

		if _, err := waitFunctionUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for Lambda Function (%s) configuration update: %w", d.Id(), err)
		}
	}

	codeUpdate := needsFunctionCodeUpdate(d)
	if codeUpdate {
		codeReq := &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(d.Id()),
		}

		if d.HasChange("architectures") {
			architectures := d.Get("architectures").([]interface{})
			if len(architectures) > 0 {
				codeReq.Architectures = flex.ExpandStringList(architectures)
			}
		}

		if v, ok := d.GetOk("filename"); ok {
			// Grab an exclusive lock so that we're only reading one function into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			conns.GlobalMutexKV.Lock(keyMutex)
			defer conns.GlobalMutexKV.Unlock(keyMutex)
			file, err := loadFileContent(v.(string))
			if err != nil {
				return fmt.Errorf("unable to load %q: %w", v.(string), err)
			}
			codeReq.ZipFile = file
		} else if v, ok := d.GetOk("image_uri"); ok {
			codeReq.ImageUri = aws.String(v.(string))
		} else {
			s3Bucket, _ := d.GetOk("s3_bucket")
			s3Key, _ := d.GetOk("s3_key")
			s3ObjectVersion, versionOk := d.GetOk("s3_object_version")

			codeReq.S3Bucket = aws.String(s3Bucket.(string))
			codeReq.S3Key = aws.String(s3Key.(string))
			if versionOk {
				codeReq.S3ObjectVersion = aws.String(s3ObjectVersion.(string))
			}
		}

		log.Printf("[DEBUG] Send Update Lambda Function Code request: %#v", codeReq)

		_, err := conn.UpdateFunctionCode(codeReq)
		if err != nil {
			return fmt.Errorf("error modifying Lambda Function (%s) Code: %w", d.Id(), err)
		}

		if _, err := waitFunctionUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for Lambda Function (%s) code update: %w", d.Id(), err)
		}
	}

	if d.HasChange("reserved_concurrent_executions") {
		nc := d.Get("reserved_concurrent_executions")

		if nc.(int) >= 0 {
			log.Printf("[DEBUG] Updating Concurrency to %d for the Lambda Function %s", nc.(int), d.Id())

			concurrencyParams := &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String(d.Id()),
				ReservedConcurrentExecutions: aws.Int64(int64(d.Get("reserved_concurrent_executions").(int))),
			}

			_, err := conn.PutFunctionConcurrency(concurrencyParams)
			if err != nil {
				return fmt.Errorf("error setting Lambda Function (%s) concurrency: %w", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Removing Concurrency for the Lambda Function %s", d.Id())

			deleteConcurrencyParams := &lambda.DeleteFunctionConcurrencyInput{
				FunctionName: aws.String(d.Id()),
			}
			_, err := conn.DeleteFunctionConcurrency(deleteConcurrencyParams)
			if err != nil {
				return fmt.Errorf("error setting Lambda Function (%s) concurrency: %w", d.Id(), err)
			}
		}
	}

	publish := d.Get("publish").(bool)
	if publish && (codeUpdate || configUpdate || d.HasChange("publish")) {
		versionReq := &lambda.PublishVersionInput{
			FunctionName: aws.String(d.Id()),
		}

		var output *lambda.FunctionConfiguration
		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			var err error
			output, err = conn.PublishVersion(versionReq)

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceConflictException, "in progress") {
				log.Printf("[DEBUG] Retrying publish of Lambda function (%s) version after error: %s", d.Id(), err)
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			output, err = conn.PublishVersion(versionReq)
		}

		if err != nil {
			return fmt.Errorf("error publishing Lambda Function (%s) version: %w", d.Id(), err)
		}

		err = conn.WaitUntilFunctionUpdated(&lambda.GetFunctionConfigurationInput{
			FunctionName: output.FunctionArn,
			Qualifier:    output.Version,
		})

		if err != nil {
			return fmt.Errorf("while waiting for function (%s) update: %w", d.Id(), err)
		}
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn()

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())
	_, err := conn.DeleteFunction(&lambda.DeleteFunctionInput{
		FunctionName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Lambda Function (%s): %w", d.Id(), err)
	}

	return nil
}

func FindFunctionByName(conn *lambda.Lambda, name string) (*lambda.GetFunctionOutput, error) {
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	return findFunction(conn, input)
}

func findFunction(conn *lambda.Lambda, input *lambda.GetFunctionInput) (*lambda.GetFunctionOutput, error) {
	output, err := conn.GetFunction(input)

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

func findLatestFunctionVersionByName(conn *lambda.Lambda, name string) (*lambda.FunctionConfiguration, error) {
	input := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(name),
		MaxItems:     aws.Int64(10000),
	}
	var output *lambda.FunctionConfiguration

	err := conn.ListVersionsByFunctionPages(input, func(page *lambda.ListVersionsByFunctionOutput, lastPage bool) bool {
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
	configChanged := hasConfigChanges(d)
	functionCodeUpdated := needsFunctionCodeUpdate(d)
	if functionCodeUpdated {
		d.SetNewComputed("last_modified")
	}

	publish := d.Get("publish").(bool)
	publishChanged := d.HasChange("publish")
	if publish && (configChanged || functionCodeUpdated || publishChanged) {
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

func hasConfigChanges(d verify.ResourceDiffer) bool {
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

// loadFileContent returns contents of a file in a given path
func loadFileContent(v string) ([]byte, error) {
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

func refreshFunctionLastUpdateStatus(conn *lambda.Lambda, functionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &lambda.GetFunctionInput{
			FunctionName: aws.String(functionName),
		}

		output, err := conn.GetFunction(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Configuration == nil {
			return nil, "", nil
		}

		lastUpdateStatus := aws.StringValue(output.Configuration.LastUpdateStatus)

		if lastUpdateStatus == lambda.LastUpdateStatusFailed {
			return output.Configuration, lastUpdateStatus, fmt.Errorf("%s: %s", aws.StringValue(output.Configuration.LastUpdateStatusReasonCode), aws.StringValue(output.Configuration.LastUpdateStatusReason))
		}

		return output.Configuration, lastUpdateStatus, nil
	}
}

func statusFunctionState(conn *lambda.Lambda, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFunctionByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Configuration, aws.StringValue(output.Configuration.State), nil
	}
}

func waitFunctionCreated(conn *lambda.Lambda, name string, timeout time.Duration) (*lambda.FunctionConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.StatePending},
		Target:  []string{lambda.StateActive},
		Refresh: statusFunctionState(conn, name),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.StateReasonCode), aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
}

func waitFunctionUpdated(conn *lambda.Lambda, functionName string, timeout time.Duration) (*lambda.FunctionConfiguration, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.LastUpdateStatusInProgress},
		Target:  []string{lambda.LastUpdateStatusSuccessful},
		Refresh: refreshFunctionLastUpdateStatus(conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.FunctionConfiguration); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.StateReasonCode), aws.StringValue(output.StateReason)))

		return output, err
	}

	return nil, err
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
