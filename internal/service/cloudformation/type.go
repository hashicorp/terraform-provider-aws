package cloudformation

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTypeCreate,
		DeleteWithoutTimeout: resourceTypeDelete,
		ReadWithoutTimeout:   resourceTypeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecated_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"documentation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"is_default_version": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"logging_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 512),
								validation.StringMatch(regexp.MustCompile(`[\.\-_/#A-Za-z0-9]+`), "must contain only alphanumeric, period, hyphen, forward slash, and octothorp characters"),
							),
						},
						"log_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"provisioning_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"schema": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"schema_handler_package": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(https|s3)\:\/\/.+`), "must begin with s3:// or https://"),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.RegistryType_Values(), false),
			},
			"type_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 204),
					validation.StringMatch(regexp.MustCompile(`[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}::[A-Za-z0-9]{2,64}(::MODULE){0,1}`), "three alphanumeric character sections separated by double colons (::)"),
				),
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"visibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	typeName := d.Get("type_name").(string)
	input := &cloudformation.RegisterTypeInput{
		ClientRequestToken:   aws.String(resource.UniqueId()),
		SchemaHandlerPackage: aws.String(d.Get("schema_handler_package").(string)),
		TypeName:             aws.String(typeName),
	}

	if v, ok := d.GetOk("execution_role_arn"); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingConfig = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	output, err := conn.RegisterTypeWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error registering CloudFormation Type (%s): %w", typeName, err))
	}

	if output == nil || output.RegistrationToken == nil {
		return diag.FromErr(fmt.Errorf("error registering CloudFormation Type (%s): empty result", typeName))
	}

	registrationOutput, err := WaitTypeRegistrationProgressStatusComplete(ctx, conn, aws.StringValue(output.RegistrationToken))

	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for CloudFormation Type (%s) register: %w", typeName, err))
	}

	// Type Version ARN is not available until after registration is complete
	d.SetId(aws.StringValue(registrationOutput.TypeVersionArn))

	return resourceTypeRead(ctx, d, meta)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	output, err := FindTypeByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading CloudFormation Type (%s): %w", d.Id(), err))
	}

	typeARN, versionID, err := TypeVersionARNToTypeARNAndVersionID(d.Id())

	if err != nil {
		return diag.FromErr(fmt.Errorf("error parsing CloudFormation Type (%s) ARN: %w", d.Id(), err))
	}

	d.Set("arn", output.Arn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set("deprecated_status", output.DeprecatedStatus)
	d.Set("description", output.Description)
	d.Set("documentation_url", output.DocumentationUrl)
	d.Set("execution_role_arn", output.ExecutionRoleArn)
	d.Set("is_default_version", output.IsDefaultVersion)
	if output.LoggingConfig != nil {
		if err := d.Set("logging_config", []interface{}{flattenLoggingConfig(output.LoggingConfig)}); err != nil {
			return diag.FromErr(fmt.Errorf("error setting logging_config: %w", err))
		}
	} else {
		d.Set("logging_config", nil)
	}
	d.Set("provisioning_type", output.ProvisioningType)
	d.Set("schema", output.Schema)
	d.Set("source_url", output.SourceUrl)
	d.Set("type", output.Type)
	d.Set("type_arn", typeARN)
	d.Set("type_name", output.TypeName)
	d.Set("version_id", versionID)
	d.Set("visibility", output.Visibility)

	return nil
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	input := &cloudformation.DeregisterTypeInput{
		Arn: aws.String(d.Id()),
	}

	_, err := conn.DeregisterTypeWithContext(ctx, input)

	// Must deregister type if removing final LIVE version. This error can also occur
	// when the type is already DEPRECATED.
	if tfawserr.ErrMessageContains(err, cloudformation.ErrCodeCFNRegistryException, "is the default version and cannot be deregistered") {
		typeARN, _, err := TypeVersionARNToTypeARNAndVersionID(d.Id())

		if err != nil {
			return diag.FromErr(fmt.Errorf("error parsing CloudFormation Type (%s) ARN: %w", d.Id(), err))
		}

		input := &cloudformation.ListTypeVersionsInput{
			Arn:              aws.String(typeARN),
			DeprecatedStatus: aws.String(cloudformation.DeprecatedStatusLive),
		}

		var typeVersionSummaries []*cloudformation.TypeVersionSummary

		err = conn.ListTypeVersionsPagesWithContext(ctx, input, func(page *cloudformation.ListTypeVersionsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			typeVersionSummaries = append(typeVersionSummaries, page.TypeVersionSummaries...)

			if len(typeVersionSummaries) > 1 {
				return false
			}

			return !lastPage
		})

		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing CloudFormation Type (%s) Versions: %w", d.Id(), err))
		}

		if len(typeVersionSummaries) <= 1 {
			input := &cloudformation.DeregisterTypeInput{
				Arn: aws.String(typeARN),
			}

			_, err := conn.DeregisterTypeWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
				return nil
			}

			if err != nil {
				return diag.FromErr(fmt.Errorf("error deregistering CloudFormation Type (%s): %w", d.Id(), err))
			}

			return nil
		}
	}

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeTypeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deregistering CloudFormation Type (%s): %w", d.Id(), err))
	}

	return nil
}

func expandLoggingConfig(tfMap map[string]interface{}) *cloudformation.LoggingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudformation.LoggingConfig{}

	if v, ok := tfMap["log_group_name"].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["log_role_arn"].(string); ok && v != "" {
		apiObject.LogRoleArn = aws.String(v)
	}

	return apiObject
}

func expandOperationPreferences(tfMap map[string]interface{}) *cloudformation.StackSetOperationPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudformation.StackSetOperationPreferences{}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok {
		apiObject.FailureToleranceCount = aws.Int64(int64(v))
	}
	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok {
		apiObject.FailureTolerancePercentage = aws.Int64(int64(v))
	}
	if v, ok := tfMap["max_concurrent_count"].(int); ok {
		apiObject.MaxConcurrentCount = aws.Int64(int64(v))
	}
	if v, ok := tfMap["max_concurrent_percentage"].(int); ok {
		apiObject.MaxConcurrentPercentage = aws.Int64(int64(v))
	}
	if v, ok := tfMap["region_concurrency_type"].(string); ok && v != "" {
		apiObject.RegionConcurrencyType = aws.String(v)
	}
	if v, ok := tfMap["region_order"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RegionOrder = flex.ExpandStringSet(v)
	}

	if ftc, ftp := aws.Int64Value(apiObject.FailureToleranceCount), aws.Int64Value(apiObject.FailureTolerancePercentage); ftp == 0 {
		apiObject.FailureTolerancePercentage = nil
	} else if ftc == 0 {
		apiObject.FailureToleranceCount = nil
	}

	if mcc, mcp := aws.Int64Value(apiObject.MaxConcurrentCount), aws.Int64Value(apiObject.MaxConcurrentPercentage); mcp == 0 {
		apiObject.MaxConcurrentPercentage = nil
	} else if mcc == 0 {
		apiObject.MaxConcurrentCount = nil
	}

	return apiObject
}

func flattenLoggingConfig(apiObject *cloudformation.LoggingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupName; v != nil {
		tfMap["log_group_name"] = aws.StringValue(v)
	}

	if v := apiObject.LogRoleArn; v != nil {
		tfMap["log_role_arn"] = aws.StringValue(v)
	}

	return tfMap
}
