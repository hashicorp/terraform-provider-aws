package kafkaconnect

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorCreate,
		ReadWithoutTimeout:   resourceConnectorRead,
		UpdateWithoutTimeout: resourceConnectorUpdate,
		DeleteWithoutTimeout: resourceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"description": {
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
												"security_groups": {
													Type:     schema.TypeSet,
													Required: true,
													ForceNew: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"subnets": {
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
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      kafkaconnect.KafkaClusterClientAuthenticationTypeNone,
							ValidateFunc: validation.StringInSlice(kafkaconnect.KafkaClusterClientAuthenticationType_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      kafkaconnect.KafkaClusterEncryptionInTransitTypePlaintext,
							ValidateFunc: validation.StringInSlice(kafkaconnect.KafkaClusterEncryptionInTransitType_Values(), false),
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
									"cloudwatch_logs": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
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
												"enabled": {
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
												"bucket": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
													ForceNew: true,
												},
												"prefix": {
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
			"name": {
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
									"arn": {
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
			"version": {
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
						"arn": {
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
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	name := d.Get("name").(string)
	input := &kafkaconnect.CreateConnectorInput{
		Capacity:                         expandCapacity(d.Get("capacity").([]interface{})[0].(map[string]interface{})),
		ConnectorConfiguration:           flex.ExpandStringMap(d.Get("connector_configuration").(map[string]interface{})),
		ConnectorName:                    aws.String(name),
		KafkaCluster:                     expandKafkaCluster(d.Get("kafka_cluster").([]interface{})[0].(map[string]interface{})),
		KafkaClusterClientAuthentication: expandKafkaClusterClientAuthentication(d.Get("kafka_cluster_client_authentication").([]interface{})[0].(map[string]interface{})),
		KafkaClusterEncryptionInTransit:  expandKafkaClusterEncryptionInTransit(d.Get("kafka_cluster_encryption_in_transit").([]interface{})[0].(map[string]interface{})),
		KafkaConnectVersion:              aws.String(d.Get("kafkaconnect_version").(string)),
		Plugins:                          expandPlugins(d.Get("plugin").(*schema.Set).List()),
		ServiceExecutionRoleArn:          aws.String(d.Get("service_execution_role_arn").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.ConnectorDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_delivery"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LogDelivery = expandLogDelivery(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("worker_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.WorkerConfiguration = expandWorkerConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.CreateConnectorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MSK Connect Connector (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectorArn))

	_, err = waitConnectorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("error waiting for MSK Connect Connector (%s) create: %s", d.Id(), err)
	}

	return resourceConnectorRead(ctx, d, meta)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	connector, err := FindConnectorByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connect Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MSK Connect Connector (%s): %s", d.Id(), err)
	}

	_ = d.Set("arn", connector.ConnectorArn)
	_ = d.Set("description", connector.ConnectorDescription)
	_ = d.Set("name", connector.ConnectorName)
	_ = d.Set("version", connector.CurrentVersion)
	_ = d.Set("kafkaconnect_version", connector.KafkaConnectVersion)
	_ = d.Set("service_execution_role_arn", connector.ServiceExecutionRoleArn)

	if err := d.Set("capacity", flattenConnectorCapacity(connector.Capacity)); err != nil {
		return diag.Errorf("error setting capacity: %s", err)
	}

	if err := d.Set("connector_configuration", flattenConnectorConfiguration(connector.ConnectorConfiguration)); err != nil {
		return diag.Errorf("error setting connector_configuration: %s", err)
	}

	if err := d.Set("kafka_cluster", flattenKafkaCluster(connector.KafkaCluster)); err != nil {
		return diag.Errorf("error setting kafka_cluster: %s", err)
	}

	if err := d.Set("kafka_cluster_client_authentication", flattenKafkaClientAuthentication(connector.KafkaClusterClientAuthentication)); err != nil {
		return diag.Errorf("error setting kafka_cluster_client_authentication: %s", err)
	}

	if err := d.Set("kafka_cluster_encryption_in_transit", flattenKafkaEncryptionInTransit(connector.KafkaClusterEncryptionInTransit)); err != nil {
		return diag.Errorf("error setting kafka_cluster_encryption_in_transit: %s", err)
	}

	if err := d.Set("plugin", flattenPlugins(connector.Plugins)); err != nil {
		return diag.Errorf("error setting plugin: %s", err)
	}

	if err := d.Set("log_delivery", flattenLogDelivery(connector.LogDelivery)); err != nil {
		return diag.Errorf("error setting log_delivery: %s", err)
	}

	if err := d.Set("worker_configuration", flattenWorkerConfiguration(connector.WorkerConfiguration)); err != nil {
		return diag.Errorf("error setting worker_configuration: %s", err)
	}

	return nil
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	input := &kafkaconnect.UpdateConnectorInput{
		Capacity:       expandCapacityUpdate(d.Get("capacity").([]interface{})[0].(map[string]interface{})),
		ConnectorArn:   aws.String(d.Id()),
		CurrentVersion: aws.String(d.Get("version").(string)),
	}

	_, err := conn.UpdateConnectorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating MSK Connect Connector (%s): %s", d.Id(), err)
	}

	_, err = waitConnectorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return diag.Errorf("error waiting for MSK Connect Connector (%s) update: %s", d.Id(), err)
	}

	return resourceConnectorRead(ctx, d, meta)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	log.Printf("[DEBUG] Deleting MSK Connect Connector: %s", d.Id())
	_, err := conn.DeleteConnectorWithContext(ctx, &kafkaconnect.DeleteConnectorInput{
		ConnectorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kafkaconnect.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting MSK Connect Connector (%s): %s", d.Id(), err)
	}

	_, err = waitConnectorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("error waiting for MSK Connect Connector (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func expandCapacity(tfMap map[string]interface{}) *kafkaconnect.Capacity {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.Capacity{}

	if v, ok := tfMap["autoscaling"].([]interface{}); ok && len(v) > 0 {
		apiObject.AutoScaling = expandAutoScaling(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["provisioned_capacity"].([]interface{}); ok && len(v) > 0 {
		apiObject.ProvisionedCapacity = expandProvisionedCapacity(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAutoScaling(tfMap map[string]interface{}) *kafkaconnect.AutoScaling {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.AutoScaling{}

	if v, ok := tfMap["max_worker_count"].(int); ok && v != 0 {
		apiObject.MaxWorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["mcu_count"].(int); ok && v != 0 {
		apiObject.McuCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_worker_count"].(int); ok && v != 0 {
		apiObject.MinWorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["scale_in_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScaleInPolicy = expandScaleInPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["scale_out_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScaleOutPolicy = expandScaleOutPolicy(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandScaleInPolicy(tfMap map[string]interface{}) *kafkaconnect.ScaleInPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ScaleInPolicy{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok && v != 0 {
		apiObject.CpuUtilizationPercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandScaleOutPolicy(tfMap map[string]interface{}) *kafkaconnect.ScaleOutPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ScaleOutPolicy{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok && v != 0 {
		apiObject.CpuUtilizationPercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandProvisionedCapacity(tfMap map[string]interface{}) *kafkaconnect.ProvisionedCapacity {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ProvisionedCapacity{}

	if v, ok := tfMap["mcu_count"].(int); ok && v != 0 {
		apiObject.McuCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["worker_count"].(int); ok && v != 0 {
		apiObject.WorkerCount = aws.Int64(int64(v))
	}

	return apiObject
}

func expandCapacityUpdate(tfMap map[string]interface{}) *kafkaconnect.CapacityUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.CapacityUpdate{}

	if v, ok := tfMap["autoscaling"].([]interface{}); ok && len(v) > 0 {
		apiObject.AutoScaling = expandAutoScalingUpdate(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["provisioned_capacity"].([]interface{}); ok && len(v) > 0 {
		apiObject.ProvisionedCapacity = expandProvisionedCapacityUpdate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAutoScalingUpdate(tfMap map[string]interface{}) *kafkaconnect.AutoScalingUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.AutoScalingUpdate{}

	if v, ok := tfMap["max_worker_count"].(int); ok {
		apiObject.MaxWorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["mcu_count"].(int); ok {
		apiObject.McuCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_worker_count"].(int); ok {
		apiObject.MinWorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["scale_in_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScaleInPolicy = expandScaleInPolicyUpdate(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["scale_out_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScaleOutPolicy = expandScaleOutPolicyUpdate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandScaleInPolicyUpdate(tfMap map[string]interface{}) *kafkaconnect.ScaleInPolicyUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ScaleInPolicyUpdate{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok {
		apiObject.CpuUtilizationPercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandScaleOutPolicyUpdate(tfMap map[string]interface{}) *kafkaconnect.ScaleOutPolicyUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ScaleOutPolicyUpdate{}

	if v, ok := tfMap["cpu_utilization_percentage"].(int); ok {
		apiObject.CpuUtilizationPercentage = aws.Int64(int64(v))
	}

	return apiObject
}

func expandProvisionedCapacityUpdate(tfMap map[string]interface{}) *kafkaconnect.ProvisionedCapacityUpdate {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ProvisionedCapacityUpdate{}

	if v, ok := tfMap["mcu_count"].(int); ok {
		apiObject.McuCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["worker_count"].(int); ok {
		apiObject.WorkerCount = aws.Int64(int64(v))
	}

	return apiObject
}

func expandKafkaCluster(tfMap map[string]interface{}) *kafkaconnect.KafkaCluster {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.KafkaCluster{}

	if v, ok := tfMap["apache_kafka_cluster"].([]interface{}); ok && len(v) > 0 {
		apiObject.ApacheKafkaCluster = expandApacheKafkaCluster(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandApacheKafkaCluster(tfMap map[string]interface{}) *kafkaconnect.ApacheKafkaCluster {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.ApacheKafkaCluster{}

	if v, ok := tfMap["bootstrap_servers"].(string); ok && v != "" {
		apiObject.BootstrapServers = aws.String(v)
	}

	if v, ok := tfMap["vpc"].([]interface{}); ok && len(v) > 0 {
		apiObject.Vpc = expandVpc(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandVpc(tfMap map[string]interface{}) *kafkaconnect.Vpc {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.Vpc{}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandKafkaClusterClientAuthentication(tfMap map[string]interface{}) *kafkaconnect.KafkaClusterClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.KafkaClusterClientAuthentication{}

	if v, ok := tfMap["authentication_type"].(string); ok && v != "" {
		apiObject.AuthenticationType = aws.String(v)
	}

	return apiObject
}

func expandKafkaClusterEncryptionInTransit(tfMap map[string]interface{}) *kafkaconnect.KafkaClusterEncryptionInTransit {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.KafkaClusterEncryptionInTransit{}

	if v, ok := tfMap["encryption_type"].(string); ok && v != "" {
		apiObject.EncryptionType = aws.String(v)
	}

	return apiObject
}

func expandPlugin(tfMap map[string]interface{}) *kafkaconnect.Plugin {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.Plugin{}

	if v, ok := tfMap["custom_plugin"].([]interface{}); ok && len(v) > 0 {
		apiObject.CustomPlugin = expandCustomPlugin(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPlugins(tfList []interface{}) []*kafkaconnect.Plugin {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*kafkaconnect.Plugin

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPlugin(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCustomPlugin(tfMap map[string]interface{}) *kafkaconnect.CustomPlugin {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.CustomPlugin{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.CustomPluginArn = aws.String(v)
	}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = aws.Int64(int64(v))
	}

	return apiObject
}

func expandLogDelivery(tfMap map[string]interface{}) *kafkaconnect.LogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.LogDelivery{}

	if v, ok := tfMap["worker_log_delivery"].([]interface{}); ok && len(v) > 0 {
		apiObject.WorkerLogDelivery = expandWorkerLogDelivery(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandWorkerLogDelivery(tfMap map[string]interface{}) *kafkaconnect.WorkerLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.WorkerLogDelivery{}

	if v, ok := tfMap["cloudwatch_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.CloudWatchLogs = expandCloudWatchLogsLogDelivery(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["firehose"].([]interface{}); ok && len(v) > 0 {
		apiObject.Firehose = expandFirehoseLogDelivery(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3 = expandS3LogDelivery(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogsLogDelivery(tfMap map[string]interface{}) *kafkaconnect.CloudWatchLogsLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.CloudWatchLogsLogDelivery{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		apiObject.LogGroup = aws.String(v)
	}

	return apiObject
}

func expandFirehoseLogDelivery(tfMap map[string]interface{}) *kafkaconnect.FirehoseLogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.FirehoseLogDelivery{}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		apiObject.DeliveryStream = aws.String(v)
	}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandS3LogDelivery(tfMap map[string]interface{}) *kafkaconnect.S3LogDelivery {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.S3LogDelivery{}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandWorkerConfiguration(tfMap map[string]interface{}) *kafkaconnect.WorkerConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafkaconnect.WorkerConfiguration{}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = aws.Int64(int64(v))
	}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.WorkerConfigurationArn = aws.String(v)
	}

	return apiObject
}

func flattenConnectorCapacity(capacity *kafkaconnect.CapacityDescription) []interface{} {
	if capacity == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"autoscaling":          flattenAutoScaling(capacity.AutoScaling),
		"provisioned_capacity": flattenProvisionedCapacity(capacity.ProvisionedCapacity),
	}

	return []interface{}{m}
}

func flattenAutoScaling(scaling *kafkaconnect.AutoScalingDescription) []interface{} {
	if scaling == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"max_worker_count": scaling.MaxWorkerCount,
		"mcu_count":        scaling.McuCount,
		"min_worker_count": scaling.MinWorkerCount,
	}

	if scaling.ScaleInPolicy != nil {
		tfMap["scale_in_policy"] = flattenScaleInPolicy(scaling.ScaleInPolicy)
	}
	if scaling.ScaleOutPolicy != nil {
		tfMap["scale_out_policy"] = flattenScaleOutPolicy(scaling.ScaleOutPolicy)
	}

	return []interface{}{tfMap}
}

func flattenScaleInPolicy(policy *kafkaconnect.ScaleInPolicyDescription) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cpu_utilization_percentage": policy.CpuUtilizationPercentage,
	}
	return []interface{}{m}
}

func flattenScaleOutPolicy(policy *kafkaconnect.ScaleOutPolicyDescription) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cpu_utilization_percentage": policy.CpuUtilizationPercentage,
	}
	return []interface{}{m}
}

func flattenProvisionedCapacity(capacity *kafkaconnect.ProvisionedCapacityDescription) []interface{} {
	if capacity == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mcu_count":    capacity.McuCount,
		"worker_count": capacity.WorkerCount,
	}

	return []interface{}{m}
}

func flattenKafkaCluster(kafkaCluster *kafkaconnect.KafkaClusterDescription) []interface{} {
	if kafkaCluster == nil {
		return []interface{}{}
	}

	return []interface{}{
		flattenApacheKafkaCluster(kafkaCluster.ApacheKafkaCluster),
	}
}

func flattenApacheKafkaCluster(apacheKafkaCluster *kafkaconnect.ApacheKafkaClusterDescription) interface{} {

	if apacheKafkaCluster == nil {
		return nil
	}

	m := flattenVpc(apacheKafkaCluster.Vpc)
	m["bootstrap_servers"] = apacheKafkaCluster.BootstrapServers

	return m
}

func flattenVpc(vpc *kafkaconnect.VpcDescription) map[string]interface{} {
	subnetIds := make([]string, len(vpc.Subnets))
	for i, subnet := range vpc.Subnets {
		subnetIds[i] = aws.StringValue(subnet)
	}

	securityGroupIds := make([]string, len(vpc.SecurityGroups))
	for i, securityGroup := range vpc.SecurityGroups {
		securityGroupIds[i] = aws.StringValue(securityGroup)
	}

	return map[string]interface{}{
		"subnets":         subnetIds,
		"security_groups": securityGroupIds,
	}
}

func flattenKafkaClientAuthentication(clientAuthentication *kafkaconnect.KafkaClusterClientAuthenticationDescription) []interface{} {
	if clientAuthentication == nil {
		return nil
	}

	clientAuthMap := map[string]interface{}{
		"authentication_type": clientAuthentication.AuthenticationType,
	}

	return []interface{}{clientAuthMap}
}

func flattenKafkaEncryptionInTransit(encryptionInTransit *kafkaconnect.KafkaClusterEncryptionInTransitDescription) []interface{} {
	if encryptionInTransit == nil {
		return nil
	}

	encryptionMap := map[string]interface{}{
		"encryption_type": encryptionInTransit.EncryptionType,
	}
	return []interface{}{encryptionMap}
}

func PluginHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["arn"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["revision"].(int64)))
	return create.StringHashcode(buf.String())
}

func flattenPlugins(plugins []*kafkaconnect.PluginDescription) *schema.Set {
	var s []interface{}
	for _, plugin := range plugins {
		s = append(s, flattenPlugin(plugin))
	}
	return schema.NewSet(PluginHash, s)
}

func flattenPlugin(plugin *kafkaconnect.PluginDescription) map[string]interface{} {
	m := make(map[string]interface{})
	m["arn"] = aws.StringValue(plugin.CustomPlugin.CustomPluginArn)
	m["revision"] = aws.Int64Value(plugin.CustomPlugin.Revision)
	return m
}

func flattenConnectorConfiguration(configuration map[string]*string) map[string]interface{} {
	if len(configuration) == 0 {
		return nil
	}

	configMap := make(map[string]interface{})

	for k, v := range configuration {
		configMap[k] = v
	}

	return configMap
}

func flattenLogDelivery(delivery *kafkaconnect.LogDeliveryDescription) []interface{} {
	if delivery == nil {
		return nil
	}

	m := map[string]interface{}{
		"worker_log_delivery": flattenWorkerLogDelivery(delivery.WorkerLogDelivery),
	}

	return []interface{}{m}
}

func flattenWorkerLogDelivery(delivery *kafkaconnect.WorkerLogDeliveryDescription) []interface{} {
	if delivery == nil {
		return nil
	}

	m := map[string]interface{}{
		"cloudwatch_logs": flattenCloudWatchLogDelivery(delivery.CloudWatchLogs),
		"firehose":        flattenFirehoseLogDelivery(delivery.Firehose),
		"s3":              flattenS3LogDelivery(delivery.S3),
	}

	return []interface{}{m}
}

func flattenCloudWatchLogDelivery(cloudWatchLog *kafkaconnect.CloudWatchLogsLogDeliveryDescription) []interface{} {
	if cloudWatchLog == nil {
		return nil
	}

	m := map[string]interface{}{
		"enabled":   cloudWatchLog.Enabled,
		"log_group": cloudWatchLog.LogGroup,
	}
	return []interface{}{m}
}

func flattenFirehoseLogDelivery(firehose *kafkaconnect.FirehoseLogDeliveryDescription) []interface{} {
	if firehose == nil {
		return nil
	}

	m := map[string]interface{}{
		"enabled":         firehose.Enabled,
		"delivery_stream": firehose.DeliveryStream,
	}

	return []interface{}{m}
}

func flattenS3LogDelivery(s3 *kafkaconnect.S3LogDeliveryDescription) []interface{} {
	if s3 == nil {
		return nil
	}
	m := map[string]interface{}{
		"enabled": s3.Enabled,
		"bucket":  s3.Bucket,
		"prefix":  s3.Prefix,
	}

	return []interface{}{m}
}

func flattenWorkerConfiguration(configuration *kafkaconnect.WorkerConfigurationDescription) []interface{} {
	if configuration == nil {
		return nil
	}

	m := map[string]interface{}{
		"revision": configuration.Revision,
		"arn":      configuration.WorkerConfigurationArn,
	}

	return []interface{}{m}
}
