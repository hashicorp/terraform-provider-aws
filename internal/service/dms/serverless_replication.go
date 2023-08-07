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

// @SDKResource("aws_dms_serverless_replication", name="Serverless Replication")
// @Tags(identifierAttribute="replication_config_arn")
func ResourceServerlessReplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServerlessReplicationCreate,
		ReadWithoutTimeout:   resourceServerlessReplicationRead,
		UpdateWithoutTimeout: resourceServerlessReplicationUpdate,
		DeleteWithoutTimeout: resourceServerlessReplicationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"compute_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							ForceNew: true,
						},
						"dns_name_servers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"kms_key_id": {
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
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
							Computed: true,
							Optional: true,
						},
						"preferred_maintenance_window": {
							Type:         schema.TypeString,
							Computed:     true,
							Optional:     true,
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
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
							Optional: true,
						},
					},
				},
			},
			"start_replication": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"replication_config_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_config_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"replication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dms.MigrationTypeValue_Values(), false),
			},
			"source_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"table_mappings": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"target_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"replication_settings": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"resource_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"supplemental_settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServerlessReplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	apiObject := &dms.CreateReplicationConfigInput{
		ReplicationConfigIdentifier: aws.String(d.Get("replication_config_identifier").(string)),
		ReplicationType:             aws.String(d.Get("replication_type").(string)),
		SourceEndpointArn:           aws.String(d.Get("source_endpoint_arn").(string)),
		TableMappings:               aws.String(d.Get("table_mappings").(string)),
		TargetEndpointArn:           aws.String(d.Get("target_endpoint_arn").(string)),
		ReplicationSettings:         aws.String(d.Get("replication_settings").(string)),
		ResourceIdentifier:          aws.String(d.Get("resource_identifier").(string)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("compute_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.ComputeConfig = expandComputeConfigInput(v.([]interface{}))
	}

	// Field can only be set if the engine type is of type Neptune
	if v, ok := d.GetOk("supplemental_settings"); ok && len(v.(string)) > 0 {
		apiObject.SupplementalSettings = aws.String(d.Get("supplemental_settings").(string))
	}

	log.Println("[DEBUG] DMS create serverless replication config:", apiObject)

	_, err := conn.CreateReplicationConfigWithContext(ctx, apiObject)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Serverless Replication Config: %s", err)
	}

	d.SetId(d.Get("replication_config_identifier").(string))

	if d.Get("start_replication").(bool) {
		if err := startReplication(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return resourceServerlessReplicationRead(ctx, d, meta)
}

func resourceServerlessReplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	response, err := conn.DescribeReplicationConfigs(&dms.DescribeReplicationConfigsInput{
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

func resourceServerlessReplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	request := &dms.ModifyReplicationConfigInput{
		ReplicationConfigArn: aws.String(d.Get("replication_config_arn").(string)),
	}
	hasChanges := false

	if d.HasChange("compute_config") {
		if v, ok := d.GetOk("compute_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			request.ComputeConfig = expandComputeConfigInput(v.([]interface{}))
		}
		hasChanges = true
	}

	if d.HasChange("replication_type") {
		request.ReplicationType = aws.String(d.Get("replication_type").(string))
		hasChanges = true
	}

	if d.HasChange("source_endpoint_arn") {
		request.SourceEndpointArn = aws.String(d.Get("source_endpoint_arn").(string))
		hasChanges = true
	}

	if d.HasChange("table_mappings") {
		request.TableMappings = aws.String(d.Get("table_mappings").(string))
		hasChanges = true
	}

	if d.HasChange("target_endpoint_arn") {
		request.TargetEndpointArn = aws.String(d.Get("target_endpoint_arn").(string))
		hasChanges = true
	}

	if d.HasChange("replication_settings") {
		request.ReplicationSettings = aws.String(d.Get("replication_settings").(string))
		hasChanges = true
	}

	if d.HasChange("supplemental_settings") {
		request.SupplementalSettings = aws.String(d.Get("supplemental_settings").(string))
		hasChanges = true
	}

	if d.HasChange("start_replication") {
		if d.Get("start_replication").(bool) {
			if err := startReplication(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			if err := stopReplication(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if hasChanges {
		_, err := conn.ModifyReplicationConfigWithContext(ctx, request)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DMS Serverless Replication Config (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceServerlessReplicationRead(ctx, d, meta)...)
}

func resourceServerlessReplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	stopReplication(ctx, d.Id(), conn)

	request := &dms.DeleteReplicationConfigInput{
		ReplicationConfigArn: aws.String(d.Get("replication_config_arn").(string)),
	}

	log.Printf("[DEBUG] DMS delete serverless replication config: %#v", request)

	_, err := conn.DeleteReplicationConfigWithContext(ctx, request)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
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

func findReplicationById(ctx context.Context, id string, conn *dms.DatabaseMigrationService) (*dms.Replication, error) {
	response, err := conn.DescribeReplicationsWithContext(ctx, &dms.DescribeReplicationsInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-config-id"),
				Values: []*string{aws.String(id)},
			},
		},
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		log.Printf("[WARN] DMS Serverless Replication (%s) not found", id)
		return nil, err
	}

	if response == nil || len(response.Replications) == 0 || response.Replications[0] == nil {
		log.Printf("[WARN] DMS Serverless Replication (%s) not found", id)
		return nil, err
	}

	return response.Replications[0], nil
}

func startReplication(ctx context.Context, id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Starting DMS Serverless Replication: (%s)", id)

	replication, _ := findReplicationById(ctx, id, conn)

	startReplicationType := replicationTypeValueStartReplication
	if *replication.Status != replicationStatusReady {
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

	replication, _ := findReplicationById(ctx, id, conn)

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
