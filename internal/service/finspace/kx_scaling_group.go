// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_finspace_kx_scaling_group", name="Kx Scaling Group")
// @Tags(identifierAttribute="arn")
func ResourceKxScalingGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxScalingGroupCreate,
		ReadWithoutTimeout:   resourceKxScalingGroupRead,
		UpdateWithoutTimeout: resourceKxScalingGroupUpdate,
		DeleteWithoutTimeout: resourceKxScalingGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Hour),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(4 * time.Hour),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"host_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"clusters": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxScalingGroup     = "Kx Scaling Group"
	kxScalingGroupIDPartCount = 2
)

func resourceKxScalingGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	environmentId := d.Get("environment_id").(string)
	scalingGroupName := d.Get(names.AttrName).(string)
	idParts := []string{
		environmentId,
		scalingGroupName,
	}
	rID, err := flex.FlattenResourceId(idParts, kxScalingGroupIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxScalingGroup, d.Get(names.AttrName).(string), err)
	}
	d.SetId(rID)

	in := &finspace.CreateKxScalingGroupInput{
		EnvironmentId:      aws.String(environmentId),
		ScalingGroupName:   aws.String(scalingGroupName),
		HostType:           aws.String(d.Get("host_type").(string)),
		AvailabilityZoneId: aws.String(d.Get("availability_zone_id").(string)),
		Tags:               getTagsIn(ctx),
	}

	out, err := conn.CreateKxScalingGroup(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxScalingGroup, d.Get(names.AttrName).(string), err)
	}

	if out == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxScalingGroup, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	if _, err := waitKxScalingGroupCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxScalingGroup, d.Id(), err)
	}

	return append(diags, resourceKxScalingGroupRead(ctx, d, meta)...)
}

func resourceKxScalingGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := FindKxScalingGroupById(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxScalingGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, d.Id(), err)
	}
	d.Set(names.AttrARN, out.ScalingGroupArn)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrStatusReason, out.StatusReason)
	d.Set("created_timestamp", out.CreatedTimestamp.String())
	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())
	d.Set(names.AttrName, out.ScalingGroupName)
	d.Set("availability_zone_id", out.AvailabilityZoneId)
	d.Set("host_type", out.HostType)
	d.Set("clusters", out.Clusters)

	parts, err := flex.ExpandResourceId(d.Id(), kxUserIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxScalingGroup, d.Id(), err)
	}
	d.Set("environment_id", parts[0])

	return diags
}

func resourceKxScalingGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Tags only.
	return append(diags, resourceKxScalingGroupRead(ctx, d, meta)...)
}

func resourceKxScalingGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace KxScalingGroup %s", d.Id())
	_, err := conn.DeleteKxScalingGroup(ctx, &finspace.DeleteKxScalingGroupInput{
		ScalingGroupName: aws.String(d.Get(names.AttrName).(string)),
		EnvironmentId:    aws.String(d.Get("environment_id").(string)),
	})
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxScalingGroup, d.Id(), err)
	}

	_, err = waitKxScalingGroupDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil && !tfresource.NotFound(err) {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxScalingGroup, d.Id(), err)
	}

	return diags
}

func FindKxScalingGroupById(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxScalingGroupOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxScalingGroupIDPartCount, false)
	if err != nil {
		return nil, err
	}
	in := &finspace.GetKxScalingGroupInput{
		EnvironmentId:    aws.String(parts[0]),
		ScalingGroupName: aws.String(parts[1]),
	}

	out, err := conn.GetKxScalingGroup(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ScalingGroupName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out, nil
}

func waitKxScalingGroupCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxScalingGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxScalingGroupStatusCreating),
		Target:                    enum.Slice(types.KxScalingGroupStatusActive),
		Refresh:                   statusKxScalingGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxScalingGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxScalingGroupDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxScalingGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.KxScalingGroupStatusDeleting),
		Target:  enum.Slice(types.KxScalingGroupStatusDeleted),
		Refresh: statusKxScalingGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxScalingGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxScalingGroup(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindKxScalingGroupById(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
