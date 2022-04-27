package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("kafka_version", func(_ context.Context, old, new, meta interface{}) bool {
				return new.(string) < old.(string) // TODO: Use gversion; "10" < "2"
			}),
			customizeDiffValidateClientAuthentication,
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
						"ebs_volume_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 16384),
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
					},
				},
			},
			"client_authentication": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sasl": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"iam": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"scram": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"tls": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_authority_arns": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
						},
						"unauthenticated": {
							Type:     schema.TypeBool,
							Optional: true,
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
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_broker": {
										Type:         schema.TypeString,
										Optional:     true,
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
												"delivery_stream": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
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
												"bucket": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"enabled": {
													Type:     schema.TypeBool,
													Required: true,
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
		ClusterName:         aws.String(name),
		KafkaVersion:        aws.String(d.Get("kafka_version").(string)),
		NumberOfBrokerNodes: aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		Tags:                Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("broker_node_group_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BrokerNodeGroupInfo = expandBrokerNodeGroupInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("client_authentication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ClientAuthentication = expandClientAuthentication(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("configuration_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConfigurationInfo = expandConfigurationInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("encryption_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionInfo = expandEncryptionInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("enhanced_monitoring"); ok {
		input.EnhancedMonitoring = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingInfo = expandLoggingInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("open_monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OpenMonitoring = expandOpenMonitoringInfo(v.([]interface{})[0].(map[string]interface{}))
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

	if d.HasChanges("enhanced_monitoring", "logging_info", "open_monitoring") {
		input := &kafka.UpdateMonitoringInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			EnhancedMonitoring: aws.String(d.Get("enhanced_monitoring").(string)),
		}

		if v, ok := d.GetOk("logging_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.LoggingInfo = expandLoggingInfo(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("open_monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OpenMonitoring = expandOpenMonitoringInfo(v.([]interface{})[0].(map[string]interface{}))
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
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
		}

		if v, ok := d.GetOk("configuration_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ConfigurationInfo = expandConfigurationInfo(v.([]interface{})[0].(map[string]interface{}))
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
			if v, ok := d.GetOk("configuration_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ConfigurationInfo = expandConfigurationInfo(v.([]interface{})[0].(map[string]interface{}))
			}
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

	if d.HasChanges("encryption_info", "client_authentication") {
		input := &kafka.UpdateSecurityInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
		}

		if d.HasChange("client_authentication") {
			if v, ok := d.GetOk("client_authentication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ClientAuthentication = expandClientAuthentication(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("encryption_info") {
			if v, ok := d.GetOk("encryption_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.EncryptionInfo = expandEncryptionInfo(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		output, err := conn.UpdateSecurity(input)

		if err != nil {
			return fmt.Errorf("error updating MSK Cluster (%s) security: %w", d.Id(), err)
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

func customizeDiffValidateClientAuthentication(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	iam := diff.Get("client_authentication.0.sasl.0.iam").(bool)
	scram := diff.Get("client_authentication.0.sasl.0.scram").(bool)
	mtls := diff.Get("client_authentication.0.tls.0.enabled").(bool)
	unauth := diff.Get("client_authentication.0.unauthenticated").(bool)

	// at least one authentication option should be enabled
	if !iam && !scram && !mtls && !unauth {
		return errors.New(`at least one client-authentication option must be enabled`)
	}

	arns := diff.Get("client_authentication.0.tls.0.certificate_authority_arns").(*schema.Set)

	// mutual tls requires certificate authority arns to be set
	if mtls {
		if arns == nil || arns.Len() == 0 {
			return errors.New(`certificate authority ARNs must be specified to enable client authentication using TLS`)
		}
	}

	// tls must be enabled if certificate arns are provided
	if !mtls && arns != nil && arns.Len() > 0 {
		return errors.New(`certificate authority ARNs must be empty to disable client authentication using TLS`)
	}

	cbe := diff.Get("encryption_info.0.encryption_in_transit.0.client_broker").(string)
	ice := diff.Get("encryption_info.0.encryption_in_transit.0.in_cluster").(bool)

	// if plaintext authentication is enabled then unauthenticated access must be enabled
	if cbe == kafka.ClientBrokerTlsPlaintext && !unauth {
		return errors.New(`unauthenticated access must be set to use PLAINTEXT client-broker encryption options`)
	}

	// scram/iam & mtls all require in-cluster encryption and TLS encryption to be enabled
	if iam || scram || mtls {
		if !ice {
			return errors.New(`sasl/scram sasl/iam and mTLS authentication requires in-cluster encryption to be enabled`)
		}
		if cbe != kafka.ClientBrokerTls && cbe != kafka.ClientBrokerTlsPlaintext {
			return errors.New(`sasl/scram sasl/iam and mTLS requires TLS or TLS_PLAINTEXT client-broker encryption options`)
		}
	}
	return nil
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
		"sasl":            flattenSasl(ca.Sasl),
		"tls":             flattenTls(ca.Tls),
		"unauthenticated": flattenUnauthenticated(ca.Unauthenticated),
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
		"enabled":                    aws.BoolValue(tls.Enabled),
	}

	return []map[string]interface{}{m}
}

func flattenUnauthenticated(uauth *kafka.Unauthenticated) bool {
	if uauth == nil {
		return false
	}

	return aws.BoolValue(uauth.Enabled)
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

func expandBrokerNodeGroupInfo(tfMap map[string]interface{}) *kafka.BrokerNodeGroupInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.BrokerNodeGroupInfo{}

	if v, ok := tfMap["az_distribution"].(string); ok && v != "" {
		apiObject.BrokerAZDistribution = aws.String(v)
	}

	if v, ok := tfMap["client_subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ClientSubnets = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["instance_type"].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["ebs_volume_size"].(int); ok && v != 0 {
		apiObject.StorageInfo = &kafka.StorageInfo{
			EbsStorageInfo: &kafka.EBSStorageInfo{
				VolumeSize: aws.Int64(int64(v)),
			},
		}
	}

	return apiObject
}

func expandClientAuthentication(tfMap map[string]interface{}) *kafka.ClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.ClientAuthentication{}

	if v, ok := tfMap["sasl"].([]interface{}); ok && len(v) > 0 {
		apiObject.Sasl = expandSasl(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["tls"].([]interface{}); ok && len(v) > 0 {
		apiObject.Tls = expandTls(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["unauthenticated"].(bool); ok {
		apiObject.Unauthenticated = &kafka.Unauthenticated{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandSasl(tfMap map[string]interface{}) *kafka.Sasl {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.Sasl{}

	if v, ok := tfMap["iam"].(bool); ok {
		apiObject.Iam = &kafka.Iam{
			Enabled: aws.Bool(v),
		}
	}

	if v, ok := tfMap["scram"].(bool); ok {
		apiObject.Scram = &kafka.Scram{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandTls(tfMap map[string]interface{}) *kafka.Tls {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.Tls{}

	if v, ok := tfMap["certificate_authority_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CertificateAuthorityArnList = flex.ExpandStringSet(v)
	}

	return apiObject
}

func expandConfigurationInfo(tfMap map[string]interface{}) *kafka.ConfigurationInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.ConfigurationInfo{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = aws.Int64(int64(v))
	}

	return apiObject
}

func expandEncryptionInfo(tfMap map[string]interface{}) *kafka.EncryptionInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.EncryptionInfo{}

	if v, ok := tfMap["encryption_in_transit"].([]interface{}); ok && len(v) > 0 {
		apiObject.EncryptionInTransit = expandEncryptionInTransit(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["encryption_at_rest_kms_key_arn"].(string); ok && v != "" {
		apiObject.EncryptionAtRest = &kafka.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(v),
		}
	}

	return apiObject
}

func expandEncryptionInTransit(tfMap map[string]interface{}) *kafka.EncryptionInTransit {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.EncryptionInTransit{}

	if v, ok := tfMap["client_broker"].(string); ok && v != "" {
		apiObject.ClientBroker = aws.String(v)
	}

	if v, ok := tfMap["in_cluster"].(bool); ok {
		apiObject.InCluster = aws.Bool(v)
	}

	return apiObject
}

func expandLoggingInfo(tfMap map[string]interface{}) *kafka.LoggingInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.LoggingInfo{}

	if v, ok := tfMap["broker_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.BrokerLogs = expandBrokerLogs(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandBrokerLogs(tfMap map[string]interface{}) *kafka.BrokerLogs {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.BrokerLogs{}

	if v, ok := tfMap["broker_logs"].([]interface{}); ok && len(v) > 0 {
		apiObject.CloudWatchLogs = expandCloudWatchLogs(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["firehose"].([]interface{}); ok && len(v) > 0 {
		apiObject.Firehose = expandFirehose(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		apiObject.S3 = expandS3(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogs(tfMap map[string]interface{}) *kafka.CloudWatchLogs {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.CloudWatchLogs{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		apiObject.LogGroup = aws.String(v)
	}

	return apiObject
}

func expandFirehose(tfMap map[string]interface{}) *kafka.Firehose {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.Firehose{}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		apiObject.DeliveryStream = aws.String(v)
	}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandS3(tfMap map[string]interface{}) *kafka.S3 {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.S3{}

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

func expandOpenMonitoringInfo(tfMap map[string]interface{}) *kafka.OpenMonitoringInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.OpenMonitoringInfo{}

	if v, ok := tfMap["prometheus"].([]interface{}); ok && len(v) > 0 {
		apiObject.Prometheus = expandPrometheusInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrometheusInfo(tfMap map[string]interface{}) *kafka.PrometheusInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.PrometheusInfo{}

	if v, ok := tfMap["jmx_exporter"].([]interface{}); ok && len(v) > 0 {
		apiObject.JmxExporter = expandJmxExporterInfo(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["node_exporter"].([]interface{}); ok && len(v) > 0 {
		apiObject.NodeExporter = expandNodeExporterInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandJmxExporterInfo(tfMap map[string]interface{}) *kafka.JmxExporterInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.JmxExporterInfo{}

	if v, ok := tfMap["enabled_in_broker"].(bool); ok {
		apiObject.EnabledInBroker = aws.Bool(v)
	}

	return apiObject
}

func expandNodeExporterInfo(tfMap map[string]interface{}) *kafka.NodeExporterInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.NodeExporterInfo{}

	if v, ok := tfMap["enabled_in_broker"].(bool); ok {
		apiObject.EnabledInBroker = aws.Bool(v)
	}

	return apiObject
}
