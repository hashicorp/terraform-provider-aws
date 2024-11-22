// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dynamodb_global_table", name="Global Table")
func resourceGlobalTable() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &dynamodb.CreateGlobalTableInput{
		GlobalTableName:  aws.String(name),
		ReplicationGroup: expandReplicas(d.Get("replica").(*schema.Set).List()),
	}

	_, err := conn.CreateGlobalTable(ctx, input)

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
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	globalTableDescription, err := findGlobalTableByName(ctx, conn, d.Id())

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
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	o, n := d.GetChange("replica")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os, ns := o.(*schema.Set), n.(*schema.Set)
	replicaUpdateCreateReplicas := expandReplicaUpdateCreateReplicas(ns.Difference(os).List())
	replicaUpdateDeleteReplicas := expandReplicaUpdateDeleteReplicas(os.Difference(ns).List())

	replicaUpdates := make([]awstypes.ReplicaUpdate, 0, (len(replicaUpdateCreateReplicas) + len(replicaUpdateDeleteReplicas)))
	replicaUpdates = append(replicaUpdates, replicaUpdateCreateReplicas...)
	replicaUpdates = append(replicaUpdates, replicaUpdateDeleteReplicas...)

	input := &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  replicaUpdates,
	}

	_, err := conn.UpdateGlobalTable(ctx, input)

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
	conn := meta.(*conns.AWSClient).DynamoDBClient(ctx)

	log.Printf("[DEBUG] Deleting DynamoDB Global Table: %s", d.Id())
	_, err := conn.UpdateGlobalTable(ctx, &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  expandReplicaUpdateDeleteReplicas(d.Get("replica").(*schema.Set).List()),
	})

	if errs.IsA[*awstypes.GlobalTableNotFoundException](err) || errs.IsA[*awstypes.ReplicaNotFoundException](err) {
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

func findGlobalTableByName(ctx context.Context, conn *dynamodb.Client, name string) (*awstypes.GlobalTableDescription, error) {
	input := &dynamodb.DescribeGlobalTableInput{
		GlobalTableName: aws.String(name),
	}

	output, err := conn.DescribeGlobalTable(ctx, input)

	if errs.IsA[*awstypes.GlobalTableNotFoundException](err) {
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

func statusGlobalTable(ctx context.Context, conn *dynamodb.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGlobalTableByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.GlobalTableStatus), nil
	}
}

func waitGlobalTableCreated(ctx context.Context, conn *dynamodb.Client, name string, timeout time.Duration) (*awstypes.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.GlobalTableStatusCreating),
		Target:     enum.Slice(awstypes.GlobalTableStatusActive),
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalTableUpdated(ctx context.Context, conn *dynamodb.Client, name string, timeout time.Duration) (*awstypes.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.GlobalTableStatusUpdating),
		Target:     enum.Slice(awstypes.GlobalTableStatusActive),
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalTableDeleted(ctx context.Context, conn *dynamodb.Client, name string, timeout time.Duration) (*awstypes.GlobalTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.GlobalTableStatusActive, awstypes.GlobalTableStatusDeleting),
		Target:     []string{},
		Refresh:    statusGlobalTable(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalTableDescription); ok {
		return output, err
	}

	return nil, err
}

func expandReplicaUpdateCreateReplicas(tfList []interface{}) []awstypes.ReplicaUpdate {
	apiObjects := make([]awstypes.ReplicaUpdate, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObjects = append(apiObjects, *expandReplicaUpdateCreateReplica(tfMap))
	}

	return apiObjects
}

func expandReplicaUpdateCreateReplica(tfMap map[string]interface{}) *awstypes.ReplicaUpdate {
	apiObject := &awstypes.ReplicaUpdate{
		Create: &awstypes.CreateReplicaAction{
			RegionName: aws.String(tfMap["region_name"].(string)),
		},
	}

	return apiObject
}

func expandReplicaUpdateDeleteReplicas(tfList []interface{}) []awstypes.ReplicaUpdate {
	apiObjects := make([]awstypes.ReplicaUpdate, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObjects = append(apiObjects, *expandReplicaUpdateDeleteReplica(tfMap))
	}

	return apiObjects
}

func expandReplicaUpdateDeleteReplica(tfMap map[string]interface{}) *awstypes.ReplicaUpdate {
	apiObject := &awstypes.ReplicaUpdate{
		Delete: &awstypes.DeleteReplicaAction{
			RegionName: aws.String(tfMap["region_name"].(string)),
		},
	}

	return apiObject
}

func expandReplicas(tfList []interface{}) []awstypes.Replica {
	apiObjects := make([]awstypes.Replica, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObjects = append(apiObjects, *expandReplica(tfMap))
	}

	return apiObjects
}

func expandReplica(tfMap map[string]interface{}) *awstypes.Replica {
	apiObject := &awstypes.Replica{
		RegionName: aws.String(tfMap["region_name"].(string)),
	}

	return apiObject
}

func flattenReplicas(apiObjects []awstypes.ReplicaDescription) []interface{} {
	tfList := []interface{}{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenReplica(&apiObject))
	}

	return tfList
}

func flattenReplica(apiObject *awstypes.ReplicaDescription) map[string]interface{} {
	tfMap := make(map[string]interface{})
	tfMap["region_name"] = aws.ToString(apiObject.RegionName)

	return tfMap
}
