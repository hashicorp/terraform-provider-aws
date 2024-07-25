// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ecs_cluster", name="Cluster")
// @Tags(identifierAttribute="id")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"execute_command_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cloud_watch_encryption_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"cloud_watch_log_group_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"s3_bucket_encryption_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												names.AttrS3BucketName: {
													Type:     schema.TypeString,
													Optional: true,
												},
												names.AttrS3KeyPrefix: {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"logging": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.ExecuteCommandLoggingDefault,
										ValidateDiagFunc: enum.Validate[awstypes.ExecuteCommandLogging](),
									},
								},
							},
						},
						"managed_storage_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fargate_ephemeral_storage_kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateClusterName,
			},
			"service_connect_defaults": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrNamespace: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ClusterSettingName](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	clusterName := d.Get(names.AttrName).(string)
	input := &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 {
		input.Configuration = expandClusterConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("service_connect_defaults"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ServiceConnectDefaults = expandClusterServiceConnectDefaultsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("setting"); ok {
		input.Settings = expandClusterSettings(v.(*schema.Set))
	}

	// CreateCluster will create the ECS IAM Service Linked Role on first ECS provision
	// This process does not complete before the initial API call finishes.
	output, err := retryClusterCreate(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = retryClusterCreate(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Cluster (%s): %s", clusterName, err)
	}

	d.SetId(aws.ToString(output.Cluster.ClusterArn))

	if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
			return append(diags, resourceClusterRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ECS Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	const (
		timeout = 2 * time.Second
	)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, timeout, func() (interface{}, error) {
		return findClusterByNameOrARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Cluster (%s): %s", d.Id(), err)
	}

	cluster := outputRaw.(*awstypes.Cluster)
	d.Set(names.AttrARN, cluster.ClusterArn)
	if cluster.Configuration != nil {
		if err := d.Set(names.AttrConfiguration, flattenClusterConfiguration(cluster.Configuration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
		}
	}
	d.Set(names.AttrName, cluster.ClusterName)
	if cluster.ServiceConnectDefaults != nil {
		if err := d.Set("service_connect_defaults", []interface{}{flattenClusterServiceConnectDefaults(cluster.ServiceConnectDefaults)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting service_connect_defaults: %s", err)
		}
	} else {
		d.Set("service_connect_defaults", nil)
	}
	if err := d.Set("setting", flattenClusterSettings(cluster.Settings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting setting: %s", err)
	}

	setTagsOut(ctx, cluster.Tags)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	if d.HasChanges(names.AttrConfiguration, "service_connect_defaults", "setting") {
		input := &ecs.UpdateClusterInput{
			Cluster: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrConfiguration); ok && len(v.([]interface{})) > 0 {
			input.Configuration = expandClusterConfiguration(v.([]interface{}))
		}

		if v, ok := d.GetOk("service_connect_defaults"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ServiceConnectDefaults = expandClusterServiceConnectDefaultsRequest(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("setting"); ok {
			input.Settings = expandClusterSettings(v.(*schema.Set))
		}

		_, err := conn.UpdateCluster(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ECS Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Cluster: %s", d.Id())
	const (
		timeout = 10 * time.Minute
	)
	_, err := tfresource.RetryWhenIsOneOf4[*awstypes.ClusterContainsContainerInstancesException, *awstypes.ClusterContainsServicesException, *awstypes.ClusterContainsTasksException, *awstypes.UpdateInProgressException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(d.Id()),
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrName, d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   "ecs",
		Resource:  "cluster/" + d.Id(),
	}.String())

	return []*schema.ResourceData{d}, nil
}

func retryClusterCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateClusterInput) (*ecs.CreateClusterOutput, error) {
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateCluster(ctx, input)
	}, "Unable to assume the service linked role")

	if err != nil {
		return nil, err
	}

	return outputRaw.(*ecs.CreateClusterOutput), nil
}

func findCluster(ctx context.Context, conn *ecs.Client, input *ecs.DescribeClustersInput) (*awstypes.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClusters(ctx context.Context, conn *ecs.Client, input *ecs.DescribeClustersInput) ([]awstypes.Cluster, error) {
	output, err := conn.DescribeClusters(ctx, input)

	if errs.IsA[*awstypes.ClusterNotFoundException](err) {
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

	return output.Clusters, nil
}

func findClusterByNameOrARN(ctx context.Context, conn *ecs.Client, nameOrARN string) (*awstypes.Cluster, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: []string{nameOrARN},
		Include:  []awstypes.ClusterField{awstypes.ClusterFieldTags, awstypes.ClusterFieldConfigurations, awstypes.ClusterFieldSettings},
	}

	output, err := findCluster(ctx, conn, input)

	// Some partitions (e.g. ISO) may not support tagging.
	partition := partitionFromConn(conn)
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Include = []awstypes.ClusterField{awstypes.ClusterFieldConfigurations, awstypes.ClusterFieldSettings}

		output, err = findCluster(ctx, conn, input)
	}

	// Some partitions (e.g. ISO) may not support describe including configuration.
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Include = []awstypes.ClusterField{awstypes.ClusterFieldSettings}

		output, err = findCluster(ctx, conn, input)
	}

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == clusterStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusCluster(ctx context.Context, conn *ecs.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := findClusterByNameOrARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.ToString(cluster.Status), err
	}
}

func waitClusterAvailable(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.Cluster, error) { //nolint:unparam
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusProvisioning},
		Target:  []string{clusterStatusActive},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *ecs.Client, arn string, timeout time.Duration) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusActive, clusterStatusDeprovisioning},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Cluster); ok {
		return v, err
	}

	return nil, err
}

func expandClusterSettings(tfSet *schema.Set) []awstypes.ClusterSetting {
	tfList := tfSet.List()
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.ClusterSetting, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.ClusterSetting{
			Name:  awstypes.ClusterSettingName(tfMap[names.AttrName].(string)),
			Value: aws.String(tfMap[names.AttrValue].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandClusterServiceConnectDefaultsRequest(tfMap map[string]interface{}) *awstypes.ClusterServiceConnectDefaultsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ClusterServiceConnectDefaultsRequest{}

	if v, ok := tfMap[names.AttrNamespace].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}

	return apiObject
}

func flattenClusterServiceConnectDefaults(apiObject *awstypes.ClusterServiceConnectDefaults) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Namespace; v != nil {
		tfMap[names.AttrNamespace] = aws.ToString(v)
	}

	return tfMap
}

func flattenClusterSettings(apiObjects []awstypes.ClusterSetting) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrName:  string(apiObject.Name),
			names.AttrValue: aws.ToString(apiObject.Value),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenClusterConfiguration(apiObject *awstypes.ClusterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ExecuteCommandConfiguration != nil {
		tfMap["execute_command_configuration"] = flattenClusterConfigurationExecuteCommandConfiguration(apiObject.ExecuteCommandConfiguration)
	}

	if apiObject.ManagedStorageConfiguration != nil {
		tfMap["managed_storage_configuration"] = flattenManagedStorageConfiguration(apiObject.ManagedStorageConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenClusterConfigurationExecuteCommandConfiguration(apiObject *awstypes.ExecuteCommandConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.KmsKeyId != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	if apiObject.LogConfiguration != nil {
		tfMap["log_configuration"] = flattenClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject.LogConfiguration)
	}

	tfMap["logging"] = string(apiObject.Logging)

	return []interface{}{tfMap}
}

func flattenClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject *awstypes.ExecuteCommandLogConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["cloud_watch_encryption_enabled"] = apiObject.CloudWatchEncryptionEnabled
	tfMap["s3_bucket_encryption_enabled"] = apiObject.S3EncryptionEnabled

	if apiObject.CloudWatchLogGroupName != nil {
		tfMap["cloud_watch_log_group_name"] = aws.ToString(apiObject.CloudWatchLogGroupName)
	}

	if apiObject.S3BucketName != nil {
		tfMap[names.AttrS3BucketName] = aws.ToString(apiObject.S3BucketName)
	}

	if apiObject.S3KeyPrefix != nil {
		tfMap[names.AttrS3KeyPrefix] = aws.ToString(apiObject.S3KeyPrefix)
	}

	return []interface{}{tfMap}
}

func flattenManagedStorageConfiguration(apiObject *awstypes.ManagedStorageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FargateEphemeralStorageKmsKeyId != nil {
		tfMap["fargate_ephemeral_storage_kms_key_id"] = aws.ToString(apiObject.FargateEphemeralStorageKmsKeyId)
	}

	if apiObject.KmsKeyId != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(apiObject.KmsKeyId)
	}

	return []interface{}{tfMap}
}

func expandClusterConfiguration(tfList []interface{}) *awstypes.ClusterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return &awstypes.ClusterConfiguration{}
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.ClusterConfiguration{}

	if v, ok := tfMap["execute_command_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ExecuteCommandConfiguration = expandClusterConfigurationExecuteCommandConfiguration(v)
	}

	if v, ok := tfMap["managed_storage_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ManagedStorageConfiguration = expandManagedStorageConfiguration(v)
	}

	return apiObject
}

func expandClusterConfigurationExecuteCommandConfiguration(tfList []interface{}) *awstypes.ExecuteCommandConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return &awstypes.ExecuteCommandConfiguration{}
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.ExecuteCommandConfiguration{}

	if v, ok := tfMap["log_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.LogConfiguration = expandClusterConfigurationExecuteCommandLogConfiguration(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["logging"].(string); ok && v != "" {
		apiObject.Logging = awstypes.ExecuteCommandLogging(v)
	}

	return apiObject
}

func expandClusterConfigurationExecuteCommandLogConfiguration(tfList []interface{}) *awstypes.ExecuteCommandLogConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return &awstypes.ExecuteCommandLogConfiguration{}
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.ExecuteCommandLogConfiguration{}

	if v, ok := tfMap["cloud_watch_log_group_name"].(string); ok && v != "" {
		apiObject.CloudWatchLogGroupName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrS3BucketName].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	if v, ok := tfMap["cloud_watch_encryption_enabled"].(bool); ok {
		apiObject.CloudWatchEncryptionEnabled = v
	}

	if v, ok := tfMap["s3_bucket_encryption_enabled"].(bool); ok {
		apiObject.S3EncryptionEnabled = v
	}

	return apiObject
}

func expandManagedStorageConfiguration(tfList []interface{}) *awstypes.ManagedStorageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return &awstypes.ManagedStorageConfiguration{}
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.ManagedStorageConfiguration{}

	if v, ok := tfMap["fargate_ephemeral_storage_kms_key_id"].(string); ok && v != "" {
		apiObject.FargateEphemeralStorageKmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	return apiObject
}
