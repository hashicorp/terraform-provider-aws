// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_global_table")
func ResourceGlobalTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalTableCreate,
		ReadWithoutTimeout:   resourceGlobalTableRead,
		UpdateWithoutTimeout: resourceGlobalTableUpdate,
		DeleteWithoutTimeout: resourceGlobalTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validGlobalTableName,
			},
			"replica": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGlobalTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &dynamodb.CreateGlobalTableInput{
		GlobalTableName:  aws.String(name),
		ReplicationGroup: expandReplicas(d.Get("replica").(*schema.Set).List()),
	}

	_, err := conn.CreateGlobalTableWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Global Table (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitGlobalTableCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Global Table (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalTableRead(ctx, d, meta)...)
}

func resourceGlobalTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	globalTableDescription, err := FindGlobalTableByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DynamoDB Global Table %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Global Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, globalTableDescription.GlobalTableArn)
	d.Set(names.AttrName, globalTableDescription.GlobalTableName)
	if err := d.Set("replica", flattenReplicas(globalTableDescription.ReplicationGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting replica: %s", err)
	}

	return diags
}

func resourceGlobalTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	o, n := d.GetChange("replica")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	replicaUpdateCreateReplicas := expandReplicaUpdateCreateReplicas(ns.Difference(os).List())
	replicaUpdateDeleteReplicas := expandReplicaUpdateDeleteReplicas(os.Difference(ns).List())

	replicaUpdates := make([]*dynamodb.ReplicaUpdate, 0, (len(replicaUpdateCreateReplicas) + len(replicaUpdateDeleteReplicas)))
	replicaUpdates = append(replicaUpdates, replicaUpdateCreateReplicas...)
	replicaUpdates = append(replicaUpdates, replicaUpdateDeleteReplicas...)

	input := &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  replicaUpdates,
	}

	_, err := conn.UpdateGlobalTableWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DynamoDB Global Table (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalTableUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Global Table (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalTableRead(ctx, d, meta)...)
}

// Deleting a DynamoDB Global Table is represented by removing all replicas.
func resourceGlobalTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn(ctx)

	log.Printf("[DEBUG] Deleting DynamoDB Global Table: %s", d.Id())
	_, err := conn.UpdateGlobalTableWithContext(ctx, &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  expandReplicaUpdateDeleteReplicas(d.Get("replica").(*schema.Set).List()),
	})

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeGlobalTableNotFoundException, dynamodb.ErrCodeReplicaNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Global Table (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalTableDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DynamoDB Global Table (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindGlobalTableByName(ctx context.Context, conn *dynamodb.DynamoDB, name string) (*dynamodb.GlobalTableDescription, error) {
	input := &dynamodb.DescribeGlobalTableInput{
		GlobalTableName: aws.String(name),
	}

	output, err := conn.DescribeGlobalTableWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeGlobalTableNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GlobalTableDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GlobalTableDescription, nil
}

func statusGlobalTable(ctx context.Context, conn *dynamodb.DynamoDB, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGlobalTableByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.GlobalTableStatus), nil
	}
}

func waitGlobalTableCreated(ctx context.Context, conn *dynamodb.DynamoDB, name string, timeout time.Duration) (*dynamodb.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{dynamodb.GlobalTableStatusCreating},
		Target:     []string{dynamodb.GlobalTableStatusActive},
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalTableDeleted(ctx context.Context, conn *dynamodb.DynamoDB, name string, timeout time.Duration) (*dynamodb.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{dynamodb.GlobalTableStatusActive, dynamodb.GlobalTableStatusDeleting},
		Target:     []string{},
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalTableUpdated(ctx context.Context, conn *dynamodb.DynamoDB, name string, timeout time.Duration) (*dynamodb.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{dynamodb.GlobalTableStatusUpdating},
		Target:     []string{dynamodb.GlobalTableStatusActive},
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func expandReplicaUpdateCreateReplicas(configuredReplicas []interface{}) []*dynamodb.ReplicaUpdate {
	replicaUpdates := make([]*dynamodb.ReplicaUpdate, 0, len(configuredReplicas))
	for _, replicaRaw := range configuredReplicas {
		replica := replicaRaw.(map[string]interface{})
		replicaUpdates = append(replicaUpdates, expandReplicaUpdateCreateReplica(replica))
	}
	return replicaUpdates
}

func expandReplicaUpdateCreateReplica(configuredReplica map[string]interface{}) *dynamodb.ReplicaUpdate {
	replicaUpdate := &dynamodb.ReplicaUpdate{
		Create: &dynamodb.CreateReplicaAction{
			RegionName: aws.String(configuredReplica["region_name"].(string)),
		},
	}
	return replicaUpdate
}

func expandReplicaUpdateDeleteReplicas(configuredReplicas []interface{}) []*dynamodb.ReplicaUpdate {
	replicaUpdates := make([]*dynamodb.ReplicaUpdate, 0, len(configuredReplicas))
	for _, replicaRaw := range configuredReplicas {
		replica := replicaRaw.(map[string]interface{})
		replicaUpdates = append(replicaUpdates, expandReplicaUpdateDeleteReplica(replica))
	}
	return replicaUpdates
}

func expandReplicaUpdateDeleteReplica(configuredReplica map[string]interface{}) *dynamodb.ReplicaUpdate {
	replicaUpdate := &dynamodb.ReplicaUpdate{
		Delete: &dynamodb.DeleteReplicaAction{
			RegionName: aws.String(configuredReplica["region_name"].(string)),
		},
	}
	return replicaUpdate
}

func expandReplicas(configuredReplicas []interface{}) []*dynamodb.Replica {
	replicas := make([]*dynamodb.Replica, 0, len(configuredReplicas))
	for _, replicaRaw := range configuredReplicas {
		replica := replicaRaw.(map[string]interface{})
		replicas = append(replicas, expandReplica(replica))
	}
	return replicas
}

func expandReplica(configuredReplica map[string]interface{}) *dynamodb.Replica {
	replica := &dynamodb.Replica{
		RegionName: aws.String(configuredReplica["region_name"].(string)),
	}
	return replica
}

func flattenReplicas(replicaDescriptions []*dynamodb.ReplicaDescription) []interface{} {
	replicas := []interface{}{}
	for _, replicaDescription := range replicaDescriptions {
		replicas = append(replicas, flattenReplica(replicaDescription))
	}
	return replicas
}

func flattenReplica(replicaDescription *dynamodb.ReplicaDescription) map[string]interface{} {
	replica := make(map[string]interface{})
	replica["region_name"] = aws.StringValue(replicaDescription.RegionName)
	return replica
}
