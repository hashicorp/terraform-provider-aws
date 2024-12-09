// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_cluster", name="Cluster")
// @Tags(identifierAttribute="id")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("kafka_version", func(_ context.Context, old, new, meta interface{}) bool {
				return semver.LessThan(new.(string), old.(string))
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			"bootstrap_brokers_vpc_connectivity_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_vpc_connectivity_sasl_scram": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_vpc_connectivity_tls": {
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          types.BrokerAZDistributionDefault,
							ValidateDiagFunc: enum.Validate[types.BrokerAZDistribution](),
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
									"vpc_connectivity": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"client_authentication": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"sasl": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"iam": {
																			Type:     schema.TypeBool,
																			Optional: true,
																			Computed: true,
																		},
																		"scram": {
																			Type:     schema.TypeBool,
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
															},
															"tls": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"public_access": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[publicAccessType](),
												},
											},
										},
									},
								},
							},
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"storage_info": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ebs_storage_info": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"provisioned_throughput": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															// This feature is available for
															// storage volume larger than 10 GiB and
															// broker types kafka.m5.4xlarge and larger.
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Optional: true,
															},
															// Minimum and maximum for this varies between broker type
															// https://docs.aws.amazon.com/msk/latest/developerguide/msk-provision-throughput.html
															"volume_throughput": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntBetween(250, 2375),
															},
														},
													},
												},
												names.AttrVolumeSize: {
													Type:     schema.TypeInt,
													Optional: true,
													// https://docs.aws.amazon.com/msk/1.0/apireference/clusters.html#clusters-model-ebsstorageinfo
													ValidateFunc: validation.IntBetween(1, 16384),
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
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_info": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
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
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.ClientBrokerTls,
										ValidateDiagFunc: enum.Validate[types.ClientBroker](),
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.EnhancedMonitoringDefault,
				ValidateDiagFunc: enum.Validate[types.EnhancedMonitoring](),
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
									names.AttrCloudWatchLogs: {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
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
												names.AttrEnabled: {
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
												names.AttrBucket: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
												},
												names.AttrPrefix: {
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
			"storage_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.StorageMode](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	name := d.Get(names.AttrClusterName).(string)
	input := &kafka.CreateClusterInput{
		ClusterName:         aws.String(name),
		KafkaVersion:        aws.String(d.Get("kafka_version").(string)),
		NumberOfBrokerNodes: aws.Int32(int32(d.Get("number_of_broker_nodes").(int))),
		Tags:                getTagsIn(ctx),
	}

	var vpcConnectivity *types.VpcConnectivity
	if v, ok := d.GetOk("broker_node_group_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BrokerNodeGroupInfo = expandBrokerNodeGroupInfo(v.([]interface{})[0].(map[string]interface{}))
		// "BadRequestException: When creating a cluster, all vpcConnectivity auth schemes must be disabled (‘enabled’ : false). You can enable auth schemes after the cluster is created"
		if input.BrokerNodeGroupInfo != nil && input.BrokerNodeGroupInfo.ConnectivityInfo != nil {
			vpcConnectivity = input.BrokerNodeGroupInfo.ConnectivityInfo.VpcConnectivity
			input.BrokerNodeGroupInfo.ConnectivityInfo.VpcConnectivity = nil
		}
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
		input.EnhancedMonitoring = types.EnhancedMonitoring(v.(string))
	}

	if v, ok := d.GetOk("logging_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.LoggingInfo = expandLoggingInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("open_monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OpenMonitoring = expandOpenMonitoringInfo(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("storage_mode"); ok {
		input.StorageMode = types.StorageMode(v.(string))
	}

	output, err := conn.CreateCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Cluster (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ClusterArn))

	cluster, err := waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) create: %s", d.Id(), err)
	}

	if vpcConnectivity != nil {
		input := &kafka.UpdateConnectivityInput{
			ClusterArn: aws.String(d.Id()),
			ConnectivityInfo: &types.ConnectivityInfo{
				VpcConnectivity: vpcConnectivity,
			},
			CurrentVersion: cluster.CurrentVersion,
		}

		output, err := conn.UpdateConnectivity(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) broker connectivity: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	cluster, err := findClusterByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s): %s", d.Id(), err)
	}

	output, err := findBootstrapBrokersByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s) bootstrap brokers: %s", d.Id(), err)
	}

	clusterARN := aws.ToString(cluster.ClusterArn)
	d.Set(names.AttrARN, clusterARN)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.ToString(output.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_public_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicSaslIam)))
	d.Set("bootstrap_brokers_public_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicSaslScram)))
	d.Set("bootstrap_brokers_public_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringPublicTls)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringTls)))
	d.Set("bootstrap_brokers_vpc_connectivity_sasl_iam", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivitySaslIam)))
	d.Set("bootstrap_brokers_vpc_connectivity_sasl_scram", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivitySaslScram)))
	d.Set("bootstrap_brokers_vpc_connectivity_tls", SortEndpointsString(aws.ToString(output.BootstrapBrokerStringVpcConnectivityTls)))
	if cluster.BrokerNodeGroupInfo != nil {
		if err := d.Set("broker_node_group_info", []interface{}{flattenBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting broker_node_group_info: %s", err)
		}
	} else {
		d.Set("broker_node_group_info", nil)
	}
	if cluster.ClientAuthentication != nil {
		if err := d.Set("client_authentication", []interface{}{flattenClientAuthentication(cluster.ClientAuthentication)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting client_authentication: %s", err)
		}
	} else {
		d.Set("client_authentication", nil)
	}
	d.Set(names.AttrClusterName, cluster.ClusterName)
	clusterUUID, _ := clusterUUIDFromARN(clusterARN)
	d.Set("cluster_uuid", clusterUUID)
	if cluster.CurrentBrokerSoftwareInfo != nil {
		if err := d.Set("configuration_info", []interface{}{flattenBrokerSoftwareInfo(cluster.CurrentBrokerSoftwareInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration_info: %s", err)
		}
	} else {
		d.Set("configuration_info", nil)
	}
	d.Set("current_version", cluster.CurrentVersion)
	d.Set("enhanced_monitoring", cluster.EnhancedMonitoring)
	if cluster.EncryptionInfo != nil {
		if err := d.Set("encryption_info", []interface{}{flattenEncryptionInfo(cluster.EncryptionInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting encryption_info: %s", err)
		}
	} else {
		d.Set("encryption_info", nil)
	}
	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)
	if cluster.LoggingInfo != nil {
		if err := d.Set("logging_info", []interface{}{flattenLoggingInfo(cluster.LoggingInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_info: %s", err)
		}
	} else {
		d.Set("logging_info", nil)
	}
	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)
	if cluster.OpenMonitoring != nil {
		if err := d.Set("open_monitoring", []interface{}{flattenOpenMonitoring(cluster.OpenMonitoring)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting open_monitoring: %s", err)
		}
	} else {
		d.Set("open_monitoring", nil)
	}
	d.Set("storage_mode", cluster.StorageMode)
	d.Set("zookeeper_connect_string", SortEndpointsString(aws.ToString(cluster.ZookeeperConnectString)))
	d.Set("zookeeper_connect_string_tls", SortEndpointsString(aws.ToString(cluster.ZookeeperConnectStringTls)))

	setTagsOut(ctx, cluster.Tags)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	if d.HasChange("broker_node_group_info.0.connectivity_info") {
		input := &kafka.UpdateConnectivityInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
		}

		if v, ok := d.GetOk("broker_node_group_info.0.connectivity_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ConnectivityInfo = expandConnectivityInfo(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateConnectivity(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) broker connectivity: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("broker_node_group_info.0.instance_type") {
		input := &kafka.UpdateBrokerTypeInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			TargetInstanceType: aws.String(d.Get("broker_node_group_info.0.instance_type").(string)),
		}

		output, err := conn.UpdateBrokerType(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) broker type: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("broker_node_group_info.0.storage_info") {
		input := &kafka.UpdateBrokerStorageInput{
			ClusterArn:     aws.String(d.Id()),
			CurrentVersion: aws.String(d.Get("current_version").(string)),
			TargetBrokerEBSVolumeInfo: []types.BrokerEBSVolumeInfo{{
				KafkaBrokerNodeId: aws.String("All"),
				VolumeSizeGB:      aws.Int32(int32(d.Get("broker_node_group_info.0.storage_info.0.ebs_storage_info.0.volume_size").(int))),
			}},
		}

		if v, ok := d.GetOk("broker_node_group_info.0.storage_info.0.ebs_storage_info.0.provisioned_throughput"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.TargetBrokerEBSVolumeInfo[0].ProvisionedThroughput = expandProvisionedThroughput(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateBrokerStorage(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) broker storage: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChange("number_of_broker_nodes") {
		input := &kafka.UpdateBrokerCountInput{
			ClusterArn:                aws.String(d.Id()),
			CurrentVersion:            aws.String(d.Get("current_version").(string)),
			TargetNumberOfBrokerNodes: aws.Int32(int32(d.Get("number_of_broker_nodes").(int))),
		}

		output, err := conn.UpdateBrokerCount(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) broker count: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("enhanced_monitoring", "logging_info", "open_monitoring") {
		input := &kafka.UpdateMonitoringInput{
			ClusterArn:         aws.String(d.Id()),
			CurrentVersion:     aws.String(d.Get("current_version").(string)),
			EnhancedMonitoring: types.EnhancedMonitoring(d.Get("enhanced_monitoring").(string)),
		}

		if v, ok := d.GetOk("logging_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.LoggingInfo = expandLoggingInfo(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("open_monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OpenMonitoring = expandOpenMonitoringInfo(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.UpdateMonitoring(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) monitoring: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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

		output, err := conn.UpdateClusterConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) configuration: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s) complete: %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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

		output, err := conn.UpdateClusterKafkaVersion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) Kafka version: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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

		output, err := conn.UpdateSecurity(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Cluster (%s) security: %s", d.Id(), err)
		}

		clusterOperationARN := aws.ToString(output.ClusterOperationArn)

		if _, err := waitClusterOperationCompleted(ctx, conn, clusterOperationARN, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) operation (%s): %s", d.Id(), clusterOperationARN, err)
		}

		// refresh the current_version attribute after each update
		if err := refreshClusterVersion(ctx, d, meta); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[DEBUG] Deleting MSK Cluster: %s", d.Id())
	_, err := conn.DeleteCluster(ctx, &kafka.DeleteClusterInput{
		ClusterArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func refreshClusterVersion(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	cluster, err := findClusterByARN(ctx, conn, d.Id())

	if err != nil {
		return fmt.Errorf("reading MSK Cluster (%s): %w", d.Id(), err)
	}

	d.Set("current_version", cluster.CurrentVersion)

	return nil
}

func findClusterByARN(ctx context.Context, conn *kafka.Client, arn string) (*types.ClusterInfo, error) {
	input := &kafka.DescribeClusterInput{
		ClusterArn: aws.String(arn),
	}

	output, err := conn.DescribeCluster(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClusterInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClusterInfo, nil
}

func findClusterV2ByARN(ctx context.Context, conn *kafka.Client, arn string) (*types.Cluster, error) {
	input := &kafka.DescribeClusterV2Input{
		ClusterArn: aws.String(arn),
	}

	output, err := conn.DescribeClusterV2(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClusterInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClusterInfo, nil
}

func findClusterOperationByARN(ctx context.Context, conn *kafka.Client, arn string) (*types.ClusterOperationInfo, error) {
	input := &kafka.DescribeClusterOperationInput{
		ClusterOperationArn: aws.String(arn),
	}

	output, err := conn.DescribeClusterOperation(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ClusterOperationInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ClusterOperationInfo, nil
}

func findBootstrapBrokersByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.GetBootstrapBrokersOutput, error) {
	input := &kafka.GetBootstrapBrokersInput{
		ClusterArn: aws.String(arn),
	}

	output, err := conn.GetBootstrapBrokers(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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

func statusClusterState(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterV2ByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusClusterOperationState(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterOperationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.OperationState), nil
	}
}

func waitClusterCreated(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClusterStateCreating),
		Target:  enum.Slice(types.ClusterStateActive),
		Refresh: statusClusterState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		if state, stateInfo := output.State, output.StateInfo; state == types.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*types.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClusterStateDeleting),
		Target:  []string{},
		Refresh: statusClusterState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Cluster); ok {
		if state, stateInfo := output.State, output.StateInfo; state == types.ClusterStateFailed && stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterOperationCompleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*types.ClusterOperationInfo, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterOperationStatePending, clusterOperationStateUpdateInProgress},
		Target:  []string{clusterOperationStateUpdateComplete},
		Refresh: statusClusterOperationState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ClusterOperationInfo); ok {
		if state, errorInfo := aws.ToString(output.OperationState), output.ErrorInfo; state == clusterOperationStateUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(errorInfo.ErrorCode), aws.ToString(errorInfo.ErrorString)))
		}

		return output, err
	}

	return nil, err
}

func clusterUUIDFromARN(clusterARN string) (string, error) {
	parsedARN, err := arn.Parse(clusterARN)
	if err != nil {
		return "", err
	}

	// arn:${Partition}:kafka:${Region}:${Account}:cluster/${ClusterName}/${Uuid}
	parts := strings.Split(parsedARN.Resource, "/")
	if len(parts) != 3 || parts[0] != "cluster" || parts[1] == "" || parts[2] == "" {
		return "", fmt.Errorf("invalid MSK Cluster ARN (%s)", clusterARN)
	}
	return parts[2], nil
}

func expandBrokerNodeGroupInfo(tfMap map[string]interface{}) *types.BrokerNodeGroupInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BrokerNodeGroupInfo{}

	if v, ok := tfMap["az_distribution"].(string); ok && v != "" {
		apiObject.BrokerAZDistribution = types.BrokerAZDistribution(v)
	}

	if v, ok := tfMap["client_subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ClientSubnets = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["connectivity_info"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ConnectivityInfo = expandConnectivityInfo(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["storage_info"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StorageInfo = expandStorageInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandConnectivityInfo(tfMap map[string]interface{}) *types.ConnectivityInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ConnectivityInfo{}

	if v, ok := tfMap["public_access"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PublicAccess = expandPublicAccess(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vpc_connectivity"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.VpcConnectivity = expandVPCConnectivity(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandStorageInfo(tfMap map[string]interface{}) *types.StorageInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StorageInfo{}

	if v, ok := tfMap["ebs_storage_info"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EbsStorageInfo = expandEBSStorageInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandEBSStorageInfo(tfMap map[string]interface{}) *types.EBSStorageInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EBSStorageInfo{}

	if v, ok := tfMap["provisioned_throughput"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ProvisionedThroughput = expandProvisionedThroughput(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int32(int32(v))
	}

	return apiObject
}

func expandProvisionedThroughput(tfMap map[string]interface{}) *types.ProvisionedThroughput {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ProvisionedThroughput{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["volume_throughput"].(int); ok && v != 0 {
		apiObject.VolumeThroughput = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPublicAccess(tfMap map[string]interface{}) *types.PublicAccess {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PublicAccess{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandVPCConnectivity(tfMap map[string]interface{}) *types.VpcConnectivity {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConnectivity{}

	if v, ok := tfMap["client_authentication"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ClientAuthentication = expandVPCConnectivityClientAuthentication(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandVPCConnectivityClientAuthentication(tfMap map[string]interface{}) *types.VpcConnectivityClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConnectivityClientAuthentication{}

	if v, ok := tfMap["sasl"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Sasl = expandVPCConnectivitySASL(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["tls"].(bool); ok {
		apiObject.Tls = &types.VpcConnectivityTls{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandVPCConnectivitySASL(tfMap map[string]interface{}) *types.VpcConnectivitySasl {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConnectivitySasl{}

	if v, ok := tfMap["iam"].(bool); ok {
		apiObject.Iam = &types.VpcConnectivityIam{
			Enabled: aws.Bool(v),
		}
	}

	if v, ok := tfMap["scram"].(bool); ok {
		apiObject.Scram = &types.VpcConnectivityScram{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandClientAuthentication(tfMap map[string]interface{}) *types.ClientAuthentication {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ClientAuthentication{}

	if v, ok := tfMap["sasl"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Sasl = expandSASL(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["tls"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Tls = expandTLS(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["unauthenticated"].(bool); ok {
		apiObject.Unauthenticated = &types.Unauthenticated{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandSASL(tfMap map[string]interface{}) *types.Sasl {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Sasl{}

	if v, ok := tfMap["iam"].(bool); ok {
		apiObject.Iam = &types.Iam{
			Enabled: aws.Bool(v),
		}
	}

	if v, ok := tfMap["scram"].(bool); ok {
		apiObject.Scram = &types.Scram{
			Enabled: aws.Bool(v),
		}
	}

	return apiObject
}

func expandTLS(tfMap map[string]interface{}) *types.Tls {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Tls{}

	if v, ok := tfMap["certificate_authority_arns"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CertificateAuthorityArnList = flex.ExpandStringValueSet(v)
		apiObject.Enabled = aws.Bool(true)
	} else {
		apiObject.Enabled = aws.Bool(false)
	}

	return apiObject
}

func expandConfigurationInfo(tfMap map[string]interface{}) *types.ConfigurationInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ConfigurationInfo{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["revision"].(int); ok && v != 0 {
		apiObject.Revision = aws.Int64(int64(v))
	}

	return apiObject
}

func expandEncryptionInfo(tfMap map[string]interface{}) *types.EncryptionInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EncryptionInfo{}

	if v, ok := tfMap["encryption_in_transit"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionInTransit = expandEncryptionInTransit(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["encryption_at_rest_kms_key_arn"].(string); ok && v != "" {
		apiObject.EncryptionAtRest = &types.EncryptionAtRest{
			DataVolumeKMSKeyId: aws.String(v),
		}
	}

	return apiObject
}

func expandEncryptionInTransit(tfMap map[string]interface{}) *types.EncryptionInTransit {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EncryptionInTransit{}

	if v, ok := tfMap["client_broker"].(string); ok && v != "" {
		apiObject.ClientBroker = types.ClientBroker(v)
	}

	if v, ok := tfMap["in_cluster"].(bool); ok {
		apiObject.InCluster = aws.Bool(v)
	}

	return apiObject
}

func expandLoggingInfo(tfMap map[string]interface{}) *types.LoggingInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.LoggingInfo{}

	if v, ok := tfMap["broker_logs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.BrokerLogs = expandBrokerLogs(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandBrokerLogs(tfMap map[string]interface{}) *types.BrokerLogs {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BrokerLogs{}

	if v, ok := tfMap[names.AttrCloudWatchLogs].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchLogs = expandCloudWatchLogs(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["firehose"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Firehose = expandFirehose(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3 = expandS3(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchLogs(tfMap map[string]interface{}) *types.CloudWatchLogs {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CloudWatchLogs{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		apiObject.LogGroup = aws.String(v)
	}

	return apiObject
}

func expandFirehose(tfMap map[string]interface{}) *types.Firehose {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Firehose{}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		apiObject.DeliveryStream = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandS3(tfMap map[string]interface{}) *types.S3 {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.S3{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandOpenMonitoringInfo(tfMap map[string]interface{}) *types.OpenMonitoringInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.OpenMonitoringInfo{}

	if v, ok := tfMap["prometheus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Prometheus = expandPrometheusInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrometheusInfo(tfMap map[string]interface{}) *types.PrometheusInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PrometheusInfo{}

	if v, ok := tfMap["jmx_exporter"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.JmxExporter = expandJmxExporterInfo(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["node_exporter"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.NodeExporter = expandNodeExporterInfo(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandJmxExporterInfo(tfMap map[string]interface{}) *types.JmxExporterInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.JmxExporterInfo{}

	if v, ok := tfMap["enabled_in_broker"].(bool); ok {
		apiObject.EnabledInBroker = aws.Bool(v)
	}

	return apiObject
}

func expandNodeExporterInfo(tfMap map[string]interface{}) *types.NodeExporterInfo {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.NodeExporterInfo{}

	if v, ok := tfMap["enabled_in_broker"].(bool); ok {
		apiObject.EnabledInBroker = aws.Bool(v)
	}

	return apiObject
}

func flattenBrokerNodeGroupInfo(apiObject *types.BrokerNodeGroupInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"az_distribution": apiObject.BrokerAZDistribution,
	}

	if v := apiObject.ClientSubnets; v != nil {
		tfMap["client_subnets"] = v
	}

	if v := apiObject.ConnectivityInfo; v != nil {
		tfMap["connectivity_info"] = []interface{}{flattenConnectivityInfo(v)}
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap[names.AttrInstanceType] = aws.ToString(v)
	}

	if v := apiObject.SecurityGroups; v != nil {
		tfMap[names.AttrSecurityGroups] = v
	}

	if v := apiObject.StorageInfo; v != nil {
		tfMap["storage_info"] = flattenStorageInfo(v)
	}

	return tfMap
}

func flattenConnectivityInfo(apiObject *types.ConnectivityInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.PublicAccess; v != nil {
		tfMap["public_access"] = []interface{}{flattenPublicAccess(v)}
	}

	if v := apiObject.VpcConnectivity; v != nil {
		tfMap["vpc_connectivity"] = []interface{}{flattenVPCConnectivity(v)}
	}

	return tfMap
}

func flattenStorageInfo(apiObject *types.StorageInfo) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EbsStorageInfo; v != nil {
		tfMap["ebs_storage_info"] = flattenEBSStorageInfo(v)
	}

	return []interface{}{tfMap}
}

func flattenEBSStorageInfo(apiObject *types.EBSStorageInfo) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ProvisionedThroughput; v != nil {
		tfMap["provisioned_throughput"] = flattenProvisionedThroughput(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap[names.AttrVolumeSize] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}

func flattenProvisionedThroughput(apiObject *types.ProvisionedThroughput) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.VolumeThroughput; v != nil {
		tfMap["volume_throughput"] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}

func flattenPublicAccess(apiObject *types.PublicAccess) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func flattenVPCConnectivity(apiObject *types.VpcConnectivity) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if v := apiObject.ClientAuthentication; v != nil {
		tfMap["client_authentication"] = []interface{}{flattenVPCConnectivityClientAuthentication(v)}
	}

	return tfMap
}

func flattenVPCConnectivityClientAuthentication(apiObject *types.VpcConnectivityClientAuthentication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Sasl; v != nil {
		tfMap["sasl"] = []interface{}{(flattenVPCConnectivitySASL(v))}
	}

	if v := apiObject.Tls; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["tls"] = aws.ToBool(v)
		}
	}

	return tfMap
}

func flattenVPCConnectivitySASL(apiObject *types.VpcConnectivitySasl) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Iam; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["iam"] = aws.ToBool(v)
		}
	}

	if v := apiObject.Scram; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["scram"] = aws.ToBool(v)
		}
	}

	return tfMap
}

func flattenClientAuthentication(apiObject *types.ClientAuthentication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Sasl; v != nil {
		tfMap["sasl"] = []interface{}{flattenSASL(v)}
	}

	if v := apiObject.Tls; v != nil {
		tfMap["tls"] = []interface{}{flattenTLS(v)}
	}

	if v := apiObject.Unauthenticated; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["unauthenticated"] = aws.ToBool(v)
		}
	}

	return tfMap
}

func flattenSASL(apiObject *types.Sasl) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Iam; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["iam"] = aws.ToBool(v)
		}
	}

	if v := apiObject.Scram; v != nil {
		if v := v.Enabled; v != nil {
			tfMap["scram"] = aws.ToBool(v)
		}
	}

	return tfMap
}

func flattenTLS(apiObject *types.Tls) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CertificateAuthorityArnList; v != nil && aws.ToBool(apiObject.Enabled) {
		tfMap["certificate_authority_arns"] = v
	}

	return tfMap
}

func flattenBrokerSoftwareInfo(apiObject *types.BrokerSoftwareInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ConfigurationArn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	if v := apiObject.ConfigurationRevision; v != nil {
		tfMap["revision"] = aws.ToInt64(v)
	}

	return tfMap
}

func flattenEncryptionInfo(apiObject *types.EncryptionInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EncryptionAtRest; v != nil {
		if v := v.DataVolumeKMSKeyId; v != nil {
			tfMap["encryption_at_rest_kms_key_arn"] = aws.ToString(v)
		}
	}

	if v := apiObject.EncryptionInTransit; v != nil {
		tfMap["encryption_in_transit"] = []interface{}{flattenEncryptionInTransit(v)}
	}

	return tfMap
}

func flattenEncryptionInTransit(apiObject *types.EncryptionInTransit) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"client_broker": apiObject.ClientBroker,
	}

	if v := apiObject.InCluster; v != nil {
		tfMap["in_cluster"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenLoggingInfo(apiObject *types.LoggingInfo) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BrokerLogs; v != nil {
		tfMap["broker_logs"] = []interface{}{flattenBrokerLogs(v)}
	}

	return tfMap
}

func flattenBrokerLogs(apiObject *types.BrokerLogs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap[names.AttrCloudWatchLogs] = []interface{}{flattenCloudWatchLogs(v)}
	}

	if v := apiObject.Firehose; v != nil {
		tfMap["firehose"] = []interface{}{flattenFirehose(v)}
	}

	if v := apiObject.S3; v != nil {
		tfMap["s3"] = []interface{}{flattenS3(v)}
	}

	return tfMap
}

func flattenCloudWatchLogs(apiObject *types.CloudWatchLogs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.ToString(v)
	}

	return tfMap
}

func flattenFirehose(apiObject *types.Firehose) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.ToString(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	return tfMap
}

func flattenS3(apiObject *types.S3) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Enabled; v != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return tfMap
}

func flattenOpenMonitoring(apiObject *types.OpenMonitoring) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Prometheus; v != nil {
		tfMap["prometheus"] = []interface{}{flattenPrometheus(v)}
	}

	return tfMap
}

func flattenPrometheus(apiObject *types.Prometheus) map[string]interface{} {
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

func flattenJmxExporter(apiObject *types.JmxExporter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EnabledInBroker; v != nil {
		tfMap["enabled_in_broker"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenNodeExporter(apiObject *types.NodeExporter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EnabledInBroker; v != nil {
		tfMap["enabled_in_broker"] = aws.ToBool(v)
	}

	return tfMap
}
