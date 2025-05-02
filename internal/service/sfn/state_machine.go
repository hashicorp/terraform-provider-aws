// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sfn_state_machine", name="State Machine")
// @Tags(identifierAttribute="id")
func resourceStateMachine() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStateMachineCreate,
		ReadWithoutTimeout:   resourceStateMachineRead,
		UpdateWithoutTimeout: resourceStateMachineUpdate,
		DeleteWithoutTimeout: resourceStateMachineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024*1024), // 1048576
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_data_key_reuse_period_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(60, 900),
						},
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrLoggingConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_execution_data": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"level": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogLevel](),
						},
						"log_destination": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 80),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "the name should only contain 0-9, A-Z, a-z, - and _"),
				),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 80-id.UniqueIDSuffixLength),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "the name should only contain 0-9, A-Z, a-z, - and _"),
				),
			},
			"publish": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"revision_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"state_machine_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tracing_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.StateMachineTypeStandard,
				ValidateDiagFunc: enum.Validate[awstypes.StateMachineType](),
			},
			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			stateMachineDefinitionValidate,
			stateMachineUpdateComputedAttributesOnPublish,
		),
	}
}

func resourceStateMachineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &sfn.CreateStateMachineInput{
		Definition: aws.String(d.Get("definition").(string)),
		Name:       aws.String(name),
		Publish:    d.Get("publish").(bool),
		RoleArn:    aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:       getTagsIn(ctx),
		Type:       awstypes.StateMachineType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.LoggingConfiguration = expandLoggingConfiguration(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("tracing_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.TracingConfiguration = expandTracingConfiguration(v.([]any)[0].(map[string]any))
	}

	// This is done to deal with IAM eventual consistency.
	// Note: the instance may be in a deleting mode, hence the retry
	// when creating the step function. This can happen when we are
	// updating the resource (since there is no update API call).
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateStateMachine(ctx, input)
	}, "StateMachineDeleting", "AccessDeniedException")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Step Functions State Machine (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*sfn.CreateStateMachineOutput).StateMachineArn))

	return append(diags, resourceStateMachineRead(ctx, d, meta)...)
}

func resourceStateMachineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	output, err := findStateMachineByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Step Functions State Machine (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Step Functions State Machine (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.StateMachineArn)
	if output.CreationDate != nil {
		d.Set(names.AttrCreationDate, aws.ToTime(output.CreationDate).Format(time.RFC3339))
	} else {
		d.Set(names.AttrCreationDate, nil)
	}
	d.Set("definition", output.Definition)
	d.Set(names.AttrDescription, output.Description)
	if output.EncryptionConfiguration != nil {
		if err := d.Set(names.AttrEncryptionConfiguration, []any{flattenEncryptionConfiguration(output.EncryptionConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrEncryptionConfiguration, nil)
	}
	if output.LoggingConfiguration != nil {
		if err := d.Set(names.AttrLoggingConfiguration, []any{flattenLoggingConfiguration(output.LoggingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_configuration: %s", err)
		}
	} else {
		d.Set(names.AttrLoggingConfiguration, nil)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set("publish", d.Get("publish").(bool))
	d.Set("revision_id", output.RevisionId)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set(names.AttrStatus, output.Status)
	if output.TracingConfiguration != nil {
		if err := d.Set("tracing_configuration", []any{flattenTracingConfiguration(output.TracingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tracing_configuration: %s", err)
		}
	} else {
		d.Set("tracing_configuration", nil)
	}
	d.Set(names.AttrType, output.Type)

	input := &sfn.ListStateMachineVersionsInput{
		StateMachineArn: aws.String(d.Id()),
	}
	listVersionsOutput, err := conn.ListStateMachineVersions(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Step Functions State Machine (%s) Versions: %s", d.Id(), err)
	}

	// The results are sorted in descending order of the version creation time.
	// https://docs.aws.amazon.com/step-functions/latest/apireference/API_ListStateMachineVersions.html
	if len(listVersionsOutput.StateMachineVersions) > 0 {
		d.Set("state_machine_version_arn", listVersionsOutput.StateMachineVersions[0].StateMachineVersionArn)
	} else {
		d.Set("state_machine_version_arn", nil)
	}

	return diags
}

func resourceStateMachineUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// "You must include at least one of definition or roleArn or you will receive a MissingRequiredParameter error"
		publish := d.Get("publish").(bool)
		input := &sfn.UpdateStateMachineInput{
			Definition:      aws.String(d.Get("definition").(string)),
			Publish:         publish,
			RoleArn:         aws.String(d.Get(names.AttrRoleARN).(string)),
			StateMachineArn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrEncryptionConfiguration) {
			if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.EncryptionConfiguration = expandEncryptionConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange(names.AttrLoggingConfiguration) {
			if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.LoggingConfiguration = expandLoggingConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("tracing_configuration") {
			if v, ok := d.GetOk("tracing_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.TracingConfiguration = expandTracingConfiguration(v.([]any)[0].(map[string]any))
			}
		}

		if publish {
			input.VersionDescription = aws.String(d.Get("version_description").(string))
		}

		_, err := conn.UpdateStateMachine(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Step Functions State Machine (%s): %s", d.Id(), err)
		}

		// Handle eventual consistency after update.
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError { // nosemgrep:ci.helper-schema-retry-RetryContext-without-TimeoutError-check
			output, err := findStateMachineByARN(ctx, conn, d.Id())

			if err != nil {
				return retry.NonRetryableError(err)
			}

			if d.HasChange("definition") && !verify.JSONBytesEqual([]byte(aws.ToString(output.Definition)), []byte(d.Get("definition").(string))) ||
				d.HasChange(names.AttrRoleARN) && aws.ToString(output.RoleArn) != d.Get(names.AttrRoleARN).(string) ||
				//d.HasChange("publish") && aws.Bool(output.Publish) != d.Get("publish").(bool) ||
				d.HasChange("tracing_configuration.0.enabled") && output.TracingConfiguration != nil && output.TracingConfiguration.Enabled != d.Get("tracing_configuration.0.enabled").(bool) ||
				d.HasChange("logging_configuration.0.include_execution_data") && output.LoggingConfiguration != nil && output.LoggingConfiguration.IncludeExecutionData != d.Get("logging_configuration.0.include_execution_data").(bool) ||
				d.HasChange("logging_configuration.0.level") && output.LoggingConfiguration != nil && string(output.LoggingConfiguration.Level) != d.Get("logging_configuration.0.level").(string) ||
				d.HasChange("encryption_configuration.0.kms_key_id") && output.EncryptionConfiguration != nil && output.EncryptionConfiguration.KmsKeyId != nil && aws.ToString(output.EncryptionConfiguration.KmsKeyId) != d.Get("encryption_configuration.0.kms_key_id") ||
				d.HasChange("encryption_configuration.0.encryption_type") && output.EncryptionConfiguration != nil && string(output.EncryptionConfiguration.Type) != d.Get("encryption_configuration.0.encryption_type").(string) ||
				d.HasChange("encryption_configuration.0.kms_data_key_reuse_period_seconds") && output.EncryptionConfiguration != nil && output.EncryptionConfiguration.KmsDataKeyReusePeriodSeconds != nil && aws.ToInt32(output.EncryptionConfiguration.KmsDataKeyReusePeriodSeconds) != int32(d.Get("encryption_configuration.0.kms_data_key_reuse_period_seconds").(int)) {
				return retry.RetryableError(fmt.Errorf("Step Functions State Machine (%s) eventual consistency", d.Id()))
			}

			return nil
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Step Functions State Machine (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStateMachineRead(ctx, d, meta)...)
}

func resourceStateMachineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	log.Printf("[DEBUG] Deleting Step Functions State Machine: %s", d.Id())
	_, err := conn.DeleteStateMachine(ctx, &sfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Step Functions State Machine (%s): %s", d.Id(), err)
	}

	if _, err := waitStateMachineDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Step Functions State Machine (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStateMachineByARN(ctx context.Context, conn *sfn.Client, arn string) (*sfn.DescribeStateMachineOutput, error) {
	input := &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(arn),
	}

	output, err := conn.DescribeStateMachine(ctx, input)

	if errs.IsA[*awstypes.StateMachineDoesNotExist](err) {
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

func statusStateMachine(ctx context.Context, conn *sfn.Client, stateMachineArn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findStateMachineByARN(ctx, conn, stateMachineArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitStateMachineDeleted(ctx context.Context, conn *sfn.Client, stateMachineArn string, timeout time.Duration) (*sfn.DescribeStateMachineOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateMachineStatusActive, awstypes.StateMachineStatusDeleting),
		Target:  []string{},
		Refresh: statusStateMachine(ctx, conn, stateMachineArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sfn.DescribeStateMachineOutput); ok {
		return output, err
	}

	return nil, err
}

func expandLoggingConfiguration(tfMap map[string]any) *awstypes.LoggingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LoggingConfiguration{}

	if v, ok := tfMap["include_execution_data"].(bool); ok {
		apiObject.IncludeExecutionData = v
	}

	if v, ok := tfMap["level"].(string); ok && v != "" {
		apiObject.Level = awstypes.LogLevel(v)
	}

	if v, ok := tfMap["log_destination"].(string); ok && v != "" {
		apiObject.Destinations = []awstypes.LogDestination{{
			CloudWatchLogsLogGroup: &awstypes.CloudWatchLogsLogGroup{
				LogGroupArn: aws.String(v),
			},
		}}
	}

	return apiObject
}

func flattenLoggingConfiguration(apiObject *awstypes.LoggingConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"include_execution_data": apiObject.IncludeExecutionData,
		"level":                  apiObject.Level,
	}

	if v := apiObject.Destinations; len(v) > 0 {
		tfMap["log_destination"] = aws.ToString(v[0].CloudWatchLogsLogGroup.LogGroupArn)
	}

	return tfMap
}

func expandTracingConfiguration(tfMap map[string]any) *awstypes.TracingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TracingConfiguration{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	return apiObject
}

func flattenTracingConfiguration(apiObject *awstypes.TracingConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	return tfMap
}

func stateMachineUpdateComputedAttributesOnPublish(_ context.Context, d *schema.ResourceDiff, meta any) error {
	if publish := d.Get("publish").(bool); publish && stateMachineNeedsConfigUpdate(d) {
		d.SetNewComputed("revision_id")
		d.SetNewComputed("state_machine_version_arn")
	}
	return nil
}

func stateMachineNeedsConfigUpdate(d sdkv2.ResourceDiffer) bool {
	for k, attr := range resourceStateMachine().Schema {
		if attr.ForceNew {
			continue
		}
		if attr.Computed && !attr.Optional {
			continue
		}

		if d.HasChange(k) {
			return true
		}
	}
	return false
}

func stateMachineDefinitionValidate(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	conn := meta.(*conns.AWSClient).SFNClient(ctx)

	if d.HasChange("definition") {
		definition := d.Get("definition").(string)
		if definition == "" {
			return nil
		}

		input := &sfn.ValidateStateMachineDefinitionInput{
			Definition: aws.String(definition),
			Type:       awstypes.StateMachineType(d.Get(names.AttrType).(string)),
		}

		output, err := conn.ValidateStateMachineDefinition(ctx, input)

		if err != nil {
			return fmt.Errorf("validating Step Functions State Machine definition: %w", err)
		}

		if result := output.Result; result != awstypes.ValidateStateMachineDefinitionResultCodeOk {
			errs := tfslices.ApplyToAll(output.Diagnostics, func(v awstypes.ValidateStateMachineDefinitionDiagnostic) error {
				return fmt.Errorf("%s (%s): %s", v.Severity, aws.ToString(v.Code), aws.ToString(v.Message))
			})

			return fmt.Errorf("invalid Step Functions State Machine definition: %w", errors.Join(errs...))
		}
	}

	return nil
}
