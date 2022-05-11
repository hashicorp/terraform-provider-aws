package ecs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	ecsClusterTimeoutDelete = 10 * time.Minute
	ecsClusterTimeoutUpdate = 10 * time.Minute
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceClusterCreate,
		Read:   resourceClusterRead,
		Update: resourceClusterUpdate,
		Delete: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceClusterImport,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateClusterName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_providers": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_ecs_cluster_capacity_providers resource instead",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
												"s3_bucket_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"s3_bucket_encryption_enabled": {
													Type:     schema.TypeBool,
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
			"default_capacity_provider_strategy": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				Deprecated: "Use the aws_ecs_cluster_capacity_providers resource instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating ECS cluster %s", clusterName)

	input := &ecs.CreateClusterInput{
		ClusterName:                     aws.String(clusterName),
		DefaultCapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
	}

	if v, ok := d.GetOk("capacity_providers"); ok {
		input.CapacityProviders = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("setting"); ok {
		input.Settings = expandClusterSettings(v.(*schema.Set))
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
		input.Configuration = expandClusterConfiguration(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	// CreateCluster will create the ECS IAM Service Linked Role on first ECS provision
	// This process does not complete before the initial API call finishes.
	out, err := retryClusterCreate(conn, input)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if input.Tags != nil && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] ECS tagging failed creating Cluster (%s) with tags: %s. Trying create without tags.", clusterName, err)
		input.Tags = nil

		out, err = retryClusterCreate(conn, input)
	}

	if err != nil {
		return fmt.Errorf("failed creating ECS Cluster (%s): %w", clusterName, err)
	}

	log.Printf("[DEBUG] ECS cluster %s created", aws.StringValue(out.Cluster.ClusterArn))

	d.SetId(aws.StringValue(out.Cluster.ClusterArn))

	if _, err := waitClusterAvailable(context.Background(), conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available while creating: %w", d.Id(), err)
	}

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, d.Id(), nil, tags)

		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			// If default tags only, log and continue. Otherwise, error.
			log.Printf("[WARN] ECS tagging failed adding tags after create for Cluster (%s): %s", d.Id(), err)
			return resourceClusterRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed adding tags after create for Cluster (%s): %w", d.Id(), err)
		}
	}

	return resourceClusterRead(d, meta)
}

func resourceClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var cluster *ecs.Cluster
	err := resource.Retry(clusterReadTimeout, func() *resource.RetryError {
		var err error
		cluster, err = FindClusterByNameOrARN(context.Background(), conn, d.Id())

		if d.IsNewResource() && tfresource.NotFound(err) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		cluster, err = FindClusterByNameOrARN(context.Background(), conn, d.Id())
	}

	if tfresource.NotFound(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS Cluster (%s): %s", d.Id(), err)
	}

	// Status==INACTIVE means deleted cluster
	if aws.StringValue(cluster.Status) == "INACTIVE" {
		log.Printf("[WARN] ECS Cluster (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("name", cluster.ClusterName)

	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return fmt.Errorf("error setting capacity_providers: %w", err)
	}
	if err := d.Set("default_capacity_provider_strategy", flattenCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting default_capacity_provider_strategy: %w", err)
	}

	if err := d.Set("setting", flattenClusterSettings(cluster.Settings)); err != nil {
		return fmt.Errorf("error setting setting: %w", err)
	}

	if cluster.Configuration != nil {
		if err := d.Set("configuration", flattenClusterConfiguration(cluster.Configuration)); err != nil {
			return fmt.Errorf("error setting configuration: %w", err)
		}
	}

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
	conn := meta.(*conns.AWSClient).ECSConn

	if d.HasChanges("setting", "configuration") {
		input := ecs.UpdateClusterInput{
			Cluster: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("setting"); ok {
			input.Settings = expandClusterSettings(v.(*schema.Set))
		}

		if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
			input.Configuration = expandClusterConfiguration(v.([]interface{}))
		}

		_, err := conn.UpdateCluster(&input)
		if err != nil {
			return fmt.Errorf("error changing ECS cluster (%s): %w", d.Id(), err)
		}

		if _, err := waitClusterAvailable(context.Background(), conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available while updating setting and configuration: %w", d.Id(), err)
		}
	}

	if d.HasChanges("capacity_providers", "default_capacity_provider_strategy") {
		input := ecs.PutClusterCapacityProvidersInput{
			Cluster:                         aws.String(d.Id()),
			CapacityProviders:               flex.ExpandStringSet(d.Get("capacity_providers").(*schema.Set)),
			DefaultCapacityProviderStrategy: expandCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
		}

		err := retryClusterCapacityProvidersPut(context.Background(), conn, &input)

		if err != nil {
			return fmt.Errorf("error changing ECS cluster capacity provider settings (%s): %w", d.Id(), err)
		}

		if _, err := waitClusterAvailable(context.Background(), conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available while updating capacity_providers, default_capacity_provider_strategy: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Id(), o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] ECS tagging failed updating tags for Cluster (%s): %s", d.Id(), err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("ECS tagging failed updating tags for Cluster (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func resourceClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	log.Printf("[DEBUG] Deleting ECS cluster %s", d.Id())
	input := &ecs.DeleteClusterInput{
		Cluster: aws.String(d.Id()),
	}
	err := resource.Retry(ecsClusterTimeoutDelete, func() *resource.RetryError {
		_, err := conn.DeleteCluster(input)

		if err == nil {
			log.Printf("[DEBUG] ECS cluster %s deleted", d.Id())
			return nil
		}

		if tfawserr.ErrCodeEquals(err, "ClusterContainsContainerInstancesException") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, "ClusterContainsServicesException") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, "ClusterContainsTasksException") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeUpdateInProgressException) {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCluster(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting ECS cluster: %s", err)
	}

	if _, err := waitClusterDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Cluster (%s) to become Deleted: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] ECS cluster %q deleted", d.Id())
	return nil
}

func retryClusterCreate(conn *ecs.ECS, input *ecs.CreateClusterInput) (*ecs.CreateClusterOutput, error) {
	var output *ecs.CreateClusterOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateCluster(input)

		if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "Unable to assume the service linked role") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateCluster(input)
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
