package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(clusterCreateDefaultTimeout),
			Update: schema.DefaultTimeout(clusterUpdateDefaultTimeout),
			Delete: schema.DefaultTimeout(clusterDeleteDefaultTimeout),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("kafka_version", func(_ context.Context, old, new, meta interface{}) bool {
				return new.(string) < old.(string)
			}),
			verify.SetTagsDiff,
		),
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"az_distribution": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      kafka.BrokerAZDistributionDefault,
							ValidateFunc: validation.StringInSlice(kafka.BrokerAZDistribution_Values(), false),
						},
						"client_subnets": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ebs_volume_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 16384),
						},
					},
				},
			},
			"client_authentication": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sasl": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scram": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
									"iam": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
									},
								},
							},
							ConflictsWith: []string{"client_authentication.0.tls"},
						},
						"tls": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_authority_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"configuration_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"revision": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"current_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				ForceNew:         true,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_at_rest_kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"encryption_in_transit": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							ForceNew:         true,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_broker": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										Default:      kafka.ClientBrokerTls,
										ValidateFunc: validation.StringInSlice(kafka.ClientBroker_Values(), false),
									},
									"in_cluster": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  true,
									},
								},
							},
						},
					},
				},
			},
			"enhanced_monitoring": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      kafka.EnhancedMonitoringDefault,
				ValidateFunc: validation.StringInSlice(kafka.EnhancedMonitoring_Values(), true),
			},
			"kafka_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"open_monitoring": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prometheus": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"jmx_exporter": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled_in_broker": {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"node_exporter": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enabled_in_broker": {
													Type:     schema.TypeBool,
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
			"logging_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker_logs": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloudwatch_logs": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
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
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
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
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"zookeeper_connect_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zookeeper_connect_string_tls": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("cluster_name").(string)
	input := &kafka.CreateClusterInput{
		BrokerNodeGroupInfo:  expandClusterBrokerNodeGroupInfo(d.Get("broker_node_group_info").([]interface{})),
		ClientAuthentication: expandClusterClientAuthentication(d.Get("client_authentication").([]interface{})),
		ClusterName:          aws.String(name),
		ConfigurationInfo:    expandClusterConfigurationInfo(d.Get("configuration_info").([]interface{})),
		EncryptionInfo:       expandClusterEncryptionInfo(d.Get("encryption_info").([]interface{})),
		EnhancedMonitoring:   aws.String(d.Get("enhanced_monitoring").(string)),
		KafkaVersion:         aws.String(d.Get("kafka_version").(string)),
		NumberOfBrokerNodes:  aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		OpenMonitoring:       expandOpenMonitoring(d.Get("open_monitoring").([]interface{})),
		LoggingInfo:          expandLoggingInfo(d.Get("logging_info").([]interface{})),
		Tags:                 Tags(tags.IgnoreAWS()),
	}

	output, err := conn.CreateCluster(input)

	if err != nil {
		return fmt.Errorf("error creating MSK Cluster (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	_, err = waitClusterCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for MSK Cluster (%s) create: %w", d.Id(), err)
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s): %w", d.Id(), err)
	}

	output, err := conn.GetBootstrapBrokers(&kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error reading MSK Cluster (%s) bootstrap brokers: %w", d.Id(), err)
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.StringValue(output.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringTls)))

	if err := d.Set("broker_node_group_info", flattenBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo)); err != nil {
		return fmt.Errorf("error setting broker_node_group_info: %w", err)
	}

	if err := d.Set("client_authentication", flattenClientAuthentication(cluster.ClientAuthentication)); err != nil {
		return fmt.Errorf("error setting configuration_info: %w", err)
	}

	d.Set("cluster_name", cluster.ClusterName)

	if err := d.Set("configuration_info", flattenConfigurationInfo(cluster.CurrentBrokerSoftwareInfo)); err != nil {
		return fmt.Errorf("error setting configuration_info: %w", err)
	}

	d.Set("current_version", cluster.CurrentVersion)
	d.Set("enhanced_monitoring", cluster.EnhancedMonitoring)

	if err := d.Set("encryption_info", flattenEncryptionInfo(cluster.EncryptionInfo)); err != nil {
		return fmt.Errorf("error setting encryption_info: %w", err)
	}

	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)

	if err := d.Set("logging_info", flattenLoggingInfo(cluster.LoggingInfo)); err != nil {
		return fmt.Errorf("error setting logging_info: %w", err)
	}

	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)

	if err := d.Set("open_monitoring", flattenOpenMonitoring(cluster.OpenMonitoring)); err != nil {
		return fmt.Errorf("error setting open_monitoring: %w", err)
	}

	d.Set("zookeeper_connect_string", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectString)))
	d.Set("zookeeper_connect_string_tls", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectStringTls)))

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	if d.HasChange("broker_node_group_info.0.ebs_volume_size") {
		input := &kafka.UpdateBrokerStorageInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
			TargetBrokerEBSVolumeInfo: []*kafka.BrokerEBSVolumeInfo{
				{
					KafkaBrokerNodeId: aws.String("All"),
					VolumeSizeGB:      aws.Int64(int64(d.Get("broker_node_group_info.0.ebs_volume_size").(int))),
				},
			},
		}

		output, err := conn.UpdateBrokerStorage(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker storage: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("broker_node_group_info.0.instance_type") {
		input := &kafka.UpdateBrokerTypeInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			TargetInstanceType: aws.String(d.Get("broker_node_group_info.0.instance_type").(string)),
		}

		output, err := conn.UpdateBrokerType(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker type: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("number_of_broker_nodes") {
		input := &kafka.UpdateBrokerCountInput{
			ClusterArn:                aws.String(d.Id()),
			CurrentVersion:            aws.String(d.Get("current_version").(string)),
			TargetNumberOfBrokerNodes: aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		}

		output, err := conn.UpdateBrokerCount(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) broker count: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChanges("enhanced_monitoring", "open_monitoring", "logging_info") {
		input := &kafka.UpdateMonitoringInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			EnhancedMonitoring: aws.String(d.Get("enhanced_monitoring").(string)),
			OpenMonitoring:     expandOpenMonitoring(d.Get("open_monitoring").([]interface{})),
			LoggingInfo:        expandLoggingInfo(d.Get("logging_info").([]interface{})),
		}

		output, err := conn.UpdateMonitoring(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) monitoring: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("configuration_info") && !d.HasChange("kafka_version") {
		input := &kafka.UpdateClusterConfigurationInput{
			ClusterArn:        aws.String(d.Id()),
			ConfigurationInfo: expandClusterConfigurationInfo(d.Get("configuration_info").([]interface{})),
			CurrentVersion:    aws.String(d.Get("current_version").(string)),
		}

		output, err := conn.UpdateClusterConfiguration(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) configuration: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("kafka_version") {
		input := &kafka.UpdateClusterKafkaVersionInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			TargetKafkaVersion: aws.String(d.Get("kafka_version").(string)),
		}

		if d.HasChange("configuration_info") {
			input.ConfigurationInfo = expandClusterConfigurationInfo(d.Get("configuration_info").([]interface{}))
		}

		output, err := conn.UpdateClusterKafkaVersion(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) kafka version: %w", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for MSK Cluster (%s) operation (%s): %w", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	log.Printf("[DEBUG] Deleting MSK Cluster: %s", d.Id())
	_, err := conn.DeleteCluster(&kafka.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting MSK Cluster (%s): %w", d.Id(), err)
	}

	_, err = waitClusterDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for MSK Cluster (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandClusterBrokerNodeGroupInfo(l []interface{}) *kafka.BrokerNodeGroupInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	bngi := &kafka.BrokerNodeGroupInfo{
		BrokerAZDistribution: aws.String(m["az_distribution"].(string)),
		ClientSubnets:        flex.ExpandStringSet(m["client_subnets"].(*schema.Set)),
		InstanceType:         aws.String(m["instance_type"].(string)),
		SecurityGroups:       flex.ExpandStringSet(m["security_groups"].(*schema.Set)),
		StorageInfo: &kafka.StorageInfo{
			EbsStorageInfo: &kafka.EBSStorageInfo{
				VolumeSize: aws.Int64(int64(m["ebs_volume_size"].(int))),
			},
		},
	}

	return bngi
}

func expandClusterClientAuthentication(l []interface{}) *kafka.ClientAuthentication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &kafka.ClientAuthentication{}

	if v, ok := m["sasl"].([]interface{}); ok {
		ca.Sasl = expandClusterSasl(v)
	}

	if v, ok := m["tls"].([]interface{}); ok {
		ca.Tls = expandClusterTls(v)
	}

	return ca
}

func expandClusterConfigurationInfo(l []interface{}) *kafka.ConfigurationInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ci := &kafka.ConfigurationInfo{
		Arn:      aws.String(m["arn"].(string)),
		Revision: aws.Int64(int64(m["revision"].(int))),
	}

	return ci
}

func expandClusterEncryptionInfo(l []interface{}) *kafka.EncryptionInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ei := &kafka.EncryptionInfo{
		EncryptionInTransit: expandClusterEncryptionInTransit(m["encryption_in_transit"].([]interface{})),
	}

	if v, ok := m["encryption_at_rest_kms_key_arn"]; ok && v.(string) != "" {
		ei.EncryptionAtRest = &kafka.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(v.(string)),
		}
	}

	return ei
}

func expandClusterEncryptionInTransit(l []interface{}) *kafka.EncryptionInTransit {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	eit := &kafka.EncryptionInTransit{
		ClientBroker: aws.String(m["client_broker"].(string)),
		InCluster:    aws.Bool(m["in_cluster"].(bool)),
	}

	return eit
}

func expandClusterSasl(l []interface{}) *kafka.Sasl {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sasl := &kafka.Sasl{}

	if v, ok := tfMap["scram"].(bool); ok {
		sasl.Scram = &kafka.Scram{
			Enabled: aws.Bool(v),
		}
	}

	if v, ok := tfMap["iam"].(bool); ok {
		sasl.Iam = &kafka.Iam{
			Enabled: aws.Bool(v),
		}
	}

	return sasl
}

func expandClusterTls(l []interface{}) *kafka.Tls {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	tls := &kafka.Tls{
		CertificateAuthorityArnList: flex.ExpandStringSet(m["certificate_authority_arns"].(*schema.Set)),
	}

	return tls
}

func expandOpenMonitoring(l []interface{}) *kafka.OpenMonitoringInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	openMonitoring := &kafka.OpenMonitoringInfo{
		Prometheus: expandOpenMonitoringPrometheus(m["prometheus"].([]interface{})),
	}

	return openMonitoring
}

func expandOpenMonitoringPrometheus(l []interface{}) *kafka.PrometheusInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	prometheus := &kafka.PrometheusInfo{
		JmxExporter:  expandOpenMonitoringPrometheusJmxExporter(m["jmx_exporter"].([]interface{})),
		NodeExporter: expandOpenMonitoringPrometheusNodeExporter(m["node_exporter"].([]interface{})),
	}

	return prometheus
}

func expandOpenMonitoringPrometheusJmxExporter(l []interface{}) *kafka.JmxExporterInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	jmxExporter := &kafka.JmxExporterInfo{
		EnabledInBroker: aws.Bool(m["enabled_in_broker"].(bool)),
	}

	return jmxExporter
}

func expandOpenMonitoringPrometheusNodeExporter(l []interface{}) *kafka.NodeExporterInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nodeExporter := &kafka.NodeExporterInfo{
		EnabledInBroker: aws.Bool(m["enabled_in_broker"].(bool)),
	}

	return nodeExporter
}

func expandLoggingInfo(l []interface{}) *kafka.LoggingInfo {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	loggingInfo := &kafka.LoggingInfo{
		BrokerLogs: expandLoggingInfoBrokerLogs(m["broker_logs"].([]interface{})),
	}

	return loggingInfo
}

func expandLoggingInfoBrokerLogs(l []interface{}) *kafka.BrokerLogs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	brokerLogs := &kafka.BrokerLogs{
		CloudWatchLogs: expandLoggingInfoBrokerLogsCloudWatchLogs(m["cloudwatch_logs"].([]interface{})),
		Firehose:       expandLoggingInfoBrokerLogsFirehose(m["firehose"].([]interface{})),
		S3:             expandLoggingInfoBrokerLogsS3(m["s3"].([]interface{})),
	}

	return brokerLogs
}

func expandLoggingInfoBrokerLogsCloudWatchLogs(l []interface{}) *kafka.CloudWatchLogs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	cloudWatchLogs := &kafka.CloudWatchLogs{
		Enabled:  aws.Bool(m["enabled"].(bool)),
		LogGroup: aws.String(m["log_group"].(string)),
	}

	return cloudWatchLogs
}

func expandLoggingInfoBrokerLogsFirehose(l []interface{}) *kafka.Firehose {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	firehose := &kafka.Firehose{
		Enabled:        aws.Bool(m["enabled"].(bool)),
		DeliveryStream: aws.String(m["delivery_stream"].(string)),
	}

	return firehose
}

func expandLoggingInfoBrokerLogsS3(l []interface{}) *kafka.S3 {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	s3 := &kafka.S3{
		Enabled: aws.Bool(m["enabled"].(bool)),
		Bucket:  aws.String(m["bucket"].(string)),
		Prefix:  aws.String(m["prefix"].(string)),
	}

	return s3
}

func flattenBrokerNodeGroupInfo(b *kafka.BrokerNodeGroupInfo) []map[string]interface{} {

	if b == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"az_distribution": aws.StringValue(b.BrokerAZDistribution),
		"client_subnets":  aws.StringValueSlice(b.ClientSubnets),
		"instance_type":   aws.StringValue(b.InstanceType),
		"security_groups": aws.StringValueSlice(b.SecurityGroups),
	}
	if b.StorageInfo != nil {
		if b.StorageInfo.EbsStorageInfo != nil {
			m["ebs_volume_size"] = int(aws.Int64Value(b.StorageInfo.EbsStorageInfo.VolumeSize))
		}
	}
	return []map[string]interface{}{m}
}

func flattenClientAuthentication(ca *kafka.ClientAuthentication) []map[string]interface{} {
	if ca == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"sasl": flattenSasl(ca.Sasl),
		"tls":  flattenTls(ca.Tls),
	}

	return []map[string]interface{}{m}
}

func flattenConfigurationInfo(bsi *kafka.BrokerSoftwareInfo) []map[string]interface{} {
	if bsi == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"arn":      aws.StringValue(bsi.ConfigurationArn),
		"revision": aws.Int64Value(bsi.ConfigurationRevision),
	}

	return []map[string]interface{}{m}
}

func flattenEncryptionInfo(e *kafka.EncryptionInfo) []map[string]interface{} {
	if e == nil || e.EncryptionAtRest == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"encryption_at_rest_kms_key_arn": aws.StringValue(e.EncryptionAtRest.DataVolumeKMSKeyId),
		"encryption_in_transit":          flattenEncryptionInTransit(e.EncryptionInTransit),
	}

	return []map[string]interface{}{m}
}

func flattenEncryptionInTransit(eit *kafka.EncryptionInTransit) []map[string]interface{} {
	if eit == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_broker": aws.StringValue(eit.ClientBroker),
		"in_cluster":    aws.BoolValue(eit.InCluster),
	}

	return []map[string]interface{}{m}
}

func flattenSasl(sasl *kafka.Sasl) []map[string]interface{} {
	if sasl == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"scram": flattenSaslScram(sasl.Scram),
		"iam":   flattenSaslIam(sasl.Iam),
	}

	return []map[string]interface{}{m}
}

func flattenSaslScram(scram *kafka.Scram) bool {
	if scram == nil {
		return false
	}

	return aws.BoolValue(scram.Enabled)
}

func flattenSaslIam(iam *kafka.Iam) bool {
	if iam == nil {
		return false
	}

	return aws.BoolValue(iam.Enabled)
}

func flattenTls(tls *kafka.Tls) []map[string]interface{} {
	if tls == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"certificate_authority_arns": aws.StringValueSlice(tls.CertificateAuthorityArnList),
	}

	return []map[string]interface{}{m}
}

func flattenOpenMonitoring(e *kafka.OpenMonitoring) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"prometheus": flattenOpenMonitoringPrometheus(e.Prometheus),
	}

	return []map[string]interface{}{m}
}

func flattenOpenMonitoringPrometheus(e *kafka.Prometheus) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"jmx_exporter":  flattenOpenMonitoringPrometheusJmxExporter(e.JmxExporter),
		"node_exporter": flattenOpenMonitoringPrometheusNodeExporter(e.NodeExporter),
	}

	return []map[string]interface{}{m}
}

func flattenOpenMonitoringPrometheusJmxExporter(e *kafka.JmxExporter) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled_in_broker": aws.BoolValue(e.EnabledInBroker),
	}

	return []map[string]interface{}{m}
}

func flattenOpenMonitoringPrometheusNodeExporter(e *kafka.NodeExporter) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled_in_broker": aws.BoolValue(e.EnabledInBroker),
	}

	return []map[string]interface{}{m}
}

func flattenLoggingInfo(e *kafka.LoggingInfo) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"broker_logs": flattenLoggingInfoBrokerLogs(e.BrokerLogs),
	}

	return []map[string]interface{}{m}
}

func flattenLoggingInfoBrokerLogs(e *kafka.BrokerLogs) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logs": flattenLoggingInfoBrokerLogsCloudWatchLogs(e.CloudWatchLogs),
		"firehose":        flattenLoggingInfoBrokerLogsFirehose(e.Firehose),
		"s3":              flattenLoggingInfoBrokerLogsS3(e.S3),
	}

	return []map[string]interface{}{m}
}

func flattenLoggingInfoBrokerLogsCloudWatchLogs(e *kafka.CloudWatchLogs) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":   aws.BoolValue(e.Enabled),
		"log_group": aws.StringValue(e.LogGroup),
	}

	return []map[string]interface{}{m}
}

func flattenLoggingInfoBrokerLogsFirehose(e *kafka.Firehose) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":         aws.BoolValue(e.Enabled),
		"delivery_stream": aws.StringValue(e.DeliveryStream),
	}

	return []map[string]interface{}{m}
}

func flattenLoggingInfoBrokerLogsS3(e *kafka.S3) []map[string]interface{} {
	if e == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(e.Enabled),
		"bucket":  aws.StringValue(e.Bucket),
		"prefix":  aws.StringValue(e.Prefix),
	}

	return []map[string]interface{}{m}
}
