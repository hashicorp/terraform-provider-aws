package aws

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	docDBGlobalClusterRemovalTimeout = 2 * time.Minute
)

func resourceAwsDocDBGlobalCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsDocDBGlobalClusterCreate,
		ReadWithoutTimeout:   resourceAwsDocDBGlobalClusterRead,
		UpdateWithoutTimeout: resourceAwsDocDBGlobalClusterUpdate,
		DeleteWithoutTimeout: resourceAwsDocDBGlobalClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ConflictsWith: []string{"source_db_cluster_identifier"},
				ValidateFunc: validation.StringInSlice([]string{
					"docdb",
				}, false),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				ConflictsWith: []string{"engine"},
				RequiredWith:  []string{"force_destroy"},
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

func resourceAwsDocDBGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).docdbconn

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

	// Prevent the following error and keep the previous default,
	// since we cannot have Engine default after adding SourceDBClusterIdentifier:
	// InvalidParameterValue: When creating standalone global cluster, value for engineName should be specified
	if input.Engine == nil && input.SourceDBClusterIdentifier == nil {
		input.Engine = aws.String("docdb")
	}

	output, err := conn.CreateGlobalClusterWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating DocDB Global Cluster: %w", err))
	}

	d.SetId(aws.StringValue(output.GlobalCluster.GlobalClusterIdentifier))

	if err := waitForDocDBGlobalClusterCreation(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) availability: %w", d.Id(), err))
	}

	return resourceAwsDocDBGlobalClusterRead(ctx, d, meta)
}

func resourceAwsDocDBGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).docdbconn

	globalCluster, err := docDBDescribeGlobalCluster(ctx, conn, d.Id())

	if isAWSErr(err, docdb.ErrCodeGlobalClusterNotFoundFault, "") {
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

	if aws.StringValue(globalCluster.Status) == "deleting" || aws.StringValue(globalCluster.Status) == "deleted" {
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

	if err := d.Set("global_cluster_members", flattenDocDBGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting global_cluster_members: %w", err))
	}

	_ = d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	_ = d.Set("storage_encrypted", globalCluster.StorageEncrypted)

	return nil
}

func resourceAwsDocDBGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).docdbconn

	input := &docdb.ModifyGlobalClusterInput{
		DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Updating DocDB Global Cluster (%s): %s", d.Id(), input)

	if d.HasChange("engine_version") {
		if err := resourceAwsDocDBGlobalClusterUpgradeEngineVersion(ctx, d, conn); err != nil {
			return diag.FromErr(err)
		}
	}

	_, err := conn.ModifyGlobalClusterWithContext(ctx, input)

	if isAWSErr(err, docdb.ErrCodeGlobalClusterNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating DocDB Global Cluster: %w", err))
	}

	if err := waitForDocDBGlobalClusterUpdate(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) update: %w", d.Id(), err))
	}

	return resourceAwsDocDBGlobalClusterRead(ctx, d, meta)
}

func resourceAwsDocDBGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).docdbconn

	if d.Get("force_destroy").(bool) {
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

			if err := waitForDocDBGlobalClusterRemoval(ctx, conn, dbClusterArn); err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for DocDB Cluster (%s) removal from DocDB Global Cluster (%s): %w", dbClusterArn, d.Id(), err))
			}
		}
	}

	input := &docdb.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DocDB Global Cluster (%s): %s", d.Id(), input)

	// Allow for eventual consistency
	err := resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteGlobalClusterWithContext(ctx, input)

		if isAWSErr(err, docdb.ErrCodeInvalidGlobalClusterStateFault, "is not empty") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteGlobalClusterWithContext(ctx, input)
	}

	if isAWSErr(err, docdb.ErrCodeGlobalClusterNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting DocDB Global Cluster: %w", err))
	}

	if err := waitForDocDBGlobalClusterDeletion(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DocDB Global Cluster (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func flattenDocDBGlobalClusterMembers(apiObjects []*docdb.GlobalClusterMember) []interface{} {
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

func docDBDescribeGlobalCluster(ctx context.Context, conn *docdb.DocDB, globalClusterID string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	log.Printf("[DEBUG] Reading DocDB Global Cluster (%s): %s", globalClusterID, input)
	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			if aws.StringValue(gc.GlobalClusterIdentifier) == globalClusterID {
				globalCluster = gc
				return false
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func docDBDescribeGlobalClusterFromDbClusterARN(ctx context.Context, conn *docdb.DocDB, dbClusterARN string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		Filters: []*docdb.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: []*string{aws.String(dbClusterARN)},
			},
		},
	}

	log.Printf("[DEBUG] Reading DocDB Global Clusters: %s", input)
	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			for _, globalClusterMember := range gc.GlobalClusterMembers {
				if aws.StringValue(globalClusterMember.DBClusterArn) == dbClusterARN {
					globalCluster = gc
					return false
				}
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func docDBGlobalClusterRefreshFunc(ctx context.Context, conn *docdb.DocDB, globalClusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalCluster, err := docDBDescribeGlobalCluster(ctx, conn, globalClusterID)

		if isAWSErr(err, docdb.ErrCodeGlobalClusterNotFoundFault, "") {
			return nil, "deleted", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading DocDB Global Cluster (%s): %w", globalClusterID, err)
		}

		if globalCluster == nil {
			return nil, "deleted", nil
		}

		return globalCluster, aws.StringValue(globalCluster.Status), nil
	}
}

func waitForDocDBGlobalClusterCreation(ctx context.Context, conn *docdb.DocDB, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"available"},
		Refresh: docDBGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForDocDBGlobalClusterUpdate(ctx context.Context, conn *docdb.DocDB, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"modifying", "upgrading"},
		Target:  []string{"available"},
		Refresh: docDBGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout: 10 * time.Minute,
		Delay:   30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForDocDBGlobalClusterDeletion(ctx context.Context, conn *docdb.DocDB, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
		},
		Target:         []string{"deleted"},
		Refresh:        docDBGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) deletion", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}

func waitForDocDBGlobalClusterRemoval(ctx context.Context, conn *docdb.DocDB, dbClusterIdentifier string) error {
	var globalCluster *docdb.GlobalCluster
	stillExistsErr := fmt.Errorf("DocDB Cluster still exists in DocDB Global Cluster")

	err := resource.RetryContext(ctx, docDBGlobalClusterRemovalTimeout, func() *resource.RetryError {
		var err error

		globalCluster, err = docDBDescribeGlobalClusterFromDbClusterARN(ctx, conn, dbClusterIdentifier)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if globalCluster != nil {
			return resource.RetryableError(stillExistsErr)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = docDBDescribeGlobalClusterFromDbClusterARN(ctx, conn, dbClusterIdentifier)
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
func resourceAwsDocDBGlobalClusterUpgradeEngineVersion(ctx context.Context, d *schema.ResourceData, conn *docdb.DocDB) error {
	log.Printf("[DEBUG] Upgrading DocDB Global Cluster (%s) engine version: %s", d.Id(), d.Get("engine_version"))
	err := resourceAwsDocDBGlobalClusterUpgradeMinorEngineVersion(ctx, d.Get("global_cluster_members").(*schema.Set), d.Get("engine_version").(string), conn)
	if err != nil {
		return err
	}
	globalCluster, err := docDBDescribeGlobalCluster(ctx, conn, d.Id())
	if err != nil {
		return err
	}
	for _, clusterMember := range globalCluster.GlobalClusterMembers {
		err := waitForDocDBClusterUpdate(conn, resourceAwsDocDBGlobalClusterGetIdByArn(ctx, conn, aws.StringValue(clusterMember.DBClusterArn)), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceAwsDocDBGlobalClusterGetIdByArn(ctx context.Context, conn *docdb.DocDB, arn string) string {
	result, err := conn.DescribeDBClustersWithContext(ctx, &docdb.DescribeDBClustersInput{})
	if err != nil {
		return ""
	}
	for _, cluster := range result.DBClusters {
		if aws.StringValue(cluster.DBClusterArn) == arn {
			return aws.StringValue(cluster.DBClusterIdentifier)
		}
	}
	return ""
}

func resourceAwsDocDBGlobalClusterUpgradeMinorEngineVersion(ctx context.Context, clusterMembers *schema.Set, engineVersion string, conn *docdb.DocDB) error {
	for _, clusterMemberRaw := range clusterMembers.List() {
		clusterMember := clusterMemberRaw.(map[string]interface{})
		if clusterMemberArn, ok := clusterMember["db_cluster_arn"]; ok && clusterMemberArn.(string) != "" {
			modInput := &docdb.ModifyDBClusterInput{
				ApplyImmediately:    aws.Bool(true),
				DBClusterIdentifier: aws.String(clusterMemberArn.(string)),
				EngineVersion:       aws.String(engineVersion),
			}
			err := resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
				_, err := conn.ModifyDBClusterWithContext(ctx, modInput)
				if err != nil {
					if isAWSErr(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
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
