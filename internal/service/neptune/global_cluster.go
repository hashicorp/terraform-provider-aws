// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package neptune

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_global_cluster", name="Global Cluster")
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
			Create: schema.DefaultTimeout(5 * time.Minute),
			// Update timeout equal to aws_neptune_cluster's Update timeout value
			// as updating a global cluster can result in a cluster modification.
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDeletionProtection: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrEngine: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrEngine, "source_db_cluster_identifier"},
				ValidateFunc: validation.StringInSlice(engine_Values(), false),
			},
			names.AttrEngineVersion: {
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrEngine, "source_db_cluster_identifier"},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	globalClusterID := d.Get("global_cluster_identifier").(string)
	input := &neptune.CreateGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	if v, ok := d.GetOk(names.AttrDeletionProtection); ok {
		input.DeletionProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrEngine); ok {
		input.Engine = aws.String(v.(string))
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

	var output *neptune.CreateGlobalClusterOutput
	var err error
	for l := backoff.NewLoop(d.Timeout(schema.TimeoutCreate)); l.Continue(ctx); {
		output, err = conn.CreateGlobalCluster(ctx, input)

		if tfawserr.ErrMessageContains(err, errCodeInvalidGlobalClusterStateFault, "in progress") {
			continue
		}

		break
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.ToString(output.GlobalCluster.GlobalClusterIdentifier))

	if _, err := waitGlobalClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Global Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	globalCluster, err := findGlobalClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Neptune Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Global Cluster (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, globalCluster.GlobalClusterArn)
	d.Set(names.AttrDeletionProtection, globalCluster.DeletionProtection)
	d.Set(names.AttrEngine, globalCluster.Engine)
	d.Set(names.AttrEngineVersion, globalCluster.EngineVersion)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting global_cluster_members: %s", err)
	}
	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set(names.AttrStorageEncrypted, globalCluster.StorageEncrypted)

	return diags
}

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	if d.HasChange(names.AttrDeletionProtection) {
		input := &neptune.ModifyGlobalClusterInput{
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			GlobalClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.ModifyGlobalCluster(ctx, input)

		if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Global Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitGlobalClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Global Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrEngineVersion) {
		if err := globalClusterUpgradeEngineVersion(ctx, conn, d, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Global Cluster (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	// Remove any members from the global cluster.
	for _, tfMapRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		if clusterARN, ok := tfMap["db_cluster_arn"].(string); ok && clusterARN != "" {
			if err := removeClusterFromGlobalCluster(ctx, conn, clusterARN, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting Neptune Global Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.InvalidGlobalClusterStateFault](ctx, d.Timeout(schema.TimeoutDelete), func(ctx context.Context) (any, error) {
		return conn.DeleteGlobalCluster(ctx, &neptune.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: aws.String(d.Id()),
		})
	}, "is not empty")

	if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Global Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Global Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

// globalClusterUpgradeEngineVersion upgrades the global cluster and its members to the
// specified engine version. It first attempts a major version upgrade. If that fails with an
// error indicating only minor upgrades are supported, it attempts a minor version upgrade for
// all members of the global cluster. This is necessary because ModifyGlobalCluster only
// supports major version upgrades; minor upgrades must be performed on each member cluster
// individually.
func globalClusterUpgradeEngineVersion(ctx context.Context, conn *neptune.Client, d *schema.ResourceData, timeout time.Duration) error {
	log.Printf("[DEBUG] Upgrading Neptune Global Cluster (%s) engine version: %s", d.Id(), d.Get(names.AttrEngineVersion))

	err := globalClusterUpgradeMajorEngineVersion(ctx, conn, d.Id(), d.Get(names.AttrEngineVersion).(string), timeout)

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "doesn't support minor version upgrades") {
		if err := globalClusterUpgradeMinorEngineVersion(ctx, conn, d.Id(), d.Get(names.AttrEngineVersion).(string), d.Get("global_cluster_members").(*schema.Set), timeout); err != nil {
			return fmt.Errorf("upgrading minor version of Neptune Global Cluster (%s): %w", d.Id(), err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("upgrading major version of Neptune Global Cluster (%s): %w", d.Id(), err)
	}
	return nil
}

func globalClusterUpgradeMajorEngineVersion(ctx context.Context, conn *neptune.Client, globalClusterID, engineVersion string, timeout time.Duration) error {
	input := &neptune.ModifyGlobalClusterInput{
		AllowMajorVersionUpgrade: aws.Bool(true),
		EngineVersion:            aws.String(engineVersion),
		GlobalClusterIdentifier:  aws.String(globalClusterID),
	}
	_, err := tfresource.RetryWhen(ctx, timeout,
		func(ctx context.Context) (any, error) {
			return conn.ModifyGlobalCluster(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
				return false, err
			}
			if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "doesn't support minor version upgrades") {
				return false, err // NOT retryable, indicates minor upgrade or wrong order
			}
			return err != nil, err
		},
	)
	if err != nil {
		return fmt.Errorf("modifying Neptune Global Cluster (%s) EngineVersion: %w", globalClusterID, err)
	}
	globalCluster, err := findGlobalClusterByID(ctx, conn, globalClusterID)
	if err != nil {
		return fmt.Errorf("after major engine_version upgrade to Neptune Global Cluster (%s): %w", globalClusterID, err)
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
		optFn := func(o *neptune.Options) { o.Region = clusterRegion }
		if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
			return fmt.Errorf("waiting for Neptune Global Cluster (%s) member (%s) update: %w", globalClusterID, clusterID, err)
		}
	}
	return err
}

func globalClusterUpgradeMinorEngineVersion(ctx context.Context, conn *neptune.Client, globalClusterID, engineVersion string, clusterMembers *schema.Set, timeout time.Duration) error {
	log.Printf("[INFO] Performing Neptune Global Cluster (%s) minor version (%s) upgrade", globalClusterID, engineVersion)
	var (
		primaryMember    map[string]any
		secondaryMembers []map[string]any
	)
	for _, tfMapRaw := range clusterMembers.List() {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}
		if isWriter, ok := tfMap["is_writer"].(bool); ok && isWriter {
			primaryMember = tfMap
		} else {
			secondaryMembers = append(secondaryMembers, tfMap)
		}
	}
	var wg sync.WaitGroup
	errChan := make(chan error, len(secondaryMembers))
	// Upgrade all secondary clusters in parallel
	for _, tfMap := range secondaryMembers {
		wg.Add(1)
		go func(tfMap map[string]any) {
			defer wg.Done()
			memberARN, ok := tfMap["db_cluster_arn"].(string)
			if !ok || memberARN == "" {
				return
			}
			clusterID, clusterRegion, err := clusterIDAndRegionFromARN(memberARN)
			if err != nil || clusterID == "" {
				errChan <- err
				return
			}
			optFn := func(o *neptune.Options) { o.Region = clusterRegion }
			if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
				errChan <- err
				return
			}
			input := &neptune.ModifyDBClusterInput{
				ApplyImmediately:    aws.Bool(true),
				DBClusterIdentifier: aws.String(clusterID),
				EngineVersion:       aws.String(engineVersion),
			}
			_, err = tfresource.RetryWhen(ctx, timeout,
				func(ctx context.Context) (any, error) {
					return conn.ModifyDBCluster(ctx, input, optFn)
				},
				func(err error) (bool, error) {
					if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
						return true, err
					}
					if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "Cannot modify engine version without a primary instance in DB cluster") {
						return false, err
					}
					if errs.IsA[*awstypes.InvalidDBClusterStateFault](err) {
						return true, err
					}
					return false, err
				},
			)
			if err != nil {
				errChan <- err
				return
			}
			if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
				errChan <- err
				return
			}
		}(tfMap)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	// Upgrade primary member last
	if primaryMember != nil {
		memberARN, ok := primaryMember["db_cluster_arn"].(string)
		if ok && memberARN != "" {
			clusterID, clusterRegion, err := clusterIDAndRegionFromARN(memberARN)
			if err != nil || clusterID == "" {
				return err
			}
			optFn := func(o *neptune.Options) { o.Region = clusterRegion }
			if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
				return err
			}
			input := &neptune.ModifyDBClusterInput{
				ApplyImmediately:    aws.Bool(true),
				DBClusterIdentifier: aws.String(clusterID),
				EngineVersion:       aws.String(engineVersion),
			}
			_, err = tfresource.RetryWhen(ctx, timeout,
				func(ctx context.Context) (any, error) {
					return conn.ModifyDBCluster(ctx, input, optFn)
				},
				func(err error) (bool, error) {
					if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions") {
						return true, err
					}
					if errs.IsAErrorMessageContains[*awstypes.InvalidDBClusterStateFault](err, "Cannot modify engine version without a primary instance in DB cluster") {
						return false, err
					}
					if errs.IsA[*awstypes.InvalidDBClusterStateFault](err) {
						return true, err
					}
					if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "upgrade global replicas first before upgrading the primary member") {
						return false, err
					}
					return false, err
				},
			)
			if err != nil {
				return err
			}
			if _, err := waitGlobalClusterMemberUpdated(ctx, conn, clusterID, timeout, optFn); err != nil {
				return err
			}
		}
	}
	globalCluster, err := findGlobalClusterByID(ctx, conn, globalClusterID)
	if err != nil {
		return fmt.Errorf("after minor engine_version upgrade to Neptune Global Cluster (%s) members: %w", globalClusterID, err)
	}
	if aws.ToString(globalCluster.EngineVersion) != engineVersion {
		log.Printf("[DEBUG] Neptune Global Cluster (%s) upgrade did not take effect, trying again", globalClusterID)
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
		// NOTE: Neptune DB Clusters and Instances have "rds" ARNs!
		if (parsedARN.Service != "neptune" && parsedARN.Service != "rds") || parts[0] != "cluster" {
			return "", "", fmt.Errorf("wrong ARN (%s) for a Neptune DB Cluster", clusterARN)
		}
		dbi = parts[1]
	}
	return dbi, parsedARN.Region, nil
}

func waitGlobalClusterMemberUpdated(ctx context.Context, conn *neptune.Client, id string, timeout time.Duration, optFns ...func(*neptune.Options)) (*awstypes.DBCluster, error) { //nolint:unparam
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
		Refresh:    statusDBCluster(conn, id, false, optFns...),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.DBCluster); ok {
		return output, err
	}

	return nil, err
}

func findGlobalClusterByID(ctx context.Context, conn *neptune.Client, id string) (*awstypes.GlobalCluster, error) {
	input := &neptune.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(id),
	}
	output, err := findGlobalCluster(ctx, conn, input, tfslices.PredicateTrue[awstypes.GlobalCluster]())

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == globalClusterStatusDeleted {
		return nil, &retry.NotFoundError{
			Message: status,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalClusterIdentifier) != id {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findGlobalClusterByClusterARN(ctx context.Context, conn *neptune.Client, arn string) (*awstypes.GlobalCluster, error) {
	input := &neptune.DescribeGlobalClustersInput{}

	return findGlobalCluster(ctx, conn, input, func(v awstypes.GlobalCluster) bool {
		return slices.ContainsFunc(v.GlobalClusterMembers, func(v awstypes.GlobalClusterMember) bool {
			return aws.ToString(v.DBClusterArn) == arn
		})
	})
}

func findGlobalCluster(ctx context.Context, conn *neptune.Client, input *neptune.DescribeGlobalClustersInput, filter tfslices.Predicate[awstypes.GlobalCluster]) (*awstypes.GlobalCluster, error) {
	output, err := findGlobalClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGlobalClusters(ctx context.Context, conn *neptune.Client, input *neptune.DescribeGlobalClustersInput, filter tfslices.Predicate[awstypes.GlobalCluster]) ([]awstypes.GlobalCluster, error) {
	var output []awstypes.GlobalCluster

	pages := neptune.NewDescribeGlobalClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusGlobalCluster(conn *neptune.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGlobalClusterByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitGlobalClusterCreated(ctx context.Context, conn *neptune.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusCreating},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterUpdated(ctx context.Context, conn *neptune.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusModifying, globalClusterStatusUpgrading},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterDeleted(ctx context.Context, conn *neptune.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{globalClusterStatusAvailable, globalClusterStatusDeleting},
		Target:         []string{},
		Refresh:        statusGlobalCluster(conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func flattenGlobalClusterMembers(apiObjects []awstypes.GlobalClusterMember) []any {
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
