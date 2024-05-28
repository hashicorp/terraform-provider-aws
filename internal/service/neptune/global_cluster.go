// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
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

// @SDKResource("aws_neptune_global_cluster")
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

func resourceGlobalClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

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

	output, err := conn.CreateGlobalClusterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Global Cluster (%s): %s", globalClusterID, err)
	}

	d.SetId(aws.StringValue(output.GlobalCluster.GlobalClusterIdentifier))

	if _, err := waitGlobalClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Global Cluster (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	globalCluster, err := FindGlobalClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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

func resourceGlobalClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	if d.HasChange(names.AttrDeletionProtection) {
		input := &neptune.ModifyGlobalClusterInput{
			DeletionProtection:      aws.Bool(d.Get(names.AttrDeletionProtection).(bool)),
			GlobalClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.ModifyGlobalClusterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeGlobalClusterNotFoundFault) {
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
		engineVersion := d.Get(names.AttrEngineVersion).(string)

		for _, tfMapRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			if clusterARN, ok := tfMap["db_cluster_arn"].(string); ok && clusterARN != "" {
				cluster, err := findClusterByARN(ctx, conn, clusterARN)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster (%s): %s", clusterARN, err)
				}

				clusterID := aws.StringValue(cluster.DBClusterIdentifier)
				input := &neptune.ModifyDBClusterInput{
					ApplyImmediately:    aws.Bool(true),
					DBClusterIdentifier: aws.String(clusterID),
					EngineVersion:       aws.String(engineVersion),
				}

				_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
					return conn.ModifyDBClusterWithContext(ctx, input)
				}, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions")

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster (%s) engine version: %s", clusterID, err)
				}

				if _, err := waitDBClusterAvailable(ctx, conn, clusterID, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster (%s) update: %s", clusterID, err)
				}
			}
		}
	}

	return append(diags, resourceGlobalClusterRead(ctx, d, meta)...)
}

func resourceGlobalClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

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

	log.Printf("[DEBUG] Deleting Neptune Global Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteGlobalClusterWithContext(ctx, &neptune.DeleteGlobalClusterInput{
			GlobalClusterIdentifier: aws.String(d.Id()),
		})
	}, neptune.ErrCodeInvalidGlobalClusterStateFault, "is not empty")

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeGlobalClusterNotFoundFault) {
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

func FindGlobalClusterByID(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.GlobalCluster, error) {
	input := &neptune.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(id),
	}
	output, err := findGlobalCluster(ctx, conn, input, tfslices.PredicateTrue[*neptune.GlobalCluster]())

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == globalClusterStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.GlobalClusterIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findGlobalClusterByClusterARN(ctx context.Context, conn *neptune.Neptune, arn string) (*neptune.GlobalCluster, error) {
	input := &neptune.DescribeGlobalClustersInput{}

	return findGlobalCluster(ctx, conn, input, func(v *neptune.GlobalCluster) bool {
		return slices.ContainsFunc(v.GlobalClusterMembers, func(v *neptune.GlobalClusterMember) bool {
			return aws.StringValue(v.DBClusterArn) == arn
		})
	})
}

func findGlobalCluster(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeGlobalClustersInput, filter tfslices.Predicate[*neptune.GlobalCluster]) (*neptune.GlobalCluster, error) {
	output, err := findGlobalClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findGlobalClusters(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeGlobalClustersInput, filter tfslices.Predicate[*neptune.GlobalCluster]) ([]*neptune.GlobalCluster, error) {
	var output []*neptune.GlobalCluster

	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *neptune.DescribeGlobalClustersOutput, lastPage bool) bool {
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

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeGlobalClusterNotFoundFault) {
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

func statusGlobalCluster(ctx context.Context, conn *neptune.Neptune, id string) retry.StateRefreshFunc {
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

func waitGlobalClusterCreated(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusCreating},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterUpdated(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{globalClusterStatusModifying, globalClusterStatusUpgrading},
		Target:  []string{globalClusterStatusAvailable},
		Refresh: statusGlobalCluster(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func waitGlobalClusterDeleted(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.GlobalCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{globalClusterStatusAvailable, globalClusterStatusDeleting},
		Target:         []string{},
		Refresh:        statusGlobalCluster(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.GlobalCluster); ok {
		return output, err
	}

	return nil, err
}

func flattenGlobalClusterMembers(apiObjects []*neptune.GlobalClusterMember) []interface{} {
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
