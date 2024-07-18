// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_replication_config", name="Replication Config")
// @Tags(identifierAttribute="id")
func resourceReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationConfigCreate,
		ReadWithoutTimeout:   resourceReplicationConfigRead,
		UpdateWithoutTimeout: resourceReplicationConfigUpdate,
		DeleteWithoutTimeout: resourceReplicationConfigDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"dns_name_servers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrKMSKeyID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"max_capacity_units": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_capacity_units": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"multi_az": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						names.AttrPreferredMaintenanceWindow: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidOnceAWeekWindowFormat,
						},
						"replication_subnet_group_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validReplicationSubnetGroupID,
						},
						names.AttrVPCSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"replication_config_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// "replication_settings" is equivalent to "replication_task_settings" on "aws_dms_replication_task"
			// All changes to this field and supporting tests should be mirrored in "aws_dms_replication_task"
			"replication_settings": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.AllDiag(
					validation.ToDiagFunc(validation.StringIsJSON),
					validateReplicationSettings,
				),
				DiffSuppressFunc:      suppressEquivalentTaskSettings,
				DiffSuppressOnRefresh: true,
			},
			"replication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MigrationTypeValue](),
			},
			"resource_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"source_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"start_replication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"supplemental_settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"table_mappings": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	replicationConfigID := d.Get("replication_config_identifier").(string)
	input := &dms.CreateReplicationConfigInput{
		ReplicationConfigIdentifier: aws.String(replicationConfigID),
		ReplicationType:             awstypes.MigrationTypeValue(d.Get("replication_type").(string)),
		SourceEndpointArn:           aws.String(d.Get("source_endpoint_arn").(string)),
		TableMappings:               aws.String(d.Get("table_mappings").(string)),
		Tags:                        getTagsIn(ctx),
		TargetEndpointArn:           aws.String(d.Get("target_endpoint_arn").(string)),
	}

	if v, ok := d.GetOk("compute_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ComputeConfig = expandComputeConfigInput(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("replication_settings"); ok {
		input.ReplicationSettings = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_identifier"); ok {
		input.ResourceIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supplemental_settings"); ok {
		input.SupplementalSettings = aws.String(v.(string))
	}

	output, err := conn.CreateReplicationConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Config (%s): %s", replicationConfigID, err)
	}

	d.SetId(aws.ToString(output.ReplicationConfig.ReplicationConfigArn))

	if d.Get("start_replication").(bool) {
		if err := startReplication(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicationConfigRead(ctx, d, meta)...)
}

func resourceReplicationConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	replicationConfig, err := findReplicationConfigByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, replicationConfig.ReplicationConfigArn)
	if err := d.Set("compute_config", flattenComputeConfig(replicationConfig.ComputeConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_config: %s", err)
	}
	d.Set("replication_config_identifier", replicationConfig.ReplicationConfigIdentifier)
	d.Set("replication_settings", replicationConfig.ReplicationSettings)
	d.Set("replication_type", replicationConfig.ReplicationType)
	d.Set("source_endpoint_arn", replicationConfig.SourceEndpointArn)
	d.Set("supplemental_settings", replicationConfig.SupplementalSettings)
	d.Set("table_mappings", replicationConfig.TableMappings)
	d.Set("target_endpoint_arn", replicationConfig.TargetEndpointArn)

	return diags
}

func resourceReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "start_replication") {
		if err := stopReplication(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &dms.ModifyReplicationConfigInput{
			ReplicationConfigArn: aws.String(d.Id()),
		}

		if d.HasChange("compute_config") {
			if v, ok := d.GetOk("compute_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ComputeConfig = expandComputeConfigInput(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("replication_settings") {
			input.ReplicationSettings = aws.String(d.Get("replication_settings").(string))
		}

		if d.HasChange("replication_type") {
			input.ReplicationType = awstypes.MigrationTypeValue(d.Get("replication_type").(string))
		}

		if d.HasChange("source_endpoint_arn") {
			input.SourceEndpointArn = aws.String(d.Get("source_endpoint_arn").(string))
		}

		if d.HasChange("supplemental_settings") {
			input.SupplementalSettings = aws.String(d.Get("supplemental_settings").(string))
		}

		if d.HasChange("table_mappings") {
			input.TableMappings = aws.String(d.Get("table_mappings").(string))
		}

		if d.HasChange("target_endpoint_arn") {
			input.TargetEndpointArn = aws.String(d.Get("target_endpoint_arn").(string))
		}

		_, err := conn.ModifyReplicationConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DMS Replication Config (%s): %s", d.Id(), err)
		}

		if d.Get("start_replication").(bool) {
			if err := startReplication(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("start_replication") {
		var f func(context.Context, *dms.Client, string, time.Duration) error
		if d.Get("start_replication").(bool) {
			f = startReplication
		} else {
			f = stopReplication
		}
		if err := f(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicationConfigRead(ctx, d, meta)...)
}

func resourceReplicationConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if err := stopReplication(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting DMS Replication Config: %s", d.Id())
	_, err := conn.DeleteReplicationConfig(ctx, &dms.DeleteReplicationConfigInput{
		ReplicationConfigArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Config (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for  DMS Replication Config (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findReplicationConfigByARN(ctx context.Context, conn *dms.Client, arn string) (*awstypes.ReplicationConfig, error) {
	input := &dms.DescribeReplicationConfigsInput{
		Filters: []awstypes.Filter{{
			Name:   aws.String("replication-config-arn"),
			Values: []string{arn},
		}},
	}

	return findReplicationConfig(ctx, conn, input)
}

func findReplicationConfig(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationConfigsInput) (*awstypes.ReplicationConfig, error) {
	output, err := findReplicationConfigs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationConfigs(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationConfigsInput) ([]awstypes.ReplicationConfig, error) {
	var output []awstypes.ReplicationConfig

	pages := dms.NewDescribeReplicationConfigsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReplicationConfigs...)
	}

	return output, nil
}

func findReplicationByReplicationConfigARN(ctx context.Context, conn *dms.Client, arn string) (*awstypes.Replication, error) {
	input := &dms.DescribeReplicationsInput{
		Filters: []awstypes.Filter{{
			Name:   aws.String("replication-config-arn"),
			Values: []string{arn},
		}},
	}

	return findReplication(ctx, conn, input)
}

func findReplication(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationsInput) (*awstypes.Replication, error) {
	output, err := findReplications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplications(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationsInput) ([]awstypes.Replication, error) {
	var output []awstypes.Replication

	pages := dms.NewDescribeReplicationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Replications...)
	}

	return output, nil
}

func statusReplication(ctx context.Context, conn *dms.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationByReplicationConfigARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func setLastReplicationError(err error, replication *awstypes.Replication) {
	var errs []error

	errs = append(errs, tfslices.ApplyToAll(replication.FailureMessages, func(v string) error {
		return errors.New(v)
	})...)
	if v := aws.ToString(replication.StopReason); v != "" {
		errs = append(errs, errors.New(v))
	}

	tfresource.SetLastError(err, errors.Join(errs...))
}

func waitReplicationRunning(ctx context.Context, conn *dms.Client, arn string, timeout time.Duration) (*awstypes.Replication, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			replicationStatusReady,
			replicationStatusInitialising,
			replicationStatusMetadataResources,
			replicationStatusTestingConnection,
			replicationStatusFetchingMetadata,
			replicationStatusCalculatingCapacity,
			replicationStatusProvisioningCapacity,
			replicationStatusReplicationStarting,
		},
		Target:     []string{replicationStatusRunning, replicationStatusStopped},
		Refresh:    statusReplication(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Replication); ok {
		setLastReplicationError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationStopped(ctx context.Context, conn *dms.Client, arn string, timeout time.Duration) (*awstypes.Replication, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationStatusStopping, replicationStatusRunning},
		Target:     []string{replicationStatusStopped},
		Refresh:    statusReplication(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      60 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Replication); ok {
		setLastReplicationError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationDeleted(ctx context.Context, conn *dms.Client, arn string, timeout time.Duration) (*awstypes.Replication, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusDeleting, replicationStatusStopped},
		Target:     []string{},
		Refresh:    statusReplication(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Replication); ok {
		setLastReplicationError(err, output)
		return output, err
	}

	return nil, err
}

func startReplication(ctx context.Context, conn *dms.Client, arn string, timeout time.Duration) error {
	replication, err := findReplicationByReplicationConfigARN(ctx, conn, arn)

	if err != nil {
		return fmt.Errorf("reading DMS Replication Config (%s) replication: %s", arn, err)
	}

	replicationStatus := aws.ToString(replication.Status)
	if replicationStatus == replicationStatusRunning {
		return nil
	}

	startReplicationType := replicationTypeValueStartReplication
	if replicationStatus != replicationStatusReady {
		startReplicationType = replicationTypeValueResumeProcessing
	}
	input := &dms.StartReplicationInput{
		ReplicationConfigArn: aws.String(arn),
		StartReplicationType: aws.String(startReplicationType),
	}

	_, err = conn.StartReplication(ctx, input)

	if err != nil {
		return fmt.Errorf("starting DMS Serverless Replication (%s): %w", arn, err)
	}

	if _, err := waitReplicationRunning(ctx, conn, arn, timeout); err != nil {
		return fmt.Errorf("waiting for DMS Serverless Replication (%s) start: %w", arn, err)
	}

	return nil
}

func stopReplication(ctx context.Context, conn *dms.Client, arn string, timeout time.Duration) error {
	replication, err := findReplicationByReplicationConfigARN(ctx, conn, arn)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading DMS Replication Config (%s) replication: %s", arn, err)
	}

	if replicationStatus := aws.ToString(replication.Status); replicationStatus == replicationStatusStopped || replicationStatus == replicationStatusCreated || replicationStatus == replicationStatusFailed {
		return nil
	}

	input := &dms.StopReplicationInput{
		ReplicationConfigArn: aws.String(arn),
	}

	_, err = conn.StopReplication(ctx, input)

	if err != nil {
		return fmt.Errorf("stopping DMS Serverless Replication (%s): %w", arn, err)
	}

	if _, err := waitReplicationStopped(ctx, conn, arn, timeout); err != nil {
		return fmt.Errorf("waiting for DMS Serverless Replication (%s) stop: %w", arn, err)
	}

	return nil
}

func flattenComputeConfig(apiObject *awstypes.ComputeConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrAvailabilityZone:           aws.ToString(apiObject.AvailabilityZone),
		"dns_name_servers":                   aws.ToString(apiObject.DnsNameServers),
		names.AttrKMSKeyID:                   aws.ToString(apiObject.KmsKeyId),
		"max_capacity_units":                 aws.ToInt32(apiObject.MaxCapacityUnits),
		"min_capacity_units":                 aws.ToInt32(apiObject.MinCapacityUnits),
		"multi_az":                           aws.ToBool(apiObject.MultiAZ),
		names.AttrPreferredMaintenanceWindow: aws.ToString(apiObject.PreferredMaintenanceWindow),
		"replication_subnet_group_id":        aws.ToString(apiObject.ReplicationSubnetGroupId),
		names.AttrVPCSecurityGroupIDs:        apiObject.VpcSecurityGroupIds,
	}

	return []interface{}{tfMap}
}

func expandComputeConfigInput(tfMap map[string]interface{}) *awstypes.ComputeConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ComputeConfig{}

	if v, ok := tfMap[names.AttrAvailabilityZone].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["dns_name_servers"].(string); ok && v != "" {
		apiObject.DnsNameServers = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["max_capacity_units"].(int); ok && v != 0 {
		apiObject.MaxCapacityUnits = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_capacity_units"].(int); ok && v != 0 {
		apiObject.MinCapacityUnits = aws.Int32(int32(v))
	}

	if v, ok := tfMap["multi_az"].(bool); ok {
		apiObject.MultiAZ = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrPreferredMaintenanceWindow].(string); ok && v != "" {
		apiObject.PreferredMaintenanceWindow = aws.String(v)
	}

	if v, ok := tfMap["replication_subnet_group_id"].(string); ok && v != "" {
		apiObject.ReplicationSubnetGroupId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVPCSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.VpcSecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}
