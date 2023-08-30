// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
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
									"kms_key_id": {
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
												"s3_bucket_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"s3_key_prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"logging": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ecs.ExecuteCommandLogging_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"name": {
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
						"namespace": {
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
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.ClusterSettingName_Values(), false),
						},
						"value": {
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
	d.Set("name", d.Id())
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	clusterName := d.Get("name").(string)
	input := &ecs.CreateClusterInput{
		ClusterName: aws.String(clusterName),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
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
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = retryClusterCreate(ctx, conn, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECS Cluster (%s): %s", clusterName, err)
	}

	d.SetId(aws.StringValue(output.Cluster.ClusterArn))

	if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, clusterReadTimeout, func() (interface{}, error) {
		return FindClusterByNameOrARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Cluster (%s): %s", d.Id(), err)
	}

	cluster := outputRaw.(*ecs.Cluster)
	d.Set("arn", cluster.ClusterArn)
	if cluster.Configuration != nil {
		if err := d.Set("configuration", flattenClusterConfiguration(cluster.Configuration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
		}
	}
	d.Set("name", cluster.ClusterName)
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
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	if d.HasChanges("configuration", "service_connect_defaults", "setting") {
		input := &ecs.UpdateClusterInput{
			Cluster: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
			input.Configuration = expandClusterConfiguration(v.([]interface{}))
		}

		if v, ok := d.GetOk("service_connect_defaults"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ServiceConnectDefaults = expandClusterServiceConnectDefaultsRequest(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("setting"); ok {
			input.Settings = expandClusterSettings(v.(*schema.Set))
		}

		_, err := conn.UpdateClusterWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating ECS Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitClusterAvailable(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for ECS Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return nil
}

func FindClusterByNameOrARN(ctx context.Context, conn *ecs.ECS, nameOrARN string) (*ecs.Cluster, error) {
	input := &ecs.DescribeClustersInput{
		Clusters: aws.StringSlice([]string{nameOrARN}),
		Include:  aws.StringSlice([]string{ecs.ClusterFieldTags, ecs.ClusterFieldConfigurations, ecs.ClusterFieldSettings}),
	}

	output, err := conn.DescribeClustersWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tagging.
	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Include = aws.StringSlice([]string{ecs.ClusterFieldConfigurations, ecs.ClusterFieldSettings})

		output, err = conn.DescribeClustersWithContext(ctx, input)
	}

	// Some partitions (e.g. ISO) may not support describe including configuration.
	if errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Include = aws.StringSlice([]string{ecs.ClusterFieldSettings})

		output, err = conn.DescribeClustersWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 || output.Clusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	if status := aws.StringValue(output.Clusters[0].Status); status == clusterStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output.Clusters[0], nil
}

func statusCluster(ctx context.Context, conn *ecs.ECS, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByNameOrARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.StringValue(cluster.Status), err
	}
}

func waitClusterAvailable(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.Cluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusProvisioning},
		Target:  []string{clusterStatusActive},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterAvailableTimeout,
		Delay:   clusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{clusterStatusActive, clusterStatusDeprovisioning},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSConn(ctx)

	log.Printf("[DEBUG] Deleting ECS Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, clusterDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteClusterWithContext(ctx, &ecs.DeleteClusterInput{
			Cluster: aws.String(d.Id()),
		})
	},
		ecs.ErrCodeClusterContainsContainerInstancesException,
		ecs.ErrCodeClusterContainsServicesException,
		ecs.ErrCodeClusterContainsTasksException,
		ecs.ErrCodeUpdateInProgressException,
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECS Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ECS Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func retryClusterCreate(ctx context.Context, conn *ecs.ECS, input *ecs.CreateClusterInput) (*ecs.CreateClusterOutput, error) {
	var output *ecs.CreateClusterOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateClusterWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "Unable to assume the service linked role") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateClusterWithContext(ctx, input)
	}

	return output, err
}

func expandClusterSettings(configured *schema.Set) []*ecs.ClusterSetting {
	list := configured.List()
	if len(list) == 0 {
		return nil
	}

	settings := make([]*ecs.ClusterSetting, 0, len(list))

	for _, raw := range list {
		data := raw.(map[string]interface{})

		setting := &ecs.ClusterSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		settings = append(settings, setting)
	}

	return settings
}

func expandClusterServiceConnectDefaultsRequest(tfMap map[string]interface{}) *ecs.ClusterServiceConnectDefaultsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ecs.ClusterServiceConnectDefaultsRequest{}

	if v, ok := tfMap["namespace"].(string); ok && v != "" {
		apiObject.Namespace = aws.String(v)
	}

	return apiObject
}

func flattenClusterServiceConnectDefaults(apiObject *ecs.ClusterServiceConnectDefaults) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Namespace; v != nil {
		tfMap["namespace"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenClusterSettings(list []*ecs.ClusterSetting) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{
			"name":  aws.StringValue(setting.Name),
			"value": aws.StringValue(setting.Value),
		}

		result = append(result, l)
	}
	return result
}

func flattenClusterConfiguration(apiObject *ecs.ClusterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ExecuteCommandConfiguration != nil {
		tfMap["execute_command_configuration"] = flattenClusterConfigurationExecuteCommandConfiguration(apiObject.ExecuteCommandConfiguration)
	}
	return []interface{}{tfMap}
}

func flattenClusterConfigurationExecuteCommandConfiguration(apiObject *ecs.ExecuteCommandConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.KmsKeyId != nil {
		tfMap["kms_key_id"] = aws.StringValue(apiObject.KmsKeyId)
	}

	if apiObject.LogConfiguration != nil {
		tfMap["log_configuration"] = flattenClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject.LogConfiguration)
	}

	if apiObject.Logging != nil {
		tfMap["logging"] = aws.StringValue(apiObject.Logging)
	}

	return []interface{}{tfMap}
}

func flattenClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject *ecs.ExecuteCommandLogConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["cloud_watch_encryption_enabled"] = aws.BoolValue(apiObject.CloudWatchEncryptionEnabled)
	tfMap["s3_bucket_encryption_enabled"] = aws.BoolValue(apiObject.S3EncryptionEnabled)

	if apiObject.CloudWatchLogGroupName != nil {
		tfMap["cloud_watch_log_group_name"] = aws.StringValue(apiObject.CloudWatchLogGroupName)
	}

	if apiObject.S3BucketName != nil {
		tfMap["s3_bucket_name"] = aws.StringValue(apiObject.S3BucketName)
	}

	if apiObject.S3KeyPrefix != nil {
		tfMap["s3_key_prefix"] = aws.StringValue(apiObject.S3KeyPrefix)
	}

	return []interface{}{tfMap}
}

func expandClusterConfiguration(nc []interface{}) *ecs.ClusterConfiguration {
	if len(nc) == 0 {
		return &ecs.ClusterConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ClusterConfiguration{}
	if v, ok := raw["execute_command_configuration"].([]interface{}); ok && len(v) > 0 {
		config.ExecuteCommandConfiguration = expandClusterConfigurationExecuteCommandConfiguration(v)
	}

	return config
}

func expandClusterConfigurationExecuteCommandConfiguration(nc []interface{}) *ecs.ExecuteCommandConfiguration {
	if len(nc) == 0 {
		return &ecs.ExecuteCommandConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ExecuteCommandConfiguration{}
	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandClusterConfigurationExecuteCommandLogConfiguration(v)
	}

	if v, ok := raw["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}

	if v, ok := raw["logging"].(string); ok && v != "" {
		config.Logging = aws.String(v)
	}

	return config
}

func expandClusterConfigurationExecuteCommandLogConfiguration(nc []interface{}) *ecs.ExecuteCommandLogConfiguration {
	if len(nc) == 0 {
		return &ecs.ExecuteCommandLogConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ExecuteCommandLogConfiguration{}

	if v, ok := raw["cloud_watch_log_group_name"].(string); ok && v != "" {
		config.CloudWatchLogGroupName = aws.String(v)
	}

	if v, ok := raw["s3_bucket_name"].(string); ok && v != "" {
		config.S3BucketName = aws.String(v)
	}

	if v, ok := raw["s3_key_prefix"].(string); ok && v != "" {
		config.S3KeyPrefix = aws.String(v)
	}

	if v, ok := raw["cloud_watch_encryption_enabled"].(bool); ok {
		config.CloudWatchEncryptionEnabled = aws.Bool(v)
	}

	if v, ok := raw["s3_bucket_encryption_enabled"].(bool); ok {
		config.S3EncryptionEnabled = aws.Bool(v)
	}

	return config
}
