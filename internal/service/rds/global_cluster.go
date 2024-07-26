// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_global_cluster")
func ResourceGlobalCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGlobalClusterCreate,
		ReadWithoutTimeout:   resourceGlobalClusterRead,
		UpdateWithoutTimeout: resourceGlobalClusterUpdate,
		DeleteWithoutTimeout: resourceGlobalClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrEngine: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_db_cluster_identifier"},
				ValidateFunc:  validation.StringInSlice(globalClusterEngine_Values(), false),
			},
			"engine_lifecycle_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(engineLifecycleSupport_Values(), false),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"engine_version_actual": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"global_cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
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
				ConflictsWith: []string{names.AttrEngine},
				RequiredWith:  []string{names.AttrForceDestroy},
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	globalClusterID := d.Get("global_cluster_identifier").(string)
	input := &rds.CreateGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	if v, ok := d.GetOk(names.AttrDatabaseName); ok {
		input.DatabaseName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDeletionProtection); ok {
		input.DeletionProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_lifecycle_support"); ok {
		input.EngineLifecycleSupport = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_db_cluster_identifier"); ok {
		input.SourceDBClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrStorageEncrypted); ok {
		input.StorageEncrypted = aws.Bool(v.(bool))
	}

	// Prevent the following error and keep the previous default,
	// since we cannot have Engine default after adding SourceDBClusterIdentifier:
	// InvalidParameterValue: When creating standalone global cluster, value for engineName should be specified
	if input.Engine == nil && input.SourceDBClusterIdentifier == nil {
		input.Engine = aws.String(globalClusterEngineAurora)
	}

	output, err := conn.CreateGlobalClusterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.StringValue(output.GlobalCluster.GlobalClusterIdentifier))

	if err := waitGlobalClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	globalCluster, err := FindGlobalClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Global Cluster (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, globalCluster.GlobalClusterArn)
	d.Set(names.AttrDatabaseName, globalCluster.DatabaseName)
	d.Set(names.AttrDeletionProtection, globalCluster.DeletionProtection)
	d.Set(names.AttrEngine, globalCluster.Engine)
	d.Set("engine_lifecycle_support", globalCluster.EngineLifecycleSupport)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}
	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set(names.AttrStorageEncrypted, globalCluster.StorageEncrypted)

	oldEngineVersion := d.Get(names.AttrEngineVersion).(string)
	newEngineVersion := aws.StringValue(globalCluster.EngineVersion)

	// For example a configured engine_version of "5.6.10a" and a returned engine_version of "5.6.global_10a".
	if oldParts, newParts := strings.Split(oldEngineVersion, "."), strings.Split(newEngineVersion, "."); len(oldParts) == 3 && //nolint:gocritic // Ignore 'badCond'
		len(oldParts) == len(newParts) &&
		oldParts[0] == newParts[0] &&
		oldParts[1] == newParts[1] &&
		strings.HasSuffix(newParts[2], oldParts[2]) {
		d.Set(names.AttrEngineVersion, oldEngineVersion)
		d.Set("engine_version_actual", newEngineVersion)
	} else {
		d.Set(names.AttrEngineVersion, newEngineVersion)
		d.Set("engine_version_actual", newEngineVersion)
	}

	return diags
}

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.ModifyGlobalClusterInput{
		DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrEngineVersion) {
		if err := globalClusterUpgradeEngineVersion(ctx, d, meta, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Global Cluster (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Updating RDS Global Cluster (%s): %s", d.Id(), input)
	_, err := conn.ModifyGlobalClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating RDS Global Cluster (%s): %s", d.Id(), err)
	}

	if err := waitGlobalClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutDelete))

	if d.Get(names.AttrForceDestroy).(bool) {
		log.Printf("[DEBUG] Removing cluster members from RDS Global Cluster: %s", d.Id())

		// The writer cluster must be removed last
		var writerARN string
		globalClusterMembers := d.Get("global_cluster_members").(*schema.Set)
		if globalClusterMembers.Len() > 0 {
			for _, globalClusterMemberRaw := range globalClusterMembers.List() {
				globalClusterMember, ok := globalClusterMemberRaw.(map[string]interface{})

				if !ok {
					continue
				}

				dbClusterArn, ok := globalClusterMember["db_cluster_arn"].(string)

				if !ok {
					continue
				}

				if globalClusterMember["is_writer"].(bool) {
					writerARN = dbClusterArn
					continue
				}

				input := &rds.RemoveFromGlobalClusterInput{
					DbClusterIdentifier:     aws.String(dbClusterArn),
					GlobalClusterIdentifier: aws.String(d.Id()),
				}

				_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)

				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
					continue
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "removing RDS DB Cluster (%s) from Global Cluster (%s): %s", dbClusterArn, d.Id(), err)
				}

				if err := waitForGlobalClusterRemoval(ctx, conn, dbClusterArn, deadline.Remaining()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) removal from RDS Global Cluster (%s): %s", dbClusterArn, d.Id(), err)
				}
			}

			input := &rds.RemoveFromGlobalClusterInput{
				DbClusterIdentifier:     aws.String(writerARN),
				GlobalClusterIdentifier: aws.String(d.Id()),
			}

			_, err := conn.RemoveFromGlobalClusterWithContext(ctx, input)
			if err != nil {
				if !tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
					return sdkdiag.AppendErrorf(diags, "removing RDS DB Cluster (%s) from Global Cluster (%s): %s", writerARN, d.Id(), err)
				}
			}

			if err := waitForGlobalClusterRemoval(ctx, conn, writerARN, deadline.Remaining()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) removal from RDS Global Cluster (%s): %s", writerARN, d.Id(), err)
			}
		}
	}

	input := &rds.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting RDS Global Cluster: %s", d.Id())

	// Allow for eventual consistency
	// InvalidGlobalClusterStateFault: Global Cluster arn:aws:rds::123456789012:global-cluster:tf-acc-test-5618525093076697001-0 is not empty
	const (
		// GlobalClusterClusterDeleteTimeout is the timeout for actual deletion of the cluster
		// This operation will be quick if successful
		globalClusterClusterDeleteTimeout = 5 * time.Minute
	)
	var timeout time.Duration
	if x, y := deadline.Remaining(), globalClusterClusterDeleteTimeout; x < y {
		timeout = x
	} else {
		timeout = y
	}
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err := conn.DeleteGlobalClusterWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidGlobalClusterStateFault, "is not empty") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteGlobalClusterWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Global Cluster: %s", err)
	}

	if err := waitGlobalClusterDeleted(ctx, conn, d.Id(), deadline.Remaining()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindGlobalClusterByDBClusterARN(ctx context.Context, conn *rds.RDS, dbClusterARN string) (*rds.GlobalCluster, error) {
	input := &rds.DescribeGlobalClustersInput{
		Filters: []*rds.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: aws.StringSlice([]string{dbClusterARN}),
			},
		},
	}

	return findGlobalCluster(ctx, conn, input, func(v *rds.GlobalCluster) bool {
		for _, v := range v.GlobalClusterMembers {
			if aws.StringValue(v.DBClusterArn) == dbClusterARN {
				return true
			}
		}
		return false
	})
}

func FindGlobalClusterByID(ctx context.Context, conn *rds.RDS, id string) (*rds.GlobalCluster, error) {
	input := &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(id),
	}
	output, err := findGlobalCluster(ctx, conn, input, tfslices.PredicateTrue[*rds.GlobalCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findGlobalCluster(ctx context.Context, conn *rds.RDS, input *rds.DescribeGlobalClustersInput, filter tfslices.Predicate[*rds.GlobalCluster]) (*rds.GlobalCluster, error) {
	output, err := findGlobalClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findGlobalClusters(ctx context.Context, conn *rds.RDS, input *rds.DescribeGlobalClustersInput, filter tfslices.Predicate[*rds.GlobalCluster]) ([]*rds.GlobalCluster, error) {
	var output []*rds.GlobalCluster

	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GlobalClusters {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
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

func flattenGlobalClusterMembers(apiObjects []*rds.GlobalClusterMember) []interface{} {
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

func statusGlobalCluster(ctx context.Context, conn *rds.RDS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGlobalClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitGlobalClusterCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{GlobalClusterStatusCreating},
		Target:  []string{GlobalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitGlobalClusterUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{GlobalClusterStatusModifying, GlobalClusterStatusUpgrading},
		Target:  []string{GlobalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitGlobalClusterDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{GlobalClusterStatusAvailable, GlobalClusterStatusDeleting},
		Target:         []string{},
		Refresh:        statusGlobalCluster(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitForGlobalClusterRemoval(ctx context.Context, conn *rds.RDS, dbClusterARN string, timeout time.Duration) error {
	_, err := tfresource.RetryUntilNotFound(ctx, timeout, func() (interface{}, error) {
		return FindGlobalClusterByDBClusterARN(ctx, conn, dbClusterARN)
	})

	return err
}

// globalClusterUpgradeEngineVersion upgrades the engine version of the RDS Global Cluster, accommodating
// either a MAJOR or MINOR version upgrade. Given only the old and new versions, determining whether to
// perform a MAJOR or MINOR upgrade is challenging. Instead of attempting to parse numerous combinations
// of engines and versions, we initially attempt a major upgrade. If AWS returns an error indicating that
// a minor version upgrade is supported ("InvalidParameterValue"/"only supports Major Version Upgrades"),
// we infer that a minor upgrade will suffice. Therefore, it's crucial to recognize that this error serves
// as more than just an error; it guides our decision between major and minor upgrades.

// IMPORTANT: Altering the error handling in `globalClusterUpgradeMajorEngineVersion` can disrupt the
// logic, including the handling of errors that signify the need for a minor version upgrade.
func globalClusterUpgradeEngineVersion(ctx context.Context, d *schema.ResourceData, meta interface{}, timeout time.Duration) error {
	log.Printf("[DEBUG] Upgrading RDS Global Cluster (%s) engine version: %s", d.Id(), d.Get(names.AttrEngineVersion))

	err := globalClusterUpgradeMajorEngineVersion(ctx, meta, d.Id(), d.Get(names.AttrEngineVersion).(string), timeout)

	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "only supports Major Version Upgrades") {
		err = globalClusterUpgradeMinorEngineVersion(ctx, meta, d.Get("global_cluster_members").(*schema.Set), d.Id(), d.Get(names.AttrEngineVersion).(string), timeout)

		if err != nil {
			return fmt.Errorf("while upgrading minor version of RDS Global Cluster (%s): %w", d.Id(), err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("while upgrading major version of RDS Global Cluster (%s): %w", d.Id(), err)
	}

	return nil
}

// globalClusterUpgradeMajorEngineVersion attempts a major version upgrade. If that fails, it returns an
// error that is more than just an error but also indicates a minor version should be tried. It's IMPORTANT
// to not handle this like any other error, such as, for example, retrying the error, since it indicates
// another branch of logic. Please use caution when updating!!
func globalClusterUpgradeMajorEngineVersion(ctx context.Context, meta interface{}, clusterID string, engineVersion string, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.ModifyGlobalClusterInput{
		AllowMajorVersionUpgrade: aws.Bool(true),
		EngineVersion:            aws.String(engineVersion),
		GlobalClusterIdentifier:  aws.String(clusterID),
	}

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		_, err := conn.ModifyGlobalClusterWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
				return retry.NonRetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "only supports Major Version Upgrades") {
				return retry.NonRetryableError(err) // NOT retryable !! AND indicates this should be a minor version upgrade
			}

			return retry.RetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.ModifyGlobalClusterWithContext(ctx, input)
	}

	if err != nil {
		return fmt.Errorf("while upgrading major version of RDS Global Cluster (%s): %w", clusterID, err)
	}

	globalCluster, err := FindGlobalClusterByID(ctx, conn, clusterID)

	if err != nil {
		return fmt.Errorf("while upgrading major version of RDS Global Cluster (%s): %w", clusterID, err)
	}

	for _, clusterMember := range globalCluster.GlobalClusterMembers {
		arnID := aws.StringValue(clusterMember.DBClusterArn)

		if arnID == "" {
			continue
		}

		dbi, clusterRegion, err := ClusterIDRegionFromARN(arnID)
		if err != nil {
			return fmt.Errorf("while upgrading RDS Global Cluster Cluster major engine version: %w", err)
		}

		if dbi == "" {
			continue
		}

		// Clusters may not all be in the same region.
		useConn := meta.(*conns.AWSClient).RDSConnForRegion(ctx, clusterRegion)

		if err := WaitForClusterUpdate(ctx, useConn, dbi, timeout); err != nil {
			return fmt.Errorf("failed to update engine_version, waiting for RDS Global Cluster (%s) to update: %s", dbi, err)
		}
	}

	return err
}

func globalClusterUpgradeMinorEngineVersion(ctx context.Context, meta interface{}, clusterMembers *schema.Set, clusterID, engineVersion string, timeout time.Duration) error {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[INFO] Performing RDS Global Cluster (%s) minor version (%s) upgrade", clusterID, engineVersion)

	leelooMultiPass := false // only one pass is needed

	for _, clusterMemberRaw := range clusterMembers.List() {
		clusterMember := clusterMemberRaw.(map[string]interface{})

		// DBClusterIdentifier supposedly can be either ARN or ID, and both used to work,
		// but as of now, only ID works
		if clusterMemberArn, ok := clusterMember["db_cluster_arn"]; !ok || clusterMemberArn.(string) == "" {
			continue
		}

		arnID := clusterMember["db_cluster_arn"].(string)

		dbi, clusterRegion, err := ClusterIDRegionFromARN(arnID)
		if err != nil {
			return fmt.Errorf("while upgrading RDS Global Cluster Cluster minor engine version: %w", err)
		}

		if dbi == "" {
			continue
		}

		useConn := meta.(*conns.AWSClient).RDSConnForRegion(ctx, clusterRegion)

		// pre-wait for the cluster to be in a state where it can be updated
		if err := WaitForClusterUpdate(ctx, useConn, dbi, timeout); err != nil {
			return fmt.Errorf("failed to update engine_version, waiting for RDS Global Cluster Cluster (%s) to update: %s", dbi, err)
		}

		modInput := &rds.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(true),
			DBClusterIdentifier: aws.String(dbi),
			EngineVersion:       aws.String(engineVersion),
		}

		log.Printf("[INFO] Performing RDS Global Cluster (%s) Cluster (%s) minor version (%s) upgrade", clusterID, dbi, engineVersion)

		err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
			_, err := useConn.ModifyDBClusterWithContext(ctx, modInput)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
					return retry.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "Cannot modify engine version without a primary instance in DB cluster") {
					return retry.NonRetryableError(err)
				}

				if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidDBClusterStateFault) {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err := useConn.ModifyDBClusterWithContext(ctx, modInput)
			if err != nil {
				return err
			}
		}

		if tfawserr.ErrMessageContains(err, "InvalidGlobalClusterStateFault", "is upgrading") {
			leelooMultiPass = true
			continue
		}

		if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "upgrade global replicas first") {
			leelooMultiPass = true
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to update engine_version on RDS Global Cluster Cluster (%s): %s", dbi, err)
		}

		log.Printf("[INFO] Waiting for RDS Global Cluster (%s) Cluster (%s) minor version (%s) upgrade", clusterID, dbi, engineVersion)
		if err := WaitForClusterUpdate(ctx, useConn, dbi, timeout); err != nil {
			return fmt.Errorf("failed to update engine_version, waiting for RDS Global Cluster Cluster (%s) to update: %s", dbi, err)
		}
	}

	globalCluster, err := FindGlobalClusterByID(ctx, conn, clusterID)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeGlobalClusterNotFoundFault) {
		return fmt.Errorf("after upgrading engine_version, could not find RDS Global Cluster (%s): %s", clusterID, err)
	}

	if err != nil {
		return fmt.Errorf("after minor engine_version upgrade to RDS Global Cluster (%s): %s", clusterID, err)
	}

	if globalCluster == nil {
		return fmt.Errorf("after minor engine_version upgrade to RDS Global Cluster (%s): empty response", clusterID)
	}

	if leelooMultiPass || aws.StringValue(globalCluster.EngineVersion) != engineVersion {
		log.Printf("[DEBUG] RDS Global Cluster (%s) upgrade did not take effect, trying again", clusterID)

		return globalClusterUpgradeMinorEngineVersion(ctx, meta, clusterMembers, clusterID, engineVersion, timeout)
	}

	return nil
}

func ClusterIDRegionFromARN(arnID string) (string, string, error) {
	parsedARN, err := arn.Parse(arnID)
	if err != nil {
		return "", "", fmt.Errorf("could not parse ARN (%s): %w", arnID, err)
	}

	dbi := ""

	if parsedARN.Resource != "" {
		parts := strings.Split(parsedARN.Resource, ":")

		if len(parts) < 2 {
			return "", "", fmt.Errorf("could not get DB Cluster ID from parsing ARN (%s): %w", arnID, err)
		}

		if parsedARN.Service != rds.EndpointsID || parts[0] != "cluster" {
			return "", "", fmt.Errorf("wrong ARN (%s) for a DB Cluster", arnID)
		}

		dbi = parts[1]
	}

	return dbi, parsedARN.Region, nil
}

var resourceClusterUpdatePendingStates = []string{
	"backing-up",
	"configuring-iam-database-auth",
	"modifying",
	"renaming",
	"resetting-master-credentials",
	"upgrading",
}

func WaitForClusterUpdate(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    resourceClusterUpdatePendingStates,
		Target:     []string{"available"},
		Refresh:    resourceClusterStateRefreshFunc(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceClusterStateRefreshFunc(ctx context.Context, conn *rds.RDS, dbClusterIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDBClustersWithContext(ctx, &rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(dbClusterIdentifier),
		})

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
			return 42, "destroyed", nil
		}

		if err != nil {
			return nil, "", err
		}

		var dbc *rds.DBCluster

		for _, c := range resp.DBClusters {
			if aws.StringValue(c.DBClusterIdentifier) == dbClusterIdentifier {
				dbc = c
			}
		}

		if dbc == nil {
			return 42, "destroyed", nil
		}

		if dbc.Status != nil {
			log.Printf("[DEBUG] DB Cluster status (%s): %s", dbClusterIdentifier, *dbc.Status)
		}

		return dbc, aws.StringValue(dbc.Status), nil
	}
}
