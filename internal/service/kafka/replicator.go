// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_replicator", name="Replicator")
// @Tags(identifierAttribute="id")
func ResourceReplicator() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"current_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
						"vpc_config": {
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
									"subnet_ids": {
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

const (
	ResNameReplicator = "Replicator"
)

func resourceReplicatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	in := &kafka.CreateReplicatorInput{
		KafkaClusters:           expandClusters(d.Get("kafka_cluster").([]interface{})),
		ReplicationInfoList:     expandReplicationInfoList(d.Get("replication_info_list").([]interface{})),
		ReplicatorName:          aws.String(d.Get("replicator_name").(string)),
		ServiceExecutionRoleArn: aws.String(d.Get("service_execution_role_arn").(string)),
		Tags:                    getTagsInV2(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateReplicator(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionCreating, ResNameReplicator, d.Get("replicator_name").(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionCreating, ResNameReplicator, d.Get("replicator_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.ReplicatorArn))

	if _, err := waitReplicatorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionWaitingForCreation, ResNameReplicator, d.Id(), err)
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	out, err := findReplicatorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kafka Replicator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionReading, ResNameReplicator, d.Id(), err)
	}

	sourceAlias := out.ReplicationInfoList[0].SourceKafkaClusterAlias
	targetAlias := out.ReplicationInfoList[0].TargetKafkaClusterAlias
	clustersArn := out.KafkaClusters

	var sourceARN *string
	var targetARN *string

	for _, arn := range clustersArn {
		clusterAlias := aws.ToString(arn.KafkaClusterAlias)
		if clusterAlias == aws.ToString(sourceAlias) {
			sourceARN = arn.AmazonMskCluster.MskClusterArn
		} else if clusterAlias == aws.ToString(targetAlias) {
			targetARN = arn.AmazonMskCluster.MskClusterArn
		}
	}

	d.Set("arn", out.ReplicatorArn)
	d.Set("current_version", out.CurrentVersion)
	d.Set("replicator_name", out.ReplicatorName)
	d.Set("description", out.ReplicatorDescription)
	d.Set("service_execution_role_arn", out.ServiceExecutionRoleArn)
	d.Set("kafka_cluster", flattenClusters(out.KafkaClusters))
	d.Set("replication_info_list", flattenReplicationInfoList(out.ReplicationInfoList, sourceARN, targetARN))

	setTagsOutV2(ctx, out.Tags)

	return diags
}

func resourceReplicatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		in := &kafka.UpdateReplicationInfoInput{
			ReplicatorArn:         aws.String(d.Id()),
			CurrentVersion:        aws.String(d.Get("current_version").(string)),
			SourceKafkaClusterArn: aws.String(d.Get("replication_info_list.0.source_kafka_cluster_arn").(string)),
			TargetKafkaClusterArn: aws.String(d.Get("replication_info_list.0.target_kafka_cluster_arn").(string)),
		}

		if d.HasChanges("replication_info_list.0.consumer_group_replication") {
			if v, ok := d.GetOk("replication_info_list.0.consumer_group_replication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				in.ConsumerGroupReplication = expandConsumerGroupReplicationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChanges("replication_info_list.0.topic_replication") {
			if v, ok := d.GetOk("replication_info_list.0.topic_replication"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				in.TopicReplication = expandTopicReplicationUpdate(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		log.Printf("[DEBUG] Updating Kafka Replicator (%s): %#v", d.Id(), in)

		out, err := conn.UpdateReplicationInfo(ctx, in)

		if err != nil {
			return create.AppendDiagError(diags, names.Kafka, create.ErrActionUpdating, ResNameReplicator, d.Id(), err)
		}

		if _, err := waitReplicatorUpdated(ctx, conn, aws.ToString(out.ReplicatorArn), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.Kafka, create.ErrActionWaitingForUpdate, ResNameReplicator, d.Id(), err)
		}
	}

	return append(diags, resourceReplicatorRead(ctx, d, meta)...)
}

func resourceReplicatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting Kafka Replicator %s", d.Id())

	_, err := conn.DeleteReplicator(ctx, &kafka.DeleteReplicatorInput{
		ReplicatorArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionDeleting, ResNameReplicator, d.Id(), err)
	}

	if _, err := waitReplicatorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.Kafka, create.ErrActionWaitingForDeletion, ResNameReplicator, d.Id(), err)
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
	if out, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		return out, err
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
	if out, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		return out, err
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
	if out, ok := outputRaw.(*kafka.DescribeReplicatorOutput); ok {
		return out, err
	}

	return nil, err
}

func statusReplicator(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findReplicatorByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.ReplicatorState), nil
	}
}

func findReplicatorByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeReplicatorOutput, error) {
	in := &kafka.DescribeReplicatorInput{
		ReplicatorArn: aws.String(arn),
	}

	out, err := conn.DescribeReplicator(ctx, in)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenReplicationInfoList(apiObjects []types.ReplicationInfoDescription, sourceCluster, targetCluster *string) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplicationInfo(apiObject, sourceCluster, targetCluster))
	}

	return tfList
}

func flattenReplicationInfo(apiObject types.ReplicationInfoDescription, sourceCluster, targetCluster *string) map[string]interface{} {
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
		tfMap["consumer_groups_to_replicate"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.ConsumerGroupsToExclude; v != nil {
		tfMap["consumer_groups_to_exclude"] = flex.FlattenStringValueSet(v)
	}

	if aws.ToBool(apiObject.SynchroniseConsumerGroupOffsets) {
		tfMap["synchronise_consumer_group_offsets"] = apiObject.SynchroniseConsumerGroupOffsets
	}

	if aws.ToBool(apiObject.DetectAndCopyNewConsumerGroups) {
		tfMap["detect_and_copy_new_consumer_groups"] = apiObject.DetectAndCopyNewConsumerGroups
	}

	return tfMap
}

func flattenTopicReplication(apiObject *types.TopicReplication) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TopicsToReplicate; v != nil {
		tfMap["topics_to_replicate"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.TopicsToExclude; v != nil {
		tfMap["topics_to_exclude"] = flex.FlattenStringValueSet(v)
	}

	if aws.ToBool(apiObject.CopyTopicConfigurations) {
		tfMap["copy_topic_configurations"] = apiObject.CopyTopicConfigurations
	}

	if aws.ToBool(apiObject.CopyAccessControlListsForTopics) {
		tfMap["copy_access_control_lists_for_topics"] = apiObject.CopyAccessControlListsForTopics
	}

	if aws.ToBool(apiObject.DetectAndCopyNewTopics) {
		tfMap["detect_and_copy_new_topics"] = apiObject.CopyAccessControlListsForTopics
	}

	return tfMap
}

func flattenClusters(apiObjects []types.KafkaClusterDescription) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenCluster(apiObject))
	}

	return tfList
}

func flattenCluster(apiObject types.KafkaClusterDescription) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AmazonMskCluster; v != nil {
		tfMap["amazon_msk_cluster"] = []interface{}{flattenAmazonCluster(v)}
	}

	if v := apiObject.VpcConfig; v != nil {
		tfMap["vpc_config"] = []interface{}{flattenClusterClientVPCConfig(v)}
	}

	return tfMap
}

func flattenClusterClientVPCConfig(apiObject *types.KafkaClusterClientVpcConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_groups_ids"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = flex.FlattenStringValueSet(v)
	}

	return tfMap
}

func flattenAmazonCluster(apiObject *types.AmazonMskCluster) map[string]interface{} {
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

func expandReplicationInfoList(tfList []interface{}) []types.ReplicationInfo {
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

	if v, ok := tfMap["topics_to_replicate"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.TopicsToReplicate = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["topics_to_exclude"].(*schema.Set); ok && v.Len() > 0 {
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

func expandClusters(tfList []interface{}) []types.KafkaCluster {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.KafkaCluster

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandCluster(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCluster(tfMap map[string]interface{}) types.KafkaCluster {
	apiObject := types.KafkaCluster{}

	if v, ok := tfMap["vpc_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.VpcConfig = expandClusterClientVPCConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["amazon_msk_cluster"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AmazonMskCluster = expandAmazonCluster(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandClusterClientVPCConfig(tfMap map[string]interface{}) *types.KafkaClusterClientVpcConfig {
	apiObject := &types.KafkaClusterClientVpcConfig{}

	if v, ok := tfMap["security_groups_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandAmazonCluster(tfMap map[string]interface{}) *types.AmazonMskCluster {
	apiObject := &types.AmazonMskCluster{}

	if v, ok := tfMap["msk_cluster_arn"].(string); ok && v != "" {
		apiObject.MskClusterArn = aws.String(v)
	}

	return apiObject
}
