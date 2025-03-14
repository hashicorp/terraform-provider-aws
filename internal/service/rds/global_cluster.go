// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_global_cluster", name="Global Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/rds/types;types.GlobalCluster")
func resourceGlobalCluster() *schema.Resource {
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
				Computed: true,
				ForceNew: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	globalClusterID := d.Get("global_cluster_identifier").(string)
	input := &rds.CreateGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
		Tags:                    getTagsIn(ctx),
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

	output, err := conn.CreateGlobalCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.ToString(output.GlobalCluster.GlobalClusterIdentifier))

	if _, err := waitGlobalClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	globalCluster, err := findGlobalClusterByID(ctx, conn, d.Id())

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
	d.Set(names.AttrEndpoint, globalCluster.Endpoint)
	d.Set(names.AttrEngine, globalCluster.Engine)
	d.Set("engine_lifecycle_support", globalCluster.EngineLifecycleSupport)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}
	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set(names.AttrStorageEncrypted, globalCluster.StorageEncrypted)

	oldEngineVersion, newEngineVersion := d.Get(names.AttrEngineVersion).(string), aws.ToString(globalCluster.EngineVersion)

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

	setTagsOut(ctx, globalCluster.TagList)

	return diags
}

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange(names.AttrEngineVersion) {
		if err := globalClusterUpgradeEngineVersion(ctx, conn, d, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &rds.ModifyGlobalClusterInput{
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			GlobalClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.ModifyGlobalCluster(ctx, input)

		if errs.IsA[*types.GlobalClusterNotFoundFault](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Global Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitGlobalClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)
	deadline := tfresource.NewDeadline(d.Timeout(schema.TimeoutDelete))

	if d.Get(names.AttrForceDestroy).(bool) {
		log.Printf("[DEBUG] Removing cluster members from RDS Global Cluster: %s", d.Id())

		// The writer cluster must be removed last.
		var writerARN string
		globalClusterMembers := d.Get("global_cluster_members").(*schema.Set)
		if globalClusterMembers.Len() > 0 {
			for _, tfMapRaw := range globalClusterMembers.List() {
				tfMap, ok := tfMapRaw.(map[string]any)
				if !ok {
					continue
				}

				dbClusterARN, ok := tfMap["db_cluster_arn"].(string)
				if !ok {
					continue
				}

				if tfMap["is_writer"].(bool) {
					writerARN = dbClusterARN
					continue
				}

				input := &rds.RemoveFromGlobalClusterInput{
					DbClusterIdentifier:     aws.String(dbClusterARN),
					GlobalClusterIdentifier: aws.String(d.Id()),
				}

				_, err := conn.RemoveFromGlobalCluster(ctx, input)

				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
					continue
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "removing RDS DB Cluster (%s) from RDS Global Cluster (%s): %s", dbClusterARN, d.Id(), err)
				}

				if _, err := waitGlobalClusterMemberRemoved(ctx, conn, dbClusterARN, deadline.Remaining()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) removal from RDS Global Cluster (%s): %s", dbClusterARN, d.Id(), err)
				}
			}

			input := &rds.RemoveFromGlobalClusterInput{
				DbClusterIdentifier:     aws.String(writerARN),
				GlobalClusterIdentifier: aws.String(d.Id()),
			}

			_, err := conn.RemoveFromGlobalCluster(ctx, input)

			if err != nil {
				if !tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "is not found in global cluster") {
					return sdkdiag.AppendErrorf(diags, "removing RDS DB Cluster (%s) from RDS Global Cluster (%s): %s", writerARN, d.Id(), err)
				}
			}

			if _, err := waitGlobalClusterMemberRemoved(ctx, conn, writerARN, deadline.Remaining()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster (%s) removal from RDS Global Cluster (%s): %s", writerARN, d.Id(), err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting RDS Global Cluster: %s", d.Id())

	// Allow for eventual consistency
	// InvalidGlobalClusterStateFault: Global Cluster arn:aws:rds::123456789012:global-cluster:tf-acc-test-5618525093076697001-0 is not empty
	const (
		// GlobalClusterClusterDeleteTimeout is the timeout for actual deletion of the cluster
		// This operation will be quick if successful
		globalClusterClusterDeleteTimeout = 5 * time.Minute
	)
	timeout := max(deadline.Remaining(), globalClusterClusterDeleteTimeout)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidGlobalClusterStateFault](ctx, timeout, func() (any, error) {
		return conn.DeleteGlobalCluster(ctx, &rds.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: aws.String(d.Id()),
		})
	}, "is not empty")

	if errs.IsA[*types.GlobalClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Global Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalClusterDeleted(ctx, conn, d.Id(), deadline.Remaining()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Global Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGlobalClusterByDBClusterARN(ctx context.Context, conn *rds.Client, dbClusterARN string) (*types.GlobalCluster, error) {
	input := &rds.DescribeGlobalClustersInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: []string{dbClusterARN},
			},
		},
	}

	return findGlobalCluster(ctx, conn, input, func(v *types.GlobalCluster) bool {
		return slices.ContainsFunc(v.GlobalClusterMembers, func(v types.GlobalClusterMember) bool {
			return aws.ToString(v.DBClusterArn) == dbClusterARN
		})
	})
}

func findGlobalClusterByID(ctx context.Context, conn *rds.Client, id string) (*types.GlobalCluster, error) {
	input := &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(id),
	}
	output, err := findGlobalCluster(ctx, conn, input, tfslices.PredicateTrue[*types.GlobalCluster]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findGlobalCluster(ctx context.Context, conn *rds.Client, input *rds.DescribeGlobalClustersInput, filter tfslices.Predicate[*types.GlobalCluster]) (*types.GlobalCluster, error) {
	output, err := findGlobalClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGlobalClusters(ctx context.Context, conn *rds.Client, input *rds.DescribeGlobalClustersInput, filter tfslices.Predicate[*types.GlobalCluster]) ([]types.GlobalCluster, error) {
	var output []types.GlobalCluster

	pages := rds.NewDescribeGlobalClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.GlobalClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusGlobalCluster(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findGlobalClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitGlobalClusterCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusCreating},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterUpdated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusModifying, globalClusterStatusUpgrading},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{globalClusterStatusAvailable, globalClusterStatusDeleting},
		Target:         []string{},
		Refresh:        statusGlobalCluster(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterMemberRemoved(ctx context.Context, conn *rds.Client, dbClusterARN string, timeout time.Duration) (*types.GlobalCluster, error) { //nolint:unparam
	outputRaw, err := tfresource.RetryUntilNotFound(ctx, timeout, func() (any, error) {
		return findGlobalClusterByDBClusterARN(ctx, conn, dbClusterARN)
	})

	if output, ok := outputRaw.(*types.GlobalCluster); ok {
		return output, err
	}

	return nil, err
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
func globalClusterUpgradeEngineVersion(ctx context.Context, conn *rds.Client, d *schema.ResourceData, timeout time.Duration) error {
	log.Printf("[DEBUG] Upgrading RDS Global Cluster (%s) engine version: %s", d.Id(), d.Get(names.AttrEngineVersion))

	err := globalClusterUpgradeMajorEngineVersion(ctx, conn, d.Id(), d.Get(names.AttrEngineVersion).(string), timeout)

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "only supports Major Version Upgrades") {
		if err := globalClusterUpgradeMinorEngineVersion(ctx, conn, d.Id(), d.Get(names.AttrEngineVersion).(string), d.Get("global_cluster_members").(*schema.Set), timeout); err != nil {
			return fmt.Errorf("upgrading minor version of RDS Global Cluster (%s): %w", d.Id(), err)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("upgrading major version of RDS Global Cluster (%s): %w", d.Id(), err)
	}

	return nil
}

// globalClusterUpgradeMajorEngineVersion attempts a major version upgrade. If that fails, it returns an
// error that is more than just an error but also indicates a minor version should be tried. It's IMPORTANT
// to not handle this like any other error, such as, for example, retrying the error, since it indicates
// another branch of logic. Please use caution when updating!!
func globalClusterUpgradeMajorEngineVersion(ctx context.Context, conn *rds.Client, globalClusterID, engineVersion string, timeout time.Duration) error {
	input := &rds.ModifyGlobalClusterInput{
		AllowMajorVersionUpgrade: aws.Bool(true),
		EngineVersion:            aws.String(engineVersion),
		GlobalClusterIdentifier:  aws.String(globalClusterID),
	}

	_, err := tfresource.RetryWhen(ctx, timeout,
		func() (any, error) {
			return conn.ModifyGlobalCluster(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*types.GlobalClusterNotFoundFault](err) {
				return false, err
			}

			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "only supports Major Version Upgrades") {
				return false, err // NOT retryable !! AND indicates this should be a minor version upgrade
			}

			// Any other errors are retryable.
			return err != nil, err
		},
	)

	if err != nil {
		return fmt.Errorf("modifying RDS Global Cluster (%s) EngineVersion: %w", globalClusterID, err)
	}

	globalCluster, err := findGlobalClusterByID(ctx, conn, globalClusterID)

	if err != nil {
		return fmt.Errorf("after major engine_version upgrade to RDS Global Cluster (%s): %w", globalClusterID, err)
	}

	for _, clusterMember := range globalCluster.GlobalClusterMembers {
		memberARN := aws.ToString(clusterMember.DBClusterArn)

		if memberARN == "" {
			continue
		}

		clusterID, clusterRegion, err := clusterIDAndRegionFromARN(memberARN)
		if err != nil {
			return err
		}

		if clusterID == "" {
			continue
		}

		optFn := func(o *rds.Options) {
			o.Region = clusterRegion
		}

		if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
			return fmt.Errorf("waiting for RDS Global Cluster (%s) member (%s) update: %w", globalClusterID, clusterID, err)
		}
	}

	return err
}

func globalClusterUpgradeMinorEngineVersion(ctx context.Context, conn *rds.Client, globalClusterID, engineVersion string, clusterMembers *schema.Set, timeout time.Duration) error {
	log.Printf("[INFO] Performing RDS Global Cluster (%s) minor version (%s) upgrade", globalClusterID, engineVersion)

	leelooMultiPass := false // only one pass is needed

	for _, tfMapRaw := range clusterMembers.List() {
		tfMap := tfMapRaw.(map[string]any)

		// DBClusterIdentifier supposedly can be either ARN or ID, and both used to work,
		// but as of now, only ID works.
		if memberARN, ok := tfMap["db_cluster_arn"]; !ok || memberARN.(string) == "" {
			continue
		}

		memberARN := tfMap["db_cluster_arn"].(string)

		clusterID, clusterRegion, err := clusterIDAndRegionFromARN(memberARN)
		if err != nil {
			return err
		}

		if clusterID == "" {
			continue
		}

		optFn := func(o *rds.Options) {
			o.Region = clusterRegion
		}

		// pre-wait for the cluster to be in a state where it can be updated
		if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
			return fmt.Errorf("waiting for RDS Global Cluster (%s) member (%s) update: %w", globalClusterID, clusterID, err)
		}

		input := &rds.ModifyDBClusterInput{
			ApplyImmediately:    aws.Bool(true),
			DBClusterIdentifier: aws.String(clusterID),
			EngineVersion:       aws.String(engineVersion),
		}

		log.Printf("[INFO] Performing RDS Global Cluster (%s) Cluster (%s) minor version (%s) upgrade", globalClusterID, clusterID, engineVersion)
		_, err = tfresource.RetryWhen(ctx, timeout,
			func() (any, error) {
				return conn.ModifyDBCluster(ctx, input, optFn)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
					return true, err
				}

				if errs.IsAErrorMessageContains[*types.InvalidDBClusterStateFault](err, "Cannot modify engine version without a primary instance in DB cluster") {
					return false, err
				}

				if errs.IsA[*types.InvalidDBClusterStateFault](err) {
					return true, err
				}

				return false, err
			},
		)

		if errs.IsAErrorMessageContains[*types.InvalidGlobalClusterStateFault](err, "is upgrading") {
			leelooMultiPass = true
			continue
		}

		if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "upgrade global replicas first") {
			leelooMultiPass = true
			continue
		}

		if err != nil {
			return fmt.Errorf("modifying RDS Global Cluster (%s) member (%s) EngineVersion: %w", globalClusterID, clusterID, err)
		}

		if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
			return fmt.Errorf("waiting for RDS Global Cluster (%s) member (%s) update: %w", globalClusterID, clusterID, err)
		}
	}

	globalCluster, err := findGlobalClusterByID(ctx, conn, globalClusterID)

	if err != nil {
		return fmt.Errorf("after minor engine_version upgrade to RDS Global Cluster (%s) members: %w", globalClusterID, err)
	}

	if leelooMultiPass || aws.ToString(globalCluster.EngineVersion) != engineVersion {
		log.Printf("[DEBUG] RDS Global Cluster (%s) upgrade did not take effect, trying again", globalClusterID)

		return globalClusterUpgradeMinorEngineVersion(ctx, conn, globalClusterID, engineVersion, clusterMembers, timeout)
	}

	return nil
}

func clusterIDAndRegionFromARN(clusterARN string) (string, string, error) {
	parsedARN, err := arn.Parse(clusterARN)
	if err != nil {
		return "", "", fmt.Errorf("could not parse ARN (%s): %w", clusterARN, err)
	}

	dbi := ""

	if parsedARN.Resource != "" {
		parts := strings.Split(parsedARN.Resource, ":")

		if len(parts) < 2 {
			return "", "", fmt.Errorf("could not get DB Cluster ID from parsing ARN (%s): %w", clusterARN, err)
		}

		if parsedARN.Service != "rds" || parts[0] != "cluster" {
			return "", "", fmt.Errorf("wrong ARN (%s) for a DB Cluster", clusterARN)
		}

		dbi = parts[1]
	}

	return dbi, parsedARN.Region, nil
}

func waitGlobalClusterMemberUpdated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration, optFns ...func(*rds.Options)) (*types.DBCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			clusterStatusBackingUp,
			clusterStatusConfiguringIAMDatabaseAuth,
			clusterStatusModifying,
			clusterStatusRenaming,
			clusterStatusResettingMasterCredentials,
			clusterStatusUpgrading,
		},
		Target:     []string{clusterStatusAvailable},
		Refresh:    statusDBCluster(ctx, conn, id, false, optFns...),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func flattenGlobalClusterMembers(apiObjects []types.GlobalClusterMember) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"db_cluster_arn": aws.ToString(apiObject.DBClusterArn),
			"is_writer":      aws.ToBool(apiObject.IsWriter),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
