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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
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
							Type:         schema.TypeList,
							MaxItems:     1,
							Optional:     true,
							ExactlyOneOf: []string{"capacity.0.autoscaling", "capacity.0.provisioned_capacity"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_worker_count": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"mcu_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1,
										ValidateFunc: validation.IntInSlice([]int{1, 2, 4, 8}),
									},
									"min_worker_count": {
										Type:     schema.TypeInt,
										Required: true,
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
						},
						"provisioned_capacity": {
							Type:         schema.TypeList,
							MaxItems:     1,
							Optional:     true,
							ExactlyOneOf: []string{"capacity.0.autoscaling", "capacity.0.provisioned_capacity"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mcu_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      1,
										ValidateFunc: validation.IntInSlice([]int{1, 2, 4, 8}),
									},
									"worker_count": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"configuration": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bootstrap_servers": {
										Type:     schema.TypeString,
										Required: true,
									},
									"vpc": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"security_group_ids": {
													Type: schema.TypeSet,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Required: true,
												},
												"subnet_ids": {
													Type: schema.TypeSet,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
													Required: true,
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
			"kafkaconnect_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_authentication": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      kafkaconnect.KafkaClusterClientAuthenticationTypeNone,
							ValidateFunc: validation.StringInSlice(kafkaconnect.KafkaClusterClientAuthenticationType_Values(), false),
						},
					},
				},
			},
			"encryption_in_transit": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      kafkaconnect.KafkaClusterEncryptionInTransitTypePlaintext,
							ValidateFunc: validation.StringInSlice(kafkaconnect.KafkaClusterEncryptionInTransitType_Values(), false),
						},
					},
				},
			},
			"custom_plugin": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"revision": {
							Type:     schema.TypeInt,
							Required: true,
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"log_group": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"firehose": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"delivery_stream": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"bucket": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"prefix": {
													Type:     schema.TypeString,
													Optional: true,
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
			"worker_configuration": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
							ForceNew:     true,
						},
						"revision": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	name := d.Get("name").(string)

	input := &kafkaconnect.CreateConnectorInput{
		ConnectorName:                    aws.String(name),
		ConnectorDescription:             aws.String(d.Get("description").(string)),
		KafkaConnectVersion:              aws.String(d.Get("kafkaconnect_version").(string)),
		ServiceExecutionRoleArn:          aws.String(d.Get("service_execution_role_arn").(string)),
		Capacity:                         expandCapacity(d.Get("capacity").([]interface{})),
		ConnectorConfiguration:           expandConnectorConfiguration(d.Get("configuration").(map[string]interface{})),
		KafkaCluster:                     expandKafkaCluster(d.Get("kafka_cluster").([]interface{})),
		KafkaClusterClientAuthentication: expandKafkaClientAuthentication(d.Get("client_authentication").([]interface{})),
		KafkaClusterEncryptionInTransit:  expandKafkaEncryptionInTransit(d.Get("encryption_in_transit").([]interface{})),
		Plugins:                          expandPlugins(d.Get("custom_plugin").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("log_delivery"); ok {
		input.LogDelivery = expandLogDelivery(v.([]interface{}))
	}

	if v, ok := d.GetOk("worker_configuration"); ok {
		input.WorkerConfiguration = expandWorkerConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateConnectorWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating MSK Connect Connector (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectorArn))

	_, err = waitConnectorCreatedWithContext(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("error waiting for MSK Connect Connector (%s) create: %s", d.Id(), err)
	}

	return resourceConnectorRead(ctx, d, meta)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	connector, err := FindConnectorByARN(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] MSK Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading MSK Connector (%s): %s", d.Id(), err)
	}

	_ = d.Set("arn", connector.ConnectorArn)
	_ = d.Set("description", connector.ConnectorDescription)
	_ = d.Set("name", connector.ConnectorName)
	_ = d.Set("state", connector.ConnectorState)
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

	if err := d.Set("client_authentication", flattenKafkaClientAuthentication(connector.KafkaClusterClientAuthentication)); err != nil {
		return diag.Errorf("error setting client_authentication: %s", err)
	}

	if err := d.Set("encryption_in_transit", flattenKafkaEncryptionInTransit(connector.KafkaClusterEncryptionInTransit)); err != nil {
		return diag.Errorf("error setting encryption_in_transit: %s", err)
	}

	if err := d.Set("encryption_in_transit", flattenKafkaEncryptionInTransit(connector.KafkaClusterEncryptionInTransit)); err != nil {
		return diag.Errorf("error setting encryption_in_transit: %s", err)
	}

	if err := d.Set("custom_plugin", flattenPlugins(connector.Plugins)); err != nil {
		return diag.Errorf("error setting custom_plugin: %s", err)
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

	if d.HasChanges("capacity.0.autoscaling", "capacity.0.provisioned_capacity") {
		input := &kafkaconnect.UpdateConnectorInput{
			Capacity:       expandCapacityUpdate(d.Get("capacity").([]interface{})),
			ConnectorArn:   aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("version").(string)),
		}

		output, err := conn.UpdateConnectorWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("error updating MSK Kafka Connector (%s) capacity: %s", d.Id(), err)
		}

		connectorARN := aws.StringValue(output.ConnectorArn)

		_, err = waitConnectorOperationCompletedWithContext(ctx, conn, connectorARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("error waiting for MSK Kafka Connector (%s) operation (%s): %s", d.Id(), connectorARN, err)
		}
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

	_, err = waitConnectorDeletedWithContext(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("error waiting for MSK Connect Connector (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func expandCapacity(tfList []interface{}) *kafkaconnect.Capacity {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.Capacity{
		AutoScaling:         expandAutoscaling(tfMap["autoscaling"].([]interface{})),
		ProvisionedCapacity: expandProvisionedCapacity(tfMap["provisioned_capacity"].([]interface{})),
	}
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

func expandAutoscaling(tfList []interface{}) *kafkaconnect.AutoScaling {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.AutoScaling{
		MaxWorkerCount: aws.Int64(int64(tfMap["max_worker_count"].(int))),
		McuCount:       aws.Int64(int64(tfMap["mcu_count"].(int))),
		MinWorkerCount: aws.Int64(int64(tfMap["min_worker_count"].(int))),
		ScaleInPolicy:  expandScaleInPolicy(tfMap["scale_in_policy"].([]interface{})),
		ScaleOutPolicy: expandScaleOutPolicy(tfMap["scale_out_policy"].([]interface{})),
	}
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

func expandScaleInPolicy(tfList []interface{}) *kafkaconnect.ScaleInPolicy {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleInPolicy{
		CpuUtilizationPercentage: aws.Int64(int64(tfMap["cpu_utilization_percentage"].(int))),
	}
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

func expandScaleOutPolicy(tfList []interface{}) *kafkaconnect.ScaleOutPolicy {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleOutPolicy{
		CpuUtilizationPercentage: aws.Int64(int64(tfMap["cpu_utilization_percentage"].(int))),
	}
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

func expandProvisionedCapacity(tfList []interface{}) *kafkaconnect.ProvisionedCapacity {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ProvisionedCapacity{
		McuCount:    aws.Int64(int64(tfMap["mcu_count"].(int))),
		WorkerCount: aws.Int64(int64(tfMap["worker_count"].(int))),
	}
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

func expandCapacityUpdate(tfList []interface{}) *kafkaconnect.CapacityUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.CapacityUpdate{
		AutoScaling:         expandAutoscalingUpdate(tfMap["autoscaling"].([]interface{})),
		ProvisionedCapacity: expandProvisionedCapacityUpdate(tfMap["provisioned_capacity"].([]interface{})),
	}
}

func expandAutoscalingUpdate(tfList []interface{}) *kafkaconnect.AutoScalingUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.AutoScalingUpdate{
		MaxWorkerCount: aws.Int64(int64(tfMap["max_worker_count"].(int))),
		McuCount:       aws.Int64(int64(tfMap["mcu_count"].(int))),
		MinWorkerCount: aws.Int64(int64(tfMap["min_worker_count"].(int))),
		ScaleInPolicy:  expandScaleInPolicyUpdate(tfMap["scale_in_policy"].([]interface{})),
		ScaleOutPolicy: expandScaleOutPolicyUpdate(tfMap["scale_out_policy"].([]interface{})),
	}
}

func expandScaleInPolicyUpdate(tfList []interface{}) *kafkaconnect.ScaleInPolicyUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleInPolicyUpdate{
		CpuUtilizationPercentage: aws.Int64(int64(tfMap["cpu_utilization_percentage"].(int))),
	}
}

func expandScaleOutPolicyUpdate(tfList []interface{}) *kafkaconnect.ScaleOutPolicyUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleOutPolicyUpdate{
		CpuUtilizationPercentage: aws.Int64(int64(tfMap["cpu_utilization_percentage"].(int))),
	}
}

func expandProvisionedCapacityUpdate(tfList []interface{}) *kafkaconnect.ProvisionedCapacityUpdate {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ProvisionedCapacityUpdate{
		McuCount:    aws.Int64(int64(tfMap["mcu_count"].(int))),
		WorkerCount: aws.Int64(int64(tfMap["worker_count"].(int))),
	}
}

func expandKafkaCluster(tfList []interface{}) *kafkaconnect.KafkaCluster {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.KafkaCluster{
		ApacheKafkaCluster: expandApacheKafkaCluster(tfMap["apache_kafka_cluster"].([]interface{})),
	}
}

func flattenKafkaCluster(kafkaCluster *kafkaconnect.KafkaClusterDescription) []interface{} {
	if kafkaCluster == nil {
		return []interface{}{}
	}

	return []interface{}{
		flattenApacheKafkaCluster(kafkaCluster.ApacheKafkaCluster),
	}
}

func expandApacheKafkaCluster(tfList []interface{}) *kafkaconnect.ApacheKafkaCluster {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.ApacheKafkaCluster{
		BootstrapServers: aws.String(tfMap["bootstrap_servers"].(string)),
		Vpc:              expandVpc(tfMap["vpc"].([]interface{})),
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

func expandVpc(tfList []interface{}) *kafkaconnect.Vpc {
	if len(tfList) == 0 {
		return nil
	}
	tfMap := tfList[0].(map[string]interface{})
	subnetsList := tfMap["subnet_ids"].(*schema.Set).List()
	subnets := make([]string, len(subnetsList))

	for i, subnet := range subnetsList {
		subnets[i] = subnet.(string)
	}

	sgsList := tfMap["security_group_ids"].(*schema.Set).List()
	securityGroups := make([]string, len(sgsList))

	for i, sg := range sgsList {
		securityGroups[i] = sg.(string)
	}

	return &kafkaconnect.Vpc{
		SecurityGroups: aws.StringSlice(securityGroups),
		Subnets:        aws.StringSlice(subnets),
	}
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
		"subnet_ids":         subnetIds,
		"security_group_ids": securityGroupIds,
	}
}

func expandKafkaClientAuthentication(clientAuthList []interface{}) *kafkaconnect.KafkaClusterClientAuthentication {
	if len(clientAuthList) == 0 {
		return nil
	}

	clientAuthMap := clientAuthList[0].(map[string]interface{})

	return &kafkaconnect.KafkaClusterClientAuthentication{
		AuthenticationType: aws.String(clientAuthMap["authentication_type"].(string)),
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

func expandKafkaEncryptionInTransit(encryptionList []interface{}) *kafkaconnect.KafkaClusterEncryptionInTransit {
	if len(encryptionList) == 0 {
		return nil
	}

	encryptionMap := encryptionList[0].(map[string]interface{})

	return &kafkaconnect.KafkaClusterEncryptionInTransit{
		EncryptionType: aws.String(encryptionMap["encryption_type"].(string)),
	}
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

func expandPlugins(tfList []interface{}) []*kafkaconnect.Plugin {

	pluginList := make([]*kafkaconnect.Plugin, 0)

	for _, plugin := range tfList {
		pluginMap := plugin.(map[string]interface{})

		customPlugin := &kafkaconnect.Plugin{
			CustomPlugin: &kafkaconnect.CustomPlugin{
				CustomPluginArn: aws.String(pluginMap["arn"].(string)),
				Revision:        aws.Int64(int64(pluginMap["revision"].(int))),
			},
		}
		pluginList = append(pluginList, customPlugin)
	}

	return pluginList
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

func expandConnectorConfiguration(tfMap map[string]interface{}) map[string]*string {
	configMap := make(map[string]*string)

	for k, v := range tfMap {
		configMap[k] = aws.String(v.(string))
	}

	return configMap
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

func expandLogDelivery(tfList []interface{}) *kafkaconnect.LogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.LogDelivery{
		WorkerLogDelivery: expandWorkerLogDelivery(tfMap["worker_log_delivery"].([]interface{})),
	}
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

func expandWorkerLogDelivery(tfList []interface{}) *kafkaconnect.WorkerLogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.WorkerLogDelivery{
		CloudWatchLogs: expandCloudwatchLogDelivery(tfMap["cloudwatch"].([]interface{})),
		Firehose:       expandFirehoseLogDelivery(tfMap["firehose"].([]interface{})),
		S3:             expandS3LogDelivery(tfMap["s3"].([]interface{})),
	}
}

func flattenWorkerLogDelivery(delivery *kafkaconnect.WorkerLogDeliveryDescription) []interface{} {
	if delivery == nil {
		return nil
	}

	m := map[string]interface{}{
		"cloudwatch": flattenCloudWatchLogDelivery(delivery.CloudWatchLogs),
		"firehose":   flattenFirehoseLogDelivery(delivery.Firehose),
		"s3":         flattenS3LogDelivery(delivery.S3),
	}

	return []interface{}{m}
}

func expandCloudwatchLogDelivery(tfList []interface{}) *kafkaconnect.CloudWatchLogsLogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	logDelivery := &kafkaconnect.CloudWatchLogsLogDelivery{
		Enabled: aws.Bool(tfMap["enabled"].(bool)),
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		logDelivery.LogGroup = aws.String(v)
	}

	return logDelivery
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

func expandFirehoseLogDelivery(tfList []interface{}) *kafkaconnect.FirehoseLogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	logDelivery := &kafkaconnect.FirehoseLogDelivery{
		Enabled: aws.Bool(tfMap["enabled"].(bool)),
	}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		logDelivery.DeliveryStream = aws.String(v)
	}

	return logDelivery
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

func expandS3LogDelivery(tfList []interface{}) *kafkaconnect.S3LogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	logDelivery := &kafkaconnect.S3LogDelivery{
		Enabled: aws.Bool(tfMap["enabled"].(bool)),
	}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		logDelivery.Bucket = aws.String(v)
	}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		logDelivery.Prefix = aws.String(v)
	}

	return logDelivery
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

func expandWorkerConfiguration(tfList []interface{}) *kafkaconnect.WorkerConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	workerConfiguration := &kafkaconnect.WorkerConfiguration{
		Revision:               aws.Int64(tfMap["revision"].(int64)),
		WorkerConfigurationArn: aws.String(tfMap["arn"].(string)),
	}

	return workerConfiguration
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
