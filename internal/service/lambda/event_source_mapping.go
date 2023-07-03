// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"log"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_lambda_event_source_mapping")
func ResourceEventSourceMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventSourceMappingCreate,
		ReadWithoutTimeout:   resourceEventSourceMappingRead,
		UpdateWithoutTimeout: resourceEventSourceMappingUpdate,
		DeleteWithoutTimeout: resourceEventSourceMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"amazon_managed_kafka_event_source_config": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				MaxItems:      1,
				ConflictsWith: []string{"self_managed_event_source", "self_managed_kafka_event_source_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"consumer_group_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 200),
						},
					},
				},
			},
			"batch_size": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// When AWS repurposed EventSourceMapping for use with SQS they kept
					// the default for BatchSize at 100 for Kinesis and DynamoDB, but made
					// the default 10 for SQS.  As such, we had to make batch_size optional.
					// Because of this, we need to ensure that if someone doesn't have
					// batch_size specified that it is not treated as a diff for those
					if new != "" && new != "0" {
						return false
					}

					var serviceName string
					if v, ok := d.GetOk("event_source_arn"); ok {
						eventSourceARN, err := arn.Parse(v.(string))
						if err != nil {
							return false
						}

						serviceName = eventSourceARN.Service
					} else if _, ok := d.GetOk("self_managed_event_source"); ok {
						serviceName = "kafka"
					}

					switch serviceName {
					case "dynamodb", "kinesis", "kafka", "mq", "rds":
						return old == "100"
					case "sqs":
						return old == "10"
					}

					return old == new
				},
			},
			"bisect_batch_on_function_error": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"destination_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_failure": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"document_db_event_source_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"collection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"full_document": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      lambda.FullDocumentDefault,
							ValidateFunc: validation.StringInSlice(lambda.FullDocument_Values(), false),
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"event_source_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"event_source_arn", "self_managed_event_source"},
			},
			"filter_criteria": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pattern": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
									},
								},
							},
						},
					},
				},
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Using function name or ARN should not be shown as a diff.
					// Try to convert the old and new values from ARN to function name
					oldFunctionName, oldFunctionNameErr := GetFunctionNameFromARN(old)
					newFunctionName, newFunctionNameErr := GetFunctionNameFromARN(new)
					return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
				},
			},
			"function_response_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lambda.FunctionResponseType_Values(), false),
				},
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_processing_result": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maximum_batching_window_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"maximum_record_age_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.IntInSlice([]int{-1}),
					validation.IntBetween(60, 604_800),
				),
			},
			"maximum_retry_attempts": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(-1, 10_000),
			},
			"parallelization_factor": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 10),
				Computed:     true,
			},
			"queues": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1000),
				},
			},
			"scaling_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_concurrency": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 1000),
						},
					},
				},
			},
			"self_managed_event_source": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoints": {
							Type:     schema.TypeMap,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if k == "self_managed_event_source.0.endpoints.KAFKA_BOOTSTRAP_SERVERS" {
									// AWS returns the bootstrap brokers in sorted order.
									olds := strings.Split(old, ",")
									sort.Strings(olds)
									news := strings.Split(new, ",")
									sort.Strings(news)

									return reflect.DeepEqual(olds, news)
								}

								return old == new
							},
						},
					},
				},
				ExactlyOneOf: []string{"event_source_arn", "self_managed_event_source"},
			},
			"self_managed_kafka_event_source_config": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				MaxItems:      1,
				ConflictsWith: []string{"event_source_arn", "amazon_managed_kafka_event_source_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"consumer_group_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 200),
						},
					},
				},
			},
			"source_access_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 22,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lambda.SourceAccessType_Values(), false),
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"starting_position": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(lambda.EventSourcePosition_Values(), false),
			},
			"starting_position_timestamp": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state_transition_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"topics": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 249),
				},
			},
			"tumbling_window_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 900),
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEventSourceMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	functionName := d.Get("function_name").(string)
	input := &lambda.CreateEventSourceMappingInput{
		Enabled:      aws.Bool(d.Get("enabled").(bool)),
		FunctionName: aws.String(functionName),
	}

	var target string

	if v, ok := d.GetOk("amazon_managed_kafka_event_source_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AmazonManagedKafkaEventSourceConfig = expandAmazonManagedKafkaEventSourceConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("batch_size"); ok {
		input.BatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("bisect_batch_on_function_error"); ok {
		input.BisectBatchOnFunctionError = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("destination_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DestinationConfig = expandDestinationConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("document_db_event_source_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DocumentDBEventSourceConfig = expandDocumentDBEventSourceConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("event_source_arn"); ok {
		v := v.(string)

		input.EventSourceArn = aws.String(v)
		target = v
	}

	if v, ok := d.GetOk("filter_criteria"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.FilterCriteria = expandFilterCriteria(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("function_response_types"); ok && v.(*schema.Set).Len() > 0 {
		input.FunctionResponseTypes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("maximum_batching_window_in_seconds"); ok {
		input.MaximumBatchingWindowInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("maximum_record_age_in_seconds"); ok {
		input.MaximumRecordAgeInSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOkExists("maximum_retry_attempts"); ok {
		input.MaximumRetryAttempts = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("parallelization_factor"); ok {
		input.ParallelizationFactor = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("queues"); ok && len(v.([]interface{})) > 0 {
		input.Queues = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ScalingConfig = expandScalingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("self_managed_event_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SelfManagedEventSource = expandSelfManagedEventSource(v.([]interface{})[0].(map[string]interface{}))

		target = "Self-Managed Apache Kafka"
	}

	if v, ok := d.GetOk("self_managed_kafka_event_source_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SelfManagedKafkaEventSourceConfig = expandSelfManagedKafkaEventSourceConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("source_access_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.SourceAccessConfigurations = expandSourceAccessConfigurations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("starting_position"); ok {
		input.StartingPosition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("starting_position_timestamp"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))

		input.StartingPositionTimestamp = aws.Time(t)
	}

	if v, ok := d.GetOk("topics"); ok && v.(*schema.Set).Len() > 0 {
		input.Topics = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tumbling_window_in_seconds"); ok {
		input.TumblingWindowInSeconds = aws.Int64(int64(v.(int)))
	}

	// IAM profiles and roles can take some time to propagate in AWS:
	//  http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	// Error creating Lambda function: InvalidParameterValueException: The
	// function defined for the task cannot be assumed by Lambda.
	//
	// The role may exist, but the permissions may not have propagated, so we retry.
	eventSourceMappingConfiguration, err := retryEventSourceMapping(ctx, func() (*lambda.EventSourceMappingConfiguration, error) {
		return conn.CreateEventSourceMappingWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lambda Event Source Mapping (%s): %s", target, err)
	}

	d.SetId(aws.StringValue(eventSourceMappingConfiguration.UUID))

	if _, err := waitEventSourceMappingCreate(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Event Source Mapping (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEventSourceMappingRead(ctx, d, meta)...)
}

func resourceEventSourceMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	eventSourceMappingConfiguration, err := FindEventSourceMappingConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[DEBUG] Lambda Event Source Mapping (%s) not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lambda Event Source Mapping (%s): %s", d.Id(), err)
	}

	if eventSourceMappingConfiguration.AmazonManagedKafkaEventSourceConfig != nil {
		if err := d.Set("amazon_managed_kafka_event_source_config", []interface{}{flattenAmazonManagedKafkaEventSourceConfig(eventSourceMappingConfiguration.AmazonManagedKafkaEventSourceConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting amazon_managed_kafka_event_source_config: %s", err)
		}
	} else {
		d.Set("amazon_managed_kafka_event_source_config", nil)
	}
	d.Set("batch_size", eventSourceMappingConfiguration.BatchSize)
	d.Set("bisect_batch_on_function_error", eventSourceMappingConfiguration.BisectBatchOnFunctionError)
	if eventSourceMappingConfiguration.DestinationConfig != nil {
		if err := d.Set("destination_config", []interface{}{flattenDestinationConfig(eventSourceMappingConfiguration.DestinationConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting destination_config: %s", err)
		}
	} else {
		d.Set("destination_config", nil)
	}
	if eventSourceMappingConfiguration.DocumentDBEventSourceConfig != nil {
		if err := d.Set("document_db_event_source_config", []interface{}{flattenDocumentDBEventSourceConfig(eventSourceMappingConfiguration.DocumentDBEventSourceConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting document_db_event_source_config: %s", err)
		}
	} else {
		d.Set("document_db_event_source_config", nil)
	}
	d.Set("event_source_arn", eventSourceMappingConfiguration.EventSourceArn)
	if v := eventSourceMappingConfiguration.FilterCriteria; v != nil {
		if err := d.Set("filter_criteria", []interface{}{flattenFilterCriteria(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting filter criteria: %s", err)
		}
	} else {
		d.Set("filter_criteria", nil)
	}
	d.Set("function_arn", eventSourceMappingConfiguration.FunctionArn)
	d.Set("function_name", eventSourceMappingConfiguration.FunctionArn)
	d.Set("function_response_types", aws.StringValueSlice(eventSourceMappingConfiguration.FunctionResponseTypes))
	if eventSourceMappingConfiguration.LastModified != nil {
		d.Set("last_modified", aws.TimeValue(eventSourceMappingConfiguration.LastModified).Format(time.RFC3339))
	} else {
		d.Set("last_modified", nil)
	}
	d.Set("last_processing_result", eventSourceMappingConfiguration.LastProcessingResult)
	d.Set("maximum_batching_window_in_seconds", eventSourceMappingConfiguration.MaximumBatchingWindowInSeconds)
	d.Set("maximum_record_age_in_seconds", eventSourceMappingConfiguration.MaximumRecordAgeInSeconds)
	d.Set("maximum_retry_attempts", eventSourceMappingConfiguration.MaximumRetryAttempts)
	d.Set("parallelization_factor", eventSourceMappingConfiguration.ParallelizationFactor)
	d.Set("queues", aws.StringValueSlice(eventSourceMappingConfiguration.Queues))
	if v := eventSourceMappingConfiguration.ScalingConfig; v != nil {
		if err := d.Set("scaling_config", []interface{}{flattenScalingConfig(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting scaling_config: %s", err)
		}
	} else {
		d.Set("scaling_config", nil)
	}
	if eventSourceMappingConfiguration.SelfManagedEventSource != nil {
		if err := d.Set("self_managed_event_source", []interface{}{flattenSelfManagedEventSource(eventSourceMappingConfiguration.SelfManagedEventSource)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting self_managed_event_source: %s", err)
		}
	} else {
		d.Set("self_managed_event_source", nil)
	}
	if eventSourceMappingConfiguration.SelfManagedKafkaEventSourceConfig != nil {
		if err := d.Set("self_managed_kafka_event_source_config", []interface{}{flattenSelfManagedKafkaEventSourceConfig(eventSourceMappingConfiguration.SelfManagedKafkaEventSourceConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting self_managed_kafka_event_source_config: %s", err)
		}
	} else {
		d.Set("self_managed_kafka_event_source_config", nil)
	}
	if err := d.Set("source_access_configuration", flattenSourceAccessConfigurations(eventSourceMappingConfiguration.SourceAccessConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source_access_configuration: %s", err)
	}
	d.Set("starting_position", eventSourceMappingConfiguration.StartingPosition)
	if eventSourceMappingConfiguration.StartingPositionTimestamp != nil {
		d.Set("starting_position_timestamp", aws.TimeValue(eventSourceMappingConfiguration.StartingPositionTimestamp).Format(time.RFC3339))
	} else {
		d.Set("starting_position_timestamp", nil)
	}
	d.Set("state", eventSourceMappingConfiguration.State)
	d.Set("state_transition_reason", eventSourceMappingConfiguration.StateTransitionReason)
	d.Set("topics", aws.StringValueSlice(eventSourceMappingConfiguration.Topics))
	d.Set("tumbling_window_in_seconds", eventSourceMappingConfiguration.TumblingWindowInSeconds)
	d.Set("uuid", eventSourceMappingConfiguration.UUID)

	switch state := d.Get("state").(string); state {
	case eventSourceMappingStateEnabled, eventSourceMappingStateEnabling:
		d.Set("enabled", true)
	case eventSourceMappingStateDisabled, eventSourceMappingStateDisabling:
		d.Set("enabled", false)
	default:
		log.Printf("[WARN] Lambda Event Source Mapping (%s) is neither enabled nor disabled, but %s", d.Id(), state)
		d.Set("enabled", nil)
	}

	return diags
}

func resourceEventSourceMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	input := &lambda.UpdateEventSourceMappingInput{
		UUID: aws.String(d.Id()),
	}

	if d.HasChange("batch_size") {
		input.BatchSize = aws.Int64(int64(d.Get("batch_size").(int)))
	}

	if d.HasChange("bisect_batch_on_function_error") {
		input.BisectBatchOnFunctionError = aws.Bool(d.Get("bisect_batch_on_function_error").(bool))
	}

	if d.HasChange("destination_config") {
		if v, ok := d.GetOk("destination_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DestinationConfig = expandDestinationConfig(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if d.HasChange("document_db_event_source_config") {
		if v, ok := d.GetOk("document_db_event_source_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DocumentDBEventSourceConfig = expandDocumentDBEventSourceConfig(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if d.HasChange("enabled") {
		input.Enabled = aws.Bool(d.Get("enabled").(bool))
	}

	if d.HasChange("filter_criteria") {
		if v, ok := d.GetOk("filter_criteria"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.FilterCriteria = expandFilterCriteria(v.([]interface{})[0].(map[string]interface{}))
		} else {
			// AWS ignores the removal if this is left as nil.
			input.FilterCriteria = &lambda.FilterCriteria{}
		}
	}

	if d.HasChange("function_name") {
		input.FunctionName = aws.String(d.Get("function_name").(string))
	}

	if d.HasChange("function_response_types") {
		input.FunctionResponseTypes = flex.ExpandStringSet(d.Get("function_response_types").(*schema.Set))
	}

	if d.HasChange("maximum_batching_window_in_seconds") {
		input.MaximumBatchingWindowInSeconds = aws.Int64(int64(d.Get("maximum_batching_window_in_seconds").(int)))
	}

	if d.HasChange("maximum_record_age_in_seconds") {
		input.MaximumRecordAgeInSeconds = aws.Int64(int64(d.Get("maximum_record_age_in_seconds").(int)))
	}

	if d.HasChange("maximum_retry_attempts") {
		input.MaximumRetryAttempts = aws.Int64(int64(d.Get("maximum_retry_attempts").(int)))
	}

	if d.HasChange("parallelization_factor") {
		input.ParallelizationFactor = aws.Int64(int64(d.Get("parallelization_factor").(int)))
	}

	if d.HasChange("scaling_config") {
		if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ScalingConfig = expandScalingConfig(v.([]interface{})[0].(map[string]interface{}))
		} else {
			// AWS ignores the removal if this is left as nil.
			input.ScalingConfig = &lambda.ScalingConfig{}
		}
	}

	if d.HasChange("source_access_configuration") {
		if v, ok := d.GetOk("source_access_configuration"); ok && v.(*schema.Set).Len() > 0 {
			input.SourceAccessConfigurations = expandSourceAccessConfigurations(v.(*schema.Set).List())
		}
	}

	if d.HasChange("tumbling_window_in_seconds") {
		input.TumblingWindowInSeconds = aws.Int64(int64(d.Get("tumbling_window_in_seconds").(int)))
	}

	_, err := retryEventSourceMapping(ctx, func() (*lambda.EventSourceMappingConfiguration, error) {
		return conn.UpdateEventSourceMappingWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Lambda Event Source Mapping (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSourceMappingUpdate(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Event Source Mapping (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceEventSourceMappingRead(ctx, d, meta)...)
}

func resourceEventSourceMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LambdaConn(ctx)

	log.Printf("[INFO] Deleting Lambda Event Source Mapping: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, eventSourceMappingPropagationTimeout, func() (interface{}, error) {
		return conn.DeleteEventSourceMappingWithContext(ctx, &lambda.DeleteEventSourceMappingInput{
			UUID: aws.String(d.Id()),
		})
	}, lambda.ErrCodeResourceInUseException)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lambda Event Source Mapping (%s): %s", d.Id(), err)
	}

	if _, err := waitEventSourceMappingDelete(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lambda Event Source Mapping (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandDestinationConfig(tfMap map[string]interface{}) *lambda.DestinationConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.DestinationConfig{}

	if v, ok := tfMap["on_failure"].([]interface{}); ok && len(v) > 0 {
		apiObject.OnFailure = expandOnFailure(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandDocumentDBEventSourceConfig(tfMap map[string]interface{}) *lambda.DocumentDBEventSourceConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.DocumentDBEventSourceConfig{}

	if v, ok := tfMap["collection_name"].(string); ok && v != "" {
		apiObject.CollectionName = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["full_document"].(string); ok && v != "" {
		apiObject.FullDocument = aws.String(v)
	}

	return apiObject
}

func expandOnFailure(tfMap map[string]interface{}) *lambda.OnFailure {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.OnFailure{}

	if v, ok := tfMap["destination_arn"].(string); ok {
		apiObject.Destination = aws.String(v)
	}

	return apiObject
}

func flattenDestinationConfig(apiObject *lambda.DestinationConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnFailure; v != nil {
		tfMap["on_failure"] = []interface{}{flattenOnFailure(v)}
	}

	return tfMap
}

func flattenDocumentDBEventSourceConfig(apiObject *lambda.DocumentDBEventSourceConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CollectionName; v != nil {
		tfMap["collection_name"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.FullDocument; v != nil {
		tfMap["full_document"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenOnFailure(apiObject *lambda.OnFailure) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Destination; v != nil {
		tfMap["destination_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func expandSelfManagedEventSource(tfMap map[string]interface{}) *lambda.SelfManagedEventSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.SelfManagedEventSource{}

	if v, ok := tfMap["endpoints"].(map[string]interface{}); ok && len(v) > 0 {
		m := map[string][]*string{}

		for k, v := range v {
			m[k] = aws.StringSlice(strings.Split(v.(string), ","))
		}

		apiObject.Endpoints = m
	}

	return apiObject
}

func flattenSelfManagedEventSource(apiObject *lambda.SelfManagedEventSource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Endpoints; v != nil {
		m := map[string]string{}

		for k, v := range v {
			m[k] = strings.Join(aws.StringValueSlice(v), ",")
		}

		tfMap["endpoints"] = m
	}

	return tfMap
}

func expandAmazonManagedKafkaEventSourceConfig(tfMap map[string]interface{}) *lambda.AmazonManagedKafkaEventSourceConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.AmazonManagedKafkaEventSourceConfig{}

	if v, ok := tfMap["consumer_group_id"].(string); ok && v != "" {
		apiObject.ConsumerGroupId = aws.String(v)
	}

	return apiObject
}

func flattenAmazonManagedKafkaEventSourceConfig(apiObject *lambda.AmazonManagedKafkaEventSourceConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConsumerGroupId; v != nil {
		tfMap["consumer_group_id"] = aws.StringValue(v)
	}

	return tfMap
}

func expandSelfManagedKafkaEventSourceConfig(tfMap map[string]interface{}) *lambda.SelfManagedKafkaEventSourceConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.SelfManagedKafkaEventSourceConfig{}

	if v, ok := tfMap["consumer_group_id"].(string); ok && v != "" {
		apiObject.ConsumerGroupId = aws.String(v)
	}

	return apiObject
}

func flattenSelfManagedKafkaEventSourceConfig(apiObject *lambda.SelfManagedKafkaEventSourceConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConsumerGroupId; v != nil {
		tfMap["consumer_group_id"] = aws.StringValue(v)
	}

	return tfMap
}

func expandSourceAccessConfiguration(tfMap map[string]interface{}) *lambda.SourceAccessConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.SourceAccessConfiguration{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["uri"].(string); ok && v != "" {
		apiObject.URI = aws.String(v)
	}

	return apiObject
}

func expandSourceAccessConfigurations(tfList []interface{}) []*lambda.SourceAccessConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*lambda.SourceAccessConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSourceAccessConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenSourceAccessConfiguration(apiObject *lambda.SourceAccessConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.URI; v != nil {
		tfMap["uri"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSourceAccessConfigurations(apiObjects []*lambda.SourceAccessConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenSourceAccessConfiguration(apiObject))
	}

	return tfList
}

func expandFilterCriteria(tfMap map[string]interface{}) *lambda.FilterCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.FilterCriteria{}

	if v, ok := tfMap["filter"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Filters = expandFilters(v.List())
	}

	return apiObject
}

func flattenFilterCriteria(apiObject *lambda.FilterCriteria) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Filters; len(v) > 0 {
		tfMap["filter"] = flattenFilters(v)
	}

	return tfMap
}

func expandFilters(tfList []interface{}) []*lambda.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*lambda.Filter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenFilters(apiObjects []*lambda.Filter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenFilter(apiObject))
	}

	return tfList
}

func expandFilter(tfMap map[string]interface{}) *lambda.Filter {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.Filter{}

	if v, ok := tfMap["pattern"].(string); ok {
		// The API permits patterns of length >= 0, so accept the empty string.
		apiObject.Pattern = aws.String(v)
	}

	return apiObject
}

func flattenFilter(apiObject *lambda.Filter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Pattern; v != nil {
		tfMap["pattern"] = aws.StringValue(v)
	}

	return tfMap
}

func expandScalingConfig(tfMap map[string]interface{}) *lambda.ScalingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &lambda.ScalingConfig{}

	if v, ok := tfMap["maximum_concurrency"].(int); ok && v != 0 {
		apiObject.MaximumConcurrency = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenScalingConfig(apiObject *lambda.ScalingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaximumConcurrency; v != nil {
		tfMap["maximum_concurrency"] = int(aws.Int64Value(v))
	}

	return tfMap
}

func findEventSourceMappingConfiguration(ctx context.Context, conn *lambda.Lambda, input *lambda.GetEventSourceMappingInput) (*lambda.EventSourceMappingConfiguration, error) {
	output, err := conn.GetEventSourceMappingWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
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

func FindEventSourceMappingConfigurationByID(ctx context.Context, conn *lambda.Lambda, uuid string) (*lambda.EventSourceMappingConfiguration, error) {
	input := &lambda.GetEventSourceMappingInput{
		UUID: aws.String(uuid),
	}

	return findEventSourceMappingConfiguration(ctx, conn, input)
}

func statusEventSourceMappingState(ctx context.Context, conn *lambda.Lambda, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		eventSourceMappingConfiguration, err := FindEventSourceMappingConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return eventSourceMappingConfiguration, aws.StringValue(eventSourceMappingConfiguration.State), nil
	}
}

const (
	eventSourceMappingCreateTimeout      = 10 * time.Minute
	eventSourceMappingUpdateTimeout      = 10 * time.Minute
	eventSourceMappingDeleteTimeout      = 5 * time.Minute
	eventSourceMappingPropagationTimeout = 5 * time.Minute
)

func waitEventSourceMappingCreate(ctx context.Context, conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventSourceMappingStateCreating, eventSourceMappingStateDisabling, eventSourceMappingStateEnabling},
		Target:  []string{eventSourceMappingStateDisabled, eventSourceMappingStateEnabled},
		Refresh: statusEventSourceMappingState(ctx, conn, id),
		Timeout: eventSourceMappingCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateTransitionReason)))

		return output, err
	}

	return nil, err
}

func waitEventSourceMappingDelete(ctx context.Context, conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventSourceMappingStateDeleting},
		Target:  []string{},
		Refresh: statusEventSourceMappingState(ctx, conn, id),
		Timeout: eventSourceMappingDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateTransitionReason)))

		return output, err
	}

	return nil, err
}

func waitEventSourceMappingUpdate(ctx context.Context, conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eventSourceMappingStateDisabling, eventSourceMappingStateEnabling, eventSourceMappingStateUpdating},
		Target:  []string{eventSourceMappingStateDisabled, eventSourceMappingStateEnabled},
		Refresh: statusEventSourceMappingState(ctx, conn, id),
		Timeout: eventSourceMappingUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateTransitionReason)))

		return output, err
	}

	return nil, err
}

func retryEventSourceMapping(ctx context.Context, f func() (*lambda.EventSourceMappingConfiguration, error)) (*lambda.EventSourceMappingConfiguration, error) {
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return f()
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "cannot be assumed by Lambda") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "execution role does not have permissions") {
				return true, err
			}

			if tfawserr.ErrMessageContains(err, lambda.ErrCodeInvalidParameterValueException, "ensure the role can perform") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return nil, err
	}

	return outputRaw.(*lambda.EventSourceMappingConfiguration), err
}
