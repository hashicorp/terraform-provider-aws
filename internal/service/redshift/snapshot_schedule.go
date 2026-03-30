// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_snapshot_schedule", name="Snapshot Schedule")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/redshift/types;awstypes;awstypes.SnapshotSchedule")
// @Testing(importIgnore="force_destroy")
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

	identifier := create.Name(ctx, d.Get(names.AttrIdentifier).(string), d.Get("identifier_prefix").(string))
	input := redshift.CreateSnapshotScheduleInput{
		ScheduleIdentifier:  aws.String(identifier),
		ScheduleDefinitions: flex.ExpandStringValueSet(d.Get("definitions").(*schema.Set)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.ScheduleDescription = aws.String(v.(string))
	}

	output, err := conn.CreateSnapshotSchedule(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Snapshot Schedule (%s): %s", identifier, err)
	}

	d.SetId(aws.ToString(output.ScheduleIdentifier))

	return append(diags, resourceSnapshotScheduleRead(ctx, d, meta)...)
}

func resourceSnapshotScheduleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	snapshotSchedule, err := findSnapshotScheduleByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Schedule (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, snapshotScheduleARN(ctx, c, d.Id()))
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
		input := redshift.ModifySnapshotScheduleInput{
			ScheduleDefinitions: flex.ExpandStringValueSet(d.Get("definitions").(*schema.Set)),
			ScheduleIdentifier:  aws.String(d.Id()),
		}

		_, err := conn.ModifySnapshotSchedule(ctx, &input)

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
		diags = append(diags, disassociateAllSnaphotSchedules(ctx, conn, d.Id())...)

		if diags.HasError() {
			return diags
		}
	}

	log.Printf("[DEBUG] Deleting Redshift Snapshot Schedule: %s", d.Id())
	input := redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteSnapshotSchedule(ctx, &input)

	if errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule (%s): %s", d.Id(), err)
	}

	return diags
}

func disassociateAllSnaphotSchedules(ctx context.Context, conn *redshift.Client, id string) diag.Diagnostics {
	var diags diag.Diagnostics

	snapshotSchedule, err := findSnapshotScheduleByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Snapshot Schedule (%s): %s", id, err)
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		clusterIdentifier := aws.ToString(associatedCluster.ClusterIdentifier)
		_, err = conn.ModifyClusterSnapshotSchedule(ctx, &redshift.ModifyClusterSnapshotScheduleInput{
			DisassociateSchedule: aws.Bool(true),
			ClusterIdentifier:    aws.String(clusterIdentifier),
			ScheduleIdentifier:   aws.String(id),
		})

		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.SnapshotScheduleNotFoundFault](err) {
			continue
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Redshift Snapshot Schedule Association (%s): %s", snapshotScheduleAssociationCreateResourceID(clusterIdentifier, id), err)
		}
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		clusterIdentifier := aws.ToString(associatedCluster.ClusterIdentifier)
		if _, err := waitSnapshotScheduleAssociationDeleted(ctx, conn, clusterIdentifier, id); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Snapshot Schedule Association (%s) delete: %s", snapshotScheduleAssociationCreateResourceID(clusterIdentifier, id), err)
		}
	}

	return diags
}

func snapshotScheduleARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "snapshotschedule:"+id)
}
