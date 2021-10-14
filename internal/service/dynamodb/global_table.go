package dynamodb

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceGlobalTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlobalTableCreate,
		Read:   resourceGlobalTableRead,
		Update: resourceGlobalTableUpdate,
		Delete: resourceGlobalTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceGlobalTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	globalTableName := d.Get("name").(string)

	input := &dynamodb.CreateGlobalTableInput{
		GlobalTableName:  aws.String(globalTableName),
		ReplicationGroup: expandReplicas(d.Get("replica").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Creating DynamoDB Global Table: %#v", input)
	_, err := conn.CreateGlobalTable(input)
	if err != nil {
		return err
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
		Refresh:    resourceGlobalTableStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceGlobalTableRead(d, meta)
}

func resourceGlobalTableRead(d *schema.ResourceData, meta interface{}) error {
	globalTableDescription, err := resourceGlobalTableRetrieve(d, meta)

	if err != nil {
		return err
	}
	if globalTableDescription == nil {
		log.Printf("[WARN] DynamoDB Global Table %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return flattenGlobalTable(d, globalTableDescription)
}

func resourceGlobalTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

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
		if _, err := conn.UpdateGlobalTable(input); err != nil {
			return err
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
			Refresh:    resourceGlobalTableStateRefreshFunc(d, meta),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			return err
		}
	}

	return nil
}

// Deleting a DynamoDB Global Table is represented by removing all replicas.
func resourceGlobalTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	input := &dynamodb.UpdateGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
		ReplicaUpdates:  expandReplicaUpdateDeleteReplicas(d.Get("replica").(*schema.Set).List()),
	}
	log.Printf("[DEBUG] Deleting DynamoDB Global Table: %#v", input)
	if _, err := conn.UpdateGlobalTable(input); err != nil {
		return err
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
		Refresh:    resourceGlobalTableStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 10 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceGlobalTableRetrieve(d *schema.ResourceData, meta interface{}) (*dynamodb.GlobalTableDescription, error) {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	input := &dynamodb.DescribeGlobalTableInput{
		GlobalTableName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Retrieving DynamoDB Global Table: %#v", input)

	output, err := conn.DescribeGlobalTable(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeGlobalTableNotFoundException, "") {
			return nil, nil
		}
		return nil, fmt.Errorf("Error retrieving DynamoDB Global Table: %s", err)
	}

	return output.GlobalTableDescription, nil
}

func resourceGlobalTableStateRefreshFunc(
	d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gtd, err := resourceGlobalTableRetrieve(d, meta)

		if err != nil {
			log.Printf("Error on retrieving DynamoDB Global Table when waiting: %s", err)
			return nil, "", err
		}

		if gtd == nil {
			return nil, "", nil
		}

		if gtd.GlobalTableStatus != nil {
			log.Printf("[DEBUG] Status for DynamoDB Global Table %s: %s", d.Id(), *gtd.GlobalTableStatus)
		}

		return gtd, *gtd.GlobalTableStatus, nil
	}
}

func flattenGlobalTable(d *schema.ResourceData, globalTableDescription *dynamodb.GlobalTableDescription) error {
	var err error

	d.Set("arn", globalTableDescription.GlobalTableArn)
	d.Set("name", globalTableDescription.GlobalTableName)

	err = d.Set("replica", flattenReplicas(globalTableDescription.ReplicationGroup))
	return err
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
	replica["region_name"] = *replicaDescription.RegionName
	return replica
}
