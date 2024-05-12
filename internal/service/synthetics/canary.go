// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"github.com/mitchellh/go-homedir"
)

const canaryMutex = `aws_synthetics_canary`

// @SDKResource("aws_synthetics_canary", name="Canary")
// @Tags(identifierAttribute="arn")
func ResourceCanary() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCanaryCreate,
		ReadWithoutTimeout:   resourceCanaryRead,
		UpdateWithoutTimeout: resourceCanaryUpdate,
		DeleteWithoutTimeout: resourceCanaryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifact_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_encryption": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_mode": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.EncryptionMode](),
									},
									names.AttrKMSKeyARN: {
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
			"artifact_s3_location": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimPrefix(new, "s3://") == old
				},
			},
			"delete_lambda": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"engine_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrExecutionRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"failure_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"handler": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 21),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z_\-]+$`), "must contain only lowercase alphanumeric, hyphen, or underscore."),
				),
			},
			"run_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_tracing": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"environment_variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"memory_in_mb": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.All(
								validation.IntDivisibleBy(64),
								validation.IntAtLeast(960),
							),
						},
						"timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(3, 14*60),
							Default:      840,
						},
					},
				},
			},
			"runtime_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrS3Bucket: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{"s3_key"},
			},
			"s3_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{names.AttrS3Bucket},
			},
			"s3_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
			},
			names.AttrSchedule: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrExpression: {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return (new == "rate(0 minute)" || new == "rate(0 minutes)") && old == "rate(0 hour)"
							},
						},
					},
				},
			},
			"source_location_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_canary": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"success_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timeline": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_modified": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_started": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_stopped": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"zip_file": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrS3Bucket, "s3_key", "s3_version"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCanaryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &synthetics.CreateCanaryInput{
		ArtifactS3Location: aws.String(d.Get("artifact_s3_location").(string)),
		ExecutionRoleArn:   aws.String(d.Get(names.AttrExecutionRoleARN).(string)),
		Name:               aws.String(name),
		RuntimeVersion:     aws.String(d.Get("runtime_version").(string)),
		Tags:               getTagsIn(ctx),
	}

	if code, err := expandCanaryCode(d); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Synthetics Canary (%s): %s", name, err)
	} else {
		input.Code = code
	}

	if v, ok := d.GetOk("run_config"); ok {
		input.RunConfig = expandCanaryRunConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("artifact_config"); ok {
		input.ArtifactConfig = expandCanaryArtifactConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrSchedule); ok {
		input.Schedule = expandCanarySchedule(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok {
		input.VpcConfig = expandCanaryVPCConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("failure_retention_period"); ok {
		input.FailureRetentionPeriodInDays = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("success_retention_period"); ok {
		input.SuccessRetentionPeriodInDays = aws.Int32(int32(v.(int)))
	}

	output, err := conn.CreateCanary(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Synthetics Canary (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Canary.Name))

	// Underlying IAM eventual consistency errors can occur after the creation
	// operation. The goal is only retry these types of errors up to the IAM
	// timeout. Since the creation process is asynchronous and can take up to
	// its own timeout, we store a stop time upfront for checking.
	// Real-life experience shows that double the standard IAM propagation time is required.
	propagationTimeout := propagationTimeout * 2
	iamwaiterStopTime := time.Now().Add(propagationTimeout)

	_, err = tfresource.RetryWhen(ctx, propagationTimeout+canaryCreatedTimeout,
		func() (interface{}, error) {
			return retryCreateCanary(ctx, conn, d, input)
		},
		func(err error) (bool, error) {
			// Only retry IAM eventual consistency errors up to that timeout.
			if err != nil && time.Now().Before(iamwaiterStopTime) {
				// This error synthesized from the Status object and not an AWS SDK Go error type.
				return strings.Contains(err.Error(), "The role defined for the function cannot be assumed by Lambda"), err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Synthetics Canary (%s): waiting for completion: %s", name, err)
	}

	if d.Get("start_canary").(bool) {
		if err := startCanary(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Synthetics Canary (%s): %s", name, err)
		}
	}

	return append(diags, resourceCanaryRead(ctx, d, meta)...)
}

func resourceCanaryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	canary, err := FindCanaryByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Synthetics Canary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Synthetics Canary (%s): %s", d.Id(), err)
	}

	canaryArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "synthetics",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("canary:%s", aws.ToString(canary.Name)),
	}.String()
	d.Set(names.AttrARN, canaryArn)
	d.Set("artifact_s3_location", canary.ArtifactS3Location)
	d.Set("engine_arn", canary.EngineArn)
	d.Set(names.AttrExecutionRoleARN, canary.ExecutionRoleArn)
	d.Set("failure_retention_period", canary.FailureRetentionPeriodInDays)
	d.Set("handler", canary.Code.Handler)
	d.Set(names.AttrName, canary.Name)
	d.Set("runtime_version", canary.RuntimeVersion)
	d.Set("source_location_arn", canary.Code.SourceLocationArn)
	d.Set(names.AttrStatus, canary.Status.State)
	d.Set("success_retention_period", canary.SuccessRetentionPeriodInDays)

	if err := d.Set(names.AttrVPCConfig, flattenCanaryVPCConfig(canary.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc config: %s", err)
	}

	runConfig := &awstypes.CanaryRunConfigInput{}
	if v, ok := d.GetOk("run_config"); ok {
		runConfig = expandCanaryRunConfig(v.([]interface{}))
	}

	if err := d.Set("run_config", flattenCanaryRunConfig(canary.RunConfig, runConfig.EnvironmentVariables)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting run config: %s", err)
	}

	if err := d.Set(names.AttrSchedule, flattenCanarySchedule(canary.Schedule)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
	}

	if err := d.Set("timeline", flattenCanaryTimeline(canary.Timeline)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
	}

	if err := d.Set("artifact_config", flattenCanaryArtifactConfig(canary.ArtifactConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting artifact_config: %s", err)
	}

	setTagsOut(ctx, canary.Tags)

	return diags
}

func resourceCanaryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "start_canary") {
		input := &synthetics.UpdateCanaryInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrVPCConfig) {
			input.VpcConfig = expandCanaryVPCConfig(d.Get(names.AttrVPCConfig).([]interface{}))
		}

		if d.HasChange("artifact_config") {
			input.ArtifactConfig = expandCanaryArtifactConfig(d.Get("artifact_config").([]interface{}))
		}

		if d.HasChange("runtime_version") {
			input.RuntimeVersion = aws.String(d.Get("runtime_version").(string))
		}

		if d.HasChanges("handler", "zip_file", names.AttrS3Bucket, "s3_key", "s3_version") {
			if code, err := expandCanaryCode(d); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
			} else {
				input.Code = code
			}
		}

		if d.HasChange("run_config") {
			input.RunConfig = expandCanaryRunConfig(d.Get("run_config").([]interface{}))
		}

		if d.HasChange("artifact_s3_location") {
			input.ArtifactS3Location = aws.String(d.Get("artifact_s3_location").(string))
		}

		if d.HasChange(names.AttrSchedule) {
			input.Schedule = expandCanarySchedule(d.Get(names.AttrSchedule).([]interface{}))
		}

		if d.HasChange("success_retention_period") {
			_, n := d.GetChange("success_retention_period")
			input.SuccessRetentionPeriodInDays = aws.Int32(int32(n.(int)))
		}

		if d.HasChange("failure_retention_period") {
			_, n := d.GetChange("failure_retention_period")
			input.FailureRetentionPeriodInDays = aws.Int32(int32(n.(int)))
		}

		if d.HasChange(names.AttrExecutionRoleARN) {
			_, n := d.GetChange(names.AttrExecutionRoleARN)
			input.ExecutionRoleArn = aws.String(n.(string))
		}

		status := d.Get(names.AttrStatus).(string)
		if status == string(awstypes.CanaryStateRunning) {
			if err := stopCanary(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
			}
		}

		_, err := conn.UpdateCanary(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
		}

		if status != string(awstypes.CanaryStateReady) {
			if _, err := waitCanaryStopped(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): waiting for Canary to stop: %s", d.Id(), err)
			}
		} else {
			if _, err := waitCanaryReady(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): waiting for Canary to be ready: %s", d.Id(), err)
			}
		}

		if d.Get("start_canary").(bool) {
			if err := startCanary(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("start_canary") {
		status := d.Get(names.AttrStatus).(string)
		if d.Get("start_canary").(bool) {
			if status != string(awstypes.CanaryStateRunning) {
				if err := startCanary(ctx, d.Id(), conn); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
				}
			}
		} else {
			if status == string(awstypes.CanaryStateRunning) {
				if err := stopCanary(ctx, d.Id(), conn); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Synthetics Canary (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceCanaryRead(ctx, d, meta)...)
}

func resourceCanaryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SyntheticsClient(ctx)

	if status := d.Get(names.AttrStatus).(string); status == string(awstypes.CanaryStateRunning) {
		if err := stopCanary(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Synthetics Canary (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Synthetics Canary: (%s)", d.Id())
	_, err := conn.DeleteCanary(ctx, &synthetics.DeleteCanaryInput{
		Name:         aws.String(d.Id()),
		DeleteLambda: d.Get("delete_lambda").(bool),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Synthetics Canary (%s): %s", d.Id(), err)
	}

	_, err = waitCanaryDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Synthetics Canary (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func expandCanaryCode(d *schema.ResourceData) (*awstypes.CanaryCodeInput, error) {
	codeConfig := &awstypes.CanaryCodeInput{
		Handler: aws.String(d.Get("handler").(string)),
	}

	if v, ok := d.GetOk("zip_file"); ok {
		conns.GlobalMutexKV.Lock(canaryMutex)
		defer conns.GlobalMutexKV.Unlock(canaryMutex)
		file, err := loadFileContent(v.(string))
		if err != nil {
			return nil, fmt.Errorf("unable to load %q: %w", v.(string), err)
		}
		codeConfig.ZipFile = file
	} else {
		codeConfig.S3Bucket = aws.String(d.Get(names.AttrS3Bucket).(string))
		codeConfig.S3Key = aws.String(d.Get("s3_key").(string))

		if v, ok := d.GetOk("s3_version"); ok {
			codeConfig.S3Version = aws.String(v.(string))
		}
	}

	return codeConfig, nil
}

func expandCanaryArtifactConfig(l []interface{}) *awstypes.ArtifactConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.ArtifactConfigInput{}

	if v, ok := m["s3_encryption"].([]interface{}); ok && len(v) > 0 {
		config.S3Encryption = expandCanaryS3EncryptionConfig(v)
	}

	return config
}

func flattenCanaryArtifactConfig(config *awstypes.ArtifactConfigOutput) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if config.S3Encryption != nil {
		m["s3_encryption"] = flattenCanaryS3EncryptionConfig(config.S3Encryption)
	}

	return []interface{}{m}
}

func expandCanaryS3EncryptionConfig(l []interface{}) *awstypes.S3EncryptionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.S3EncryptionConfig{}

	if v, ok := m["encryption_mode"].(string); ok && v != "" {
		config.EncryptionMode = awstypes.EncryptionMode(v)
	}

	if v, ok := m[names.AttrKMSKeyARN].(string); ok && v != "" {
		config.KmsKeyArn = aws.String(v)
	}

	return config
}

func flattenCanaryS3EncryptionConfig(config *awstypes.S3EncryptionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if config.EncryptionMode != "" {
		m["encryption_mode"] = string(config.EncryptionMode)
	}

	if config.KmsKeyArn != nil {
		m[names.AttrKMSKeyARN] = aws.ToString(config.KmsKeyArn)
	}

	return []interface{}{m}
}

func expandCanarySchedule(l []interface{}) *awstypes.CanaryScheduleInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &awstypes.CanaryScheduleInput{
		Expression: aws.String(m[names.AttrExpression].(string)),
	}

	if v, ok := m["duration_in_seconds"]; ok {
		codeConfig.DurationInSeconds = aws.Int64(int64(v.(int)))
	}

	return codeConfig
}

func flattenCanarySchedule(canarySchedule *awstypes.CanaryScheduleOutput) []interface{} {
	if canarySchedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrExpression:  aws.ToString(canarySchedule.Expression),
		"duration_in_seconds": aws.ToInt64(canarySchedule.DurationInSeconds),
	}

	return []interface{}{m}
}

func expandCanaryRunConfig(l []interface{}) *awstypes.CanaryRunConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &awstypes.CanaryRunConfigInput{
		TimeoutInSeconds: aws.Int32(int32(m["timeout_in_seconds"].(int))),
	}

	if v, ok := m["memory_in_mb"].(int); ok && v > 0 {
		codeConfig.MemoryInMB = aws.Int32(int32(v))
	}

	if v, ok := m["active_tracing"].(bool); ok {
		codeConfig.ActiveTracing = aws.Bool(v)
	}

	if vars, ok := m["environment_variables"].(map[string]interface{}); ok && len(vars) > 0 {
		codeConfig.EnvironmentVariables = flex.ExpandStringValueMap(vars)
	}

	return codeConfig
}

func flattenCanaryRunConfig(canaryCodeOut *awstypes.CanaryRunConfigOutput, envVars map[string]string) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"timeout_in_seconds": aws.ToInt32(canaryCodeOut.TimeoutInSeconds),
		"memory_in_mb":       aws.ToInt32(canaryCodeOut.MemoryInMB),
		"active_tracing":     aws.ToBool(canaryCodeOut.ActiveTracing),
	}

	if envVars != nil {
		m["environment_variables"] = envVars
	}

	return []interface{}{m}
}

func flattenCanaryVPCConfig(canaryVpcOutput *awstypes.VpcConfigOutput) []interface{} {
	if canaryVpcOutput == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrSubnetIDs:        flex.FlattenStringValueSet(canaryVpcOutput.SubnetIds),
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(canaryVpcOutput.SecurityGroupIds),
		names.AttrVPCID:            aws.ToString(canaryVpcOutput.VpcId),
	}

	return []interface{}{m}
}

func expandCanaryVPCConfig(l []interface{}) *awstypes.VpcConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &awstypes.VpcConfigInput{
		SubnetIds:        flex.ExpandStringValueSet(m[names.AttrSubnetIDs].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
	}

	return codeConfig
}

func flattenCanaryTimeline(timeline *awstypes.CanaryTimeline) []interface{} {
	if timeline == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"created": aws.ToTime(timeline.Created).Format(time.RFC3339),
	}

	if timeline.LastModified != nil {
		m["last_modified"] = aws.ToTime(timeline.LastModified).Format(time.RFC3339)
	}

	if timeline.LastStarted != nil {
		m["last_started"] = aws.ToTime(timeline.LastStarted).Format(time.RFC3339)
	}

	if timeline.LastStopped != nil {
		m["last_stopped"] = aws.ToTime(timeline.LastStopped).Format(time.RFC3339)
	}

	return []interface{}{m}
}

func startCanary(ctx context.Context, name string, conn *synthetics.Client) error {
	log.Printf("[DEBUG] Starting Synthetics Canary: (%s)", name)
	_, err := conn.StartCanary(ctx, &synthetics.StartCanaryInput{
		Name: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("starting Synthetics Canary: %w", err)
	}

	_, err = waitCanaryRunning(ctx, conn, name)

	if err != nil {
		return fmt.Errorf("starting Synthetics Canary: waiting for completion: %w", err)
	}

	return nil
}

func stopCanary(ctx context.Context, name string, conn *synthetics.Client) error {
	log.Printf("[DEBUG] Stopping Synthetics Canary: (%s)", name)
	_, err := conn.StopCanary(ctx, &synthetics.StopCanaryInput{
		Name: aws.String(name),
	})

	if errs.IsA[*awstypes.ConflictException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("stopping Synthetics Canary: %w", err)
	}

	_, err = waitCanaryStopped(ctx, conn, name)

	if err != nil {
		return fmt.Errorf("stopping Synthetics Canary: waiting for completion: %w", err)
	}

	return nil
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
