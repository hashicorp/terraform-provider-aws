// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_msk_replicator", name="Replicator")
// @Tags(identifierAttribute="id")
func resourceReplicator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicatorCreate,
		ReadWithoutTimeout:   resourceReplicatorRead,
		UpdateWithoutTimeout: resourceReplicatorUpdate,
		DeleteWithoutTimeout: resourceReplicatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
							Required: true,
							MaxItems: 1,
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
						names.AttrVPCConfig: {
							Type:     schema.TypeList,
							Required: true,
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
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
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
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
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
									"topics_to_exclude": {
										Type:     schema.TypeSet,
										Optional: true,
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
									"consumer_groups_to_exclude": {
										Type:     schema.TypeSet,
										Optional: true,
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
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	name := d.Get("replicator_name").(string)
	input := &kafka.CreateReplicatorInput{
		KafkaClusters:           expandKafkaClusters(d.Get("kafka_cluster").([]interface{})),
		ReplicationInfoList:     expandReplicationInfos(d.Get("replication_info_list").([]interface{})),
		ReplicatorName:          aws.String(name),
		ServiceExecutionRoleArn: aws.String(d.Get("service_execution_role_arn").(string)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateReplicator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK Replicator (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ReplicatorArn))

	if _, err := waitReplicatorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK Replicator (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	output, err := findReplicatorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kafka Replicator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Replicator (%s): %s", d.Id(), err)
	}

	sourceAlias := aws.ToString(output.ReplicationInfoList[0].SourceKafkaClusterAlias)
	targetAlias := aws.ToString(output.ReplicationInfoList[0].TargetKafkaClusterAlias)
	var sourceARN, targetARN *string

	for _, cluster := range output.KafkaClusters {
		if clusterAlias := aws.ToString(cluster.KafkaClusterAlias); clusterAlias == sourceAlias {
			sourceARN = cluster.AmazonMskCluster.MskClusterArn
		} else if clusterAlias == targetAlias {
			targetARN = cluster.AmazonMskCluster.MskClusterArn
		}
	}

	d.Set(names.AttrARN, output.ReplicatorArn)
	d.Set("current_version", output.CurrentVersion)
	d.Set(names.AttrDescription, output.ReplicatorDescription)
	d.Set("kafka_cluster", flattenKafkaClusterDescriptions(output.KafkaClusters))
	d.Set("replication_info_list", flattenReplicationInfoDescriptions(output.ReplicationInfoList, sourceARN, targetARN))
	d.Set("replicator_name", output.ReplicatorName)
	d.Set("service_execution_role_arn", output.ServiceExecutionRoleArn)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceReplicatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &kafka.UpdateReplicationInfoInput{
			CurrentVersion:        aws.String(d.Get("current_version").(string)),
			ReplicatorArn:         aws.String(d.Id()),
			SourceKafkaClusterArn: aws.String(d.Get("replication_info_list.0.source_kafka_cluster_arn").(string)),
			TargetKafkaClusterArn: aws.String(d.Get("replication_info_list.0.target_kafka_cluster_arn").(string)),
		}

		if d.HasChanges("replication_info_list.0.consumer_group_replication") {
			if v, ok := d.GetOk("replication_info_list.0.consumer_group_replication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ConsumerGroupReplication = expandConsumerGroupReplicationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChanges("replication_info_list.0.topic_replication") {
			if v, ok := d.GetOk("replication_info_list.0.topic_replication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.TopicReplication = expandTopicReplicationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		_, err := conn.UpdateReplicationInfo(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating MSK Replicator (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicatorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for MSK Replicator (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting MSK Replicator: %s", d.Id())
	_, err := conn.DeleteReplicator(ctx, &kafka.DeleteReplicatorInput{
		ReplicatorArn: aws.String(d.Id()),
	})

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
		Refresh: statusReplicator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitReplicatorUpdated(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeReplicatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReplicatorStateUpdating),
		Target:  enum.Slice(types.ReplicatorStateRunning),
		Refresh: statusReplicator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitReplicatorDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeReplicatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ReplicatorStateRunning, types.ReplicatorStateDeleting),
		Target:  []string{},
		Refresh: statusReplicator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		if stateInfo := output.StateInfo; stateInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(stateInfo.Code), aws.ToString(stateInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

func statusReplicator(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicatorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ReplicatorState), nil
	}
}

func findReplicatorByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeReplicatorOutput, error) {
	input := &kafka.DescribeReplicatorInput{
		ReplicatorArn: aws.String(arn),
	}

	output, err := conn.DescribeReplicator(ctx, input)

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

func flattenReplicationInfoDescriptions(apiObjects []types.ReplicationInfoDescription, sourceCluster, targetCluster *string) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicationInfoDescription(apiObject, sourceCluster, targetCluster))
	}

	return tfList
}

func flattenReplicationInfoDescription(apiObject types.ReplicationInfoDescription, sourceCluster, targetCluster *string) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := sourceCluster; v != nil {
		tfMap["source_kafka_cluster_arn"] = aws.ToString(v)
	}

	if v := targetCluster; v != nil {
		tfMap["target_kafka_cluster_arn"] = aws.ToString(v)
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
		tfMap["topic_replication"] = []interface{}{flattenTopicReplication(v)}
	}

	if v := apiObject.ConsumerGroupReplication; v != nil {
		tfMap["consumer_group_replication"] = []interface{}{flattenConsumerGroupReplication(v)}
	}

	return tfMap
}

func flattenConsumerGroupReplication(apiObject *types.ConsumerGroupReplication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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

func flattenTopicReplication(apiObject *types.TopicReplication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

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
		tfMap["starting_position"] = []interface{}{flattenReplicationStartingPosition(v)}
	}

	if v := apiObject.TopicsToReplicate; v != nil {
		tfMap["topics_to_replicate"] = v
	}

	if v := apiObject.TopicsToExclude; v != nil {
		tfMap["topics_to_exclude"] = v
	}

	return tfMap
}

func flattenReplicationStartingPosition(apiObject *types.ReplicationStartingPosition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenKafkaClusterDescriptions(apiObjects []types.KafkaClusterDescription) []interface{} { // nosemgrep:ci.kafka-in-func-name
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenKafkaClusterDescription(apiObject))
	}

	return tfList
}

func flattenKafkaClusterDescription(apiObject types.KafkaClusterDescription) map[string]interface{} { // nosemgrep:ci.kafka-in-func-name
	tfMap := map[string]interface{}{}

	if v := apiObject.AmazonMskCluster; v != nil {
		tfMap["amazon_msk_cluster"] = []interface{}{flattenAmazonMSKCluster(v)}
	}

	if v := apiObject.VpcConfig; v != nil {
		tfMap[names.AttrVPCConfig] = []interface{}{flattenKafkaClusterClientVPCConfig(v)}
	}

	return tfMap
}

func flattenKafkaClusterClientVPCConfig(apiObject *types.KafkaClusterClientVpcConfig) map[string]interface{} { // nosemgrep:ci.kafka-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_groups_ids"] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap[names.AttrSubnetIDs] = v
	}

	return tfMap
}

func flattenAmazonMSKCluster(apiObject *types.AmazonMskCluster) map[string]interface{} { // nosemgrep:ci.msk-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"msk_cluster_arn": apiObject.MskClusterArn,
	}

	return tfMap
}

func expandConsumerGroupReplicationUpdate(tfMap map[string]interface{}) *types.ConsumerGroupReplicationUpdate {
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

func expandTopicReplicationUpdate(tfMap map[string]interface{}) *types.TopicReplicationUpdate {
	apiObject := &types.TopicReplicationUpdate{}

	if v, ok := tfMap["topics_to_replicate"].(*schema.Set); ok {
		apiObject.TopicsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["topics_to_exclude"].(*schema.Set); ok {
		apiObject.TopicsToExclude = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["copy_topic_configurations"].(bool); ok {
		apiObject.CopyTopicConfigurations = aws.Bool(v)
	}

	if v, ok := tfMap["copy_access_control_lists_for_topics"].(bool); ok {
		apiObject.CopyAccessControlListsForTopics = aws.Bool(v)
	}

	if v, ok := tfMap["detect_and_copy_new_topics"].(bool); ok {
		apiObject.DetectAndCopyNewTopics = aws.Bool(v)
	}

	return apiObject
}

func expandReplicationInfos(tfList []interface{}) []types.ReplicationInfo {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ReplicationInfo

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandReplicationInfo(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandReplicationInfo(tfMap map[string]interface{}) types.ReplicationInfo {
	apiObject := types.ReplicationInfo{}

	if v, ok := tfMap["source_kafka_cluster_arn"].(string); ok {
		apiObject.SourceKafkaClusterArn = aws.String(v)
	}

	if v, ok := tfMap["target_kafka_cluster_arn"].(string); ok {
		apiObject.TargetKafkaClusterArn = aws.String(v)
	}

	if v, ok := tfMap["target_compression_type"].(string); ok {
		apiObject.TargetCompressionType = types.TargetCompressionType(v)
	}

	if v, ok := tfMap["topic_replication"].([]interface{}); ok {
		apiObject.TopicReplication = expandTopicReplication(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["consumer_group_replication"].([]interface{}); ok {
		apiObject.ConsumerGroupReplication = expandConsumerGroupReplication(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandConsumerGroupReplication(tfMap map[string]interface{}) *types.ConsumerGroupReplication {
	apiObject := &types.ConsumerGroupReplication{}

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

func expandTopicReplication(tfMap map[string]interface{}) *types.TopicReplication {
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

	if v, ok := tfMap["starting_position"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StartingPosition = expandReplicationStartingPosition(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["topics_to_replicate"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TopicsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["topics_to_exclude"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TopicsToExclude = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandReplicationStartingPosition(tfMap map[string]interface{}) *types.ReplicationStartingPosition {
	apiObject := &types.ReplicationStartingPosition{}

	if v, ok := tfMap[names.AttrType].(string); ok {
		apiObject.Type = types.ReplicationStartingPositionType(v)
	}

	return apiObject
}

func expandKafkaClusters(tfList []interface{}) []types.KafkaCluster { // nosemgrep:ci.kafka-in-func-name
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.KafkaCluster

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandKafkaCluster(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKafkaCluster(tfMap map[string]interface{}) types.KafkaCluster { // nosemgrep:ci.kafka-in-func-name
	apiObject := types.KafkaCluster{}

	if v, ok := tfMap[names.AttrVPCConfig].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.VpcConfig = expandKafkaClusterClientVPCConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["amazon_msk_cluster"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmazonMskCluster = expandAmazonMSKCluster(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandKafkaClusterClientVPCConfig(tfMap map[string]interface{}) *types.KafkaClusterClientVpcConfig { // nosemgrep:ci.kafka-in-func-name
	apiObject := &types.KafkaClusterClientVpcConfig{}

	if v, ok := tfMap["security_groups_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandAmazonMSKCluster(tfMap map[string]interface{}) *types.AmazonMskCluster { // nosemgrep:ci.msk-in-func-name
	apiObject := &types.AmazonMskCluster{}

	if v, ok := tfMap["msk_cluster_arn"].(string); ok && v != "" {
		apiObject.MskClusterArn = aws.String(v)
	}

	return apiObject
}
