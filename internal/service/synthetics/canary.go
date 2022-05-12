package synthetics

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/mitchellh/go-homedir"
)

const awsMutexCanary = `aws_synthetics_canary`

func ResourceCanary() *schema.Resource {
	return &schema.Resource{
		Create: resourceCanaryCreate,
		Read:   resourceCanaryRead,
		Update: resourceCanaryUpdate,
		Delete: resourceCanaryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(synthetics.EncryptionMode_Values(), false),
									},
									"kms_key_arn": {
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
			"engine_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role_arn": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 21),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z_\-]+$`), "must contain only lowercase alphanumeric, hyphen, or underscore."),
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
			"s3_bucket": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{"s3_key"},
			},
			"s3_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{"s3_bucket"},
			},
			"s3_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
			},
			"schedule": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return new == "rate(0 minute)" && old == "rate(0 hour)"
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"success_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"zip_file": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"s3_bucket", "s3_key", "s3_version"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCanaryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &synthetics.CreateCanaryInput{
		ArtifactS3Location: aws.String(d.Get("artifact_s3_location").(string)),
		ExecutionRoleArn:   aws.String(d.Get("execution_role_arn").(string)),
		Name:               aws.String(name),
		RuntimeVersion:     aws.String(d.Get("runtime_version").(string)),
	}

	if code, err := expandCanaryCode(d); err != nil {
		return err
	} else {
		input.Code = code
	}

	if v, ok := d.GetOk("run_config"); ok {
		input.RunConfig = expandCanaryRunConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("artifact_config"); ok {
		input.ArtifactConfig = expandCanaryArtifactConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandCanarySchedule(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandCanaryVPCConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("failure_retention_period"); ok {
		input.FailureRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("success_retention_period"); ok {
		input.SuccessRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Synthetics Canary: %s", input)
	output, err := conn.CreateCanary(input)

	if err != nil {
		return fmt.Errorf("error creating Synthetics Canary (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Canary.Name))

	// Underlying IAM eventual consistency errors can occur after the creation
	// operation. The goal is only retry these types of errors up to the IAM
	// timeout. Since the creation process is asynchronous and can take up to
	// its own timeout, we store a stop time upfront for checking.
	// Real-life experience shows that double the standard IAM propagation time is required.
	propagationTimeout := propagationTimeout * 2
	iamwaiterStopTime := time.Now().Add(propagationTimeout)

	_, err = tfresource.RetryWhen(
		propagationTimeout+canaryCreatedTimeout,
		func() (interface{}, error) {
			return retryCreateCanary(conn, d, input)
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
		return fmt.Errorf("error waiting for Synthetics Canary (%s) create: %w", d.Id(), err)
	}

	if d.Get("start_canary").(bool) {
		if err := startCanary(d.Id(), conn); err != nil {
			return err
		}
	}

	return resourceCanaryRead(d, meta)
}

func resourceCanaryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	canary, err := FindCanaryByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Synthetics Canary (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Synthetics Canary (%s): %w", d.Id(), err)
	}

	canaryArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   synthetics.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("canary:%s", aws.StringValue(canary.Name)),
	}.String()
	d.Set("arn", canaryArn)
	d.Set("artifact_s3_location", canary.ArtifactS3Location)
	d.Set("engine_arn", canary.EngineArn)
	d.Set("execution_role_arn", canary.ExecutionRoleArn)
	d.Set("failure_retention_period", canary.FailureRetentionPeriodInDays)
	d.Set("handler", canary.Code.Handler)
	d.Set("name", canary.Name)
	d.Set("runtime_version", canary.RuntimeVersion)
	d.Set("source_location_arn", canary.Code.SourceLocationArn)
	d.Set("status", canary.Status.State)
	d.Set("success_retention_period", canary.SuccessRetentionPeriodInDays)

	if err := d.Set("vpc_config", flattenCanaryVPCConfig(canary.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc config: %w", err)
	}

	runConfig := &synthetics.CanaryRunConfigInput{}
	if v, ok := d.GetOk("run_config"); ok {
		runConfig = expandCanaryRunConfig(v.([]interface{}))
	}

	if err := d.Set("run_config", flattenCanaryRunConfig(canary.RunConfig, runConfig.EnvironmentVariables)); err != nil {
		return fmt.Errorf("error setting run config: %w", err)
	}

	if err := d.Set("schedule", flattenCanarySchedule(canary.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	if err := d.Set("timeline", flattenCanaryTimeline(canary.Timeline)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	if err := d.Set("artifact_config", flattenCanaryArtifactConfig(canary.ArtifactConfig)); err != nil {
		return fmt.Errorf("error setting artifact_config: %w", err)
	}

	tags := KeyValueTags(canary.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCanaryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn

	if d.HasChangesExcept("tags", "tags_all", "start_canary") {
		input := &synthetics.UpdateCanaryInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("vpc_config") {
			input.VpcConfig = expandCanaryVPCConfig(d.Get("vpc_config").([]interface{}))
		}

		if d.HasChange("artifact_config") {
			input.ArtifactConfig = expandCanaryArtifactConfig(d.Get("artifact_config").([]interface{}))
		}

		if d.HasChange("runtime_version") {
			input.RuntimeVersion = aws.String(d.Get("runtime_version").(string))
		}

		if d.HasChanges("handler", "zip_file", "s3_bucket", "s3_key", "s3_version") {
			if code, err := expandCanaryCode(d); err != nil {
				return err
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

		if d.HasChange("schedule") {
			input.Schedule = expandCanarySchedule(d.Get("schedule").([]interface{}))
		}

		if d.HasChange("success_retention_period") {
			_, n := d.GetChange("success_retention_period")
			input.SuccessRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		}

		if d.HasChange("failure_retention_period") {
			_, n := d.GetChange("failure_retention_period")
			input.FailureRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		}

		if d.HasChange("execution_role_arn") {
			_, n := d.GetChange("execution_role_arn")
			input.ExecutionRoleArn = aws.String(n.(string))
		}

		status := d.Get("status").(string)
		if status == synthetics.CanaryStateRunning {
			if err := stopCanary(d.Id(), conn); err != nil {
				return err
			}
		}

		log.Printf("[DEBUG] Updating Synthetics Canary: %s", input)
		_, err := conn.UpdateCanary(input)

		if err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s): %w", d.Id(), err)
		}

		if status != synthetics.CanaryStateReady {
			if _, err := waitCanaryStopped(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) stop: %w", d.Id(), err)
			}
		} else {
			if _, err := waitCanaryReady(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) ready: %w", d.Id(), err)
			}
		}

		if d.Get("start_canary").(bool) {
			if err := startCanary(d.Id(), conn); err != nil {
				return err
			}
		}
	}

	if d.HasChange("start_canary") {
		status := d.Get("status").(string)
		if d.Get("start_canary").(bool) {
			if status != synthetics.CanaryStateRunning {
				if err := startCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		} else {
			if status == synthetics.CanaryStateRunning {
				if err := stopCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceCanaryRead(d, meta)
}

func resourceCanaryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn

	if status := d.Get("status").(string); status == synthetics.CanaryStateRunning {
		if err := stopCanary(d.Id(), conn); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Synthetics Canary: (%s)", d.Id())
	_, err := conn.DeleteCanary(&synthetics.DeleteCanaryInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Synthetics Canary (%s): %w", d.Id(), err)
	}

	_, err = waitCanaryDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandCanaryCode(d *schema.ResourceData) (*synthetics.CanaryCodeInput, error) {
	codeConfig := &synthetics.CanaryCodeInput{
		Handler: aws.String(d.Get("handler").(string)),
	}

	if v, ok := d.GetOk("zip_file"); ok {
		conns.GlobalMutexKV.Lock(awsMutexCanary)
		defer conns.GlobalMutexKV.Unlock(awsMutexCanary)
		file, err := loadFileContent(v.(string))
		if err != nil {
			return nil, fmt.Errorf("unable to load %q: %w", v.(string), err)
		}
		codeConfig.ZipFile = file
	} else {
		codeConfig.S3Bucket = aws.String(d.Get("s3_bucket").(string))
		codeConfig.S3Key = aws.String(d.Get("s3_key").(string))

		if v, ok := d.GetOk("s3_version"); ok {
			codeConfig.S3Version = aws.String(v.(string))
		}
	}

	return codeConfig, nil
}

func expandCanaryArtifactConfig(l []interface{}) *synthetics.ArtifactConfigInput_ {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &synthetics.ArtifactConfigInput_{}

	if v, ok := m["s3_encryption"].([]interface{}); ok && len(v) > 0 {
		config.S3Encryption = expandCanaryS3EncryptionConfig(v)
	}

	return config
}

func flattenCanaryArtifactConfig(config *synthetics.ArtifactConfigOutput_) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if config.S3Encryption != nil {
		m["s3_encryption"] = flattenCanaryS3EncryptionConfig(config.S3Encryption)
	}

	return []interface{}{m}
}

func expandCanaryS3EncryptionConfig(l []interface{}) *synthetics.S3EncryptionConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &synthetics.S3EncryptionConfig{}

	if v, ok := m["encryption_mode"].(string); ok && v != "" {
		config.EncryptionMode = aws.String(v)
	}

	if v, ok := m["kms_key_arn"].(string); ok && v != "" {
		config.KmsKeyArn = aws.String(v)
	}

	return config
}

func flattenCanaryS3EncryptionConfig(config *synthetics.S3EncryptionConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if config.EncryptionMode != nil {
		m["encryption_mode"] = aws.StringValue(config.EncryptionMode)
	}

	if config.KmsKeyArn != nil {
		m["kms_key_arn"] = aws.StringValue(config.KmsKeyArn)
	}

	return []interface{}{m}
}

func expandCanarySchedule(l []interface{}) *synthetics.CanaryScheduleInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.CanaryScheduleInput{
		Expression: aws.String(m["expression"].(string)),
	}

	if v, ok := m["duration_in_seconds"]; ok {
		codeConfig.DurationInSeconds = aws.Int64(int64(v.(int)))
	}

	return codeConfig
}

func flattenCanarySchedule(canarySchedule *synthetics.CanaryScheduleOutput) []interface{} {
	if canarySchedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"expression":          aws.StringValue(canarySchedule.Expression),
		"duration_in_seconds": aws.Int64Value(canarySchedule.DurationInSeconds),
	}

	return []interface{}{m}
}

func expandCanaryRunConfig(l []interface{}) *synthetics.CanaryRunConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.CanaryRunConfigInput{
		TimeoutInSeconds: aws.Int64(int64(m["timeout_in_seconds"].(int))),
	}

	if v, ok := m["memory_in_mb"].(int); ok && v > 0 {
		codeConfig.MemoryInMB = aws.Int64(int64(v))
	}

	if v, ok := m["active_tracing"].(bool); ok {
		codeConfig.ActiveTracing = aws.Bool(v)
	}

	if vars, ok := m["environment_variables"].(map[string]interface{}); ok && len(vars) > 0 {
		codeConfig.EnvironmentVariables = flex.ExpandStringMap(vars)
	}

	return codeConfig
}

func flattenCanaryRunConfig(canaryCodeOut *synthetics.CanaryRunConfigOutput, envVars map[string]*string) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"timeout_in_seconds": aws.Int64Value(canaryCodeOut.TimeoutInSeconds),
		"memory_in_mb":       aws.Int64Value(canaryCodeOut.MemoryInMB),
		"active_tracing":     aws.BoolValue(canaryCodeOut.ActiveTracing),
	}

	if envVars != nil {
		m["environment_variables"] = aws.StringValueMap(envVars)
	}

	return []interface{}{m}
}

func flattenCanaryVPCConfig(canaryVpcOutput *synthetics.VpcConfigOutput) []interface{} {
	if canaryVpcOutput == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"subnet_ids":         flex.FlattenStringSet(canaryVpcOutput.SubnetIds),
		"security_group_ids": flex.FlattenStringSet(canaryVpcOutput.SecurityGroupIds),
		"vpc_id":             aws.StringValue(canaryVpcOutput.VpcId),
	}

	return []interface{}{m}
}

func expandCanaryVPCConfig(l []interface{}) *synthetics.VpcConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.VpcConfigInput{
		SubnetIds:        flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
	}

	return codeConfig
}

func flattenCanaryTimeline(timeline *synthetics.CanaryTimeline) []interface{} {
	if timeline == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"created": aws.TimeValue(timeline.Created).Format(time.RFC3339),
	}

	if timeline.LastModified != nil {
		m["last_modified"] = aws.TimeValue(timeline.LastModified).Format(time.RFC3339)
	}

	if timeline.LastStarted != nil {
		m["last_started"] = aws.TimeValue(timeline.LastStarted).Format(time.RFC3339)
	}

	if timeline.LastStopped != nil {
		m["last_stopped"] = aws.TimeValue(timeline.LastStopped).Format(time.RFC3339)
	}

	return []interface{}{m}
}

func startCanary(name string, conn *synthetics.Synthetics) error {
	log.Printf("[DEBUG] Starting Synthetics Canary: (%s)", name)
	_, err := conn.StartCanary(&synthetics.StartCanaryInput{
		Name: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error starting Synthetics Canary (%s): %w", name, err)
	}

	_, err = waitCanaryRunning(conn, name)

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) start: %w", name, err)
	}

	return nil
}

func stopCanary(name string, conn *synthetics.Synthetics) error {
	log.Printf("[DEBUG] Stopping Synthetics Canary: (%s)", name)
	_, err := conn.StopCanary(&synthetics.StopCanaryInput{
		Name: aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeConflictException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error stopping Synthetics Canary (%s): %w", name, err)
	}

	_, err = waitCanaryStopped(conn, name)

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) stop: %w", name, err)
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
