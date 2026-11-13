// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_replicator", name="Replicator")
// @Tags(identifierAttribute="id")
// @ArnIdentity
// @Testing(preIdentityVersion="v6.49.0")
// @Testing(tagsTest=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/kafka;kafka.DescribeReplicatorOutput")
// @Testing(preCheck="testAccPreCheck")
func resourceReplicator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicatorCreate,
		ReadWithoutTimeout:   resourceReplicatorRead,
		UpdateWithoutTimeout: resourceReplicatorUpdate,
		DeleteWithoutTimeout: resourceReplicatorDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"current_version": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				"kafka_cluster": {
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					MinItems: 2,
					MaxItems: 2,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"amazon_msk_cluster": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								// Mutual exclusivity with apache_kafka_cluster is enforced per
								// entry by validateKafkaClusterKind; ExactlyOneOf cannot be used
								// here because it would traverse the multi-item kafka_cluster list.
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"msk_cluster_arn": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							"apache_kafka_cluster": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"apache_kafka_cluster_id": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"bootstrap_broker_string": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
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
										"mtls": {
											Type:     schema.TypeList,
											Optional: true,
											ForceNew: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"secret_arn": {
														Type:         schema.TypeString,
														Required:     true,
														ForceNew:     true,
														ValidateFunc: verify.ValidARN,
													},
												},
											},
										},
										"sasl_scram": {
											Type:     schema.TypeList,
											Optional: true,
											ForceNew: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"mechanism": {
														Type:             schema.TypeString,
														Required:         true,
														ForceNew:         true,
														ValidateDiagFunc: enum.Validate[types.KafkaClusterSaslScramMechanism](),
													},
													"secret_arn": {
														Type:         schema.TypeString,
														Required:     true,
														ForceNew:     true,
														ValidateFunc: verify.ValidARN,
													},
												},
											},
										},
									},
								},
							},
							"encryption_in_transit": {
								Type:     schema.TypeList,
								Optional: true,
								ForceNew: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"root_ca_certificate": {
											Type:         schema.TypeString,
											Required:     true,
											ForceNew:     true,
											ValidateFunc: verify.ValidARN,
										},
									},
								},
							},
							names.AttrVPCConfig: {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"security_groups_ids": {
											Type:     schema.TypeSet,
											Optional: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										names.AttrSubnetIDs: {
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
						},
					},
				},
				"log_delivery": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"replicator_log_delivery": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrCloudWatchLogs: {
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
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
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
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
											Type:     schema.TypeList,
											Optional: true,
											MaxItems: 1,
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
				"replication_info_list": {
					Type:     schema.TypeList,
					Required: true,
					ForceNew: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"source_kafka_cluster_alias": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"source_kafka_cluster_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
								ExactlyOneOf: []string{
									"replication_info_list.0.source_kafka_cluster_arn",
									"replication_info_list.0.source_kafka_cluster_id",
								},
							},
							"source_kafka_cluster_id": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ExactlyOneOf: []string{
									"replication_info_list.0.source_kafka_cluster_arn",
									"replication_info_list.0.source_kafka_cluster_id",
								},
							},
							"target_compression_type": {
								Type:     schema.TypeString,
								Required: true,
								ForceNew: true,
							},
							"target_kafka_cluster_alias": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"target_kafka_cluster_arn": {
								Type:         schema.TypeString,
								Optional:     true,
								ForceNew:     true,
								ValidateFunc: verify.ValidARN,
								ExactlyOneOf: []string{
									"replication_info_list.0.target_kafka_cluster_arn",
									"replication_info_list.0.target_kafka_cluster_id",
								},
							},
							"target_kafka_cluster_id": {
								Type:     schema.TypeString,
								Optional: true,
								ForceNew: true,
								ExactlyOneOf: []string{
									"replication_info_list.0.target_kafka_cluster_arn",
									"replication_info_list.0.target_kafka_cluster_id",
								},
							},
							"topic_replication": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"copy_access_control_lists_for_topics": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
										"copy_topic_configurations": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
										"detect_and_copy_new_topics": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
										"starting_position": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrType: {
														Type:             schema.TypeString,
														Optional:         true,
														ForceNew:         true,
														ValidateDiagFunc: enum.Validate[types.ReplicationStartingPositionType](),
													},
												},
											},
										},
										"topic_name_configuration": {
											Type:     schema.TypeList,
											Optional: true,
											Computed: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrType: {
														Type:             schema.TypeString,
														Optional:         true,
														ForceNew:         true,
														ValidateDiagFunc: enum.Validate[types.ReplicationTopicNameConfigurationType](),
													},
												},
											},
										},
										"topics_to_exclude": {
											Type:     schema.TypeSet,
											Optional: true,
											Computed: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"topics_to_replicate": {
											Type:     schema.TypeSet,
											Required: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
									},
								},
							},
							"consumer_group_replication": {
								Type:     schema.TypeList,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"consumer_group_offset_sync_mode": {
											Type:             schema.TypeString,
											Optional:         true,
											Computed:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[types.ConsumerGroupOffsetSyncMode](),
										},
										"consumer_groups_to_exclude": {
											Type:     schema.TypeSet,
											Optional: true,
											Computed: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"consumer_groups_to_replicate": {
											Type:     schema.TypeSet,
											Required: true,
											Elem: &schema.Schema{
												Type: schema.TypeString,
											},
										},
										"detect_and_copy_new_consumer_groups": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
										"synchronise_consumer_group_offsets": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  true,
										},
									},
								},
							},
						},
					},
				},
				"replicator_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"service_execution_role_arn": {
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
					Required:     true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
		CustomizeDiff: customdiff.Sequence(
			validateKafkaClusterKind,
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				var diags diag.Diagnostics
				cloudwatchLogsBlock := "log_delivery.0.replicator_log_delivery.0.cloudwatch_logs.0"
				firehoseLogBlock := "log_delivery.0.replicator_log_delivery.0.firehose.0"
				s3LogBlock := "log_delivery.0.replicator_log_delivery.0.s3.0"
				if v, ok := d.Get(fmt.Sprintf("%s.%s", cloudwatchLogsBlock, names.AttrEnabled)).(bool); ok && !v {
					if _, ok := d.GetOk(fmt.Sprintf("%s.%s", cloudwatchLogsBlock, "log_group")); ok {
						diags = sdkdiag.AppendErrorf(diags, "cannot specify log_group when CloudWatch Logs logging is disabled")
					}
				}
				if v, ok := d.Get(fmt.Sprintf("%s.%s", firehoseLogBlock, names.AttrEnabled)).(bool); ok && !v {
					if _, ok := d.GetOk(fmt.Sprintf("%s.%s", firehoseLogBlock, "delivery_stream")); ok {
						diags = sdkdiag.AppendErrorf(diags, "cannot specify delivery_stream when Firehose logging is disabled")
					}
				}
				if v, ok := d.Get(fmt.Sprintf("%s.%s", s3LogBlock, names.AttrEnabled)).(bool); ok && !v {
					if _, ok := d.GetOk(fmt.Sprintf("%s.%s", s3LogBlock, names.AttrBucket)); ok {
						diags = sdkdiag.AppendErrorf(diags, "cannot specify bucket when S3 logging is disabled")
					}
					if _, ok := d.GetOk(fmt.Sprintf("%s.%s", s3LogBlock, names.AttrPrefix)); ok {
						diags = sdkdiag.AppendErrorf(diags, "cannot specify prefix when S3 logging is disabled")
					}
				}
				if diags.HasError() {
					return sdkdiag.DiagnosticsError(diags)
				}

				return nil
			},
		),
	}
}

// validateKafkaClusterKind enforces per-entry constraints on the kafka_cluster block that
// SDKv2 schema options cannot express: exactly one of amazon_msk_cluster / apache_kafka_cluster
// per entry, and that client_authentication, encryption_in_transit, and vpc_config are only
// placed on the cluster kind that supports them. ExactlyOneOf does not expand reliably across
// list indices for a multi-entry list, so the checks run here instead.
func validateKafkaClusterKind(ctx context.Context, d *schema.ResourceDiff, meta any) error { // nosemgrep:ci.kafka-in-func-name
	var diags diag.Diagnostics

	for i := range d.Get("kafka_cluster").([]any) {
		amazon := len(d.Get(fmt.Sprintf("kafka_cluster.%d.amazon_msk_cluster", i)).([]any)) > 0
		apache := len(d.Get(fmt.Sprintf("kafka_cluster.%d.apache_kafka_cluster", i)).([]any)) > 0
		clientAuth := len(d.Get(fmt.Sprintf("kafka_cluster.%d.client_authentication", i)).([]any)) > 0
		encryption := len(d.Get(fmt.Sprintf("kafka_cluster.%d.encryption_in_transit", i)).([]any)) > 0
		vpc := len(d.Get(fmt.Sprintf("kafka_cluster.%d.vpc_config", i)).([]any)) > 0

		switch {
		case amazon && apache:
			diags = sdkdiag.AppendErrorf(diags, "kafka_cluster[%d]: only one of amazon_msk_cluster, apache_kafka_cluster may be set", i)
		case !amazon && !apache:
			diags = sdkdiag.AppendErrorf(diags, "kafka_cluster[%d]: one of amazon_msk_cluster, apache_kafka_cluster must be set", i)
		}

		// client_authentication and encryption_in_transit describe how to reach a
		// self-managed Apache Kafka cluster and are only valid for an apache_kafka_cluster.
		if amazon && clientAuth {
			diags = sdkdiag.AppendErrorf(diags, "kafka_cluster[%d]: client_authentication is only valid for apache_kafka_cluster", i)
		}
		if amazon && encryption {
			diags = sdkdiag.AppendErrorf(diags, "kafka_cluster[%d]: encryption_in_transit is only valid for apache_kafka_cluster", i)
		}

		// vpc_config is supplied on the Amazon MSK entry only; the replicator reaches the
		// Apache cluster through that VPC.
		if apache && vpc {
			diags = sdkdiag.AppendErrorf(diags, "kafka_cluster[%d]: vpc_config is not valid for apache_kafka_cluster", i)
		}
	}

	if diags.HasError() {
		return sdkdiag.DiagnosticsError(diags)
	}

	return nil
}

// kafkaClusterIdentifier returns the identifier for whichever cluster kind the description
// carries: the MSK cluster ARN for an Amazon MSK cluster, or the cluster ID for an Apache
// Kafka cluster. It returns nil when neither is set, avoiding a nil dereference on
// AmazonMskCluster for Apache cluster entries.
func kafkaClusterIdentifier(desc types.KafkaClusterDescription) *string { // nosemgrep:ci.kafka-in-func-name
	if v := desc.AmazonMskCluster; v != nil {
		return v.MskClusterArn
	}

	if v := desc.ApacheKafkaCluster; v != nil {
		return v.ApacheKafkaClusterId
	}

	return nil
}

func resourceReplicatorCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	name := d.Get("replicator_name").(string)
	input := kafka.CreateReplicatorInput{
		KafkaClusters:           expandKafkaClusters(d.Get("kafka_cluster").([]any)),
		ReplicationInfoList:     expandReplicationInfos(d.Get("replication_info_list").([]any)),
		ReplicatorName:          aws.String(name),
		ServiceExecutionRoleArn: aws.String(d.Get("service_execution_role_arn").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_delivery"); ok {
		input.LogDelivery = expandLogDelivery(v.([]any))
	}

	output, err := conn.CreateReplicator(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Replicator (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ReplicatorArn))

	if _, err := waitReplicatorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Replicator (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	output, err := findReplicatorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Kafka Replicator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Replicator (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ReplicatorArn)
	d.Set("current_version", output.CurrentVersion)
	d.Set(names.AttrDescription, output.ReplicatorDescription)
	// DescribeReplicator does not round-trip an Apache Kafka cluster faithfully: it omits the
	// ApacheKafkaCluster, ClientAuthentication, and EncryptionInTransit details and echoes the
	// VpcConfig onto every entry. Reconcile the flattened response against prior state so the
	// post-apply plan is empty (kafka_cluster is entirely ForceNew, so this preserves the
	// configured values). On import there is no prior state, so the raw response is used.
	kafkaClusters := flattenKafkaClusterDescriptions(output.KafkaClusters)
	if old, ok := d.Get("kafka_cluster").([]any); ok && len(old) > 0 {
		kafkaClusters = reconcileKafkaClusters(old, kafkaClusters)
	}
	if err := d.Set("kafka_cluster", kafkaClusters); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kafka_cluster: %s", err)
	}
	if err := d.Set("log_delivery", flattenLogDelivery(output.LogDelivery)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_delivery: %s", err)
	}
	if v := output.ReplicationInfoList; len(v) > 0 {
		sourceAlias, targetAlias := aws.ToString(v[0].SourceKafkaClusterAlias), aws.ToString(v[0].TargetKafkaClusterAlias)
		var sourceCluster, targetCluster *types.KafkaClusterDescription

		for i := range output.KafkaClusters {
			switch aws.ToString(output.KafkaClusters[i].KafkaClusterAlias) {
			case sourceAlias:
				sourceCluster = &output.KafkaClusters[i]
			case targetAlias:
				targetCluster = &output.KafkaClusters[i]
			}
		}

		if err := d.Set("replication_info_list", flattenReplicationInfoDescriptions(v, sourceCluster, targetCluster)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting replication_info_list: %s", err)
		}
	} else {
		d.Set("replication_info_list", nil)
	}
	d.Set("replicator_name", output.ReplicatorName)
	d.Set("service_execution_role_arn", output.ServiceExecutionRoleArn)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceReplicatorUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := kafka.UpdateReplicationInfoInput{
			CurrentVersion:        aws.String(d.Get("current_version").(string)),
			ReplicatorArn:         aws.String(d.Id()),
			SourceKafkaClusterArn: aws.String(d.Get("replication_info_list.0.source_kafka_cluster_arn").(string)),
			TargetKafkaClusterArn: aws.String(d.Get("replication_info_list.0.target_kafka_cluster_arn").(string)),
		}

		if d.HasChanges("log_delivery") {
			input.LogDelivery = expandLogDelivery(d.Get("log_delivery").([]any))
		}

		if d.HasChanges("replication_info_list.0.consumer_group_replication") {
			if v, ok := d.GetOk("replication_info_list.0.consumer_group_replication"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ConsumerGroupReplication = expandConsumerGroupReplicationUpdate(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChanges("replication_info_list.0.topic_replication") {
			if v, ok := d.GetOk("replication_info_list.0.topic_replication"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.TopicReplication = expandTopicReplicationUpdate(v.([]any)[0].(map[string]any))
			}
		}

		_, err := conn.UpdateReplicationInfo(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Replicator (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicatorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Replicator (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting MSK Replicator: %s", d.Id())
	input := kafka.DeleteReplicatorInput{
		ReplicatorArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteReplicator(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK Replicator (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicatorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Replicator (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitReplicatorCreated(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeReplicatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReplicatorStateCreating),
		Target:  enum.Slice(types.ReplicatorStateRunning),
		Refresh: statusReplicator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitReplicatorUpdated(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeReplicatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReplicatorStateUpdating),
		Target:  enum.Slice(types.ReplicatorStateRunning),
		Refresh: statusReplicator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitReplicatorDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeReplicatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReplicatorStateRunning, types.ReplicatorStateDeleting),
		Target:  []string{},
		Refresh: statusReplicator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func statusReplicator(conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findReplicatorByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ReplicatorState), nil
	}
}

func findReplicatorByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeReplicatorOutput, error) {
	input := kafka.DescribeReplicatorInput{
		ReplicatorArn: aws.String(arn),
	}

	return findReplicator(ctx, conn, &input)
}

func findReplicator(ctx context.Context, conn *kafka.Client, input *kafka.DescribeReplicatorInput) (*kafka.DescribeReplicatorOutput, error) {
	output, err := conn.DescribeReplicator(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func flattenReplicationInfoDescriptions(apiObjects []types.ReplicationInfoDescription, sourceCluster, targetCluster *types.KafkaClusterDescription) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicationInfoDescription(apiObject, sourceCluster, targetCluster))
	}

	return tfList
}

func flattenReplicationInfoDescription(apiObject types.ReplicationInfoDescription, sourceCluster, targetCluster *types.KafkaClusterDescription) map[string]any {
	tfMap := map[string]any{}

	// An Amazon MSK cluster is referenced by ARN, an Apache Kafka cluster by ID.
	if v := sourceCluster; v != nil {
		if c := v.AmazonMskCluster; c != nil {
			tfMap["source_kafka_cluster_arn"] = aws.ToString(c.MskClusterArn)
		} else if c := v.ApacheKafkaCluster; c != nil {
			tfMap["source_kafka_cluster_id"] = aws.ToString(c.ApacheKafkaClusterId)
		}
	}

	if v := targetCluster; v != nil {
		if c := v.AmazonMskCluster; c != nil {
			tfMap["target_kafka_cluster_arn"] = aws.ToString(c.MskClusterArn)
		} else if c := v.ApacheKafkaCluster; c != nil {
			tfMap["target_kafka_cluster_id"] = aws.ToString(c.ApacheKafkaClusterId)
		}
	}

	if v := apiObject.SourceKafkaClusterAlias; v != nil {
		tfMap["source_kafka_cluster_alias"] = aws.ToString(v)
	}

	if v := apiObject.TargetKafkaClusterAlias; v != nil {
		tfMap["target_kafka_cluster_alias"] = aws.ToString(v)
	}

	if v := apiObject.TargetCompressionType; v != "" {
		tfMap["target_compression_type"] = v
	}

	if v := apiObject.TopicReplication; v != nil {
		tfMap["topic_replication"] = []any{flattenTopicReplication(v)}
	}

	if v := apiObject.ConsumerGroupReplication; v != nil {
		tfMap["consumer_group_replication"] = []any{flattenConsumerGroupReplication(v)}
	}

	return tfMap
}

func flattenConsumerGroupReplication(apiObject *types.ConsumerGroupReplication) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ConsumerGroupOffsetSyncMode; v != "" {
		tfMap["consumer_group_offset_sync_mode"] = string(v)
	}

	if v := apiObject.ConsumerGroupsToReplicate; v != nil {
		tfMap["consumer_groups_to_replicate"] = v
	}

	if v := apiObject.ConsumerGroupsToExclude; v != nil {
		tfMap["consumer_groups_to_exclude"] = v
	}

	if aws.ToBool(apiObject.DetectAndCopyNewConsumerGroups) {
		tfMap["detect_and_copy_new_consumer_groups"] = apiObject.DetectAndCopyNewConsumerGroups
	}

	if aws.ToBool(apiObject.SynchroniseConsumerGroupOffsets) {
		tfMap["synchronise_consumer_group_offsets"] = apiObject.SynchroniseConsumerGroupOffsets
	}

	return tfMap
}

func flattenTopicReplication(apiObject *types.TopicReplication) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if aws.ToBool(apiObject.CopyAccessControlListsForTopics) {
		tfMap["copy_access_control_lists_for_topics"] = apiObject.CopyAccessControlListsForTopics
	}

	if aws.ToBool(apiObject.CopyTopicConfigurations) {
		tfMap["copy_topic_configurations"] = apiObject.CopyTopicConfigurations
	}

	if aws.ToBool(apiObject.DetectAndCopyNewTopics) {
		tfMap["detect_and_copy_new_topics"] = apiObject.DetectAndCopyNewTopics
	}

	if v := apiObject.StartingPosition; v != nil {
		tfMap["starting_position"] = []any{flattenReplicationStartingPosition(v)}
	}

	if v := apiObject.TopicNameConfiguration; v != nil {
		tfMap["topic_name_configuration"] = []any{flattenReplicationTopicNameConfiguration(v)}
	}

	if v := apiObject.TopicsToReplicate; v != nil {
		tfMap["topics_to_replicate"] = v
	}

	if v := apiObject.TopicsToExclude; v != nil {
		tfMap["topics_to_exclude"] = v
	}

	return tfMap
}

func flattenReplicationStartingPosition(apiObject *types.ReplicationStartingPosition) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenReplicationTopicNameConfiguration(apiObject *types.ReplicationTopicNameConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

// reconcileKafkaClusters overlays the flattened DescribeReplicator response onto prior state,
// entry by entry (order is preserved by the API). It preserves the fields the response does
// not return for Apache Kafka clusters (apache_kafka_cluster, client_authentication,
// encryption_in_transit) and drops the vpc_config that the API echoes onto entries where the
// configuration did not set one. amazon_msk_cluster and the computed alias come from the API.
func reconcileKafkaClusters(old, flattened []any) []any { // nosemgrep:ci.kafka-in-func-name
	for i := range flattened {
		newMap, ok := flattened[i].(map[string]any)
		if !ok || i >= len(old) {
			continue
		}
		oldMap, ok := old[i].(map[string]any)
		if !ok {
			continue
		}

		for _, k := range []string{"apache_kafka_cluster", "client_authentication", "encryption_in_transit"} {
			if v, ok := oldMap[k].([]any); ok && len(v) > 0 {
				newMap[k] = v
			}
		}

		if v, ok := oldMap[names.AttrVPCConfig].([]any); !ok || len(v) == 0 {
			delete(newMap, names.AttrVPCConfig)
		}
	}

	return flattened
}

func flattenKafkaClusterDescriptions(apiObjects []types.KafkaClusterDescription) []any { // nosemgrep:ci.kafka-in-func-name
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenKafkaClusterDescription(apiObject))
	}

	return tfList
}

func flattenKafkaClusterDescription(apiObject types.KafkaClusterDescription) map[string]any { // nosemgrep:ci.kafka-in-func-name
	tfMap := map[string]any{}

	if v := apiObject.AmazonMskCluster; v != nil {
		tfMap["amazon_msk_cluster"] = []any{flattenAmazonMSKCluster(v)}
	}

	if v := apiObject.ApacheKafkaCluster; v != nil {
		tfMap["apache_kafka_cluster"] = []any{flattenApacheKafkaCluster(v)}
	}

	if v := apiObject.ClientAuthentication; v != nil {
		tfMap["client_authentication"] = []any{flattenKafkaClusterClientAuthentication(v)}
	}

	// Only surface encryption_in_transit when a custom root CA is configured. The implicit
	// TLS-only encryption sent for every Apache Kafka cluster is not user-specified and would
	// otherwise produce spurious diffs against configs that omit the block.
	if v := apiObject.EncryptionInTransit; v != nil && v.RootCaCertificate != nil {
		tfMap["encryption_in_transit"] = []any{flattenKafkaClusterEncryptionInTransit(v)}
	}

	if v := apiObject.VpcConfig; v != nil {
		tfMap[names.AttrVPCConfig] = []any{flattenKafkaClusterClientVPCConfig(v)}
	}

	return tfMap
}

func flattenApacheKafkaCluster(apiObject *types.ApacheKafkaCluster) map[string]any { // nosemgrep:ci.kafka-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ApacheKafkaClusterId; v != nil {
		tfMap["apache_kafka_cluster_id"] = aws.ToString(v)
	}

	if v := apiObject.BootstrapBrokerString; v != nil {
		tfMap["bootstrap_broker_string"] = aws.ToString(v)
	}

	return tfMap
}

func flattenKafkaClusterClientAuthentication(apiObject *types.KafkaClusterClientAuthentication) map[string]any { // nosemgrep:ci.kafka-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.MTLS; v != nil {
		tfMapMTLS := map[string]any{}

		if v := v.SecretArn; v != nil {
			tfMapMTLS["secret_arn"] = aws.ToString(v)
		}

		tfMap["mtls"] = []any{tfMapMTLS}
	}

	if v := apiObject.SaslScram; v != nil {
		tfMapSaslScram := map[string]any{}

		if v := v.Mechanism; v != "" {
			tfMapSaslScram["mechanism"] = string(v)
		}

		if v := v.SecretArn; v != nil {
			tfMapSaslScram["secret_arn"] = aws.ToString(v)
		}

		tfMap["sasl_scram"] = []any{tfMapSaslScram}
	}

	return tfMap
}

func flattenKafkaClusterEncryptionInTransit(apiObject *types.KafkaClusterEncryptionInTransit) map[string]any { // nosemgrep:ci.kafka-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.RootCaCertificate; v != nil {
		tfMap["root_ca_certificate"] = aws.ToString(v)
	}

	return tfMap
}

func flattenKafkaClusterClientVPCConfig(apiObject *types.KafkaClusterClientVpcConfig) map[string]any { // nosemgrep:ci.kafka-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_groups_ids"] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return tfMap
}

func flattenAmazonMSKCluster(apiObject *types.AmazonMskCluster) map[string]any { // nosemgrep:ci.msk-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"msk_cluster_arn": apiObject.MskClusterArn,
	}

	return tfMap
}

func flattenLogDelivery(apiObject *types.LogDelivery) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ReplicatorLogDelivery; v != nil {
		tfMap["replicator_log_delivery"] = flattenReplicatorLogDelivery(v)
	}

	return []any{tfMap}
}

func flattenReplicatorLogDelivery(apiObject *types.ReplicatorLogDelivery) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap[names.AttrCloudWatchLogs] = flattenReplicatorLogDeliveryCloudWatchLogs(v)
	}

	if v := apiObject.Firehose; v != nil {
		tfMap["firehose"] = flattenReplicatorLogDeliveryFirehose(v)
	}

	if v := apiObject.S3; v != nil {
		tfMap["s3"] = flattenReplicatorLogDeliveryS3(v)
	}

	return []any{tfMap}
}

func flattenReplicatorLogDeliveryCloudWatchLogs(apiObject *types.ReplicatorCloudWatchLogs) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func flattenReplicatorLogDeliveryFirehose(apiObject *types.ReplicatorFirehose) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func flattenReplicatorLogDeliveryS3(apiObject *types.ReplicatorS3) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.Bucket; v != nil {
		tfMap[names.AttrBucket] = aws.ToString(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return []any{tfMap}
}

func expandConsumerGroupReplicationUpdate(tfMap map[string]any) *types.ConsumerGroupReplicationUpdate {
	apiObject := &types.ConsumerGroupReplicationUpdate{}

	if v, ok := tfMap["consumer_groups_to_replicate"].(*schema.Set); ok {
		apiObject.ConsumerGroupsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["consumer_groups_to_exclude"].(*schema.Set); ok {
		apiObject.ConsumerGroupsToExclude = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["synchronise_consumer_group_offsets"].(bool); ok {
		apiObject.SynchroniseConsumerGroupOffsets = aws.Bool(v)
	}

	if v, ok := tfMap["detect_and_copy_new_consumer_groups"].(bool); ok {
		apiObject.DetectAndCopyNewConsumerGroups = aws.Bool(v)
	}

	return apiObject
}

func expandTopicReplicationUpdate(tfMap map[string]any) *types.TopicReplicationUpdate {
	apiObject := &types.TopicReplicationUpdate{}

	if v, ok := tfMap["copy_topic_configurations"].(bool); ok {
		apiObject.CopyTopicConfigurations = aws.Bool(v)
	}

	if v, ok := tfMap["copy_access_control_lists_for_topics"].(bool); ok {
		apiObject.CopyAccessControlListsForTopics = aws.Bool(v)
	}

	if v, ok := tfMap["detect_and_copy_new_topics"].(bool); ok {
		apiObject.DetectAndCopyNewTopics = aws.Bool(v)
	}

	if v, ok := tfMap["topics_to_exclude"].(*schema.Set); ok {
		apiObject.TopicsToExclude = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["topics_to_replicate"].(*schema.Set); ok {
		apiObject.TopicsToReplicate = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandReplicationInfos(tfList []any) []types.ReplicationInfo {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ReplicationInfo

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandReplicationInfo(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandReplicationInfo(tfMap map[string]any) types.ReplicationInfo {
	apiObject := types.ReplicationInfo{}

	if v, ok := tfMap["source_kafka_cluster_arn"].(string); ok && v != "" {
		apiObject.SourceKafkaClusterArn = aws.String(v)
	}

	if v, ok := tfMap["source_kafka_cluster_id"].(string); ok && v != "" {
		apiObject.SourceKafkaClusterId = aws.String(v)
	}

	if v, ok := tfMap["target_kafka_cluster_arn"].(string); ok && v != "" {
		apiObject.TargetKafkaClusterArn = aws.String(v)
	}

	if v, ok := tfMap["target_kafka_cluster_id"].(string); ok && v != "" {
		apiObject.TargetKafkaClusterId = aws.String(v)
	}

	if v, ok := tfMap["target_compression_type"].(string); ok {
		apiObject.TargetCompressionType = types.TargetCompressionType(v)
	}

	if v, ok := tfMap["topic_replication"].([]any); ok {
		apiObject.TopicReplication = expandTopicReplication(v[0].(map[string]any))
	}

	if v, ok := tfMap["consumer_group_replication"].([]any); ok {
		apiObject.ConsumerGroupReplication = expandConsumerGroupReplication(v[0].(map[string]any))
	}

	return apiObject
}

func expandConsumerGroupReplication(tfMap map[string]any) *types.ConsumerGroupReplication {
	apiObject := &types.ConsumerGroupReplication{}

	if v, ok := tfMap["consumer_group_offset_sync_mode"].(string); ok && v != "" {
		apiObject.ConsumerGroupOffsetSyncMode = types.ConsumerGroupOffsetSyncMode(v)
	}

	if v, ok := tfMap["consumer_groups_to_replicate"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ConsumerGroupsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["consumer_groups_to_exclude"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ConsumerGroupsToExclude = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["synchronise_consumer_group_offsets"].(bool); ok {
		apiObject.SynchroniseConsumerGroupOffsets = aws.Bool(v)
	}

	if v, ok := tfMap["detect_and_copy_new_consumer_groups"].(bool); ok {
		apiObject.DetectAndCopyNewConsumerGroups = aws.Bool(v)
	}

	return apiObject
}

func expandTopicReplication(tfMap map[string]any) *types.TopicReplication {
	apiObject := &types.TopicReplication{}

	if v, ok := tfMap["copy_access_control_lists_for_topics"].(bool); ok {
		apiObject.CopyAccessControlListsForTopics = aws.Bool(v)
	}

	if v, ok := tfMap["copy_topic_configurations"].(bool); ok {
		apiObject.CopyTopicConfigurations = aws.Bool(v)
	}

	if v, ok := tfMap["detect_and_copy_new_topics"].(bool); ok {
		apiObject.DetectAndCopyNewTopics = aws.Bool(v)
	}

	if v, ok := tfMap["starting_position"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.StartingPosition = expandReplicationStartingPosition(v[0].(map[string]any))
	}

	if v, ok := tfMap["topic_name_configuration"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.TopicNameConfiguration = expandReplicationTopicNameConfiguration(v[0].(map[string]any))
	}

	if v, ok := tfMap["topics_to_replicate"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TopicsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["topics_to_exclude"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TopicsToExclude = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandReplicationStartingPosition(tfMap map[string]any) *types.ReplicationStartingPosition {
	apiObject := &types.ReplicationStartingPosition{}

	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = types.ReplicationStartingPositionType(v)
	}

	return apiObject
}

func expandReplicationTopicNameConfiguration(tfMap map[string]any) *types.ReplicationTopicNameConfiguration {
	apiObject := &types.ReplicationTopicNameConfiguration{}

	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = types.ReplicationTopicNameConfigurationType(v)
	}

	return apiObject
}

func expandKafkaClusters(tfList []any) []types.KafkaCluster { // nosemgrep:ci.kafka-in-func-name
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.KafkaCluster

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandKafkaCluster(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKafkaCluster(tfMap map[string]any) types.KafkaCluster { // nosemgrep:ci.kafka-in-func-name
	apiObject := types.KafkaCluster{}

	if v, ok := tfMap[names.AttrVPCConfig].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.VpcConfig = expandKafkaClusterClientVPCConfig(v[0].(map[string]any))
	}

	if v, ok := tfMap["amazon_msk_cluster"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmazonMskCluster = expandAmazonMSKCluster(v[0].(map[string]any))
	}

	if v, ok := tfMap["apache_kafka_cluster"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ApacheKafkaCluster = expandApacheKafkaCluster(v[0].(map[string]any))
	}

	if v, ok := tfMap["client_authentication"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.ClientAuthentication = expandKafkaClusterClientAuthentication(v[0].(map[string]any))
	}

	if v, ok := tfMap["encryption_in_transit"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.EncryptionInTransit = expandKafkaClusterEncryptionInTransit(v[0].(map[string]any))
	}

	// The API requires an encryption-in-transit configuration for every Apache Kafka cluster.
	// EncryptionType only supports TLS, so default it when the user supplied no
	// encryption_in_transit block (e.g. a public-CA source that needs no custom root CA).
	if apiObject.ApacheKafkaCluster != nil && apiObject.EncryptionInTransit == nil {
		apiObject.EncryptionInTransit = &types.KafkaClusterEncryptionInTransit{
			EncryptionType: types.KafkaClusterEncryptionInTransitTypeTls,
		}
	}

	return apiObject
}

func expandApacheKafkaCluster(tfMap map[string]any) *types.ApacheKafkaCluster { // nosemgrep:ci.kafka-in-func-name
	apiObject := &types.ApacheKafkaCluster{}

	if v, ok := tfMap["apache_kafka_cluster_id"].(string); ok && v != "" {
		apiObject.ApacheKafkaClusterId = aws.String(v)
	}

	if v, ok := tfMap["bootstrap_broker_string"].(string); ok && v != "" {
		apiObject.BootstrapBrokerString = aws.String(v)
	}

	return apiObject
}

func expandKafkaClusterClientAuthentication(tfMap map[string]any) *types.KafkaClusterClientAuthentication { // nosemgrep:ci.kafka-in-func-name
	apiObject := &types.KafkaClusterClientAuthentication{}

	if v, ok := tfMap["mtls"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMapMTLS := v[0].(map[string]any)
		if v, ok := tfMapMTLS["secret_arn"].(string); ok && v != "" {
			apiObject.MTLS = &types.KafkaClusterMTLSAuthentication{
				SecretArn: aws.String(v),
			}
		}
	}

	if v, ok := tfMap["sasl_scram"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMapSaslScram := v[0].(map[string]any)
		saslScram := &types.KafkaClusterSaslScramAuthentication{}

		if v, ok := tfMapSaslScram["mechanism"].(string); ok && v != "" {
			saslScram.Mechanism = types.KafkaClusterSaslScramMechanism(v)
		}

		if v, ok := tfMapSaslScram["secret_arn"].(string); ok && v != "" {
			saslScram.SecretArn = aws.String(v)
		}

		apiObject.SaslScram = saslScram
	}

	return apiObject
}

func expandKafkaClusterEncryptionInTransit(tfMap map[string]any) *types.KafkaClusterEncryptionInTransit { // nosemgrep:ci.kafka-in-func-name
	// EncryptionType only supports TLS, so it is set implicitly rather than exposed as an
	// argument the user must always provide.
	apiObject := &types.KafkaClusterEncryptionInTransit{
		EncryptionType: types.KafkaClusterEncryptionInTransitTypeTls,
	}

	if v, ok := tfMap["root_ca_certificate"].(string); ok && v != "" {
		apiObject.RootCaCertificate = aws.String(v)
	}

	return apiObject
}

func expandKafkaClusterClientVPCConfig(tfMap map[string]any) *types.KafkaClusterClientVpcConfig { // nosemgrep:ci.kafka-in-func-name
	apiObject := &types.KafkaClusterClientVpcConfig{}

	if v, ok := tfMap["security_groups_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandAmazonMSKCluster(tfMap map[string]any) *types.AmazonMskCluster { // nosemgrep:ci.msk-in-func-name
	apiObject := &types.AmazonMskCluster{}

	if v, ok := tfMap["msk_cluster_arn"].(string); ok && v != "" {
		apiObject.MskClusterArn = aws.String(v)
	}

	return apiObject
}

func expandLogDelivery(tfList []any) *types.LogDelivery {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)

	if !ok {
		return nil
	}

	apiObject := &types.LogDelivery{}

	if v, ok := tfMap["replicator_log_delivery"].([]any); ok {
		apiObject.ReplicatorLogDelivery = expandReplicatorLogDelivery(v)
	}

	return apiObject
}

func expandReplicatorLogDelivery(tfList []any) *types.ReplicatorLogDelivery {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicatorLogDelivery{}

	if v, ok := tfMap[names.AttrCloudWatchLogs].([]any); ok {
		apiObject.CloudWatchLogs = expandReplicatorLogDeliveryCloudWatchLogs(v)
	}

	if v, ok := tfMap["firehose"].([]any); ok {
		apiObject.Firehose = expandReplicatorLogDeliveryFirehose(v)
	}

	if v, ok := tfMap["s3"].([]any); ok {
		apiObject.S3 = expandReplicatorLogDeliveryS3(v)
	}

	return apiObject
}

func expandReplicatorLogDeliveryCloudWatchLogs(tfList []any) *types.ReplicatorCloudWatchLogs {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicatorCloudWatchLogs{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		apiObject.LogGroup = aws.String(v)
	}

	return apiObject
}

func expandReplicatorLogDeliveryFirehose(tfList []any) *types.ReplicatorFirehose {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicatorFirehose{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		apiObject.DeliveryStream = aws.String(v)
	}

	return apiObject
}

func expandReplicatorLogDeliveryS3(tfList []any) *types.ReplicatorS3 {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.ReplicatorS3{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		apiObject.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}
