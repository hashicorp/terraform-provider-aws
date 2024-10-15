// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudformation_type", name="Type")
func resourceType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTypeCreate,
		DeleteWithoutTimeout: resourceTypeDelete,
		ReadWithoutTimeout:   resourceTypeRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"documentation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrExecutionRoleARN: {
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
						names.AttrLogGroupName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 512),
								validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_./#-]+`), "must contain only alphanumeric, period, hyphen, forward slash, and octothorp characters"),
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
			names.AttrSchema: {
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
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(https|s3)\:\/\/.+`), "must begin with s3:// or https://"),
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.RegistryType](),
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
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}::[0-9A-Za-z]{2,64}(::MODULE){0,1}`), "three alphanumeric character sections separated by double colons (::)"),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	typeName := d.Get("type_name").(string)
	input := &cloudformation.RegisterTypeInput{
		ClientRequestToken:   aws.String(id.UniqueId()),
		SchemaHandlerPackage: aws.String(d.Get("schema_handler_package").(string)),
		TypeName:             aws.String(typeName),
	}

	if v, ok := d.GetOk(names.AttrExecutionRoleARN); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingConfig = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrType); ok {
		input.Type = awstypes.RegistryType(v.(string))
	}

	output, err := conn.RegisterType(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "registering CloudFormation Type (%s): %s", typeName, err)
	}

	registrationOutput, err := waitTypeRegistrationProgressStatusComplete(ctx, conn, aws.ToString(output.RegistrationToken))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Type (%s) register: %s", typeName, err)
	}

	// Type Version ARN is not available until after registration is complete
	d.SetId(aws.ToString(registrationOutput.TypeVersionArn))

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	output, err := findTypeByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Type (%s): %s", d.Id(), err)
	}

	typeARN, versionID, err := typeVersionARNToTypeARNAndVersionID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set("default_version_id", output.DefaultVersionId)
	d.Set("deprecated_status", output.DeprecatedStatus)
	d.Set(names.AttrDescription, output.Description)
	d.Set("documentation_url", output.DocumentationUrl)
	d.Set(names.AttrExecutionRoleARN, output.ExecutionRoleArn)
	d.Set("is_default_version", output.IsDefaultVersion)
	if output.LoggingConfig != nil {
		if err := d.Set("logging_config", []interface{}{flattenLoggingConfig(output.LoggingConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_config: %s", err)
		}
	} else {
		d.Set("logging_config", nil)
	}
	d.Set("provisioning_type", output.ProvisioningType)
	d.Set(names.AttrSchema, output.Schema)
	d.Set("source_url", output.SourceUrl)
	d.Set(names.AttrType, output.Type)
	d.Set("type_arn", typeARN)
	d.Set("type_name", output.TypeName)
	d.Set("version_id", versionID)
	d.Set("visibility", output.Visibility)

	return diags
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	log.Printf("[INFO] Deleting CloudFormation Type: %s", d.Id())
	_, err := conn.DeregisterType(ctx, &cloudformation.DeregisterTypeInput{
		Arn: aws.String(d.Id()),
	})

	// Must deregister type if removing final LIVE version. This error can also occur
	// when the type is already DEPRECATED.
	if errs.IsAErrorMessageContains[*awstypes.CFNRegistryException](err, "is the default version and cannot be deregistered") {
		typeARN, _, err := typeVersionARNToTypeARNAndVersionID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &cloudformation.ListTypeVersionsInput{
			Arn:              aws.String(typeARN),
			DeprecatedStatus: awstypes.DeprecatedStatusLive,
		}

		var typeVersionSummaries []awstypes.TypeVersionSummary

		pages := cloudformation.NewListTypeVersionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "listing CloudFormation Type (%s) Versions: %s", d.Id(), err)
			}

			typeVersionSummaries = append(typeVersionSummaries, page.TypeVersionSummaries...)
		}

		if len(typeVersionSummaries) <= 1 {
			input := &cloudformation.DeregisterTypeInput{
				Arn: aws.String(typeARN),
			}

			_, err := conn.DeregisterType(ctx, input)

			if errs.IsA[*awstypes.TypeNotFoundException](err) {
				return diags
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deregistering CloudFormation Type (%s): %s", d.Id(), err)
			}

			return diags
		}
	}

	if errs.IsA[*awstypes.TypeNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deregistering CloudFormation Type (%s): %s", d.Id(), err)
	}

	return diags
}

func findTypeByARN(ctx context.Context, conn *cloudformation.Client, arn string) (*cloudformation.DescribeTypeOutput, error) {
	input := &cloudformation.DescribeTypeInput{
		Arn: aws.String(arn),
	}

	output, err := findType(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.DeprecatedStatus; status == awstypes.DeprecatedStatusDeprecated {
		return nil, &retry.NotFoundError{
			LastRequest: input,
			Message:     string(status),
		}
	}

	return output, nil
}

func findTypeByName(ctx context.Context, conn *cloudformation.Client, name string) (*cloudformation.DescribeTypeOutput, error) {
	input := &cloudformation.DescribeTypeInput{
		Type:     awstypes.RegistryTypeResource,
		TypeName: aws.String(name),
	}

	return findType(ctx, conn, input)
}

func findType(ctx context.Context, conn *cloudformation.Client, input *cloudformation.DescribeTypeInput) (*cloudformation.DescribeTypeOutput, error) {
	output, err := conn.DescribeType(ctx, input)

	if errs.IsA[*awstypes.TypeNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findTypeRegistrationByToken(ctx context.Context, conn *cloudformation.Client, registrationToken string) (*cloudformation.DescribeTypeRegistrationOutput, error) {
	input := &cloudformation.DescribeTypeRegistrationInput{
		RegistrationToken: aws.String(registrationToken),
	}

	output, err := conn.DescribeTypeRegistration(ctx, input)

	if errs.IsA[*awstypes.CFNRegistryException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusTypeRegistrationProgress(ctx context.Context, conn *cloudformation.Client, registrationToken string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTypeRegistrationByToken(ctx, conn, registrationToken)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ProgressStatus), nil
	}
}

func waitTypeRegistrationProgressStatusComplete(ctx context.Context, conn *cloudformation.Client, registrationToken string) (*cloudformation.DescribeTypeRegistrationOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistrationStatusInProgress),
		Target:  enum.Slice(awstypes.RegistrationStatusComplete),
		Refresh: statusTypeRegistrationProgress(ctx, conn, registrationToken),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeTypeRegistrationOutput); ok {
		return output, err
	}

	return nil, err
}

func expandLoggingConfig(tfMap map[string]interface{}) *awstypes.LoggingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LoggingConfig{}

	if v, ok := tfMap[names.AttrLogGroupName].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["log_role_arn"].(string); ok && v != "" {
		apiObject.LogRoleArn = aws.String(v)
	}

	return apiObject
}

func expandOperationPreferences(tfMap map[string]interface{}) *awstypes.StackSetOperationPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.StackSetOperationPreferences{}

	if v, ok := tfMap["failure_tolerance_count"].(int); ok {
		apiObject.FailureToleranceCount = aws.Int32(int32(v))
	}
	if v, ok := tfMap["failure_tolerance_percentage"].(int); ok {
		apiObject.FailureTolerancePercentage = aws.Int32(int32(v))
	}
	if v, ok := tfMap["max_concurrent_count"].(int); ok {
		apiObject.MaxConcurrentCount = aws.Int32(int32(v))
	}
	if v, ok := tfMap["max_concurrent_percentage"].(int); ok {
		apiObject.MaxConcurrentPercentage = aws.Int32(int32(v))
	}
	if v, ok := tfMap["concurrency_mode"].(string); ok && v != "" {
		apiObject.ConcurrencyMode = awstypes.ConcurrencyMode(v)
	}
	if v, ok := tfMap["region_concurrency_type"].(string); ok && v != "" {
		apiObject.RegionConcurrencyType = awstypes.RegionConcurrencyType(v)
	}
	if v, ok := tfMap["region_order"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RegionOrder = flex.ExpandStringValueSet(v)
	}

	if ftc, ftp := aws.ToInt32(apiObject.FailureToleranceCount), aws.ToInt32(apiObject.FailureTolerancePercentage); ftp == 0 {
		apiObject.FailureTolerancePercentage = nil
	} else if ftc == 0 {
		apiObject.FailureToleranceCount = nil
	}

	if mcc, mcp := aws.ToInt32(apiObject.MaxConcurrentCount), aws.ToInt32(apiObject.MaxConcurrentPercentage); mcp == 0 {
		apiObject.MaxConcurrentPercentage = nil
	} else if mcc == 0 {
		apiObject.MaxConcurrentCount = nil
	}

	return apiObject
}

func flattenLoggingConfig(apiObject *awstypes.LoggingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupName; v != nil {
		tfMap[names.AttrLogGroupName] = aws.ToString(v)
	}

	if v := apiObject.LogRoleArn; v != nil {
		tfMap["log_role_arn"] = aws.ToString(v)
	}

	return tfMap
}
