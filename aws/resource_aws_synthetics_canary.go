package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const awsMutexCanary = `aws_synthetics_canary`

func resourceAwsSyntheticsCanary() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSyntheticsCanaryCreate,
		Read:   resourceAwsSyntheticsCanaryRead,
		Update: resourceAwsSyntheticsCanaryUpdate,
		Delete: resourceAwsSyntheticsCanaryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 21),
			},
			"artifact_s3_location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimPrefix(old, "s3://") == new
				},
			},
			"code": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"handler": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_bucket": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"zip_file": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source_location_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"failure_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"success_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"run_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeout_in_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Required: true,
						},
						"duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
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
							Optional: true,
						},
					},
				},
			},
			"engine_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"runtime_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSyntheticsCanaryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	input := &synthetics.CreateCanaryInput{
		Name:               aws.String(d.Get("name").(string)),
		ArtifactS3Location: aws.String(d.Get("artifact_s3_location").(string)),
		ExecutionRoleArn:   aws.String(d.Get("execution_role_arn").(string)),
		RuntimeVersion:     aws.String("syn-1.0"),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().SyntheticsTags()
	}

	code, err := expandAwsSyntheticsCanaryCode(d.Get("code").([]interface{}))
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

	resp, err := conn.CreateCanary(input)
	if err != nil {
		return fmt.Errorf("error creating Synthetics Canary: %s", err)
	}

	d.SetId(aws.StringValue(resp.Canary.Name))

	return resourceAwsSyntheticsCanaryRead(d, meta)
}

func resourceAwsSyntheticsCanaryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &synthetics.GetCanaryInput{
		Name: aws.String(d.Id()),
	}

	resp, err := conn.GetCanary(input)
	if err != nil {
		return fmt.Errorf("error reading Synthetics Canary: %s", err)
	}

	canary := resp.Canary
	d.Set("name", canary.Name)
	d.Set("engine_arn", canary.EngineArn)
	d.Set("execution_role_arn", canary.ExecutionRoleArn)
	d.Set("runtime_version", canary.RuntimeVersion)
	d.Set("artifact_s3_location", canary.ArtifactS3Location)
	d.Set("failure_retention_period", canary.FailureRetentionPeriodInDays)
	d.Set("success_retention_period", canary.SuccessRetentionPeriodInDays)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "synthetics",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("canary/%s", aws.StringValue(canary.Name)),
	}.String()

	d.Set("arn", arn)

	if err := d.Set("code", flattenAwsSyntheticsCanaryCode(canary.Code)); err != nil {
		return fmt.Errorf("error setting code: %s", err)
	}

	if err := d.Set("vpc_config", flattenAwsSyntheticsCanaryVpcConfig(canary.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc config: %s", err)
	}

	if err := d.Set("run_config", flattenAwsSyntheticsCanaryRunConfig(canary.RunConfig)); err != nil {
		return fmt.Errorf("error setting run config: %s", err)
	}

	if err := d.Set("schedule", flattenAwsSyntheticsCanarySchedule(canary.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.SyntheticsKeyValueTags(canary.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSyntheticsCanaryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	input := &synthetics.UpdateCanaryInput{
		Name: aws.String(d.Id()),
	}

	updateFlag := false

	if d.HasChange("vpc_config") {
		input.VpcConfig = expandAwsSyntheticsCanaryVpcConfig(d.Get("vpc_config").([]interface{}))
		updateFlag = true
	}

	if d.HasChange("run_config") {
		input.RunConfig = expandAwsSyntheticsCanaryRunConfig(d.Get("run_config").([]interface{}))
		updateFlag = true
	}

	if d.HasChange("schedule") {
		input.Schedule = expandAwsSyntheticsCanarySchedule(d.Get("schedule").([]interface{}))
		updateFlag = true
	}

	if d.HasChange("success_retention_period") {
		_, n := d.GetChange("success_retention_period")
		input.SuccessRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		updateFlag = true
	}

	if d.HasChange("failure_retention_period") {
		_, n := d.GetChange("failure_retention_period")
		input.FailureRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		updateFlag = true
	}

	if d.HasChange("execution_role_arn") {
		_, n := d.GetChange("execution_role_arn")
		input.ExecutionRoleArn = aws.String(n.(string))
		updateFlag = true
	}

	if updateFlag {
		_, err := conn.UpdateCanary(input)
		if err != nil {
			return fmt.Errorf("error updating Synthetics Canary: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SyntheticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsSyntheticsCanaryRead(d, meta)
}

func resourceAwsSyntheticsCanaryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	input := &synthetics.DeleteCanaryInput{
		Name: aws.String(d.Id()),
	}

	_, err := conn.DeleteCanary(input)
	if err != nil {
		if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Synthetics Canary: %s", err)
	}

	return nil
}

func expandAwsSyntheticsCanaryCode(l []interface{}) (*synthetics.CanaryCodeInput, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.CanaryCodeInput{
		Handler: aws.String(m["handler"].(string)),
	}

	if v, ok := m["zip_file"]; ok {
		awsMutexKV.Lock(awsMutexCanary)
		defer awsMutexKV.Unlock(awsMutexCanary)
		file, err := loadFileContent(v.(string))
		if err != nil {
			return nil, fmt.Errorf("Unable to load %q: %s", v.(string), err)
		}

		codeConfig.ZipFile = file
	}

	return codeConfig, nil
}

func flattenAwsSyntheticsCanaryCode(canaryCodeOut *synthetics.CanaryCodeOutput) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"handler":             aws.StringValue(canaryCodeOut.Handler),
		"source_location_arn": aws.StringValue(canaryCodeOut.SourceLocationArn),
	}

	return []interface{}{m}
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

	return codeConfig
}

func flattenAwsSyntheticsCanaryRunConfig(canaryCodeOut *synthetics.CanaryRunConfigOutput) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"timeout_in_seconds": aws.Int64Value(canaryCodeOut.TimeoutInSeconds),
	}

	return []interface{}{m}
}

func flattenAwsSyntheticsCanaryVpcConfig(canaryVpcOutput *synthetics.VpcConfigOutput) []interface{} {
	if canaryVpcOutput == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"subnet_ids":         flattenStringSet(canaryVpcOutput.SubnetIds),
		"security_group_ids": flattenStringSet(canaryVpcOutput.SecurityGroupIds),
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
		SubnetIds:        expandStringSet(m["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
	}

	return codeConfig
}
