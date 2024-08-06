// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_cluster", name="Cluster")
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

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
			"broker_node_group_info": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"az_distribution": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_subnets": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"connectivity_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vpc_connectivity": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"client_authentication": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"sasl": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"iam": {
																			Type:     schema.TypeBool,
																			Computed: true,
																		},
																		"scram": {
																			Type:     schema.TypeBool,
																			Computed: true,
																		},
																	},
																},
															},
															"tls": {
																Type:     schema.TypeBool,
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
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrType: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"storage_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ebs_storage_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"provisioned_throughput": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Computed: true,
															},
															"volume_throughput": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},
												names.AttrVolumeSize: {
													Type:     schema.TypeInt,
													Computed: true,
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
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_broker_nodes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
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

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName := d.Get(names.AttrClusterName).(string)
	input := &kafka.ListClustersInput{
		ClusterNameFilter: aws.String(clusterName),
	}
	cluster, err := findCluster(ctx, conn, input, func(v *types.ClusterInfo) bool {
		return aws.ToString(v.ClusterName) == clusterName
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s): %s", clusterName, err)
	}

	clusterARN := aws.ToString(cluster.ClusterArn)
	bootstrapBrokersOutput, err := findBootstrapBrokersByARN(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s) bootstrap brokers: %s", clusterARN, err)
	}

	d.SetId(clusterARN)
	d.Set(names.AttrARN, clusterARN)
	d.Set("bootstrap_brokers", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerString)))
	d.Set("bootstrap_brokers_public_sasl_iam", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringPublicSaslIam)))
	d.Set("bootstrap_brokers_public_sasl_scram", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringPublicSaslScram)))
	d.Set("bootstrap_brokers_public_tls", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringPublicTls)))
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringSaslIam)))
	d.Set("bootstrap_brokers_sasl_scram", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringSaslScram)))
	d.Set("bootstrap_brokers_tls", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringTls)))
	if cluster.BrokerNodeGroupInfo != nil {
		if err := d.Set("broker_node_group_info", []interface{}{flattenBrokerNodeGroupInfo(cluster.BrokerNodeGroupInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting broker_node_group_info: %s", err)
		}
	} else {
		d.Set("broker_node_group_info", nil)
	}
	d.Set(names.AttrClusterName, cluster.ClusterName)
	clusterUUID, _ := clusterUUIDFromARN(clusterARN)
	d.Set("cluster_uuid", clusterUUID)
	d.Set("kafka_version", cluster.CurrentBrokerSoftwareInfo.KafkaVersion)
	d.Set("number_of_broker_nodes", cluster.NumberOfBrokerNodes)
	d.Set("zookeeper_connect_string", SortEndpointsString(aws.ToString(cluster.ZookeeperConnectString)))
	d.Set("zookeeper_connect_string_tls", SortEndpointsString(aws.ToString(cluster.ZookeeperConnectStringTls)))

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func findCluster(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersInput, filter tfslices.Predicate[*types.ClusterInfo]) (*types.ClusterInfo, error) {
	output, err := findClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func findClusters(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersInput, filter tfslices.Predicate[*types.ClusterInfo]) ([]types.ClusterInfo, error) {
	var output []types.ClusterInfo

	pages := kafka.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ClusterInfoList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
