// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &athena.CreateWorkGroupInput{
		Configuration: expandWorkGroupConfiguration(d.Get(names.AttrConfiguration).([]interface{})),
		Name:          aws.String(name),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateWorkGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena WorkGroup (%s): %s", name, err)
	}

	d.SetId(name)

	if v := types.WorkGroupState(d.Get(names.AttrState).(string)); v == types.WorkGroupStateDisabled {
		input := &athena.UpdateWorkGroupInput{
			State:     v,
			WorkGroup: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	wg, err := findWorkGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena WorkGroup (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "athena",
		AccountID: meta.(*conns.AWSClient).AccountID,
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

func resourceWorkGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &athena.UpdateWorkGroupInput{
			WorkGroup: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange(names.AttrConfiguration) {
			input.ConfigurationUpdates = expandWorkGroupConfigurationUpdates(d.Get(names.AttrConfiguration).([]interface{}))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrState) {
			input.State = types.WorkGroupState(d.Get(names.AttrState).(string))
		}

		_, err := conn.UpdateWorkGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkGroupRead(ctx, d, meta)...)
}

func resourceWorkGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	input := &athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrForceDestroy); ok {
		input.RecursiveDeleteOption = aws.Bool(v.(bool))
	}

	_, err := conn.DeleteWorkGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkGroupByName(ctx context.Context, conn *athena.Client, name string) (*types.WorkGroup, error) {
	input := &athena.GetWorkGroupInput{
		WorkGroup: aws.String(name),
	}

	output, err := conn.GetWorkGroup(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "is not found") {
		return nil, &retry.NotFoundError{
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

func expandWorkGroupConfiguration(l []interface{}) *types.WorkGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &types.WorkGroupConfiguration{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configuration.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configuration.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		configuration.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configuration.ExecutionRole = aws.String(v)
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandWorkGroupResultConfiguration(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"].(bool); ok {
		configuration.RequesterPaysEnabled = aws.Bool(v)
	}

	return configuration
}

func expandWorkGroupEngineVersion(l []interface{}) *types.EngineVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	engineVersion := &types.EngineVersion{}

	if v, ok := m["selected_engine_version"].(string); ok && v != "" {
		engineVersion.SelectedEngineVersion = aws.String(v)
	}

	return engineVersion
}

func expandWorkGroupConfigurationUpdates(l []interface{}) *types.WorkGroupConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configurationUpdates := &types.WorkGroupConfigurationUpdates{}

	if v, ok := m["bytes_scanned_cutoff_per_query"].(int); ok && v > 0 {
		configurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v))
	} else {
		configurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
	}

	if v, ok := m["enforce_workgroup_configuration"].(bool); ok {
		configurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v)
	}

	if v, ok := m[names.AttrEngineVersion].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		configurationUpdates.EngineVersion = expandWorkGroupEngineVersion(v)
	}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		configurationUpdates.ExecutionRole = aws.String(v)
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"].(bool); ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v)
	}

	if v, ok := m["result_configuration"]; ok {
		configurationUpdates.ResultConfigurationUpdates = expandWorkGroupResultConfigurationUpdates(v.([]interface{}))
	}

	if v, ok := m["requester_pays_enabled"].(bool); ok {
		configurationUpdates.RequesterPaysEnabled = aws.Bool(v)
	}

	return configurationUpdates
}

func expandWorkGroupResultConfiguration(l []interface{}) *types.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfiguration := &types.ResultConfiguration{}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		resultConfiguration.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := m["output_location"].(string); ok && v != "" {
		resultConfiguration.OutputLocation = aws.String(v)
	}

	if v, ok := m[names.AttrExpectedBucketOwner].(string); ok && v != "" {
		resultConfiguration.ExpectedBucketOwner = aws.String(v)
	}

	if v, ok := m["acl_configuration"]; ok {
		resultConfiguration.AclConfiguration = expandResultConfigurationACLConfig(v.([]interface{}))
	}

	return resultConfiguration
}

func expandWorkGroupResultConfigurationUpdates(l []interface{}) *types.ResultConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfigurationUpdates := &types.ResultConfigurationUpdates{}

	if v, ok := m[names.AttrEncryptionConfiguration]; ok {
		resultConfigurationUpdates.EncryptionConfiguration = expandWorkGroupEncryptionConfiguration(v.([]interface{}))
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
		resultConfigurationUpdates.AclConfiguration = expandResultConfigurationACLConfig(v.([]interface{}))
	} else {
		resultConfigurationUpdates.RemoveAclConfiguration = aws.Bool(true)
	}

	return resultConfigurationUpdates
}

func expandWorkGroupEncryptionConfiguration(l []interface{}) *types.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	encryptionConfiguration := &types.EncryptionConfiguration{}

	if v, ok := m["encryption_option"]; ok && v.(string) != "" {
		encryptionConfiguration.EncryptionOption = types.EncryptionOption(v.(string))
	}

	if v, ok := m[names.AttrKMSKeyARN]; ok && v.(string) != "" {
		encryptionConfiguration.KmsKey = aws.String(v.(string))
	}

	return encryptionConfiguration
}

func flattenWorkGroupConfiguration(configuration *types.WorkGroupConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bytes_scanned_cutoff_per_query":     aws.ToInt64(configuration.BytesScannedCutoffPerQuery),
		"enforce_workgroup_configuration":    aws.ToBool(configuration.EnforceWorkGroupConfiguration),
		names.AttrEngineVersion:              flattenWorkGroupEngineVersion(configuration.EngineVersion),
		"execution_role":                     aws.ToString(configuration.ExecutionRole),
		"publish_cloudwatch_metrics_enabled": aws.ToBool(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":               flattenWorkGroupResultConfiguration(configuration.ResultConfiguration),
		"requester_pays_enabled":             aws.ToBool(configuration.RequesterPaysEnabled),
	}

	return []interface{}{m}
}

func flattenWorkGroupEngineVersion(engineVersion *types.EngineVersion) []interface{} {
	if engineVersion == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"effective_engine_version": aws.ToString(engineVersion.EffectiveEngineVersion),
		"selected_engine_version":  aws.ToString(engineVersion.SelectedEngineVersion),
	}

	return []interface{}{m}
}

func flattenWorkGroupResultConfiguration(resultConfiguration *types.ResultConfiguration) []interface{} {
	if resultConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrEncryptionConfiguration: flattenWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":                 aws.ToString(resultConfiguration.OutputLocation),
	}

	if resultConfiguration.ExpectedBucketOwner != nil {
		m[names.AttrExpectedBucketOwner] = aws.ToString(resultConfiguration.ExpectedBucketOwner)
	}

	if resultConfiguration.AclConfiguration != nil {
		m["acl_configuration"] = flattenWorkGroupACLConfiguration(resultConfiguration.AclConfiguration)
	}

	return []interface{}{m}
}

func flattenWorkGroupEncryptionConfiguration(encryptionConfiguration *types.EncryptionConfiguration) []interface{} {
	if encryptionConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_option": encryptionConfiguration.EncryptionOption,
		names.AttrKMSKeyARN: aws.ToString(encryptionConfiguration.KmsKey),
	}

	return []interface{}{m}
}

func flattenWorkGroupACLConfiguration(aclConfig *types.AclConfiguration) []interface{} {
	if aclConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"s3_acl_option": aclConfig.S3AclOption,
	}

	return []interface{}{m}
}
