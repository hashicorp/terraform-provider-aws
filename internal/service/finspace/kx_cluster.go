// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_finspace_kx_cluster", name="Kx Cluster")
// @Tags(identifierAttribute="arn")
func ResourceKxCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxClusterCreate,
		ReadWithoutTimeout:   resourceKxClusterRead,
		UpdateWithoutTimeout: resourceKxClusterUpdate,
		DeleteWithoutTimeout: resourceKxClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute), // Tags only
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling_metric": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice(
								enum.Slice(types.AutoScalingMetricCpuUtilizationPercentage), true),
						},
						"max_node_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 5),
						},
						"metric_target": {
							Type:         schema.TypeFloat,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatBetween(0, 100),
						},
						"min_node_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 5),
						},
						"scale_in_cooldown_seconds": {
							Type:         schema.TypeFloat,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatBetween(0, 100000),
						},
						"scale_out_cooldown_seconds": {
							Type:         schema.TypeFloat,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatBetween(0, 100000),
						},
					},
				},
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"az_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KxAzMode](),
			},
			"cache_storage_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1200, 33600),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(8, 10),
						},
					},
				},
			},
			"capacity_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 5),
						},
						"node_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 32),
						},
					},
				},
			},
			"code": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(3, 255),
						},
						"s3_key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(3, 1024),
						},
						"s3_object_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(3, 63),
						},
					},
				},
			},
			"command_line_arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				ValidateDiagFunc: verify.ValidAllDiag(
					validation.MapKeyLenBetween(1, 50),
					validation.MapValueLenBetween(1, 50),
				),
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_configurations": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cache_type": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											"CACHE_1000",
										}, true),
									},
									"db_paths": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"changeset_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 26),
						},
						"database_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(3, 63),
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"execution_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"initialization_script": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"release_label": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 16),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"savedown_storage_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice(
								enum.Slice(types.KxSavedownStorageTypeSds01), true),
						},
						"size": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(4, 16000),
						},
					},
				},
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KxClusterType](),
			},
			"vpc_configuration": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(types.IPAddressTypeIpV4), true),
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
						},
						"vpc_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxCluster = "Kx Cluster"

	kxClusterIDPartCount = 2
)

func resourceKxClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	environmentId := d.Get("environment_id").(string)
	clusterName := d.Get("name").(string)
	idParts := []string{
		environmentId,
		clusterName,
	}
	rID, err := flex.FlattenResourceId(idParts, kxClusterIDPartCount, false)
	if err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxCluster, d.Get("name").(string), err)...)
	}
	d.SetId(rID)

	in := &finspace.CreateKxClusterInput{
		EnvironmentId:         aws.String(environmentId),
		ClusterName:           aws.String(clusterName),
		ClusterType:           types.KxClusterType(d.Get("type").(string)),
		ReleaseLabel:          aws.String(d.Get("release_label").(string)),
		AzMode:                types.KxAzMode(d.Get("az_mode").(string)),
		CapacityConfiguration: expandCapacityConfiguration(d.Get("capacity_configuration").([]interface{})),
		ClientToken:           aws.String(id.UniqueId()),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.ClusterDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("initialization_script"); ok {
		in.InitializationScript = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role"); ok {
		in.ExecutionRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_id"); ok {
		in.AvailabilityZoneId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("command_line_arguments"); ok && len(v.(map[string]interface{})) > 0 {
		in.CommandLineArguments = expandCommandLineArguments(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("vpc_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.VpcConfiguration = expandVPCConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("auto_scaling_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.AutoScalingConfiguration = expandAutoScalingConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("database"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Databases = expandDatabases(v.([]interface{}))
	}

	if v, ok := d.GetOk("savedown_storage_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.SavedownStorageConfiguration = expandSavedownStorageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("cache_storage_configurations"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.CacheStorageConfigurations = expandCacheStorageConfigurations(v.([]interface{}))
	}

	if v, ok := d.GetOk("code"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Code = expandCode(v.([]interface{}))
	}

	out, err := conn.CreateKxCluster(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionCreating, ResNameKxCluster, d.Get("name").(string), err)...)
	}

	if out == nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionCreating, ResNameKxCluster, d.Get("name").(string), errors.New("empty output"))...)
	}

	if _, err := waitKxClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxCluster, d.Id(), err)...)
	}

	return append(diags, resourceKxClusterRead(ctx, d, meta)...)
}

func resourceKxClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := findKxClusterByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxCluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionReading, ResNameKxCluster, d.Id(), err)...)
	}

	d.Set("status", out.Status)
	d.Set("status_reason", out.StatusReason)
	d.Set("created_timestamp", out.CreatedTimestamp.String())
	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())
	d.Set("name", out.ClusterName)
	d.Set("type", out.ClusterType)
	d.Set("release_label", out.ReleaseLabel)
	d.Set("description", out.ClusterDescription)
	d.Set("az_mode", out.AzMode)
	d.Set("availability_zone_id", out.AvailabilityZoneId)
	d.Set("execution_role", out.ExecutionRole)
	d.Set("initialization_script", out.InitializationScript)

	if err := d.Set("capacity_configuration", flattenCapacityConfiguration(out.CapacityConfiguration)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if err := d.Set("vpc_configuration", flattenVPCConfiguration(out.VpcConfiguration)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if err := d.Set("code", flattenCode(out.Code)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if err := d.Set("auto_scaling_configuration", flattenAutoScalingConfiguration(out.AutoScalingConfiguration)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if err := d.Set("savedown_storage_configuration", flattenSavedownStorageConfiguration(
		out.SavedownStorageConfiguration)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if err := d.Set("cache_storage_configurations", flattenCacheStorageConfigurations(
		out.CacheStorageConfigurations)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	if d.IsNewResource() {
		if err := d.Set("database", flattenDatabases(out.Databases)); err != nil {
			return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
		}
	}

	if err := d.Set("command_line_arguments", flattenCommandLineArguments(out.CommandLineArguments)); err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}

	// compose cluster ARN using environment ARN
	parts, err := flex.ExpandResourceId(d.Id(), kxUserIDPartCount, false)
	if err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}
	env, err := findKxEnvironmentByID(ctx, conn, parts[0])
	if err != nil {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionSetting, ResNameKxCluster, d.Id(), err)...)
	}
	arn := fmt.Sprintf("%s/kxCluster/%s", aws.ToString(env.EnvironmentArn), aws.ToString(out.ClusterName))
	d.Set("arn", arn)
	d.Set("environment_id", parts[0])

	return diags
}

func resourceKxClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Tags only.
	return append(diags, resourceKxClusterRead(ctx, d, meta)...)
}

func resourceKxClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace KxCluster %s", d.Id())
	_, err := conn.DeleteKxCluster(ctx, &finspace.DeleteKxClusterInput{
		ClusterName:   aws.String(d.Get("name").(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return append(diags, create.DiagError(names.FinSpace, create.ErrActionDeleting, ResNameKxCluster, d.Id(), err)...)
	}

	_, err = waitKxClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil && !tfresource.NotFound(err) {
		return append(diags, create.DiagError(names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxCluster, d.Id(), err)...)
	}

	return diags
}

func waitKxClusterCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxClusterStatusPending, types.KxClusterStatusCreating),
		Target:                    enum.Slice(types.KxClusterStatusRunning),
		Refresh:                   statusKxCluster(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxClusterDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.KxClusterStatusDeleting),
		Target:  enum.Slice(types.KxClusterStatusDeleted),
		Refresh: statusKxCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxCluster(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKxClusterByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findKxClusterByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxClusterOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxUserIDPartCount, false)
	if err != nil {
		return nil, err
	}
	in := &finspace.GetKxClusterInput{
		EnvironmentId: aws.String(parts[0]),
		ClusterName:   aws.String(parts[1]),
	}

	out, err := conn.GetKxCluster(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ClusterName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandCapacityConfiguration(tfList []interface{}) *types.CapacityConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.CapacityConfiguration{}

	if v, ok := tfMap["node_type"].(string); ok && v != "" {
		a.NodeType = aws.String(v)
	}

	if v, ok := tfMap["node_count"].(int); ok && v != 0 {
		a.NodeCount = aws.Int32(int32(v))
	}

	return a
}

func expandAutoScalingConfiguration(tfList []interface{}) *types.AutoScalingConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.AutoScalingConfiguration{}

	if v, ok := tfMap["auto_scaling_metric"].(string); ok && v != "" {
		a.AutoScalingMetric = types.AutoScalingMetric(v)
	}

	if v, ok := tfMap["min_node_count"].(int); ok && v != 0 {
		a.MinNodeCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_node_count"].(int); ok && v != 0 {
		a.MaxNodeCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["metric_target"].(float64); ok && v != 0 {
		a.MetricTarget = aws.Float64(v)
	}

	if v, ok := tfMap["scale_in_cooldown_seconds"].(float64); ok && v != 0 {
		a.ScaleInCooldownSeconds = aws.Float64(v)
	}

	if v, ok := tfMap["scale_out_cooldown_seconds"].(float64); ok && v != 0 {
		a.ScaleOutCooldownSeconds = aws.Float64(v)
	}

	return a
}

func expandSavedownStorageConfiguration(tfList []interface{}) *types.KxSavedownStorageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.KxSavedownStorageConfiguration{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = types.KxSavedownStorageType(v)
	}

	if v, ok := tfMap["size"].(int); ok && v != 0 {
		a.Size = int32(v)
	}

	return a
}

func expandVPCConfiguration(tfList []interface{}) *types.VpcConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.VpcConfiguration{}

	if v, ok := tfMap["vpc_id"].(string); ok && v != "" {
		a.VpcId = aws.String(v)
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		a.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		a.SubnetIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["ip_address_type"].(string); ok && v != "" {
		a.IpAddressType = types.IPAddressType(v)
	}

	return a
}

func expandCacheStorageConfiguration(tfMap map[string]interface{}) *types.KxCacheStorageConfiguration {
	if tfMap == nil {
		return nil
	}

	a := &types.KxCacheStorageConfiguration{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = &v
	}

	if v, ok := tfMap["size"].(int); ok {
		a.Size = aws.Int32(int32(v))
	}

	return a
}

func expandCacheStorageConfigurations(tfList []interface{}) []types.KxCacheStorageConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.KxCacheStorageConfiguration

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandCacheStorageConfiguration(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandDatabases(tfList []interface{}) []types.KxDatabaseConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.KxDatabaseConfiguration

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandDatabase(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandDatabase(tfMap map[string]interface{}) *types.KxDatabaseConfiguration {
	if tfMap == nil {
		return nil
	}

	a := &types.KxDatabaseConfiguration{}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		a.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["cache_configurations"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		a.CacheConfigurations = expandCacheConfigurations(v.([]interface{}))
	}

	if v, ok := tfMap["changeset_id"].(string); ok && v != "" {
		a.ChangesetId = aws.String(v)
	}

	return a
}

func expandCacheConfigurations(tfList []interface{}) []types.KxDatabaseCacheConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.KxDatabaseCacheConfiguration

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandCacheConfiguration(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandCacheConfiguration(tfMap map[string]interface{}) *types.KxDatabaseCacheConfiguration {
	if tfMap == nil {
		return nil
	}

	a := &types.KxDatabaseCacheConfiguration{}

	if v, ok := tfMap["cache_type"].(string); ok && v != "" {
		a.CacheType = &v
	}

	if v, ok := tfMap["db_paths"].(*schema.Set); ok && v.Len() > 0 {
		a.DbPaths = flex.ExpandStringValueSet(v)
	}

	return a
}

func expandCode(tfList []interface{}) *types.CodeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.CodeConfiguration{}

	if v, ok := tfMap["s3_bucket"].(string); ok && v != "" {
		a.S3Bucket = aws.String(v)
	}

	if v, ok := tfMap["s3_key"].(string); ok && v != "" {
		a.S3Key = aws.String(v)
	}

	if v, ok := tfMap["s3_object_version"].(string); ok && v != "" {
		a.S3ObjectVersion = aws.String(v)
	}

	return a
}

func expandCommandLineArgument(k string, v string) *types.KxCommandLineArgument {
	if k == "" || v == "" {
		return nil
	}

	a := &types.KxCommandLineArgument{
		Key:   aws.String(k),
		Value: aws.String(v),
	}
	return a
}

func expandCommandLineArguments(tfMap map[string]interface{}) []types.KxCommandLineArgument {
	if tfMap == nil {
		return nil
	}

	var s []types.KxCommandLineArgument

	for k, v := range tfMap {
		a := expandCommandLineArgument(k, v.(string))

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func flattenCapacityConfiguration(apiObject *types.CapacityConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.NodeType; v != nil {
		m["node_type"] = aws.ToString(v)
	}

	if v := apiObject.NodeCount; v != nil {
		m["node_count"] = aws.ToInt32(v)
	}

	return []interface{}{m}
}

func flattenAutoScalingConfiguration(apiObject *types.AutoScalingConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AutoScalingMetric; v != "" {
		m["auto_scaling_metric"] = v
	}

	if v := apiObject.MinNodeCount; v != nil {
		m["min_node_count"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxNodeCount; v != nil {
		m["max_node_count"] = aws.ToInt32(v)
	}

	if v := apiObject.MetricTarget; v != nil {
		m["metric_target"] = aws.ToFloat64(v)
	}

	if v := apiObject.ScaleInCooldownSeconds; v != nil {
		m["scale_in_cooldown_seconds"] = aws.ToFloat64(v)
	}

	if v := apiObject.ScaleOutCooldownSeconds; v != nil {
		m["scale_out_cooldown_seconds"] = aws.ToFloat64(v)
	}

	return []interface{}{m}
}

func flattenSavedownStorageConfiguration(apiObject *types.KxSavedownStorageConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Type; v != "" {
		m["type"] = v
	}

	if v := apiObject.Size; v >= 4 && v <= 16000 {
		m["size"] = v
	}

	return []interface{}{m}
}

func flattenVPCConfiguration(apiObject *types.VpcConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.VpcId; v != nil {
		m["vpc_id"] = aws.ToString(v)
	}

	if v := apiObject.SecurityGroupIds; v != nil {
		m["security_group_ids"] = v
	}

	if v := apiObject.SubnetIds; v != nil {
		m["subnet_ids"] = v
	}

	if v := apiObject.IpAddressType; v != "" {
		m["ip_address_type"] = string(v)
	}

	return []interface{}{m}
}

func flattenCode(apiObject *types.CodeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.S3Bucket; v != nil {
		m["s3_bucket"] = aws.ToString(v)
	}

	if v := apiObject.S3Key; v != nil {
		m["s3_key"] = aws.ToString(v)
	}

	if v := apiObject.S3ObjectVersion; v != nil {
		m["s3_object_version"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenCacheStorageConfiguration(apiObject *types.KxCacheStorageConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Type; aws.ToString(v) != "" {
		m["type"] = aws.ToString(v)
	}

	if v := apiObject.Size; v != nil {
		m["size"] = aws.ToInt32(v)
	}

	return m
}

func flattenCacheStorageConfigurations(apiObjects []types.KxCacheStorageConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenCacheStorageConfiguration(&apiObject))
	}

	return l
}

func flattenCacheConfiguration(apiObject *types.KxDatabaseCacheConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CacheType; aws.ToString(v) != "" {
		m["cache_type"] = aws.ToString(v)
	}

	if v := apiObject.DbPaths; v != nil {
		m["db_paths"] = v
	}

	return m
}

func flattenCacheConfigurations(apiObjects []types.KxDatabaseCacheConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenCacheConfiguration(&apiObject))
	}

	return l
}

func flattenDatabase(apiObject *types.KxDatabaseConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DatabaseName; v != nil {
		m["database_name"] = aws.ToString(v)
	}

	if v := apiObject.CacheConfigurations; v != nil {
		m["cache_configurations"] = flattenCacheConfigurations(v)
	}

	if v := apiObject.ChangesetId; v != nil {
		m["changeset_id"] = aws.ToString(v)
	}

	return m
}

func flattenDatabases(apiObjects []types.KxDatabaseConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenDatabase(&apiObject))
	}

	return l
}

func flattenCommandLineArguments(apiObjects []types.KxCommandLineArgument) map[string]string {
	if len(apiObjects) == 0 {
		return nil
	}

	m := make(map[string]string)

	for _, apiObject := range apiObjects {
		m[aws.ToString(apiObject.Key)] = aws.ToString(apiObject.Value)
	}

	return m
}
