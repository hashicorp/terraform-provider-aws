package kafkaconnect

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceConnectorCreate,
		Read:   resourceConnectorRead,
		Update: resourceConnectorUpdate,
		Delete: resourceConnectorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
										Type:     schema.TypeInt,
										Required: true,
									},
									"min_worker_count": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"scale_in_policy": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu_utilization_percentage": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"scale_out_policy": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu_utilization_percentage": {
													Type:     schema.TypeInt,
													Required: true,
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
										Type:     schema.TypeInt,
										Required: true,
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
			"connector_configuration": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"kafka_cluster": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootstrap_servers": {
							Type:     schema.TypeString,
							Required: true,
						},
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
						"client_authentication": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"encryption_at_transit": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_plugin": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"revision": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				Required: true,
			},
			"service_execution_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"log_delivery": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
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
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
												},
												"log_group": {
													Type:     schema.TypeString,
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
			// "worker_configuration": {
			// 	Type: schema.TypeList,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"arn": {
			// 				Type:     schema.TypeString,
			// 				Required: true,
			// 			},
			// 			"revision": {
			// 				Type:     schema.TypeString,
			// 				Required: true,
			// 			},
			// 		},
			// 	},
			// 	MaxItems: 1,
			// 	Optional: true,
			// },
		},
	}
}

func resourceConnectorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	name := d.Get("name").(string)
	kafkaCluster := d.Get("kafka_cluster").([]interface{})

	input := &kafkaconnect.CreateConnectorInput{
		ConnectorName:                    aws.String(name),
		KafkaConnectVersion:              aws.String(d.Get("version").(string)),
		ServiceExecutionRoleArn:          aws.String(d.Get("service_execution_role_arn").(string)),
		Capacity:                         expandCapacity(d.Get("capacity").([]interface{})),
		ConnectorConfiguration:           expandConnectorConfiguration(d.Get("connector_configuration").(map[string]interface{})),
		KafkaCluster:                     expandKafkaCluster(kafkaCluster),
		KafkaClusterClientAuthentication: expandKafkaClientAuthentication(kafkaCluster),
		KafkaClusterEncryptionInTransit:  expandKafkaEncryptionInTransit(kafkaCluster),
		Plugins:                          expandPlugins(d.Get("custom_plugin").(*schema.Set)),
	}

	if v, ok := d.GetOk("log_delivery"); ok {
		input.LogDelivery = expandLogDelivery(v.([]interface{}))
	}

	output, err := conn.CreateConnector(input)

	if err != nil {
		return fmt.Errorf("error creating MSK Connect Connector (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ConnectorArn))

	_, err = waitConnectorCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for MSK Connect Connector (%s) create: %w", d.Id(), err)
	}

	return resourceConnectorRead(d, meta)
}

func resourceConnectorRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceConnectorUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceConnectorRead(d, meta)
}

func resourceConnectorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	log.Printf("[DEBUG] Deleting MSK Connect Connector: %s", d.Id())
	_, err := conn.DeleteConnector(&kafkaconnect.DeleteConnectorInput{
		ConnectorArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kafkaconnect.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting MSK Connect Connector (%s): %w", d.Id(), err)
	}

	_, err = waitConnectorDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for MSK Connect Conector (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandCapacity(tfList []interface{}) *kafkaconnect.Capacity {
	if len(tfList) == 0 {
		return nil
	}

	capacity := tfList[0].(map[string]interface{})

	return &kafkaconnect.Capacity{
		AutoScaling:         expandAutoscaling(capacity["autoscaling"].([]interface{})),
		ProvisionedCapacity: expandProvisionedCapacity((capacity["provisioned_capacity"].([]interface{}))),
	}
}

func expandAutoscaling(tfList []interface{}) *kafkaconnect.AutoScaling {
	if len(tfList) == 0 {
		return nil
	}

	autoscaling := tfList[0].(map[string]interface{})

	return &kafkaconnect.AutoScaling{
		MaxWorkerCount: aws.Int64(int64(autoscaling["max_worker_count"].(int))),
		McuCount:       aws.Int64(int64(autoscaling["mcu_count"].(int))),
		MinWorkerCount: aws.Int64(int64(autoscaling["min_worker_count"].(int))),
		ScaleInPolicy:  expandScaleInPolicy(autoscaling["scale_in_policy"].([]interface{})),
		ScaleOutPolicy: expandScaleOutPolicy(autoscaling["scale_out_policy"].([]interface{})),
	}
}

func expandScaleInPolicy(tfList []interface{}) *kafkaconnect.ScaleInPolicy {
	if len(tfList) == 0 {
		return nil
	}

	policy := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleInPolicy{
		CpuUtilizationPercentage: aws.Int64(int64(policy["cpu_utilization_percentage"].(int))),
	}
}

func expandScaleOutPolicy(tfList []interface{}) *kafkaconnect.ScaleOutPolicy {
	if len(tfList) == 0 {
		return nil
	}

	policy := tfList[0].(map[string]interface{})

	return &kafkaconnect.ScaleOutPolicy{
		CpuUtilizationPercentage: aws.Int64(int64(policy["cpu_utilization_percentage"].(int))),
	}
}

func expandProvisionedCapacity(tfList []interface{}) *kafkaconnect.ProvisionedCapacity {
	if len(tfList) == 0 {
		return nil
	}

	capacity := tfList[0].(map[string]interface{})

	return &kafkaconnect.ProvisionedCapacity{
		McuCount:    aws.Int64(int64(capacity["mcu_count"].(int))),
		WorkerCount: aws.Int64(int64(capacity["worker_count"].(int))),
	}
}

func expandKafkaCluster(tfList []interface{}) *kafkaconnect.KafkaCluster {
	if len(tfList) == 0 {
		return nil
	}

	return &kafkaconnect.KafkaCluster{
		ApacheKafkaCluster: expandApacheKafkaCluster(tfList[0]),
	}
}

func expandApacheKafkaCluster(tfList interface{}) *kafkaconnect.ApacheKafkaCluster {
	cluster := tfList.(map[string]interface{})

	return &kafkaconnect.ApacheKafkaCluster{
		BootstrapServers: aws.String(cluster["bootstrap_servers"].(string)),
		Vpc:              expandVpc(cluster),
	}
}

func expandVpc(tfMap map[string]interface{}) *kafkaconnect.Vpc {
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

func expandKafkaClientAuthentication(tfList []interface{}) *kafkaconnect.KafkaClusterClientAuthentication {
	if len(tfList) == 0 {
		return nil
	}

	cluster := tfList[0].(map[string]interface{})

	clientAuthList := cluster["client_authentication"].([]interface{})
	if len(clientAuthList) == 0 {
		return nil
	}

	clientAuth := clientAuthList[0].(map[string]interface{})

	return &kafkaconnect.KafkaClusterClientAuthentication{
		AuthenticationType: aws.String(clientAuth["authentication_type"].(string)),
	}
}

func expandKafkaEncryptionInTransit(tfList []interface{}) *kafkaconnect.KafkaClusterEncryptionInTransit {
	if len(tfList) == 0 {
		return nil
	}

	cluster := tfList[0].(map[string]interface{})

	encryptionList := cluster["encryption_at_transit"].([]interface{})
	if len(encryptionList) == 0 {
		return nil
	}

	encryption := encryptionList[0].(map[string]interface{})

	return &kafkaconnect.KafkaClusterEncryptionInTransit{
		EncryptionType: aws.String(encryption["encryption_type"].(string)),
	}
}

func expandPlugins(tfSet *schema.Set) []*kafkaconnect.Plugin {
	if tfSet.Len() == 0 {
		return nil
	}

	plugins := make([]*kafkaconnect.Plugin, tfSet.Len())

	tfList := tfSet.List()

	for i, plugin := range tfList {
		pluginMap := plugin.(map[string]interface{})

		plugins[i] = &kafkaconnect.Plugin{
			CustomPlugin: &kafkaconnect.CustomPlugin{
				CustomPluginArn: aws.String(pluginMap["arn"].(string)),
				Revision:        aws.Int64(int64(pluginMap["revision"].(int))),
			},
		}
	}

	return plugins
}

func expandConnectorConfiguration(tfMap map[string]interface{}) map[string]*string {
	if len(tfMap) == 0 {
		return nil
	}

	config := make(map[string]*string)

	for k, v := range tfMap {
		config[k] = aws.String(v.(string))
	}

	return config
}

func expandLogDelivery(tfList []interface{}) *kafkaconnect.LogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	logDelivery := tfList[0].(map[string]interface{})

	return &kafkaconnect.LogDelivery{
		WorkerLogDelivery: expandWorkerLogDelivery(logDelivery["worker_log_delivery"].([]interface{})),
	}
}

func expandWorkerLogDelivery(tfList []interface{}) *kafkaconnect.WorkerLogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	return &kafkaconnect.WorkerLogDelivery{
		CloudWatchLogs: expandCloudwatchLogDelivery(tfMap["cloudwatch"].([]interface{})),
		Firehose: expandFirehoseLogDelivery(tfMap["firehose"].([]interface{})),
		S3: expandS3LogDelivery(tfMap["s3"].([]interface{})),
	}
}

func expandCloudwatchLogDelivery(tfList []interface{}) *kafkaconnect.CloudWatchLogsLogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	logDelivery := &kafkaconnect.CloudWatchLogsLogDelivery{
		Enabled:  aws.Bool(tfMap["enabled"].(bool)),
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		logDelivery.LogGroup = aws.String(v)
	}

	return logDelivery
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
