// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafkaconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
)

// @SDKResource("aws_mskconnect_connector", name="Connector")
// @Tags(identifierAttribute="arn")
func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorCreate,
		ReadWithoutTimeout:   resourceConnectorRead,
		UpdateWithoutTimeout: resourceConnectorUpdate,
		DeleteWithoutTimeout: resourceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_worker_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 10),
									},
									"mcu_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1,
										ValidateFunc: validation.IntInSlice([]int{1, 2, 4, 8}),
									},
									"min_worker_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 10),
									},
									"scale_in_policy": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu_utilization_percentage": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntBetween(1, 100),
												},
											},
										},
									},
									"scale_out_policy": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu_utilization_percentage": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntBetween(1, 100),
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{"capacity.0.autoscaling", "capacity.0.provisioned_capacity"},
						},
						"provisioned_capacity": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mcu_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1,
										ValidateFunc: validation.IntInSlice([]int{1, 2, 4, 8}),
									},
									"worker_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 10),
									},
								},
							},
							ExactlyOneOf: []string{"capacity.0.autoscaling", "capacity.0.provisioned_capacity"},
						},
					},
				},
			},
			"connector_configuration": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"kafka_cluster": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apache_kafka_cluster": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bootstrap_servers": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"vpc": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrSecurityGroups: {
													Type:     schema.TypeSet,
													Required: true,
													ForceNew: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												names.AttrSubnets: {
													Type:     schema.TypeSet,
													Required: true,
													ForceNew: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"kafka_cluster_client_authentication": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.KafkaClusterClientAuthenticationTypeNone,
							ValidateDiagFunc: enum.Validate[awstypes.KafkaClusterClientAuthenticationType](),
						},
					},
				},
			},
			"kafka_cluster_encryption_in_transit": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.KafkaClusterEncryptionInTransitTypePlaintext,
							ValidateDiagFunc: enum.Validate[awstypes.KafkaClusterEncryptionInTransitType](),
						},
					},
				},
			},
			"kafkaconnect_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"log_delivery": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"worker_log_delivery": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCloudWatchLogs: {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
												"log_group": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
									"firehose": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"delivery_stream": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucket: {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
												names.AttrPrefix: {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"plugin": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_plugin": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidARN,
									},
									"revision": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
			},
			"service_execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"worker_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"revision": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kafkaconnect.CreateConnectorInput{
		Capacity:                         expandCapacity(d.Get("capacity").([]interface{})[0].(map[string]interface{})),
		ConnectorConfiguration:           flex.ExpandStringValueMap(d.Get("connector_configuration").(map[string]interface{})),
		ConnectorName:                    aws.String(name),
		KafkaCluster:                     expandCluster(d.Get("kafka_cluster").([]interface{})[0].(map[string]interface{})),
		KafkaClusterClientAuthentication: expandClusterClientAuthentication(d.Get("kafka_cluster_client_authentication").([]interface{})[0].(map[string]interface{})),
		KafkaClusterEncryptionInTransit:  expandClusterEncryptionInTransit(d.Get("kafka_cluster_encryption_in_transit").([]interface{})[0].(map[string]interface{})),
		KafkaConnectVersion:              aws.String(d.Get("kafkaconnect_version").(string)),
		Plugins:                          expandPlugins(d.Get("plugin").(*schema.Set).List()),
		ServiceExecutionRoleArn:          aws.String(d.Get("service_execution_role_arn").(string)),
		Tags:                             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ConnectorDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_delivery"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LogDelivery = expandLogDelivery(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("worker_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.WorkerConfiguration = expandWorkerConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateConnector(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Connect Connector (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ConnectorArn))

	if _, err := waitConnectorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Connector (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	connector, err := findConnectorByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Connect Connector (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, connector.ConnectorArn)
	if connector.Capacity != nil {
		if err := d.Set("capacity", []interface{}{flattenCapacityDescription(connector.Capacity)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting capacity: %s", err)
		}
	} else {
		d.Set("capacity", nil)
	}
	d.Set("connector_configuration", connector.ConnectorConfiguration)
	d.Set(names.AttrDescription, connector.ConnectorDescription)
	if connector.KafkaCluster != nil {
		if err := d.Set("kafka_cluster", []interface{}{flattenClusterDescription(connector.KafkaCluster)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting kafka_cluster: %s", err)
		}
	} else {
		d.Set("kafka_cluster", nil)
	}
	if connector.KafkaClusterClientAuthentication != nil {
		if err := d.Set("kafka_cluster_client_authentication", []interface{}{flattenClusterClientAuthenticationDescription(connector.KafkaClusterClientAuthentication)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting kafka_cluster_client_authentication: %s", err)
		}
	} else {
		d.Set("kafka_cluster_client_authentication", nil)
	}
	if connector.KafkaClusterEncryptionInTransit != nil {
		if err := d.Set("kafka_cluster_encryption_in_transit", []interface{}{flattenClusterEncryptionInTransitDescription(connector.KafkaClusterEncryptionInTransit)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting kafka_cluster_encryption_in_transit: %s", err)
		}
	} else {
		d.Set("kafka_cluster_encryption_in_transit", nil)
	}
	d.Set("kafkaconnect_version", connector.KafkaConnectVersion)
	if connector.LogDelivery != nil {
		if err := d.Set("log_delivery", []interface{}{flattenLogDeliveryDescription(connector.LogDelivery)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting log_delivery: %s", err)
		}
	} else {
		d.Set("log_delivery", nil)
	}
	d.Set(names.AttrName, connector.ConnectorName)
	if err := d.Set("plugin", flattenPluginDescriptions(connector.Plugins)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting plugin: %s", err)
	}
	d.Set("service_execution_role_arn", connector.ServiceExecutionRoleArn)
	d.Set(names.AttrVersion, connector.CurrentVersion)
	if connector.WorkerConfiguration != nil {
		if err := d.Set("worker_configuration", []interface{}{flattenWorkerConfigurationDescription(connector.WorkerConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting worker_configuration: %s", err)
		}
	} else {
		d.Set("worker_configuration", nil)
	}

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &kafkaconnect.UpdateConnectorInput{
			Capacity:       expandCapacityUpdate(d.Get("capacity").([]interface{})[0].(map[string]interface{})),
			ConnectorArn:   aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get(names.AttrVersion).(string)),
		}

		_, err := conn.UpdateConnector(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Connect Connector (%s): %s", d.Id(), err)
		}

		if _, err := waitConnectorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Connector (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	log.Printf("[DEBUG] Deleting MSK Connect Connector: %s", d.Id())
	_, err := conn.DeleteConnector(ctx, &kafkaconnect.DeleteConnectorInput{
		ConnectorArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Connect Connector (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Connect Connector (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConnectorByARN(ctx context.Context, conn *kafkaconnect.Client, arn string) (*kafkaconnect.DescribeConnectorOutput, error) {
	input := &kafkaconnect.DescribeConnectorInput{
		ConnectorArn: aws.String(arn),
	}

	output, err := conn.DescribeConnector(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func statusConnector(ctx context.Context, conn *kafkaconnect.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConnectorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ConnectorState), nil
	}
}

func waitConnectorCreated(ctx context.Context, conn *kafkaconnect.Client, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStateCreating),
		Target:  enum.Slice(awstypes.ConnectorStateRunning),
		Refresh: statusConnector(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := output.ConnectorState, output.StateDescription; state == awstypes.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateDescription.Code), aws.ToString(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorUpdated(ctx context.Context, conn *kafkaconnect.Client, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStateUpdating),
		Target:  enum.Slice(awstypes.ConnectorStateRunning),
		Refresh: statusConnector(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := output.ConnectorState, output.StateDescription; state == awstypes.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateDescription.Code), aws.ToString(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorDeleted(ctx context.Context, conn *kafkaconnect.Client, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStateDeleting),
		Target:  []string{},
		Refresh: statusConnector(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := output.ConnectorState, output.StateDescription; state == awstypes.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateDescription.Code), aws.ToString(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func expandCapacity(tfMap map[string]interface{}) *awstypes.Capacity {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Capacity{}

	if v, ok := tfMap["autoscaling"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutoScaling = expandAutoScaling(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["provisioned_capacity"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ProvisionedCapacity = expandProvisionedCapacity(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAutoScaling(tfMap map[string]interface{}) *awstypes.AutoScaling {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutoScaling{}

	if v, ok := tfMap["max_worker_count"].(int); ok && v != 0 {
		apiObject.MaxWorkerCount = int32(v)
	}

	if v, ok := tfMap["mcu_count"].(int); ok && v != 0 {
		apiObject.McuCount = int32(v)
	}

	if v, ok := tfMap["min_worker_count"].(int); ok && v != 0 {
		apiObject.MinWorkerCount = int32(v)
	}

	if v, ok := tfMap["scale_in_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ScaleInPolicy = expandScaleInPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["scale_out_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ScaleOutPolicy = expandScaleOutPolicy(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandScaleInPolicy(tfMap map[string]interface{}) *awstypes.ScaleInPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ScaleInPolicy{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok && v != 0 {
		apiObject.CpuUtilizationPercentage = int32(v)
	}

	return apiObject
}

func expandScaleOutPolicy(tfMap map[string]interface{}) *awstypes.ScaleOutPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ScaleOutPolicy{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok && v != 0 {
		apiObject.CpuUtilizationPercentage = int32(v)
	}

	return apiObject
}

func expandProvisionedCapacity(tfMap map[string]interface{}) *awstypes.ProvisionedCapacity {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ProvisionedCapacity{}

	if v, ok := tfMap["mcu_count"].(int); ok && v != 0 {
		apiObject.McuCount = int32(v)
	}

	if v, ok := tfMap["worker_count"].(int); ok && v != 0 {
		apiObject.WorkerCount = int32(v)
	}

	return apiObject
}

func expandCapacityUpdate(tfMap map[string]interface{}) *awstypes.CapacityUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CapacityUpdate{}

	if v, ok := tfMap["autoscaling"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AutoScaling = expandAutoScalingUpdate(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["provisioned_capacity"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ProvisionedCapacity = expandProvisionedCapacityUpdate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAutoScalingUpdate(tfMap map[string]interface{}) *awstypes.AutoScalingUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AutoScalingUpdate{}

	if v, ok := tfMap["max_worker_count"].(int); ok {
		apiObject.MaxWorkerCount = int32(v)
	}

	if v, ok := tfMap["mcu_count"].(int); ok {
		apiObject.McuCount = int32(v)
	}

	if v, ok := tfMap["min_worker_count"].(int); ok {
		apiObject.MinWorkerCount = int32(v)
	}

	if v, ok := tfMap["scale_in_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ScaleInPolicy = expandScaleInPolicyUpdate(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["scale_out_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ScaleOutPolicy = expandScaleOutPolicyUpdate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandScaleInPolicyUpdate(tfMap map[string]interface{}) *awstypes.ScaleInPolicyUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ScaleInPolicyUpdate{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok {
		apiObject.CpuUtilizationPercentage = int32(v)
	}

	return apiObject
}

func expandScaleOutPolicyUpdate(tfMap map[string]interface{}) *awstypes.ScaleOutPolicyUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ScaleOutPolicyUpdate{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok {
		apiObject.CpuUtilizationPercentage = int32(v)
	}

	return apiObject
}

func expandProvisionedCapacityUpdate(tfMap map[string]interface{}) *awstypes.ProvisionedCapacityUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ProvisionedCapacityUpdate{}

	if v, ok := tfMap["mcu_count"].(int); ok {
		apiObject.McuCount = int32(v)
	}

	if v, ok := tfMap["worker_count"].(int); ok {
		apiObject.WorkerCount = int32(v)
	}

	return apiObject
}

func expandCluster(tfMap map[string]interface{}) *awstypes.KafkaCluster {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KafkaCluster{}

	if v, ok := tfMap["apache_kafka_cluster"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ApacheKafkaCluster = expandApacheCluster(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandApacheCluster(tfMap map[string]interface{}) *awstypes.ApacheKafkaCluster {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ApacheKafkaCluster{}

	if v, ok := tfMap["bootstrap_servers"].(string); ok && v != "" {
		apiObject.BootstrapServers = aws.String(v)
	}

	if v, ok := tfMap["vpc"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Vpc = expandVPC(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandVPC(tfMap map[string]interface{}) *awstypes.Vpc {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Vpc{}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnets].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandClusterClientAuthentication(tfMap map[string]interface{}) *awstypes.KafkaClusterClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KafkaClusterClientAuthentication{}

	if v, ok := tfMap["authentication_type"].(string); ok && v != "" {
		apiObject.AuthenticationType = awstypes.KafkaClusterClientAuthenticationType(v)
	}

	return apiObject
}

func expandClusterEncryptionInTransit(tfMap map[string]interface{}) *awstypes.KafkaClusterEncryptionInTransit {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KafkaClusterEncryptionInTransit{}

	if v, ok := tfMap["encryption_type"].(string); ok && v != "" {
		apiObject.EncryptionType = awstypes.KafkaClusterEncryptionInTransitType(v)
	}

	return apiObject
}

func expandPlugin(tfMap map[string]interface{}) *awstypes.Plugin {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Plugin{}

	if v, ok := tfMap["custom_plugin"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CustomPlugin = expandCustomPlugin(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPlugins(tfList []interface{}) []awstypes.Plugin {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Plugin

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandPlugin(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCustomPlugin(tfMap map[string]interface{}) *awstypes.CustomPlugin {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CustomPlugin{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.CustomPluginArn = aws.String(v)
	}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = int64(v)
	}

	return apiObject
}

func expandLogDelivery(tfMap map[string]interface{}) *awstypes.LogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LogDelivery{}

	if v, ok := tfMap["worker_log_delivery"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.WorkerLogDelivery = expandWorkerLogDelivery(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandWorkerLogDelivery(tfMap map[string]interface{}) *awstypes.WorkerLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WorkerLogDelivery{}

	if v, ok := tfMap[names.AttrCloudWatchLogs].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchLogs = expandCloudWatchLogsLogDelivery(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["firehose"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Firehose = expandFirehoseLogDelivery(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3 = expandS3LogDelivery(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogsLogDelivery(tfMap map[string]interface{}) *awstypes.CloudWatchLogsLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CloudWatchLogsLogDelivery{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		apiObject.LogGroup = aws.String(v)
	}

	return apiObject
}

func expandFirehoseLogDelivery(tfMap map[string]interface{}) *awstypes.FirehoseLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FirehoseLogDelivery{}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		apiObject.DeliveryStream = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	return apiObject
}

func expandS3LogDelivery(tfMap map[string]interface{}) *awstypes.S3LogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3LogDelivery{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = v
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandWorkerConfiguration(tfMap map[string]interface{}) *awstypes.WorkerConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WorkerConfiguration{}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = int64(v)
	}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.WorkerConfigurationArn = aws.String(v)
	}

	return apiObject
}

func flattenCapacityDescription(apiObject *awstypes.CapacityDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoScaling; v != nil {
		tfMap["autoscaling"] = []interface{}{flattenAutoScalingDescription(v)}
	}

	if v := apiObject.ProvisionedCapacity; v != nil {
		tfMap["provisioned_capacity"] = []interface{}{flattenProvisionedCapacityDescription(v)}
	}

	return tfMap
}

func flattenAutoScalingDescription(apiObject *awstypes.AutoScalingDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"max_worker_count": apiObject.MaxWorkerCount,
		"mcu_count":        apiObject.McuCount,
		"min_worker_count": apiObject.MinWorkerCount,
	}

	if v := apiObject.ScaleInPolicy; v != nil {
		tfMap["scale_in_policy"] = []interface{}{flattenScaleInPolicyDescription(v)}
	}

	if v := apiObject.ScaleOutPolicy; v != nil {
		tfMap["scale_out_policy"] = []interface{}{flattenScaleOutPolicyDescription(v)}
	}

	return tfMap
}

func flattenScaleInPolicyDescription(apiObject *awstypes.ScaleInPolicyDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cpu_utilization_percentage": apiObject.CpuUtilizationPercentage,
	}

	return tfMap
}

func flattenScaleOutPolicyDescription(apiObject *awstypes.ScaleOutPolicyDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cpu_utilization_percentage": apiObject.CpuUtilizationPercentage,
	}

	return tfMap
}

func flattenProvisionedCapacityDescription(apiObject *awstypes.ProvisionedCapacityDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"mcu_count":    apiObject.McuCount,
		"worker_count": apiObject.WorkerCount,
	}

	return tfMap
}

func flattenClusterDescription(apiObject *awstypes.KafkaClusterDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ApacheKafkaCluster; v != nil {
		tfMap["apache_kafka_cluster"] = []interface{}{flattenApacheClusterDescription(v)}
	}

	return tfMap
}

func flattenApacheClusterDescription(apiObject *awstypes.ApacheKafkaClusterDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BootstrapServers; v != nil {
		tfMap["bootstrap_servers"] = aws.ToString(v)
	}

	if v := apiObject.Vpc; v != nil {
		tfMap["vpc"] = []interface{}{flattenVPCDescription(v)}
	}

	return tfMap
}

func flattenVPCDescription(apiObject *awstypes.VpcDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap[names.AttrSecurityGroups] = v
	}

	if v := apiObject.Subnets; v != nil {
		tfMap[names.AttrSubnets] = v
	}

	return tfMap
}

func flattenClusterClientAuthenticationDescription(apiObject *awstypes.KafkaClusterClientAuthenticationDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"authentication_type": apiObject.AuthenticationType,
	}

	return tfMap
}

func flattenClusterEncryptionInTransitDescription(apiObject *awstypes.KafkaClusterEncryptionInTransitDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"encryption_type": apiObject.EncryptionType,
	}

	return tfMap
}

func flattenPluginDescription(apiObject *awstypes.PluginDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CustomPlugin; v != nil {
		tfMap["custom_plugin"] = []interface{}{flattenCustomPluginDescription(v)}
	}

	return tfMap
}

func flattenPluginDescriptions(apiObjects []awstypes.PluginDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPluginDescription(&apiObject))
	}

	return tfList
}

func flattenCustomPluginDescription(apiObject *awstypes.CustomPluginDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"revision": apiObject.Revision,
	}

	if v := apiObject.CustomPluginArn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenLogDeliveryDescription(apiObject *awstypes.LogDeliveryDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.WorkerLogDelivery; v != nil {
		tfMap["worker_log_delivery"] = []interface{}{flattenWorkerLogDeliveryDescription(v)}
	}

	return tfMap
}

func flattenWorkerLogDeliveryDescription(apiObject *awstypes.WorkerLogDeliveryDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap[names.AttrCloudWatchLogs] = []interface{}{flattenCloudWatchLogsLogDeliveryDescription(v)}
	}

	if v := apiObject.Firehose; v != nil {
		tfMap["firehose"] = []interface{}{flattenFirehoseLogDeliveryDescription(v)}
	}

	if v := apiObject.S3; v != nil {
		tfMap["s3"] = []interface{}{flattenS3LogDeliveryDescription(v)}
	}

	return tfMap
}

func flattenCloudWatchLogsLogDeliveryDescription(apiObject *awstypes.CloudWatchLogsLogDeliveryDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFirehoseLogDeliveryDescription(apiObject *awstypes.FirehoseLogDeliveryDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.ToString(v)
	}

	return tfMap
}

func flattenS3LogDeliveryDescription(apiObject *awstypes.S3LogDeliveryDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return tfMap
}

func flattenWorkerConfigurationDescription(apiObject *awstypes.WorkerConfigurationDescription) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"revision": apiObject.Revision,
	}

	if v := apiObject.WorkerConfigurationArn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}
