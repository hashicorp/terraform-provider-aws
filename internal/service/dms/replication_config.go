// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_replication_config", name="Replication Config")
// @Tags(identifierAttribute="replication_config_arn")
func ResourceReplicationConfig() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"dns_name_servers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_key_id": {
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
						"preferred_maintenance_window": {
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
						"vpc_security_group_ids": {
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
			"replication_settings": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"replication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dms.MigrationTypeValue_Values(), false),
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
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	replicationConfigID := d.Get("replication_config_identifier").(string)
	input := &dms.CreateReplicationConfigInput{
		ReplicationConfigIdentifier: aws.String(replicationConfigID),
		ReplicationType:             aws.String(d.Get("replication_type").(string)),
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

	output, err := conn.CreateReplicationConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Config (%s): %s", replicationConfigID, err)
	}

	d.SetId(aws.StringValue(output.ReplicationConfig.ReplicationConfigArn))

	if d.Get("start_replication").(bool) {
		if err := startReplication(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return resourceReplicationConfigRead(ctx, d, meta)
}

func resourceReplicationConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	replicationConfig, err := FindReplicationConfigByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Config (%s): %s", d.Id(), err)
	}

	d.Set("arn", replicationConfig.ReplicationConfigArn)
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
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "start_replication") {
		if err := stopReplication(ctx, d.Id(), conn); err != nil {
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
			input.ReplicationType = aws.String(d.Get("replication_type").(string))
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

		_, err := conn.ModifyReplicationConfigWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Replication Config (%s): %s", d.Id(), err)
		}

		if d.Get("start_replication").(bool) {
			if err := startReplication(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("start_replication") {
		var err error
		if d.Get("start_replication").(bool) {
			err = startReplication(ctx, d.Id(), conn)
		} else {
			err = stopReplication(ctx, d.Id(), conn)
		}
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicationConfigRead(ctx, d, meta)...)
}

func resourceReplicationConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if err := stopReplication(ctx, d.Id(), conn); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting DMS Replication Config: %s", d.Id())
	_, err := conn.DeleteReplicationConfigWithContext(ctx, &dms.DeleteReplicationConfigInput{
		ReplicationConfigArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Config (%s): %s", d.Id(), err)
	}

	if err := waitReplicationDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication config (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func FindReplicationConfigByARN(ctx context.Context, conn *dms.DatabaseMigrationService, arn string) (*dms.ReplicationConfig, error) {
	input := &dms.DescribeReplicationConfigsInput{
		Filters: []*dms.Filter{{
			Name:   aws.String("replication-config-arn"),
			Values: aws.StringSlice([]string{arn}),
		}},
	}

	return findReplicationConfig(ctx, conn, input)
}

func findReplicationConfig(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationConfigsInput) (*dms.ReplicationConfig, error) {
	output, err := findReplicationConfigs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findReplicationConfigs(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationConfigsInput) ([]*dms.ReplicationConfig, error) {
	var output []*dms.ReplicationConfig

	err := conn.DescribeReplicationConfigsPagesWithContext(ctx, input, func(page *dms.DescribeReplicationConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationConfigs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenComputeConfig(apiObject *dms.ComputeConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_zone":            aws.StringValue(apiObject.AvailabilityZone),
		"dns_name_servers":             aws.StringValue(apiObject.DnsNameServers),
		"kms_key_id":                   aws.StringValue(apiObject.KmsKeyId),
		"max_capacity_units":           aws.Int64Value(apiObject.MaxCapacityUnits),
		"min_capacity_units":           aws.Int64Value(apiObject.MinCapacityUnits),
		"multi_az":                     aws.BoolValue(apiObject.MultiAZ),
		"preferred_maintenance_window": aws.StringValue(apiObject.PreferredMaintenanceWindow),
		"replication_subnet_group_id":  aws.StringValue(apiObject.ReplicationSubnetGroupId),
		"vpc_security_group_ids":       flex.FlattenStringSet(apiObject.VpcSecurityGroupIds),
	}

	return []interface{}{tfMap}
}

func expandComputeConfigInput(tfMap map[string]interface{}) *dms.ComputeConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &dms.ComputeConfig{}

	if v, ok := tfMap["availability_zone"].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["dns_name_servers"].(string); ok && v != "" {
		apiObject.DnsNameServers = aws.String(v)
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["max_capacity_units"].(int); ok && v != 0 {
		apiObject.MaxCapacityUnits = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_capacity_units"].(int); ok && v != 0 {
		apiObject.MinCapacityUnits = aws.Int64(int64(v))
	}

	if v, ok := tfMap["multi_az"].(bool); ok {
		apiObject.MultiAZ = aws.Bool(v)
	}

	if v, ok := tfMap["preferred_maintenance_window"].(string); ok && v != "" {
		apiObject.PreferredMaintenanceWindow = aws.String(v)
	}

	if v, ok := tfMap["replication_subnet_group_id"].(string); ok && v != "" {
		apiObject.ReplicationSubnetGroupId = aws.String(v)
	}

	if v, ok := tfMap["vpc_security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.VpcSecurityGroupIds = flex.ExpandStringSet(v)
	}

	return apiObject
}

func startReplication(ctx context.Context, id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Starting DMS Serverless Replication: (%s)", id)

	replication, _ := FindReplicationById(ctx, id, conn)

	if aws.StringValue(replication.Status) == replicationStatusRunning {
		return nil
	}

	startReplicationType := replicationTypeValueStartReplication
	if aws.StringValue(replication.Status) != replicationStatusReady {
		startReplicationType = replicationTypeValueResumeProcessing
	}

	_, err := conn.StartReplicationWithContext(ctx, &dms.StartReplicationInput{
		ReplicationConfigArn: replication.ReplicationConfigArn,
		StartReplicationType: aws.String(startReplicationType),
	})

	if err != nil {
		return fmt.Errorf("starting DMS Serverless Replication (%s): %w", id, err)
	}

	err = waitReplicationRunning(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("waiting for DMS Serverless Replication (%s) start: %w", id, err)
	}

	return nil
}

func stopReplication(ctx context.Context, id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Stopping DMS Serverless Replication: (%s)", id)

	replication, _ := FindReplicationById(ctx, id, conn)
	status := aws.StringValue(replication.Status)

	if status == replicationStatusStopped || status == replicationStatusCreated || status == replicationStatusFailed {
		return nil
	}

	_, err := conn.StopReplicationWithContext(ctx, &dms.StopReplicationInput{
		ReplicationConfigArn: aws.String(*replication.ReplicationConfigArn),
	})

	if err != nil {
		return fmt.Errorf("stopping DMS Serverless Replication (%s): %w", id, err)
	}

	err = waitReplicationStopped(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("waiting for DMS Replication (%s) stop: %w", id, err)
	}

	return nil
}
