package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

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
				return verify.SemVerLessThan(new.(string), old.(string))
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
			"bootstrap_brokers_public_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_public_tls": {
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
						"connectivity_info": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"public_access": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringInSlice(PublicAccessType_Values(), false),
												},
											},
										},
									},
								},
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
				Optional: true,
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
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
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
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	output, err := conn.CreateClusterWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MSK Cluster (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	_, err = waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for MSK Cluster (%s) create: %s", d.Id(), err)
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	cluster, err := FindClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MSK Cluster (%s): %s", d.Id(), err)
	}

	output, err := conn.GetBootstrapBrokersWithContext(ctx, &kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("reading MSK Cluster (%s) bootstrap brokers: %s", d.Id(), err)
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.StringValue(output.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_public_sasl_iam", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringPublicSaslIam)))
	d.Set("bootstrap_brokers_public_sasl_scram", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringPublicSaslScram)))
	d.Set("bootstrap_brokers_public_tls", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringPublicTls)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.StringValue(output.BootstrapBrokerStringTls)))

	if cluster.BrokerNodeGroupInfo != nil {
		if err := d.Set("broker_node_group_info", []interface{}{flattenBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo)}); err != nil {
			return diag.Errorf("setting broker_node_group_info: %s", err)
		}
	} else {
		d.Set("broker_node_group_info", nil)
	}

	if cluster.ClientAuthentication != nil {
		if err := d.Set("client_authentication", []interface{}{flattenClientAuthentication(cluster.ClientAuthentication)}); err != nil {
			return diag.Errorf("setting client_authentication: %s", err)
		}
	} else {
		d.Set("client_authentication", nil)
	}

	d.Set("cluster_name", cluster.ClusterName)

	if cluster.CurrentBrokerSoftwareInfo != nil {
		if err := d.Set("configuration_info", []interface{}{flattenBrokerSoftwareInfo(cluster.CurrentBrokerSoftwareInfo)}); err != nil {
			return diag.Errorf("setting configuration_info: %s", err)
		}
	} else {
		d.Set("configuration_info", nil)
	}

	d.Set("current_version", cluster.CurrentVersion)
	d.Set("enhanced_monitoring", cluster.EnhancedMonitoring)

	if cluster.EncryptionInfo != nil {
		if err := d.Set("encryption_info", []interface{}{flattenEncryptionInfo(cluster.EncryptionInfo)}); err != nil {
			return diag.Errorf("setting encryption_info: %s", err)
		}
	} else {
		d.Set("encryption_info", nil)
	}

	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)

	if cluster.LoggingInfo != nil {
		if err := d.Set("logging_info", []interface{}{flattenLoggingInfo(cluster.LoggingInfo)}); err != nil {
			return diag.Errorf("setting logging_info: %s", err)
		}
	} else {
		d.Set("logging_info", nil)
	}

	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)

	if cluster.OpenMonitoring != nil {
		if err := d.Set("open_monitoring", []interface{}{flattenOpenMonitoring(cluster.OpenMonitoring)}); err != nil {
			return diag.Errorf("setting open_monitoring: %s", err)
		}
	} else {
		d.Set("open_monitoring", nil)
	}

	d.Set("zookeeper_connect_string", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectString)))
	d.Set("zookeeper_connect_string_tls", SortEndpointsString(aws.StringValue(cluster.ZookeeperConnectStringTls)))

	tags := KeyValueTags(cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

		output, err := conn.UpdateBrokerStorageWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) broker storage: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("broker_node_group_info.0.connectivity_info") {
		input := &kafka.UpdateConnectivityInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
		}

		if v, ok := d.GetOk("broker_node_group_info.0.connectivity_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ConnectivityInfo = expandConnectivityInfo(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateConnectivityWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) broker connectivity: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("broker_node_group_info.0.instance_type") {
		input := &kafka.UpdateBrokerTypeInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			TargetInstanceType: aws.String(d.Get("broker_node_group_info.0.instance_type").(string)),
		}

		output, err := conn.UpdateBrokerTypeWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) broker type: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("number_of_broker_nodes") {
		input := &kafka.UpdateBrokerCountInput{
			ClusterArn:                aws.String(d.Id()),
			CurrentVersion:            aws.String(d.Get("current_version").(string)),
			TargetNumberOfBrokerNodes: aws.Int64(int64(d.Get("number_of_broker_nodes").(int))),
		}

		output, err := conn.UpdateBrokerCountWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) broker count: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
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

		output, err := conn.UpdateMonitoringWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) monitoring: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
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

		output, err := conn.UpdateClusterConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) configuration: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
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

		output, err := conn.UpdateClusterKafkaVersionWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) Kafka version: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
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
				if input.EncryptionInfo != nil {
					input.EncryptionInfo.EncryptionAtRest = nil // "Updating encryption-at-rest settings on your cluster is not currently supported."
					if input.EncryptionInfo.EncryptionInTransit != nil {
						input.EncryptionInfo.EncryptionInTransit.InCluster = nil // "Updating the inter-broker encryption setting on your cluster is not currently supported."
					}
				}
			}
		}

		output, err := conn.UpdateSecurityWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MSK Cluster (%s) security: %s", d.Id(), err)
		}

		clusterOperationARN := aws.StringValue(output.ClusterOperationArn)

		_, err = waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return diag.Errorf("waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating MSK Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConn

	log.Printf("[DEBUG] Deleting MSK Cluster: %s", d.Id())
	_, err := conn.DeleteClusterWithContext(ctx, &kafka.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting MSK Cluster (%s): %s", d.Id(), err)
	}

	_, err = waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("waiting for MSK Cluster (%s) delete: %s", d.Id(), err)
	}

	return nil
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

	if v, ok := tfMap["connectivity_info"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConnectivityInfo = expandConnectivityInfo(v[0].(map[string]interface{}))
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

func expandConnectivityInfo(tfMap map[string]interface{}) *kafka.ConnectivityInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.ConnectivityInfo{}

	if v, ok := tfMap["public_access"].([]interface{}); ok && len(v) > 0 {
		apiObject.PublicAccess = expandPublicAccess(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPublicAccess(tfMap map[string]interface{}) *kafka.PublicAccess {
	if tfMap == nil {
		return nil
	}

	apiObject := &kafka.PublicAccess{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
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
		apiObject.Enabled = aws.Bool(true)
	} else {
		apiObject.Enabled = aws.Bool(false)
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

	if v, ok := tfMap["cloudwatch_logs"].([]interface{}); ok && len(v) > 0 {
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

func flattenBrokerNodeGroupInfo(apiObject *kafka.BrokerNodeGroupInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BrokerAZDistribution; v != nil {
		tfMap["az_distribution"] = aws.StringValue(v)
	}

	if v := apiObject.ClientSubnets; v != nil {
		tfMap["client_subnets"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ConnectivityInfo; v != nil {
		tfMap["connectivity_info"] = []interface{}{flattenConnectivityInfo(v)}
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap["instance_type"] = aws.StringValue(v)
	}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap["security_groups"] = aws.StringValueSlice(v)
	}

	if v := apiObject.StorageInfo; v != nil {
		if v := v.EbsStorageInfo; v != nil {
			if v := v.VolumeSize; v != nil {
				tfMap["ebs_volume_size"] = aws.Int64Value(v)
			}
		}
	}

	return tfMap
}

func flattenConnectivityInfo(apiObject *kafka.ConnectivityInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PublicAccess; v != nil {
		tfMap["public_access"] = []interface{}{flattenPublicAccess(v)}
	}

	return tfMap
}

func flattenPublicAccess(apiObject *kafka.PublicAccess) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenClientAuthentication(apiObject *kafka.ClientAuthentication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Sasl; v != nil {
		tfMap["sasl"] = []interface{}{flattenSasl(v)}
	}

	if v := apiObject.Tls; v != nil {
		tfMap["tls"] = []interface{}{flattenTls(v)}
	}

	if v := apiObject.Unauthenticated; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["unauthenticated"] = aws.BoolValue(v)
		}
	}

	return tfMap
}

func flattenSasl(apiObject *kafka.Sasl) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Iam; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["iam"] = aws.BoolValue(v)
		}
	}

	if v := apiObject.Scram; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["scram"] = aws.BoolValue(v)
		}
	}

	return tfMap
}

func flattenTls(apiObject *kafka.Tls) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CertificateAuthorityArnList; v != nil && aws.BoolValue(apiObject.Enabled) {
		tfMap["certificate_authority_arns"] = aws.StringValueSlice(v)
	}

	return tfMap
}

func flattenBrokerSoftwareInfo(apiObject *kafka.BrokerSoftwareInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConfigurationArn; v != nil {
		tfMap["arn"] = aws.StringValue(v)
	}

	if v := apiObject.ConfigurationRevision; v != nil {
		tfMap["revision"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenEncryptionInfo(apiObject *kafka.EncryptionInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EncryptionAtRest; v != nil {
		if v := v.DataVolumeKMSKeyId; v != nil {
			tfMap["encryption_at_rest_kms_key_arn"] = aws.StringValue(v)
		}
	}

	if v := apiObject.EncryptionInTransit; v != nil {
		tfMap["encryption_in_transit"] = []interface{}{flattenEncryptionInTransit(v)}
	}

	return tfMap
}

func flattenEncryptionInTransit(apiObject *kafka.EncryptionInTransit) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClientBroker; v != nil {
		tfMap["client_broker"] = aws.StringValue(v)
	}

	if v := apiObject.InCluster; v != nil {
		tfMap["in_cluster"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenLoggingInfo(apiObject *kafka.LoggingInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BrokerLogs; v != nil {
		tfMap["broker_logs"] = []interface{}{flattenBrokerLogs(v)}
	}

	return tfMap
}

func flattenBrokerLogs(apiObject *kafka.BrokerLogs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap["cloudwatch_logs"] = []interface{}{flattenCloudWatchLogs(v)}
	}

	if v := apiObject.Firehose; v != nil {
		tfMap["firehose"] = []interface{}{flattenFirehose(v)}
	}

	if v := apiObject.S3; v != nil {
		tfMap["s3"] = []interface{}{flattenS3(v)}
	}

	return tfMap
}

func flattenCloudWatchLogs(apiObject *kafka.CloudWatchLogs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenFirehose(apiObject *kafka.Firehose) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.StringValue(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenS3(apiObject *kafka.S3) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap["bucket"] = aws.StringValue(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap["prefix"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenOpenMonitoring(apiObject *kafka.OpenMonitoring) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Prometheus; v != nil {
		tfMap["prometheus"] = []interface{}{flattenPrometheus(v)}
	}

	return tfMap
}

func flattenPrometheus(apiObject *kafka.Prometheus) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.JmxExporter; v != nil {
		tfMap["jmx_exporter"] = []interface{}{flattenJmxExporter(v)}
	}

	if v := apiObject.NodeExporter; v != nil {
		tfMap["node_exporter"] = []interface{}{flattenNodeExporter(v)}
	}

	return tfMap
}

func flattenJmxExporter(apiObject *kafka.JmxExporter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EnabledInBroker; v != nil {
		tfMap["enabled_in_broker"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenNodeExporter(apiObject *kafka.NodeExporter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EnabledInBroker; v != nil {
		tfMap["enabled_in_broker"] = aws.BoolValue(v)
	}

	return tfMap
}
