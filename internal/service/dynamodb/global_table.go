package dynamodb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGlobalTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn()

	globalTableName := d.Get(names.AttrName).(string)

	input := &dynamodb.CreateGlobalTableInput{
		GlobalTableName:  aws.String(globalTableName),
		ReplicationGroup: expandReplicas(d.Get("replica").(*schema.Set).List()),
	}

	_, err := conn.CreateGlobalTableWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Global Table (%s): %s", globalTableName, err)
	}

	d.SetId(globalTableName)

	log.Println("[INFO] Waiting for DynamoDB Global Table to be created")
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.GlobalTableStatusCreating,
			dynamodb.GlobalTableStatusDeleting,
			dynamodb.GlobalTableStatusUpdating,
		},
		Target: []string{
			dynamodb.GlobalTableStatusActive,
		},
		Refresh:    resourceGlobalTableStateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DynamoDB Global Table (%s): waiting for completion: %s", globalTableName, err)
	}

	return append(diags, resourceGlobalTableRead(ctx, d, meta)...)
}

func resourceGlobalTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	globalTableDescription, err := resourceGlobalTableRetrieve(ctx, d, meta)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Global Table (%s): %s", d.Id(), err)
	}
	if globalTableDescription == nil {
		log.Printf("[WARN] DynamoDB Global Table %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := flattenGlobalTable(d, globalTableDescription); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DynamoDB Global Table (%s): %s", d.Id(), err)
	}
	return diags
}

func resourceGlobalTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn()

	if d.HasChange("replica") {
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
		log.Printf("[DEBUG] Updating DynamoDB Global Table: %#v", input)
		if _, err := conn.UpdateGlobalTableWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Global Table (%s): %s", d.Id(), err)
		}

		log.Println("[INFO] Waiting for DynamoDB Global Table to be updated")
		stateConf := &resource.StateChangeConf{
			Pending: []string{
				dynamodb.GlobalTableStatusCreating,
				dynamodb.GlobalTableStatusDeleting,
				dynamodb.GlobalTableStatusUpdating,
			},
			Target: []string{
				dynamodb.GlobalTableStatusActive,
			},
			Refresh:    resourceGlobalTableStateRefreshFunc(ctx, d, meta),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
		}
		_, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DynamoDB Global Table (%s): waiting for completion: %s", d.Id(), err)
		}
	}

	return diags
}

// Deleting a DynamoDB Global Table is represented by removing all replicas.
func resourceGlobalTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DynamoDBConn()

	input := &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  expandReplicaUpdateDeleteReplicas(d.Get("replica").(*schema.Set).List()),
	}
	log.Printf("[DEBUG] Deleting DynamoDB Global Table: %#v", input)
	if _, err := conn.UpdateGlobalTableWithContext(ctx, input); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Global Table (%s): %s", d.Id(), err)
	}

	log.Println("[INFO] Waiting for DynamoDB Global Table to be destroyed")
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.GlobalTableStatusActive,
			dynamodb.GlobalTableStatusCreating,
			dynamodb.GlobalTableStatusDeleting,
			dynamodb.GlobalTableStatusUpdating,
		},
		Target:     []string{},
		Refresh:    resourceGlobalTableStateRefreshFunc(ctx, d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return sdkdiag.AppendErrorf(diags, "deleting DynamoDB Global Table (%s): waiting for completion: %s", d.Id(), err)
}

func resourceGlobalTableRetrieve(ctx context.Context, d *schema.ResourceData, meta interface{}) (*dynamodb.GlobalTableDescription, error) {
	conn := meta.(*conns.AWSClient).DynamoDBConn()

	input := &dynamodb.DescribeGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Retrieving DynamoDB Global Table: %#v", input)

	output, err := conn.DescribeGlobalTableWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeGlobalTableNotFoundException) {
			return nil, nil
		}
		return nil, err
	}

	return output.GlobalTableDescription, nil
}

func resourceGlobalTableStateRefreshFunc(ctx context.Context,
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gtd, err := resourceGlobalTableRetrieve(ctx, d, meta)

		if err != nil {
			return nil, "", err
		}

		if gtd == nil {
			return nil, "", nil
		}

		return gtd, *gtd.GlobalTableStatus, nil
	}
}

func flattenGlobalTable(d *schema.ResourceData, globalTableDescription *dynamodb.GlobalTableDescription) error {
	d.Set(names.AttrARN, globalTableDescription.GlobalTableArn)
	d.Set(names.AttrName, globalTableDescription.GlobalTableName)

	return d.Set("replica", flattenReplicas(globalTableDescription.ReplicationGroup))
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
