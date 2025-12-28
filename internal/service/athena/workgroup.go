// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_athena_workgroup", name="WorkGroup")
// @Tags(identifierAttribute="arn")
func resourceWorkGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkGroupCreate,
		ReadWithoutTimeout:   resourceWorkGroupRead,
		UpdateWithoutTimeout: resourceWorkGroupUpdate,
		DeleteWithoutTimeout: resourceWorkGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: managedQueryResultsValidation,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bytes_scanned_cutoff_per_query": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.IntAtLeast(10485760),
								validation.IntInSlice([]int{0}),
							),
						},
						"customer_content_encryption_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKey: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:key/?[a-zA-Z_0-9+=,.@\-_/]+$|^arn:aws[a-z\-]*:kms:([a-z0-9\-]+):\d{12}:alias/?[a-zA-Z_0-9+=,.@\-_/]+$|^alias/[a-zA-Z0-9/_-]+$|[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`), "must be a valid KMS Key ARN or Alias"),
									},
								},
							},
						},
						"enable_minimum_encryption_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enforce_workgroup_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrEngineVersion: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"effective_engine_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"selected_engine_version": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "AUTO",
									},
								},
							},
						},
						"execution_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"identity_center_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable_identity_center": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"identity_center_instance_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"managed_query_results_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrEncryptionConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrKMSKey: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
								},
							},
						},
						"publish_cloudwatch_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"result_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acl_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_acl_option": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.S3AclOption](),
												},
											},
										},
									},
									names.AttrEncryptionConfiguration: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_option": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.EncryptionOption](),
												},
												names.AttrKMSKeyARN: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
											},
										},
									},
									names.AttrExpectedBucketOwner: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"output_location": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"requester_pays_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
				),
			},
			names.AttrState: {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.WorkGroupStateEnabled,
				ValidateDiagFunc: enum.Validate[types.WorkGroupState](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := athena.CreateWorkGroupInput{
		Configuration: expandWorkGroupConfiguration(d.Get(names.AttrConfiguration).([]any)),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateWorkGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena WorkGroup (%s): %s", name, err)
	}

	d.SetId(name)

	if v := types.WorkGroupState(d.Get(names.AttrState).(string)); v == types.WorkGroupStateDisabled {
		input := athena.UpdateWorkGroupInput{
			State:     v,
			WorkGroup: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	wg, err := findWorkGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena WorkGroup (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "athena",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("workgroup/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	if err := d.Set(names.AttrConfiguration, flattenWorkGroupConfiguration(wg.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	d.Set(names.AttrDescription, wg.Description)
	d.Set(names.AttrForceDestroy, d.Get(names.AttrForceDestroy))
	d.Set(names.AttrName, wg.Name)
	d.Set(names.AttrState, wg.State)

	return diags
}

func resourceWorkGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := athena.UpdateWorkGroupInput{
			WorkGroup: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange(names.AttrConfiguration) {
			input.ConfigurationUpdates = expandWorkGroupConfigurationUpdates(d.Get(names.AttrConfiguration).([]any))
			if d.HasChange("configuration.0.customer_content_encryption_configuration") {
				// When customer_content_encryption_configuration is changed and the expander returns nil,
				// the content_encryption_configuration needs to be removed.
				// To remove it, set RemoveCustomerContentEncryptionConfiguration to true.
				if input.ConfigurationUpdates == nil || input.ConfigurationUpdates.CustomerContentEncryptionConfiguration == nil {
					input.ConfigurationUpdates.RemoveCustomerContentEncryptionConfiguration = aws.Bool(true)
				}
			}
			if d.HasChange("configuration.0.enable_minimum_encryption_configuration") {
				// When enable_minimum_encryption_configuration is returned as nil, set it to false to disable it.
				if input.ConfigurationUpdates == nil || input.ConfigurationUpdates.EnableMinimumEncryptionConfiguration == nil {
					input.ConfigurationUpdates.EnableMinimumEncryptionConfiguration = aws.Bool(false)
				}
			}
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrState) {
			input.State = types.WorkGroupState(d.Get(names.AttrState).(string))
		}

		_, err := conn.UpdateWorkGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	log.Printf("[DEBUG] Deleting Athena WorkGroup (%s)", d.Id())
	input := athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}
	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		input.RecursiveDeleteOption = aws.Bool(v.(bool))
	}
	_, err := conn.DeleteWorkGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkGroupByName(ctx context.Context, conn *athena.Client, name string) (*types.WorkGroup, error) {
	input := athena.GetWorkGroupInput{
		WorkGroup: aws.String(name),
	}

	output, err := conn.GetWorkGroup(ctx, &input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "is not found") {
		return nil, &sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WorkGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.WorkGroup, nil
}

func expandWorkGroupConfiguration(l []any) *types.WorkGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	configuration := &types.WorkGroupConfiguration{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configuration.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	}

	if v, ok := m["customer_content_encryption_configuration"]; ok {
		configuration.CustomerContentEncryptionConfiguration = expandWorkGroupCustomerContentEncryptionConfiguration(v.([]any))
	}

	// Depending on other configurations, enable_minimum_encryption_configuration
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	if v, ok := m["enable_minimum_encryption_configuration"].(bool); ok && v {
		configuration.EnableMinimumEncryptionConfiguration = aws.Bool(v)
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configuration.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]any); ok && len(v) > 0 && v[0] != nil {
		configuration.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configuration.ExecutionRole = aws.String(v)
	}

	if v, ok := m["identity_center_configuration"]; ok {
		configuration.IdentityCenterConfiguration = expandWorkGroupIdentityCenterConfiguration(v.([]any))
	}

	if v, ok := m["managed_query_results_configuration"]; ok {
		configuration.ManagedQueryResultsConfiguration = expandWorkGroupManagedQueryResultsConfiguration(v.([]any))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandWorkGroupResultConfiguration(v.([]any))
	}

	if v, ok := m["requester_pays_enabled"].(bool); ok {
		configuration.RequesterPaysEnabled = aws.Bool(v)
	}

	return configuration
}

func expandWorkGroupCustomerContentEncryptionConfiguration(l []any) *types.CustomerContentEncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	// customerContentEncryptionConfiguration is created and returned
	// only when a KMS key is specified.
	// Otherwise, nil is returned to avoid an SDK error.
	//
	// When the KMS key is removed from the configuration during an update,
	// the returned nil is handled in the Update function via
	// RemoveCustomerContentEncryptionConfiguration.
	if v, ok := m[names.AttrKMSKey].(string); ok && v != "" {
		customerContentEncryptionConfiguration := &types.CustomerContentEncryptionConfiguration{
			KmsKey: aws.String(v),
		}
		return customerContentEncryptionConfiguration
	}
	return nil
}

func expandWorkGroupEngineVersion(l []any) *types.EngineVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	engineVersion := &types.EngineVersion{}

	if v, ok := m["selected_engine_version"].(string); ok && v != "" {
		engineVersion.SelectedEngineVersion = aws.String(v)
	}

	return engineVersion
}

func expandWorkGroupConfigurationUpdates(l []any) *types.WorkGroupConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	configurationUpdates := &types.WorkGroupConfigurationUpdates{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	} else {
		configurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
	}

	if v, ok := m["customer_content_encryption_configuration"]; ok {
		configurationUpdates.CustomerContentEncryptionConfiguration = expandWorkGroupCustomerContentEncryptionConfiguration(v.([]any))
	}

	// Depending on other configurations, enable_minimum_encryption_configuration
	// must not be specified, even when set to false.
	// Therefore, the value is set only when it is true to avoid an API error.
	// The returned nil is handled in the Update function when this setting
	// needs to be disabled.
	if v, ok := m["enable_minimum_encryption_configuration"].(bool); ok && v {
		configurationUpdates.EnableMinimumEncryptionConfiguration = aws.Bool(v)
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]any); ok && len(v) > 0 && v[0] != nil {
		configurationUpdates.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configurationUpdates.ExecutionRole = aws.String(v)
	}

	if v, ok := m["managed_query_results_configuration"]; ok {
		configurationUpdates.ManagedQueryResultsConfigurationUpdates = expandWorkGroupManagedQueryResultsConfigurationUpdates(v.([]any))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := m["result_configuration"]; ok {
		configurationUpdates.ResultConfigurationUpdates = expandWorkGroupResultConfigurationUpdates(v.([]any))
	}

	if v, ok := m["requester_pays_enabled"].(bool); ok {
		configurationUpdates.RequesterPaysEnabled = aws.Bool(v)
	}

	return configurationUpdates
}

func expandWorkGroupIdentityCenterConfiguration(l []any) *types.IdentityCenterConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	identityCenterConfiguration := &types.IdentityCenterConfiguration{}

	if v, ok := m["enable_identity_center"].(bool); ok {
		identityCenterConfiguration.EnableIdentityCenter = aws.Bool(v)
	}

	if v, ok := m["identity_center_instance_arn"].(string); ok && v != "" {
		identityCenterConfiguration.IdentityCenterInstanceArn = aws.String(v)
	}

	return identityCenterConfiguration
}

func expandWorkGroupResultConfiguration(l []any) *types.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	resultConfiguration := &types.ResultConfiguration{}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		resultConfiguration.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]any))
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfiguration.OutputLocation = aws.String(v)
	}

	if v, ok := m[names.AttrExpectedBucketOwner].(string); ok && v != "" {
		resultConfiguration.ExpectedBucketOwner = aws.String(v)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfiguration.AclConfiguration = expandResultConfigurationACLConfig(v.([]any))
	}

	return resultConfiguration
}

func expandWorkGroupResultConfigurationUpdates(l []any) *types.ResultConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	resultConfigurationUpdates := &types.ResultConfigurationUpdates{}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		resultConfigurationUpdates.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]any))
	} else {
		resultConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfigurationUpdates.OutputLocation = aws.String(v)
	} else {
		resultConfigurationUpdates.RemoveOutputLocation = aws.Bool(true)
	}

	if v, ok := m[names.AttrExpectedBucketOwner].(string); ok && v != "" {
		resultConfigurationUpdates.ExpectedBucketOwner = aws.String(v)
	} else {
		resultConfigurationUpdates.RemoveExpectedBucketOwner = aws.Bool(true)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfigurationUpdates.AclConfiguration = expandResultConfigurationACLConfig(v.([]any))
	} else {
		resultConfigurationUpdates.RemoveAclConfiguration = aws.Bool(true)
	}

	return resultConfigurationUpdates
}

func expandWorkGroupEncryptionConfiguration(l []any) *types.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	encryptionConfiguration := &types.EncryptionConfiguration{}

	if v, ok := m["encryption_option"]; ok && v.(string) != "" {
		encryptionConfiguration.EncryptionOption = types.EncryptionOption(v.(string))
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		encryptionConfiguration.KmsKey = aws.String(v.(string))
	}

	return encryptionConfiguration
}

func expandWorkGroupManagedQueryResultsConfiguration(l []any) *types.ManagedQueryResultsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	managedQueryResultsConfiguration := &types.ManagedQueryResultsConfiguration{}

	if v, ok := m[names.AttrEnabled].(bool); ok {
		managedQueryResultsConfiguration.Enabled = v
	}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		managedQueryResultsConfiguration.EncryptionConfiguration = expandManagedQueryResultsEncryptionConfiguration(v.([]any))
	}

	return managedQueryResultsConfiguration
}

func expandManagedQueryResultsEncryptionConfiguration(l []any) *types.ManagedQueryResultsEncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	if v, ok := m[names.AttrKMSKey].(string); !ok || v == "" {
		return nil
	}
	managedQueryResultsEncryptionConfiguration := &types.ManagedQueryResultsEncryptionConfiguration{}
	managedQueryResultsEncryptionConfiguration.KmsKey = aws.String(m[names.AttrKMSKey].(string))

	return managedQueryResultsEncryptionConfiguration
}

func expandWorkGroupManagedQueryResultsConfigurationUpdates(l []any) *types.ManagedQueryResultsConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	managedQueryResultsConfigurationUpdates := &types.ManagedQueryResultsConfigurationUpdates{}
	if v, ok := m[names.AttrEnabled].(bool); ok {
		managedQueryResultsConfigurationUpdates.Enabled = aws.Bool(v)
		if !v {
			managedQueryResultsConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
			return managedQueryResultsConfigurationUpdates
		}
	}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		encConfig := expandManagedQueryResultsEncryptionConfiguration(v.([]any))
		if encConfig != nil {
			managedQueryResultsConfigurationUpdates.EncryptionConfiguration = encConfig
		} else {
			managedQueryResultsConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
		}
	}

	return managedQueryResultsConfigurationUpdates
}

func flattenWorkGroupConfiguration(configuration *types.WorkGroupConfiguration) []any {
	if configuration == nil {
		return []any{}
	}

	m := map[string]any{
		"bytes_scanned_cutoff_per_query":      aws.ToInt64(configuration.BytesScannedCutoffPerQuery),
		"enforce_workgroup_configuration":     aws.ToBool(configuration.EnforceWorkGroupConfiguration),
		names.AttrEngineVersion:               flattenWorkGroupEngineVersion(configuration.EngineVersion),
		"execution_role":                      aws.ToString(configuration.ExecutionRole),
		"identity_center_configuration":       flattenWorkGroupIdentityCenterConfiguration(configuration.IdentityCenterConfiguration),
		"managed_query_results_configuration": flattenWorkGroupManagedQueryResultsConfiguration(configuration.ManagedQueryResultsConfiguration),
		"publish_cloudwatch_metrics_enabled":  aws.ToBool(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":                flattenWorkGroupResultConfiguration(configuration.ResultConfiguration),
		"requester_pays_enabled":              aws.ToBool(configuration.RequesterPaysEnabled),
		"customer_content_encryption_configuration": flattenWorkGroupCustomerContentEncryptionConfiguration(configuration.CustomerContentEncryptionConfiguration),
		"enable_minimum_encryption_configuration":   aws.ToBool(configuration.EnableMinimumEncryptionConfiguration),
func flattenWorkGroupCustomerContentEncryptionConfiguration(encryptionConfiguration *types.CustomerContentEncryptionConfiguration) []any {
	if encryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrKMSKey: aws.ToString(encryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func flattenWorkGroupEngineVersion(engineVersion *types.EngineVersion) []any {
	if engineVersion == nil {
		return []any{}
	}

	m := map[string]any{
		"effective_engine_version": aws.ToString(engineVersion.EffectiveEngineVersion),
		"selected_engine_version":  aws.ToString(engineVersion.SelectedEngineVersion),
	}

	return []any{m}
}

func flattenWorkGroupIdentityCenterConfiguration(identityCenterConfiguration *types.IdentityCenterConfiguration) []any {
	if identityCenterConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"enable_identity_center":       aws.ToBool(identityCenterConfiguration.EnableIdentityCenter),
		"identity_center_instance_arn": aws.ToString(identityCenterConfiguration.IdentityCenterInstanceArn),
	}

	return []any{m}
}

func flattenWorkGroupResultConfiguration(resultConfiguration *types.ResultConfiguration) []any {
	if resultConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEncryptionConfiguration: flattenWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":                 aws.ToString(resultConfiguration.OutputLocation),
	}

	if resultConfiguration.ExpectedBucketOwner != nil {
		m[names.AttrExpectedBucketOwner] = aws.ToString(resultConfiguration.ExpectedBucketOwner)
	}

	if resultConfiguration.AclConfiguration != nil {
		m["acl_configuration"] = flattenWorkGroupACLConfiguration(resultConfiguration.AclConfiguration)
	}

	return []any{m}
}

func flattenWorkGroupEncryptionConfiguration(encryptionConfiguration *types.EncryptionConfiguration) []any {
	if encryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		"encryption_option": encryptionConfiguration.EncryptionOption,
		names.AttrKMSKeyARN: aws.ToString(encryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func flattenWorkGroupACLConfiguration(aclConfig *types.AclConfiguration) []any {
	if aclConfig == nil {
		return []any{}
	}

	m := map[string]any{
		"s3_acl_option": aclConfig.S3AclOption,
	}

	return []any{m}
}

func flattenWorkGroupManagedQueryResultsConfiguration(managedQueryResultsConfiguration *types.ManagedQueryResultsConfiguration) []any {
	if managedQueryResultsConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled:                 aws.ToBool(&managedQueryResultsConfiguration.Enabled),
		names.AttrEncryptionConfiguration: flattenWorkGroupManagedQueryResultsEncryptionConfiguration(managedQueryResultsConfiguration.EncryptionConfiguration),
	}

	return []any{m}
}

func flattenWorkGroupManagedQueryResultsEncryptionConfiguration(managedQueryResultsEncryptionConfiguration *types.ManagedQueryResultsEncryptionConfiguration) []any {
	if managedQueryResultsEncryptionConfiguration == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrKMSKey: aws.ToString(managedQueryResultsEncryptionConfiguration.KmsKey),
	}

	return []any{m}
}

func managedQueryResultsValidation(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	configRaw, configOk := diff.GetOk(names.AttrConfiguration)
	if !configOk {
		return nil
	}

	configList, ok := configRaw.([]any)
	if !ok || len(configList) == 0 || configList[0] == nil {
		return nil
	}

	config := configList[0].(map[string]any)

	mqrEnabled := false
	if mqrRaw, mqrOk := config["managed_query_results_configuration"]; mqrOk {
		if mqrList, ok := mqrRaw.([]any); ok && len(mqrList) > 0 && mqrList[0] != nil {
			if mqrConfig, ok := mqrList[0].(map[string]any); ok {
				if enabled, ok := mqrConfig[names.AttrEnabled].(bool); ok {
					mqrEnabled = enabled
				}
			}
		}
	}

	if !mqrEnabled {
		return nil
	}

	if rcRaw, rcOk := config["result_configuration"]; rcOk {
		if rcList, ok := rcRaw.([]any); ok && len(rcList) > 0 && rcList[0] != nil {
			if rcConfig, ok := rcList[0].(map[string]any); ok {
				if outputLoc, ok := rcConfig["output_location"].(string); ok && outputLoc != "" {
					return fmt.Errorf("configuration.result_configuration.output_location cannot be specified when configuration.managed_query_results_configuration.enabled is true")
				}
			}
		}
	}

	return nil
}
