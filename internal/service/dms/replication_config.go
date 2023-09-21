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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
		input.ComputeConfig = expandComputeConfigInput(v.([]interface{}))
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

	response, err := conn.DescribeReplicationConfigsWithContext(ctx, &dms.DescribeReplicationConfigsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-config-id"),
				Values: []*string{aws.String(d.Id())},
			},
		},
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		log.Printf("[WARN] DMS Serverless Replication Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing DMS Serverless Replication Config (%s): %s", d.Id(), err)
	}

	if response == nil || len(response.ReplicationConfigs) == 0 || response.ReplicationConfigs[0] == nil {
		log.Printf("[WARN] DMS Serverless Replication Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	replicationConfig := response.ReplicationConfigs[0]

	d.Set("replication_config_arn", replicationConfig.ReplicationConfigArn)
	d.Set("replication_config_identifier", replicationConfig.ReplicationConfigIdentifier)
	d.Set("replication_type", replicationConfig.ReplicationType)
	d.Set("source_endpoint_arn", replicationConfig.SourceEndpointArn)
	d.Set("table_mappings", replicationConfig.TableMappings)
	d.Set("target_endpoint_arn", replicationConfig.TargetEndpointArn)
	d.Set("replication_settings", replicationConfig.ReplicationSettings)
	d.Set("supplemental_settings", replicationConfig.SupplementalSettings)

	if err := d.Set("compute_config", flattenComputeConfig(replicationConfig.ComputeConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting compute_config: %s", err)
	}

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
				input.ComputeConfig = expandComputeConfigInput(v.([]interface{}))
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

func flattenComputeConfig(object *dms.ComputeConfig) []interface{} {
	computeConfig := map[string]interface{}{
		"availability_zone":            aws.StringValue(object.AvailabilityZone),
		"dns_name_servers":             aws.StringValue(object.DnsNameServers),
		"kms_key_id":                   aws.StringValue(object.KmsKeyId),
		"max_capacity_units":           aws.Int64Value(object.MaxCapacityUnits),
		"min_capacity_units":           aws.Int64Value(object.MinCapacityUnits),
		"multi_az":                     aws.BoolValue(object.MultiAZ),
		"preferred_maintenance_window": aws.StringValue(object.PreferredMaintenanceWindow),
		"replication_subnet_group_id":  aws.StringValue(object.ReplicationSubnetGroupId),
		"vpc_security_group_ids":       flex.FlattenStringSet(object.VpcSecurityGroupIds),
	}

	return []interface{}{computeConfig}
}

func expandComputeConfigInput(v []interface{}) *dms.ComputeConfig {
	if v[0] == nil {
		return nil
	}

	computeConfigResource := v[0].(map[string]interface{})
	computeConfig := dms.ComputeConfig{}

	if v, ok := computeConfigResource["availability_zone"]; ok && v.(string) != "" {
		computeConfig.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := computeConfigResource["dns_name_servers"]; ok && v.(string) != "" {
		computeConfig.DnsNameServers = aws.String(v.(string))
	}

	if v, ok := computeConfigResource["kms_key_id"]; ok && v.(string) != "" {
		computeConfig.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := computeConfigResource["max_capacity_units"]; ok {
		computeConfig.MaxCapacityUnits = aws.Int64(int64(v.(int)))
	}

	if v, ok := computeConfigResource["min_capacity_units"]; ok {
		computeConfig.MinCapacityUnits = aws.Int64(int64(v.(int)))
	}

	if v, ok := computeConfigResource["multi_az"]; ok {
		computeConfig.MultiAZ = aws.Bool(v.(bool))
	}

	if v, ok := computeConfigResource["preferred_maintenance_window"]; ok && v.(string) != "" {
		computeConfig.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := computeConfigResource["replication_subnet_group_id"]; ok && v.(string) != "" {
		computeConfig.ReplicationSubnetGroupId = aws.String(v.(string))
	}

	if v := computeConfigResource["vpc_security_group_ids"].(*schema.Set); v.Len() > 0 {
		computeConfig.VpcSecurityGroupIds = flex.ExpandStringSet(v)
	}

	return &computeConfig
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
