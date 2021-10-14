package aws

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
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	iamwaiter "github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/synthetics/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/synthetics/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/mitchellh/go-homedir"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"artifact_s3_location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	code, err := expandAwsSyntheticsCanaryCode(d)

	if err != nil {
		return err
	}

	input.Code = code

	if v, ok := d.GetOk("run_config"); ok {
		input.RunConfig = expandAwsSyntheticsCanaryRunConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandAwsSyntheticsCanarySchedule(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandAwsSyntheticsCanaryVpcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("failure_retention_period"); ok {
		input.FailureRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("success_retention_period"); ok {
		input.SuccessRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SyntheticsTags()
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
	iamwaiterStopTime := time.Now().Add(iamwaiter.PropagationTimeout)

	_, err = tfresource.RetryWhen(
		iamwaiter.PropagationTimeout+waiter.CanaryCreatedTimeout,
		func() (interface{}, error) {
			return waiter.CanaryReady(conn, d.Id())
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
		if err := syntheticsStartCanary(d.Id(), conn); err != nil {
			return err
		}
	}

	return resourceCanaryRead(d, meta)
}

func resourceCanaryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	canary, err := finder.CanaryByName(conn, d.Id())

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

	if err := d.Set("vpc_config", flattenAwsSyntheticsCanaryVpcConfig(canary.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc config: %w", err)
	}

	if err := d.Set("run_config", flattenAwsSyntheticsCanaryRunConfig(canary.RunConfig)); err != nil {
		return fmt.Errorf("error setting run config: %w", err)
	}

	if err := d.Set("schedule", flattenAwsSyntheticsCanarySchedule(canary.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	if err := d.Set("timeline", flattenAwsSyntheticsCanaryTimeline(canary.Timeline)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	tags := tftags.SyntheticsKeyValueTags(canary.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

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
			input.VpcConfig = expandAwsSyntheticsCanaryVpcConfig(d.Get("vpc_config").([]interface{}))
		}

		if d.HasChange("runtime_version") {
			input.RuntimeVersion = aws.String(d.Get("runtime_version").(string))
		}

		if d.HasChanges("handler", "zip_file", "s3_bucket", "s3_key", "s3_version") {
			code, err := expandAwsSyntheticsCanaryCode(d)
			if err != nil {
				return err
			}
			input.Code = code
		}

		if d.HasChange("run_config") {
			input.RunConfig = expandAwsSyntheticsCanaryRunConfig(d.Get("run_config").([]interface{}))
		}

		if d.HasChange("schedule") {
			input.Schedule = expandAwsSyntheticsCanarySchedule(d.Get("schedule").([]interface{}))
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
			if err := syntheticsStopCanary(d.Id(), conn); err != nil {
				return err
			}
		}

		log.Printf("[DEBUG] Updating Synthetics Canary: %s", input)
		_, err := conn.UpdateCanary(input)

		if err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s): %w", d.Id(), err)
		}

		if status != synthetics.CanaryStateReady {
			if _, err := waiter.CanaryStopped(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) stop: %w", d.Id(), err)
			}
		} else {
			if _, err := waiter.CanaryReady(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) ready: %w", d.Id(), err)
			}
		}

		if d.Get("start_canary").(bool) {
			if err := syntheticsStartCanary(d.Id(), conn); err != nil {
				return err
			}
		}
	}

	if d.HasChange("start_canary") {
		status := d.Get("status").(string)
		if d.Get("start_canary").(bool) {
			if status != synthetics.CanaryStateRunning {
				if err := syntheticsStartCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		} else {
			if status == synthetics.CanaryStateRunning {
				if err := syntheticsStopCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := tftags.SyntheticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceCanaryRead(d, meta)
}

func resourceCanaryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SyntheticsConn

	if status := d.Get("status").(string); status == synthetics.CanaryStateRunning {
		if err := syntheticsStopCanary(d.Id(), conn); err != nil {
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

	_, err = waiter.CanaryDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandAwsSyntheticsCanaryCode(d *schema.ResourceData) (*synthetics.CanaryCodeInput, error) {
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

func expandAwsSyntheticsCanarySchedule(l []interface{}) *synthetics.CanaryScheduleInput {
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

func flattenAwsSyntheticsCanarySchedule(canarySchedule *synthetics.CanaryScheduleOutput) []interface{} {
	if canarySchedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"expression":          aws.StringValue(canarySchedule.Expression),
		"duration_in_seconds": aws.Int64Value(canarySchedule.DurationInSeconds),
	}

	return []interface{}{m}
}

func expandAwsSyntheticsCanaryRunConfig(l []interface{}) *synthetics.CanaryRunConfigInput {
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

	return codeConfig
}

func flattenAwsSyntheticsCanaryRunConfig(canaryCodeOut *synthetics.CanaryRunConfigOutput) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"timeout_in_seconds": aws.Int64Value(canaryCodeOut.TimeoutInSeconds),
		"memory_in_mb":       aws.Int64Value(canaryCodeOut.MemoryInMB),
		"active_tracing":     aws.BoolValue(canaryCodeOut.ActiveTracing),
	}

	return []interface{}{m}
}

func flattenAwsSyntheticsCanaryVpcConfig(canaryVpcOutput *synthetics.VpcConfigOutput) []interface{} {
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

func expandAwsSyntheticsCanaryVpcConfig(l []interface{}) *synthetics.VpcConfigInput {
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

func flattenAwsSyntheticsCanaryTimeline(timeline *synthetics.CanaryTimeline) []interface{} {
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

func syntheticsStartCanary(name string, conn *synthetics.Synthetics) error {
	log.Printf("[DEBUG] Starting Synthetics Canary: (%s)", name)
	_, err := conn.StartCanary(&synthetics.StartCanaryInput{
		Name: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error starting Synthetics Canary (%s): %w", name, err)
	}

	_, err = waiter.CanaryRunning(conn, name)

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) start: %w", name, err)
	}

	return nil
}

func syntheticsStopCanary(name string, conn *synthetics.Synthetics) error {
	log.Printf("[DEBUG] Stopping Synthetics Canary: (%s)", name)
	_, err := conn.StopCanary(&synthetics.StopCanaryInput{
		Name: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error stopping Synthetics Canary (%s): %w", name, err)
	}

	_, err = waiter.CanaryStopped(conn, name)

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
