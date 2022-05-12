package docdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceGlobalCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalClusterCreate,
		ReadWithoutTimeout:   resourceGlobalClusterRead,
		UpdateWithoutTimeout: resourceGlobalClusterUpdate,
		DeleteWithoutTimeout: resourceGlobalClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		//Timeouts will scale per number of resources in the cluster. Timeouts implemented on each resource action.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(GlobalClusterCreateTimeout),
			Update: schema.DefaultTimeout(GlobalClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(GlobalClusterDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"engine": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				AtLeastOneOf:  []string{"engine", "source_db_cluster_identifier"},
				ConflictsWith: []string{"source_db_cluster_identifier"},
				ValidateFunc:  validEngine(),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"global_cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validGlobalCusterIdentifier,
			},
			"global_cluster_members": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_cluster_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_writer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"global_cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_db_cluster_identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				AtLeastOneOf:  []string{"engine", "source_db_cluster_identifier"},
				ConflictsWith: []string{"engine"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	input := &docdb.CreateGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Get("global_cluster_identifier").(string)),
	}

	if v, ok := d.GetOk("database_name"); ok {
		input.DatabaseName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("deletion_protection"); ok {
		input.DeletionProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_db_cluster_identifier"); ok {
		input.SourceDBClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_encrypted"); ok {
		input.StorageEncrypted = aws.Bool(v.(bool))
	}

	output, err := conn.CreateGlobalClusterWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating DocDB Global Cluster: %w", err))
	}

	d.SetId(aws.StringValue(output.GlobalCluster.GlobalClusterIdentifier))

	if err := waitForGlobalClusterCreation(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) availability: %w", d.Id(), err))
	}

	return resourceGlobalClusterRead(ctx, d, meta)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	globalCluster, err := FindGlobalClusterById(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) {
		log.Printf("[WARN] DocDB Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading DocDB Global Cluster: %w", err))
	}

	if globalCluster == nil {
		log.Printf("[WARN] DocDB Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(globalCluster.Status) == GlobalClusterStatusDeleting || aws.StringValue(globalCluster.Status) == GlobalClusterStatusDeleted {
		log.Printf("[WARN] DocDB Global Cluster (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(globalCluster.Status))
		d.SetId("")
		return nil
	}

	_ = d.Set("arn", globalCluster.GlobalClusterArn)
	_ = d.Set("database_name", globalCluster.DatabaseName)
	_ = d.Set("deletion_protection", globalCluster.DeletionProtection)
	_ = d.Set("engine", globalCluster.Engine)
	_ = d.Set("engine_version", globalCluster.EngineVersion)
	_ = d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)

	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting global_cluster_members: %w", err))
	}

	_ = d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	_ = d.Set("storage_encrypted", globalCluster.StorageEncrypted)

	return nil
}

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	input := &docdb.ModifyGlobalClusterInput{
		DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Updating DocDB Global Cluster (%s): %s", d.Id(), input)

	if d.HasChange("engine_version") {
		if err := resourceGlobalClusterUpgradeEngineVersion(ctx, d, conn); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := conn.ModifyGlobalClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating DocDB Global Cluster: %w", err))
	}

	if err := waitForGlobalClusterUpdate(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) update: %w", d.Id(), err))
	}

	return resourceGlobalClusterRead(ctx, d, meta)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DocDBConn

	for _, globalClusterMemberRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
		globalClusterMember, ok := globalClusterMemberRaw.(map[string]interface{})
		if !ok {
			continue
		}

		dbClusterArn, ok := globalClusterMember["db_cluster_arn"].(string)
		if !ok {
			continue
		}

		input := &docdb.RemoveFromGlobalClusterInput{
			DbClusterIdentifier:     aws.String(dbClusterArn),
			GlobalClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)
		if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
			continue
		}
		if err != nil {
			return diag.FromErr(fmt.Errorf("error removing DocDB Cluster (%s) from Global Cluster (%s): %w", dbClusterArn, d.Id(), err))
		}

		if err := waitForGlobalClusterRemoval(ctx, conn, dbClusterArn, d.Timeout(schema.TimeoutDelete)); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for DocDB Cluster (%s) removal from DocDB Global Cluster (%s): %w", dbClusterArn, d.Id(), err))
		}
	}

	input := &docdb.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DocDB Global Cluster (%s): %s", d.Id(), input)

	// Allow for eventual consistency
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := conn.DeleteGlobalClusterWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, docdb.ErrCodeInvalidGlobalClusterStateFault, "is not empty") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteGlobalClusterWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting DocDB Global Cluster: %w", err))
	}

	if err := WaitForGlobalClusterDeletion(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func flattenGlobalClusterMembers(apiObjects []*docdb.GlobalClusterMember) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"db_cluster_arn": aws.StringValue(apiObject.DBClusterArn),
			"is_writer":      aws.BoolValue(apiObject.IsWriter),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func statusGlobalClusterRefreshFunc(ctx context.Context, conn *docdb.DocDB, globalClusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalCluster, err := FindGlobalClusterById(ctx, conn, globalClusterID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) || globalCluster == nil {
			return nil, GlobalClusterStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading DocDB Global Cluster (%s): %w", globalClusterID, err)
		}

		return globalCluster, aws.StringValue(globalCluster.Status), nil
	}
}

func waitForGlobalClusterCreation(ctx context.Context, conn *docdb.DocDB, globalClusterID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{GlobalClusterStatusCreating},
		Target:  []string{GlobalClusterStatusAvailable},
		Refresh: statusGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout: timeout,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForGlobalClusterUpdate(ctx context.Context, conn *docdb.DocDB, globalClusterID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{GlobalClusterStatusModifying, GlobalClusterStatusUpgrading},
		Target:  []string{GlobalClusterStatusAvailable},
		Refresh: statusGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForGlobalClusterRemoval(ctx context.Context, conn *docdb.DocDB, dbClusterIdentifier string, timeout time.Duration) error {
	var globalCluster *docdb.GlobalCluster
	stillExistsErr := fmt.Errorf("DocDB Cluster still exists in DocDB Global Cluster")

	err := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
		var err error

		globalCluster, err = findGlobalClusterByArn(ctx, conn, dbClusterIdentifier)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if globalCluster != nil {
			return resource.RetryableError(stillExistsErr)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = findGlobalClusterByArn(ctx, conn, dbClusterIdentifier)
	}

	if err != nil {
		return err
	}

	if globalCluster != nil {
		return stillExistsErr
	}

	return nil
}

// Updating major versions is not supported by documentDB
// To support minor version upgrades, we will upgrade all cluster members
func resourceGlobalClusterUpgradeEngineVersion(ctx context.Context, d *schema.ResourceData, conn *docdb.DocDB) error {
	log.Printf("[DEBUG] Upgrading DocDB Global Cluster (%s) engine version: %s", d.Id(), d.Get("engine_version"))
	err := resourceGlobalClusterUpgradeMinorEngineVersion(ctx, d.Get("global_cluster_members").(*schema.Set), d.Get("engine_version").(string), conn, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}
	globalCluster, err := FindGlobalClusterById(ctx, conn, d.Id())
	if err != nil {
		return err
	}
	for _, clusterMember := range globalCluster.GlobalClusterMembers {
		err := waitForClusterUpdate(conn, findGlobalClusterIdByArn(ctx, conn, aws.StringValue(clusterMember.DBClusterArn)), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceGlobalClusterUpgradeMinorEngineVersion(ctx context.Context, clusterMembers *schema.Set, engineVersion string, conn *docdb.DocDB, timeout time.Duration) error {
	for _, clusterMemberRaw := range clusterMembers.List() {
		clusterMember := clusterMemberRaw.(map[string]interface{})
		if clusterMemberArn, ok := clusterMember["db_cluster_arn"]; ok && clusterMemberArn.(string) != "" {
			modInput := &docdb.ModifyDBClusterInput{
				ApplyImmediately:    aws.Bool(true),
				DBClusterIdentifier: aws.String(clusterMemberArn.(string)),
				EngineVersion:       aws.String(engineVersion),
			}
			err := resource.RetryContext(ctx, timeout, func() *resource.RetryError {
				_, err := conn.ModifyDBClusterWithContext(ctx, modInput)
				if err != nil {
					if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if tfresource.TimedOut(err) {
				_, err := conn.ModifyDBClusterWithContext(ctx, modInput)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return fmt.Errorf("failed to update engine_version on global cluster member (%s): %w", clusterMemberArn, err)
			}
		}
	}
	return nil
}
