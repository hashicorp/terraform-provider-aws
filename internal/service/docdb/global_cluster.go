// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"
	"reflect"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_global_cluster")
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
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
				ValidateFunc:  validation.StringInSlice(engine_Values(), false),
				AtLeastOneOf:  []string{names.AttrEngine, "source_db_cluster_identifier"},
				ConflictsWith: []string{"source_db_cluster_identifier"},
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				AtLeastOneOf:  []string{names.AttrEngine, "source_db_cluster_identifier"},
				ConflictsWith: []string{names.AttrEngine},
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

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	globalClusterID := d.Get("global_cluster_identifier").(string)
	input := &docdb.CreateGlobalClusterInput{
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

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_db_cluster_identifier"); ok {
		input.SourceDBClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrStorageEncrypted); ok {
		input.StorageEncrypted = aws.Bool(v.(bool))
	}

	output, err := conn.CreateGlobalCluster(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.ToString(output.GlobalCluster.GlobalClusterIdentifier))

	if _, err := waitGlobalClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Global Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	globalCluster, err := findGlobalClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Global Cluster (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, globalCluster.GlobalClusterArn)
	d.Set(names.AttrDatabaseName, globalCluster.DatabaseName)
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

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	if d.HasChange(names.AttrDeletionProtection) {
		input := &docdb.ModifyGlobalClusterInput{
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			GlobalClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.ModifyGlobalCluster(ctx, input)

		if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DocumentDB Global Cluster: %s", err)
		}

		if _, err := waitGlobalClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Global Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrEngineVersion) {
		engineVersion := d.Get(names.AttrEngineVersion).(string)

		for _, tfMapRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			if clusterARN, ok := tfMap["db_cluster_arn"].(string); ok && clusterARN != "" {
				cluster, err := findClusterByARN(ctx, conn, clusterARN)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster (%s): %s", clusterARN, err)
				}

				clusterID := aws.ToString(cluster.DBClusterIdentifier)
				input := &docdb.ModifyDBClusterInput{
					ApplyImmediately:    aws.Bool(true),
					DBClusterIdentifier: aws.String(clusterID),
					EngineVersion:       aws.String(engineVersion),
				}

				_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
					return conn.ModifyDBCluster(ctx, input)
				}, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions")

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster (%s) engine version: %s", clusterID, err)
				}

				if _, err := waitDBClusterAvailable(ctx, conn, clusterID, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster (%s) update: %s", clusterID, err)
				}
			}
		}
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	// Remove any members from the global cluster.
	for _, tfMapRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if clusterARN, ok := tfMap["db_cluster_arn"].(string); ok && clusterARN != "" {
			if err := removeClusterFromGlobalCluster(ctx, conn, clusterARN, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	log.Printf("[DEBUG] Deleting DocumentDB Global Cluster: %s", d.Id())

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidGlobalClusterStateFault](ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteGlobalCluster(ctx, &docdb.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: aws.String(d.Id()),
		})
	}, "is not empty")

	if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Global Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitGlobalClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Global Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGlobalClusterByID(ctx context.Context, conn *docdb.Client, id string) (*awstypes.GlobalCluster, error) {
	input := &docdb.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(id),
	}
	output, err := findGlobalCluster(ctx, conn, input, tfslices.PredicateTrue[awstypes.GlobalCluster]())

	if err != nil {
		return nil, err
	}

	if status := aws.ToString(output.Status); status == globalClusterStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.GlobalClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findGlobalClusterByClusterARN(ctx context.Context, conn *docdb.Client, arn string) (*awstypes.GlobalCluster, error) {
	input := &docdb.DescribeGlobalClustersInput{}

	return findGlobalCluster(ctx, conn, input, func(v awstypes.GlobalCluster) bool {
		return slices.ContainsFunc(v.GlobalClusterMembers, func(v awstypes.GlobalClusterMember) bool {
			return aws.ToString(v.DBClusterArn) == arn
		})
	})
}

func findGlobalCluster(ctx context.Context, conn *docdb.Client, input *docdb.DescribeGlobalClustersInput, filter tfslices.Predicate[awstypes.GlobalCluster]) (*awstypes.GlobalCluster, error) {
	output, err := findGlobalClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGlobalClusters(ctx context.Context, conn *docdb.Client, input *docdb.DescribeGlobalClustersInput, filter tfslices.Predicate[awstypes.GlobalCluster]) ([]awstypes.GlobalCluster, error) {
	var output []awstypes.GlobalCluster

	pages := docdb.NewDescribeGlobalClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.GlobalClusterNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			if !reflect.ValueOf(v).IsZero() && filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusGlobalCluster(ctx context.Context, conn *docdb.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
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

func waitGlobalClusterCreated(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusCreating},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterUpdated(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusModifying, globalClusterStatusUpgrading},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterDeleted(ctx context.Context, conn *docdb.Client, id string, timeout time.Duration) (*awstypes.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{globalClusterStatusAvailable, globalClusterStatusDeleting},
		Target:         []string{},
		Refresh:        statusGlobalCluster(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func flattenGlobalClusterMembers(apiObjects []awstypes.GlobalClusterMember) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"db_cluster_arn": aws.ToString(apiObject.DBClusterArn),
			"is_writer":      aws.ToBool(apiObject.IsWriter),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
