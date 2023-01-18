package lambda

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFunctionRead,

		Schema: map[string]*schema.Schema{
			"architectures": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
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
						"target_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment": {
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
						"size": {
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
						"arn": {
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
			"kms_key_arn": {
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
			"role": {
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tracing_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	functionName := d.Get("function_name").(string)
	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	} else {
		// If no qualifier provided, set version to latest published version.
		latest, err := findLatestFunctionVersionByName(conn, functionName)

		if err != nil {
			return fmt.Errorf("reading Lambda Function (%s) latest version: %w", functionName, err)
		}

		// If no published version exists, AWS returns '$LATEST' for latestVersion
		if v := aws.StringValue(latest.Version); v != FunctionVersionLatest {
			input.Qualifier = aws.String(v)
		}
	}

	output, err := findFunction(conn, input)

	if err != nil {
		return fmt.Errorf("reading Lambda Function (%s): %w", functionName, err)
	}

	function := output.Configuration
	functionARN := aws.StringValue(function.FunctionArn)
	qualifierSuffix := fmt.Sprintf(":%s", aws.StringValue(input.Qualifier))
	versionSuffix := fmt.Sprintf(":%s", aws.StringValue(function.Version))

	d.Set("version", function.Version)

	qualifiedARN := functionARN
	if !strings.HasSuffix(functionARN, qualifierSuffix) && !strings.HasSuffix(functionARN, versionSuffix) {
		qualifiedARN = functionARN + versionSuffix
	}

	unqualifiedARN := strings.TrimSuffix(functionARN, qualifierSuffix)

	d.SetId(functionName)
	d.Set("architectures", aws.StringValueSlice(function.Architectures))
	d.Set("arn", unqualifiedARN)
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
	d.Set("qualified_arn", qualifiedARN)
	d.Set("qualified_invoke_arn", functionInvokeARN(qualifiedARN, meta))
	if output.Concurrency != nil {
		d.Set("reserved_concurrent_executions", output.Concurrency.ReservedConcurrentExecutions)
	} else {
		d.Set("reserved_concurrent_executions", -1)
	}
	d.Set("role", function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("signing_job_arn", function.SigningJobArn)
	d.Set("signing_profile_version_arn", function.SigningProfileVersionArn)
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

	if err := d.Set("tags", KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	// See resourceFunctionRead().
	if partition := meta.(*conns.AWSClient).Partition; partition == endpoints.AwsPartitionID && SignerServiceIsAvailable(meta.(*conns.AWSClient).Region) {
		var codeSigningConfigArn string

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
