// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

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
func ResourceCluster() *schema.Resource {
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

func resourceClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrName, d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Service:   "ecs",
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
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
	partition := meta.(*conns.AWSClient).Partition

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, clusterReadTimeout, func() (interface{}, error) {
		return FindClusterByNameOrARN(ctx, conn, partition, d.Id())
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

func FindClusterByNameOrARN(ctx context.Context, conn *ecs.Client, partition string, nameOrARN string) (*awstypes.Cluster, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: []string{nameOrARN},
		Include:  []awstypes.ClusterField{awstypes.ClusterFieldTags, awstypes.ClusterFieldConfigurations, awstypes.ClusterFieldSettings},
	}

	output, err := conn.DescribeClusters(ctx, input)

	// Some partitions (e.g. ISO) may not support tagging.
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Include = []awstypes.ClusterField{awstypes.ClusterFieldConfigurations, awstypes.ClusterFieldSettings}

		output, err = conn.DescribeClusters(ctx, input)
	}

	// Some partitions (e.g. ISO) may not support describe including configuration.
	if errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Include = []awstypes.ClusterField{awstypes.ClusterFieldSettings}
		output, err = conn.DescribeClusters(ctx, input)
	}

	if errs.IsA[*awstypes.ClusterNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	if status := aws.ToString(output.Clusters[0].Status); status == clusterStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return tfresource.AssertFirstValueResult(output.Clusters)
}

func statusCluster(ctx context.Context, conn *ecs.Client, clusterArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		parsedArn, err := arn.Parse(clusterArn)

		if err != nil {
			return nil, "", err
		}

		cluster, err := FindClusterByNameOrARN(ctx, conn, parsedArn.Partition, clusterArn)

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
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusProvisioning},
		Target:  []string{clusterStatusActive},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterAvailableTimeout,
		Delay:   clusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *ecs.Client, arn string) (*awstypes.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusActive, clusterStatusDeprovisioning},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Cluster); ok {
		return v, err
	}

	return nil, err
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	log.Printf("[DEBUG] Deleting ECS Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenIsOneOf4[*awstypes.ClusterContainsContainerInstancesException, *awstypes.ClusterContainsServicesException, *awstypes.ClusterContainsTasksException, *awstypes.UpdateInProgressException](ctx, clusterDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteCluster(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(d.Id()),
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func retryClusterCreate(ctx context.Context, conn *ecs.Client, input *ecs.CreateClusterInput) (*ecs.CreateClusterOutput, error) {
	var output *ecs.CreateClusterOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateCluster(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Unable to assume the service linked role") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCluster(ctx, input)
	}

	return output, err
}

func expandClusterSettings(configured *schema.Set) []awstypes.ClusterSetting {
	list := configured.List()
	if len(list) == 0 {
		return nil
	}

	settings := make([]awstypes.ClusterSetting, 0, len(list))

	for _, raw := range list {
		data := raw.(map[string]interface{})

		setting := awstypes.ClusterSetting{
			Name:  awstypes.ClusterSettingName(data[names.AttrName].(string)),
			Value: aws.String(data[names.AttrValue].(string)),
		}

		settings = append(settings, setting)
	}

	return settings
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

func flattenClusterSettings(list []awstypes.ClusterSetting) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{
			names.AttrName:  string(setting.Name),
			names.AttrValue: aws.ToString(setting.Value),
		}

		result = append(result, l)
	}
	return result
}

func flattenClusterConfiguration(apiObject *awstypes.ClusterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ExecuteCommandConfiguration != nil {
		tfMap["execute_command_configuration"] = flattenClusterConfigurationExecuteCommandConfiguration(apiObject.ExecuteCommandConfiguration)
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

func expandClusterConfiguration(nc []interface{}) *awstypes.ClusterConfiguration {
	if len(nc) == 0 || nc[0] == nil {
		return &awstypes.ClusterConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &awstypes.ClusterConfiguration{}
	if v, ok := raw["execute_command_configuration"].([]interface{}); ok && len(v) > 0 {
		config.ExecuteCommandConfiguration = expandClusterConfigurationExecuteCommandConfiguration(v)
	}

	return config
}

func expandClusterConfigurationExecuteCommandConfiguration(nc []interface{}) *awstypes.ExecuteCommandConfiguration {
	if len(nc) == 0 || nc[0] == nil {
		return &awstypes.ExecuteCommandConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &awstypes.ExecuteCommandConfiguration{}
	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandClusterConfigurationExecuteCommandLogConfiguration(v)
	}

	if v, ok := raw[names.AttrKMSKeyID].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}

	if v, ok := raw["logging"].(string); ok && v != "" {
		config.Logging = awstypes.ExecuteCommandLogging(v)
	}

	return config
}

func expandClusterConfigurationExecuteCommandLogConfiguration(nc []interface{}) *awstypes.ExecuteCommandLogConfiguration {
	if len(nc) == 0 || nc[0] == nil {
		return &awstypes.ExecuteCommandLogConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &awstypes.ExecuteCommandLogConfiguration{}

	if v, ok := raw["cloud_watch_log_group_name"].(string); ok && v != "" {
		config.CloudWatchLogGroupName = aws.String(v)
	}

	if v, ok := raw[names.AttrS3BucketName].(string); ok && v != "" {
		config.S3BucketName = aws.String(v)
	}

	if v, ok := raw[names.AttrS3KeyPrefix].(string); ok && v != "" {
		config.S3KeyPrefix = aws.String(v)
	}

	if v, ok := raw["cloud_watch_encryption_enabled"].(bool); ok {
		config.CloudWatchEncryptionEnabled = v
	}

	if v, ok := raw["s3_bucket_encryption_enabled"].(bool); ok {
		config.S3EncryptionEnabled = v
	}

	return config
}
