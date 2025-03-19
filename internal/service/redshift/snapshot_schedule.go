// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_snapshot_schedule", name="Snapshot Schedule")
// @Tags(identifierAttribute="arn")
func resourceSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotScheduleCreate,
		ReadWithoutTimeout:   resourceSnapshotScheduleRead,
		UpdateWithoutTimeout: resourceSnapshotScheduleUpdate,
		DeleteWithoutTimeout: resourceSnapshotScheduleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definitions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
			},
			"identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrIdentifier},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSnapshotScheduleCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	identifier := create.Name(d.Get(names.AttrIdentifier).(string), d.Get("identifier_prefix").(string))
	input := &redshift.CreateSnapshotScheduleInput{
		ScheduleIdentifier:  aws.String(identifier),
		ScheduleDefinitions: flex.ExpandStringValueSet(d.Get("definitions").(*schema.Set)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ScheduleDescription = aws.String(v.(string))
	}

	output, err := conn.CreateSnapshotSchedule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Schedule (%s): %s", identifier, err)
	}

	d.SetId(aws.ToString(output.ScheduleIdentifier))

	return append(diags, resourceSnapshotScheduleRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	snapshotSchedule, err := findSnapshotScheduleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Schedule (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   names.Redshift,
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("snapshotschedule:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("definitions", snapshotSchedule.ScheduleDefinitions)
	d.Set(names.AttrDescription, snapshotSchedule.ScheduleDescription)
	d.Set(names.AttrIdentifier, snapshotSchedule.ScheduleIdentifier)
	d.Set("identifier_prefix", create.NamePrefixFromName(aws.ToString(snapshotSchedule.ScheduleIdentifier)))

	setTagsOut(ctx, snapshotSchedule.Tags)

	return diags
}

func resourceSnapshotScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChange("definitions") {
		input := &redshift.ModifySnapshotScheduleInput{
			ScheduleDefinitions: flex.ExpandStringValueSet(d.Get("definitions").(*schema.Set)),
			ScheduleIdentifier:  aws.String(d.Id()),
		}

		_, err := conn.ModifySnapshotSchedule(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Snapshot Schedule (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSnapshotScheduleRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.Get(names.AttrForceDestroy).(bool) {
		diags = append(diags, snapshotScheduleDisassociateAll(ctx, conn, d.Id())...)

		if diags.HasError() {
			return diags
		}
	}

	log.Printf("[DEBUG] Deleting Redshift Snapshot Schedule: %s", d.Id())
	_, err := conn.DeleteSnapshotSchedule(ctx, &redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

func snapshotScheduleDisassociateAll(ctx context.Context, conn *redshift.Client, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	snapshotSchedule, err := findSnapshotScheduleByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Schedule (%s): %s", id, err)
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		clusterIdentifier := aws.ToString(associatedCluster.ClusterIdentifier)
		_, err = conn.ModifyClusterSnapshotSchedule(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    aws.String(clusterIdentifier),
			ScheduleIdentifier:   aws.String(id),
			DisassociateSchedule: aws.Bool(true),
		})

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule Association (%s): %s", SnapshotScheduleAssociationCreateResourceID(clusterIdentifier, id), err)
		}
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		clusterIdentifier := aws.ToString(associatedCluster.ClusterIdentifier)
		if _, err := waitSnapshotScheduleAssociationDeleted(ctx, conn, clusterIdentifier, id); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Schedule Association (%s) delete: %s", SnapshotScheduleAssociationCreateResourceID(clusterIdentifier, id), err)
		}
	}

	return diags
}

func findSnapshotScheduleByID(ctx context.Context, conn *redshift.Client, id string) (*awstypes.SnapshotSchedule, error) {
	input := &redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(id),
	}

	output, err := conn.DescribeSnapshotSchedules(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.SnapshotSchedules)
}
