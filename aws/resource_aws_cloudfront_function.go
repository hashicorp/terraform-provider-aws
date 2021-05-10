package aws

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"

	// "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lambda/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// const awsMutexLambdaKey = `aws_cloudfront_function`

// const CloudfrontFunctionVersionLatest = "$LATEST"

func resourceAwsCloudfrontFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudfrontFunctionCreate,
		Read:   resourceAwsCloudfrontFunctionRead,
		Update: resourceAwsCloudfrontFunctionUpdate,
		Delete: resourceAwsCloudfrontFunctionDelete,
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
			"filename": {
				Type:          schema.TypeString,
				Required:      true,
				ConflictsWith: []string{"s3_bucket", "s3_key", "s3_object_version", "image_uri"},
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"runtime": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.FunctionRuntime_Values(), false),
			},
			"publish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		// CustomizeDiff: customdiff.Sequence(
		// 	checkHandlerRuntimeForZipFunction,
		// 	updateComputedAttributesOnPublish,
		// 	SetTagsDiff,
		// ),
	}
}

/*
func checkHandlerRuntimeForZipFunction(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	packageType := d.Get("package_type")
	_, handlerOk := d.GetOk("handler")
	_, runtimeOk := d.GetOk("runtime")

	if packageType == lambda.PackageTypeZip && !handlerOk && !runtimeOk {
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
	}
	return nil
}

func hasConfigChanges(d resourceDiffer) bool {
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
		d.HasChange("tracing_config") ||
		d.HasChange("vpc_config") ||
		d.HasChange("runtime") ||
		d.HasChange("environment")
}
*/
// resourceAwsCloudfrontFunction maps to:
// CreateFunction in the API / SDK
func resourceAwsCloudfrontFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	functionName := d.Get("function_name").(string)

	log.Printf("[DEBUG] Creating Lambda Function %s", functionName)

	filename, hasFilename := d.GetOk("filename")

	// Grab an exclusive lock so that we're only reading one function into
	// memory at a time.
	// See https://github.com/hashicorp/terraform/issues/9364
	awsMutexKV.Lock(awsMutexLambdaKey)
	defer awsMutexKV.Unlock(awsMutexLambdaKey)
	file, err := loadFileContent(filename.(string))
	if err != nil {
		return fmt.Errorf("unable to load %q: %w", filename.(string), err)
	}
	runtime, _ := d.GetOk("runtime")

	params := &cloudfront.CreateFunctionInput{
		Code:         file,
		Comment:      aws.String(d.Get("comment").(string)),
		FunctionName: aws.String(functionName),
		//		Publish:      aws.Bool(d.Get("publish").(bool)),
	}

	if packageType == lambda.PackageTypeZip {
		params.Handler = aws.String(handler.(string))
		params.Runtime = aws.String(runtime.(string))
	}

	if tfresource.TimedOut(err) {
		_, err = conn.CreateFunction(params)
	}

	if err != nil {
		if !tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "throttled by EC2") {
			return fmt.Errorf("error creating Lambda Function (1): %w", err)
		}

		err := resource.Retry(waiter.CloudfrontFunctionExtraThrottlingTimeout, func() *resource.RetryError {
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

	if err := waitForCloudfrontFunctionCreation(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Lambda Function (%s) creation: %w", d.Id(), err)
	}

	if reservedConcurrentExecutions >= 0 {

		log.Printf("[DEBUG] Setting Concurrency to %d for the Lambda Function %s", reservedConcurrentExecutions, functionName)

		concurrencyParams := &lambda.PutFunctionConcurrencyInput{
			FunctionName:                 aws.String(functionName),
			ReservedConcurrentExecutions: aws.Int64(int64(reservedConcurrentExecutions)),
		}

		err := resource.Retry(waiter.CloudfrontFunctionPutConcurrencyTimeout, func() *resource.RetryError {
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

	return resourceAwsCloudfrontFunctionRead(d, meta)
}

// resourceAwsCloudfrontFunctionRead maps to:
// GetFunction in the API / SDK
func resourceAwsCloudfrontFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &lambda.GetFunctionInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	// qualifier for lambda function data source
	qualifier, qualifierExistance := d.GetOk("qualifier")
	if qualifierExistance {
		params.Qualifier = aws.String(qualifier.(string))
		log.Printf("[DEBUG] Fetching Lambda Function: %s:%s", d.Id(), qualifier.(string))
	} else {
		log.Printf("[DEBUG] Fetching Lambda Function: %s", d.Id())
	}

	getFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == lambda.ErrCodeResourceNotFoundException && !d.IsNewResource() {
			d.SetId("")
			return nil
		}
		return err
	}

	if getFunctionOutput.Concurrency != nil {
		d.Set("reserved_concurrent_executions", getFunctionOutput.Concurrency.ReservedConcurrentExecutions)
	} else {
		d.Set("reserved_concurrent_executions", -1)
	}

	// Tagging operations are permitted on Lambda functions only.
	// Tags on aliases and versions are not supported.
	if !qualifierExistance {
		tags := keyvaluetags.LambdaKeyValueTags(getFunctionOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	}

	// getFunctionOutput.Code.Location is a pre-signed URL pointing at the zip
	// file that we uploaded when we created the resource. You can use it to
	// download the code from AWS. The other part is
	// getFunctionOutput.Configuration which holds metadata.

	function := getFunctionOutput.Configuration

	if err := d.Set("arn", function.FunctionArn); err != nil {
		return fmt.Errorf("error setting function arn for Lambda Function: %w", err)
	}

	if err := d.Set("description", function.Description); err != nil {
		return fmt.Errorf("error setting function description for Lambda Function: %w", err)
	}

	if err := d.Set("handler", function.Handler); err != nil {
		return fmt.Errorf("error setting handler for Lambda Function: %w", err)
	}

	if err := d.Set("memory_size", function.MemorySize); err != nil {
		return fmt.Errorf("error setting memory size for Lambda Function: %w", err)
	}

	if err := d.Set("last_modified", function.LastModified); err != nil {
		return fmt.Errorf("error setting last modified time for Lambda Function: %w", err)
	}

	if err := d.Set("role", function.Role); err != nil {
		return fmt.Errorf("error setting role for Lambda Function: %w", err)
	}

	if err := d.Set("runtime", function.Runtime); err != nil {
		return fmt.Errorf("error setting runtime for Lambda Function: %w", err)
	}

	if err := d.Set("timeout", function.Timeout); err != nil {
		return fmt.Errorf("error setting timeout for Lambda Function: %w", err)
	}

	if err := d.Set("kms_key_arn", function.KMSKeyArn); err != nil {
		return fmt.Errorf("error setting KMS key arn for Lambda Function: %w", err)
	}

	if err := d.Set("source_code_hash", function.CodeSha256); err != nil {
		return fmt.Errorf("error setting CodeSha256 for Lambda Function: %w", err)
	}

	if err := d.Set("source_code_size", function.CodeSize); err != nil {
		return fmt.Errorf("error setting code size for Lambda Function: %w", err)
	}

	// Add Signing Profile Version ARN
	if err := d.Set("signing_profile_version_arn", function.SigningProfileVersionArn); err != nil {
		return fmt.Errorf("error setting signing profile version arn for Lambda Function: %w", err)
	}

	// Add Signing Job ARN
	if err := d.Set("signing_job_arn", function.SigningJobArn); err != nil {
		return fmt.Errorf("error setting signing job arn for Lambda Function: %w", err)
	}

	fileSystemConfigs := flattenLambdaFileSystemConfigs(function.FileSystemConfigs)
	log.Printf("[INFO] Setting Lambda %s file system configs %#v from API", d.Id(), fileSystemConfigs)
	if err := d.Set("file_system_config", fileSystemConfigs); err != nil {
		return fmt.Errorf("error setting file system config for Lambda Function (%s): %w", d.Id(), err)
	}

	// Add Package Type
	log.Printf("[INFO] Setting Lambda %s package type %#v from API", d.Id(), function.PackageType)
	if err := d.Set("package_type", function.PackageType); err != nil {
		return fmt.Errorf("error setting package type for Lambda Function: %w", err)
	}

	// Add Image Configuration
	imageConfig := flattenLambdaImageConfig(function.ImageConfigResponse)
	log.Printf("[INFO] Setting Lambda %s Image config %#v from API", d.Id(), imageConfig)
	if err := d.Set("image_config", imageConfig); err != nil {
		return fmt.Errorf("error setting image config for Lambda Function: %w", err)
	}

	if err := d.Set("image_uri", getFunctionOutput.Code.ImageUri); err != nil {
		return fmt.Errorf("error setting image uri for Lambda Function: %w", err)
	}

	layers := flattenLambdaLayers(function.Layers)
	log.Printf("[INFO] Setting Lambda %s Layers %#v from API", d.Id(), layers)
	if err := d.Set("layers", layers); err != nil {
		return fmt.Errorf("error setting layers for Lambda Function (%s): %w", d.Id(), err)
	}

	config := flattenLambdaVpcConfigResponse(function.VpcConfig)
	log.Printf("[INFO] Setting Lambda %s VPC config %#v from API", d.Id(), config)
	if err := d.Set("vpc_config", config); err != nil {
		return fmt.Errorf("error setting vpc_config for Lambda Function (%s): %w", d.Id(), err)
	}

	environment := flattenLambdaEnvironment(function.Environment)
	log.Printf("[INFO] Setting Lambda %s environment %#v from API", d.Id(), environment)
	if err := d.Set("environment", environment); err != nil {
		log.Printf("[ERR] Error setting environment for Lambda Function (%s): %s", d.Id(), err)
	}

	if function.DeadLetterConfig != nil && function.DeadLetterConfig.TargetArn != nil {
		d.Set("dead_letter_config", []interface{}{
			map[string]interface{}{
				"target_arn": *function.DeadLetterConfig.TargetArn,
			},
		})
	} else {
		d.Set("dead_letter_config", []interface{}{})
	}

	// Assume `PassThrough` on partitions that don't support tracing config
	tracingConfigMode := "PassThrough"
	if function.TracingConfig != nil {
		tracingConfigMode = *function.TracingConfig.Mode
	}
	d.Set("tracing_config", []interface{}{
		map[string]interface{}{
			"mode": tracingConfigMode,
		},
	})

	// Get latest version and ARN unless qualifier is specified via data source
	if qualifierExistance {
		d.Set("version", function.Version)
		d.Set("qualified_arn", function.FunctionArn)
	} else {

		// List is sorted from oldest to latest
		// so this may get costly over time :'(
		var lastVersion, lastQualifiedArn string
		err = listVersionsByFunctionPages(conn, &lambda.ListVersionsByFunctionInput{
			FunctionName: function.FunctionName,
			MaxItems:     aws.Int64(10000),
		}, func(p *lambda.ListVersionsByFunctionOutput, lastPage bool) bool {
			if lastPage {
				last := p.Versions[len(p.Versions)-1]
				lastVersion = *last.Version
				lastQualifiedArn = *last.FunctionArn
				return false
			}
			return true
		})
		if err != nil {
			return err
		}

		d.Set("version", lastVersion)
		d.Set("qualified_arn", lastQualifiedArn)
	}

	invokeArn := lambdaFunctionInvokeArn(*function.FunctionArn, meta)
	d.Set("invoke_arn", invokeArn)

	// Currently, this functionality is only enabled in AWS Commercial partition
	// and other partitions return ambiguous error codes (e.g. AccessDeniedException
	// in AWS GovCloud (US)) so we cannot just ignore the error as would typically.
	if meta.(*AWSClient).partition != endpoints.AwsPartitionID {
		return nil
	}

	// Code Signing is only supported on zip packaged lambda functions.
	var codeSigningConfigArn string

	if aws.StringValue(function.PackageType) == lambda.PackageTypeZip {
		codeSigningConfigInput := &lambda.GetFunctionCodeSigningConfigInput{
			FunctionName: aws.String(d.Id()),
		}
		getCodeSigningConfigOutput, err := conn.GetFunctionCodeSigningConfig(codeSigningConfigInput)
		if err != nil {
			return fmt.Errorf("error getting Lambda Function (%s) code signing config %w", d.Id(), err)
		}

		if getCodeSigningConfigOutput != nil {
			codeSigningConfigArn = aws.StringValue(getCodeSigningConfigOutput.CodeSigningConfigArn)
		}
	}

	d.Set("code_signing_config_arn", codeSigningConfigArn)

	return nil
}

func listVersionsByFunctionPages(c *lambda.Lambda, input *lambda.ListVersionsByFunctionInput,
	fn func(p *lambda.ListVersionsByFunctionOutput, lastPage bool) bool) error {
	for {
		page, err := c.ListVersionsByFunction(input)
		if err != nil {
			return err
		}
		lastPage := page.NextMarker == nil

		shouldContinue := fn(page, lastPage)
		if !shouldContinue || lastPage {
			break
		}
		input.Marker = page.NextMarker
	}
	return nil
}

// resourceAwsCloudfrontFunction maps to:
// DeleteFunction in the API / SDK
func resourceAwsCloudfrontFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())

	params := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	_, err := conn.DeleteFunction(params)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Lambda Function (%s): %w", d.Id(), err)
	}

	return nil
}

func needsFunctionCodeUpdate(d resourceDiffer) bool {
	return d.HasChange("filename") ||
		d.HasChange("source_code_hash") ||
		d.HasChange("s3_bucket") ||
		d.HasChange("s3_key") ||
		d.HasChange("s3_object_version") ||
		d.HasChange("image_uri")

}

// resourceAwsCloudfrontFunctionUpdate maps to:
// UpdateFunctionCode in the API / SDK
func resourceAwsCloudfrontFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

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

		if err := keyvaluetags.LambdaUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Lambda Function (%s) tags: %w", arn, err)
		}
	}

	configReq := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		configReq.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("handler") {
		configReq.Handler = aws.String(d.Get("handler").(string))
	}
	if d.HasChange("file_system_config") {
		configReq.FileSystemConfigs = make([]*lambda.FileSystemConfig, 0)
		if v, ok := d.GetOk("file_system_config"); ok && len(v.([]interface{})) > 0 {
			configReq.FileSystemConfigs = expandLambdaFileSystemConfigs(v.([]interface{}))
		}
	}
	if d.HasChange("image_config") {
		configReq.ImageConfig = &lambda.ImageConfig{}
		if v, ok := d.GetOk("image_config"); ok && len(v.([]interface{})) > 0 {
			configReq.ImageConfig = expandLambdaImageConfigs(v.([]interface{}))
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
		configReq.Layers = expandStringList(layers)
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
	if d.HasChange("tracing_config") {
		tracingConfig := d.Get("tracing_config").([]interface{})
		if len(tracingConfig) == 1 { // Schema guarantees either 0 or 1
			config := tracingConfig[0].(map[string]interface{})
			configReq.TracingConfig = &lambda.TracingConfig{
				Mode: aws.String(config["mode"].(string)),
			}
		}
	}
	if d.HasChange("vpc_config") {
		configReq.VpcConfig = &lambda.VpcConfig{
			SecurityGroupIds: []*string{},
			SubnetIds:        []*string{},
		}
		if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 {
			vpcConfig := v.([]interface{})[0].(map[string]interface{})
			configReq.VpcConfig.SecurityGroupIds = expandStringSet(vpcConfig["security_group_ids"].(*schema.Set))
			configReq.VpcConfig.SubnetIds = expandStringSet(vpcConfig["subnet_ids"].(*schema.Set))
		}
	}

	if d.HasChange("runtime") {
		configReq.Runtime = aws.String(d.Get("runtime").(string))
	}
	if d.HasChange("environment") {
		if v, ok := d.GetOk("environment"); ok {
			environments := v.([]interface{})
			environment, ok := environments[0].(map[string]interface{})
			if !ok {
				return errors.New("At least one field is expected inside environment")
			}

			if environmentVariables, ok := environment["variables"]; ok {
				variables := readEnvironmentVariables(environmentVariables.(map[string]interface{}))

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

		err := resource.Retry(waiter.CloudfrontFunctionUpdateTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
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
			err := resource.Retry(waiter.CloudfrontFunctionExtraThrottlingTimeout, func() *resource.RetryError { // nosem: helper-schema-resource-Retry-without-TimeoutError-check
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

		if err := waitForCloudfrontFunctionUpdate(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Lambda Function (%s) update: %w", d.Id(), err)
		}
	}

	codeUpdate := needsFunctionCodeUpdate(d)
	if codeUpdate {
		codeReq := &lambda.UpdateFunctionCodeInput{
			FunctionName: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("filename"); ok {
			// Grab an exclusive lock so that we're only reading one function into
			// memory at a time.
			// See https://github.com/hashicorp/terraform/issues/9364
			awsMutexKV.Lock(awsMutexLambdaKey)
			defer awsMutexKV.Unlock(awsMutexLambdaKey)
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
		err := resource.Retry(waiter.CloudfrontFunctionPublishTimeout, func() *resource.RetryError {
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

	return resourceAwsCloudfrontFunctionRead(d, meta)
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

func readEnvironmentVariables(ev map[string]interface{}) map[string]string {
	variables := make(map[string]string)
	for k, v := range ev {
		variables[k] = v.(string)
	}

	return variables
}

func lambdaFunctionInvokeArn(functionArn string, meta interface{}) string {
	return arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		AccountID: "lambda",
		Resource:  fmt.Sprintf("path/2015-03-31/functions/%s/invocations", functionArn),
	}.String()
}

func refreshCloudfrontFunctionLastUpdateStatus(conn *lambda.Lambda, functionName string) resource.StateRefreshFunc {
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

func refreshCloudfrontFunctionState(conn *lambda.Lambda, functionName string) resource.StateRefreshFunc {
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

		state := aws.StringValue(output.Configuration.State)

		if state == lambda.StateFailed {
			return output.Configuration, state, fmt.Errorf("%s: %s", aws.StringValue(output.Configuration.StateReasonCode), aws.StringValue(output.Configuration.StateReason))
		}

		return output.Configuration, state, nil
	}
}

func waitForCloudfrontFunctionCreation(conn *lambda.Lambda, functionName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.StatePending},
		Target:  []string{lambda.StateActive},
		Refresh: refreshCloudfrontFunctionState(conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForCloudfrontFunctionUpdate(conn *lambda.Lambda, functionName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{lambda.LastUpdateStatusInProgress},
		Target:  []string{lambda.LastUpdateStatusSuccessful},
		Refresh: refreshCloudfrontFunctionLastUpdateStatus(conn, functionName),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func flattenLambdaEnvironment(apiObject *lambda.EnvironmentResponse) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Variables; v != nil {
		tfMap["variables"] = aws.StringValueMap(v)
	}

	return []interface{}{tfMap}
}

func flattenLambdaFileSystemConfigs(fscList []*lambda.FileSystemConfig) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(fscList))
	for _, fsc := range fscList {
		f := make(map[string]interface{})
		f["arn"] = *fsc.Arn
		f["local_mount_path"] = *fsc.LocalMountPath

		results = append(results, f)
	}
	return results
}

func expandLambdaFileSystemConfigs(fscMaps []interface{}) []*lambda.FileSystemConfig {
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

func flattenLambdaImageConfig(response *lambda.ImageConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if response == nil || response.Error != nil || response.ImageConfig == nil {
		return nil
	}

	settings["command"] = response.ImageConfig.Command
	settings["entry_point"] = response.ImageConfig.EntryPoint
	settings["working_directory"] = response.ImageConfig.WorkingDirectory

	return []map[string]interface{}{settings}
}

func expandLambdaImageConfigs(imageConfigMaps []interface{}) *lambda.ImageConfig {
	imageConfig := &lambda.ImageConfig{}
	// only one image_config block is allowed
	if len(imageConfigMaps) == 1 && imageConfigMaps[0] != nil {
		config := imageConfigMaps[0].(map[string]interface{})
		if len(config["entry_point"].([]interface{})) > 0 {
			imageConfig.EntryPoint = expandStringList(config["entry_point"].([]interface{}))
		}
		if len(config["command"].([]interface{})) > 0 {
			imageConfig.Command = expandStringList(config["command"].([]interface{}))
		}
		imageConfig.WorkingDirectory = aws.String(config["working_directory"].(string))
	}
	return imageConfig
}
