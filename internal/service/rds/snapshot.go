// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// @SDKResource("aws_db_snapshot", name="DB Snapshot")
// @Tags(identifierAttribute="db_snapshot_arn")
// @Testing(tagsTest=false)
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCreate,
		ReadWithoutTimeout:   resourceSnapshotRead,
		UpdateWithoutTimeout: resourceSnapshotUpdate,
		DeleteWithoutTimeout: resourceSnapshotDelete,

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
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIOPS: {
				Type:     schema.TypeInt,
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
			"option_group_name": {
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
			"source_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_region": {
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
			names.AttrStorageType: {
				Type:     schema.TypeString,
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

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbSnapshotID := d.Get("db_snapshot_identifier").(string)
	input := &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: aws.String(d.Get("db_instance_identifier").(string)),
		DBSnapshotIdentifier: aws.String(dbSnapshotID),
		Tags:                 getTagsInV2(ctx),
	}

	output, err := conn.CreateDBSnapshot(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Snapshot (%s): %s", dbSnapshotID, err)
	}

	d.SetId(aws.ToString(output.DBSnapshot.DBSnapshotIdentifier))

	if _, err := waitDBSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Snapshot (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("shared_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input := &rds.ModifyDBSnapshotAttributeInput{
			AttributeName:        aws.String("restore"),
			DBSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:          flex.ExpandStringValueSet(v.(*schema.Set)),
		}

		_, err := conn.ModifyDBSnapshotAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Snapshot (%s) attribute: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSnapshotRead(ctx, d, meta)...)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	snapshot, err := findDBSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Snapshot (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(snapshot.DBSnapshotArn)
	d.Set(names.AttrAllocatedStorage, snapshot.AllocatedStorage)
	d.Set(names.AttrAvailabilityZone, snapshot.AvailabilityZone)
	d.Set("db_instance_identifier", snapshot.DBInstanceIdentifier)
	d.Set("db_snapshot_arn", arn)
	d.Set("db_snapshot_identifier", snapshot.DBSnapshotIdentifier)
	d.Set(names.AttrEncrypted, snapshot.Encrypted)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrIOPS, snapshot.Iops)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("option_group_name", snapshot.OptionGroupName)
	d.Set(names.AttrPort, snapshot.Port)
	d.Set("source_db_snapshot_identifier", snapshot.SourceDBSnapshotIdentifier)
	d.Set("source_region", snapshot.SourceRegion)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set(names.AttrStatus, snapshot.Status)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	attribute, err := findDBSnapshotAttributeByTwoPartKey(ctx, conn, d.Id(), dbSnapshotAttributeNameRestore)
	switch {
	case err == nil:
		d.Set("shared_accounts", attribute.AttributeValues)
	case tfresource.NotFound(err):
	default:
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Snapshot (%s) attribute: %s", d.Id(), err)
	}

	setTagsOutV2(ctx, snapshot.TagList)

	return diags
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange("shared_accounts") {
		o, n := d.GetChange("shared_accounts")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)
		input := &rds.ModifyDBSnapshotAttributeInput{
			AttributeName:        aws.String("restore"),
			DBSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:          flex.ExpandStringValueSet(add),
			ValuesToRemove:       flex.ExpandStringValueSet(del),
		}

		_, err := conn.ModifyDBSnapshotAttribute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Snapshot (%s) attribute: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Snapshot: %s", d.Id())
	_, err := conn.DeleteDBSnapshot(ctx, &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*types.DBSnapshotNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Snapshot (%s): %s", d.Id(), err)
	}

	return diags
}

func findDBSnapshotByID(ctx context.Context, conn *rds.Client, id string) (*types.DBSnapshot, error) {
	input := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(id),
	}
	output, err := findDBSnapshot(ctx, conn, input, tfslices.PredicateTrue[*types.DBSnapshot]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBSnapshotIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBSnapshot(ctx context.Context, conn *rds.Client, input *rds.DescribeDBSnapshotsInput, filter tfslices.Predicate[*types.DBSnapshot]) (*types.DBSnapshot, error) {
	output, err := findDBSnapshots(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBSnapshots(ctx context.Context, conn *rds.Client, input *rds.DescribeDBSnapshotsInput, filter tfslices.Predicate[*types.DBSnapshot]) ([]types.DBSnapshot, error) {
	var output []types.DBSnapshot

	pages := rds.NewDescribeDBSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBSnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBSnapshots {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBSnapshot(ctx context.Context, conn *rds.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBSnapshotByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status), nil
	}
}

func waitDBSnapshotCreated(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*types.DBSnapshot, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{dbSnapshotCreating},
		Target:     []string{dbSnapshotAvailable},
		Refresh:    statusDBSnapshot(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBSnapshot); ok {
		tfresource.SetLastError(err, fmt.Errorf("%d%% progress", aws.ToInt32(output.PercentProgress)))
		return output, err
	}

	return nil, err
}

func findDBSnapshotAttributeByTwoPartKey(ctx context.Context, conn *rds.Client, id, attributeName string) (*types.DBSnapshotAttribute, error) {
	input := &rds.DescribeDBSnapshotAttributesInput{
		DBSnapshotIdentifier: aws.String(id),
	}

	return findDBSnapshotAttribute(ctx, conn, input, func(v *types.DBSnapshotAttribute) bool {
		return aws.ToString(v.AttributeName) == attributeName
	})
}

func findDBSnapshotAttribute(ctx context.Context, conn *rds.Client, input *rds.DescribeDBSnapshotAttributesInput, filter tfslices.Predicate[*types.DBSnapshotAttribute]) (*types.DBSnapshotAttribute, error) {
	output, err := findDBSnapshotAttributes(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBSnapshotAttributes(ctx context.Context, conn *rds.Client, input *rds.DescribeDBSnapshotAttributesInput, filter tfslices.Predicate[*types.DBSnapshotAttribute]) ([]types.DBSnapshotAttribute, error) {
	output, err := conn.DescribeDBSnapshotAttributes(ctx, input)

	if errs.IsA[*types.DBSnapshotNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DBSnapshotAttributesResult == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.DBSnapshotAttributesResult.DBSnapshotAttributes, tfslices.PredicateValue(filter)), nil
}
