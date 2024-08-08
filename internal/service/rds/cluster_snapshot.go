// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_cluster_snapshot", name="DB Cluster Snapshot")
// @Tags(identifierAttribute="db_cluster_snapshot_arn")
// @Testing(tagsTest=false)
func resourceClusterSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterSnapshotCreate,
		ReadWithoutTimeout:   resourceClusterSnapshotRead,
		DeleteWithoutTimeout: resourceClusterSnapshotDelete,
		UpdateWithoutTimeout: resourceClusterSnapshotUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"db_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_cluster_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexache.MustCompile(`^[a-z]`), "must begin with a lowercase letter"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"shared_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"source_db_cluster_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	id := d.Get("db_cluster_snapshot_identifier").(string)
	input := &rds.CreateDBClusterSnapshotInput{
		DBClusterIdentifier:         aws.String(d.Get("db_cluster_identifier").(string)),
		DBClusterSnapshotIdentifier: aws.String(id),
		Tags:                        getTagsInV2(ctx),
	}

	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.InvalidDBClusterStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.CreateDBClusterSnapshot(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Cluster Snapshot (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDBClusterSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Cluster Snapshot (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("shared_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input := &rds.ModifyDBClusterSnapshotAttributeInput{
			AttributeName:               aws.String(clusterSnapshotAttributeNameRestore),
			DBClusterSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:                 flex.ExpandStringValueSet(v.(*schema.Set)),
		}

		_, err := conn.ModifyDBClusterSnapshotAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Cluster Snapshot (%s) attribute: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	snapshot, err := findDBClusterSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Snapshot (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAllocatedStorage, snapshot.AllocatedStorage)
	d.Set(names.AttrAvailabilityZones, snapshot.AvailabilityZones)
	d.Set("db_cluster_identifier", snapshot.DBClusterIdentifier)
	d.Set("db_cluster_snapshot_arn", snapshot.DBClusterSnapshotArn)
	d.Set("db_cluster_snapshot_identifier", snapshot.DBClusterSnapshotIdentifier)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set(names.AttrPort, snapshot.Port)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_cluster_snapshot_arn", snapshot.SourceDBClusterSnapshotArn)
	d.Set(names.AttrStatus, snapshot.Status)
	d.Set(names.AttrStorageEncrypted, snapshot.StorageEncrypted)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	attribute, err := findDBClusterSnapshotAttributeByTwoPartKey(ctx, conn, d.Id(), clusterSnapshotAttributeNameRestore)
	switch {
	case err == nil:
		d.Set("shared_accounts", attribute.AttributeValues)
	case tfresource.NotFound(err):
	default:
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Snapshot (%s) attribute: %s", d.Id(), err)
	}

	setTagsOutV2(ctx, snapshot.TagList)

	return diags
}

func resourceClusterSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange("shared_accounts") {
		o, n := d.GetChange("shared_accounts")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)
		input := &rds.ModifyDBClusterSnapshotAttributeInput{
			AttributeName:               aws.String(clusterSnapshotAttributeNameRestore),
			DBClusterSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:                 flex.ExpandStringValueSet(add),
			ValuesToRemove:              flex.ExpandStringValueSet(del),
		}

		_, err := conn.ModifyDBClusterSnapshotAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Cluster Snapshot (%s) attribute: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterSnapshotRead(ctx, d, meta)...)
}

func resourceClusterSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Cluster Snapshot: %s", d.Id())
	_, err := conn.DeleteDBClusterSnapshot(ctx, &rds.DeleteDBClusterSnapshotInput{
		DBClusterSnapshotIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.DBClusterSnapshotNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Cluster Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

func findDBClusterSnapshotByID(ctx context.Context, conn *rds.Client, id string) (*types.DBClusterSnapshot, error) {
	input := &rds.DescribeDBClusterSnapshotsInput{
		DBClusterSnapshotIdentifier: aws.String(id),
	}
	output, err := findDBClusterSnapshot(ctx, conn, input, tfslices.PredicateTrue[*types.DBClusterSnapshot]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterSnapshotIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBClusterSnapshot(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterSnapshotsInput, filter tfslices.Predicate[*types.DBClusterSnapshot]) (*types.DBClusterSnapshot, error) {
	output, err := findDBClusterSnapshots(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterSnapshots(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterSnapshotsInput, filter tfslices.Predicate[*types.DBClusterSnapshot]) ([]types.DBClusterSnapshot, error) {
	var output []types.DBClusterSnapshot

	pages := rds.NewDescribeDBClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBClusterSnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterSnapshots {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBClusterSnapshot(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBClusterSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDBClusterSnapshotCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBClusterSnapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterSnapshotStatusCreating},
		Target:     []string{clusterSnapshotStatusAvailable},
		Refresh:    statusDBClusterSnapshot(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBClusterSnapshot); ok {
		return output, err
	}

	return nil, err
}

func findDBClusterSnapshotAttributeByTwoPartKey(ctx context.Context, conn *rds.Client, id, attributeName string) (*types.DBClusterSnapshotAttribute, error) {
	input := &rds.DescribeDBClusterSnapshotAttributesInput{
		DBClusterSnapshotIdentifier: aws.String(id),
	}

	return findDBClusterSnapshotAttribute(ctx, conn, input, func(v *types.DBClusterSnapshotAttribute) bool {
		return aws.ToString(v.AttributeName) == attributeName
	})
}

func findDBClusterSnapshotAttribute(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterSnapshotAttributesInput, filter tfslices.Predicate[*types.DBClusterSnapshotAttribute]) (*types.DBClusterSnapshotAttribute, error) {
	output, err := findDBClusterSnapshotAttributes(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterSnapshotAttributes(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterSnapshotAttributesInput, filter tfslices.Predicate[*types.DBClusterSnapshotAttribute]) ([]types.DBClusterSnapshotAttribute, error) {
	output, err := conn.DescribeDBClusterSnapshotAttributes(ctx, input)

	if errs.IsA[*types.DBClusterSnapshotNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DBClusterSnapshotAttributesResult == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.DBClusterSnapshotAttributesResult.DBClusterSnapshotAttributes, tfslices.PredicateValue(filter)), nil
}
