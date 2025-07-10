// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_target_group_attachment", name="Target Group Attachment")
// @Testing(tagsTest=false)
func resourceTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetGroupAttachmentCreate,
		ReadWithoutTimeout:   resourceTargetGroupAttachmentRead,
		DeleteWithoutTimeout: resourceTargetGroupAttachmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrTarget: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						names.AttrPort: {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"target_group_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTargetGroupAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	targetGroupID := d.Get("target_group_identifier").(string)
	target := expandTarget(d.Get(names.AttrTarget).([]any)[0].(map[string]any))
	targetID := aws.ToString(target.Id)
	targetPort := aws.ToInt32(target.Port)
	id := targetGroupAttachmentCreateResourceID(targetGroupID, targetID, targetPort)
	input := vpclattice.RegisterTargetsInput{
		TargetGroupIdentifier: aws.String(targetGroupID),
		Targets:               []types.Target{target},
	}

	_, err := conn.RegisterTargets(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPC Lattice Target Group Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTargetGroupAttachmentCreated(ctx, conn, targetGroupID, targetID, targetPort, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Lattice Target Group Attachment (%s) create: %s", id, err)
	}

	return append(diags, resourceTargetGroupAttachmentRead(ctx, d, meta)...)
}

func resourceTargetGroupAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	// Can't use targetGroupAttachmentParseResourceID as targetID ALB ARN includes '/'.
	targetGroupID := d.Get("target_group_identifier").(string)
	target := expandTarget(d.Get(names.AttrTarget).([]any)[0].(map[string]any))
	targetID := aws.ToString(target.Id)
	targetPort := aws.ToInt32(target.Port)
	output, err := findTargetByThreePartKey(ctx, conn, targetGroupID, targetID, targetPort)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Lattice Target Group Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Lattice Target Group Attachment (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTarget, []any{flattenTargetSummary(output)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target: %s", err)
	}
	d.Set("target_group_identifier", targetGroupID)

	return diags
}

func resourceTargetGroupAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	log.Printf("[INFO] Deleting VPC Lattice Target Group Attachment: %s", d.Id())
	targetGroupID := d.Get("target_group_identifier").(string)
	target := expandTarget(d.Get(names.AttrTarget).([]any)[0].(map[string]any))
	targetID := aws.ToString(target.Id)
	targetPort := aws.ToInt32(target.Port)
	input := vpclattice.DeregisterTargetsInput{
		TargetGroupIdentifier: aws.String(targetGroupID),
		Targets: []types.Target{{
			Id: aws.String(targetID),
		}},
	}
	if targetPort > 0 {
		input.Targets[0].Port = aws.Int32(targetPort)
	}
	_, err := conn.DeregisterTargets(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPC Lattice Target Group Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTargetGroupAttachmentDeleted(ctx, conn, targetGroupID, targetID, targetPort, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Lattice Target Group Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const targetGroupAttachmentResourceIDSeparator = "/"

func targetGroupAttachmentCreateResourceID(targetGroupID, targetID string, targetPort int32) string {
	parts := []string{targetGroupID, targetID, flex.Int32ValueToStringValue(targetPort)}
	id := strings.Join(parts, targetGroupAttachmentResourceIDSeparator)

	return id
}

func findTargetByThreePartKey(ctx context.Context, conn *vpclattice.Client, targetGroupID, targetID string, targetPort int32) (*types.TargetSummary, error) {
	input := vpclattice.ListTargetsInput{
		TargetGroupIdentifier: aws.String(targetGroupID),
		Targets: []types.Target{{
			Id: aws.String(targetID),
		}},
	}
	if targetPort > 0 {
		input.Targets[0].Port = aws.Int32(targetPort)
	}

	return findTarget(ctx, conn, &input)
}

func findTarget(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListTargetsInput) (*types.TargetSummary, error) {
	output, err := findTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTargets(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListTargetsInput) ([]types.TargetSummary, error) {
	var output []types.TargetSummary

	paginator := vpclattice.NewListTargetsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Items...)
	}

	return output, nil
}

func statusTarget(ctx context.Context, conn *vpclattice.Client, targetGroupID, targetID string, targetPort int32) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findTargetByThreePartKey(ctx, conn, targetGroupID, targetID, targetPort)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTargetGroupAttachmentCreated(ctx context.Context, conn *vpclattice.Client, targetGroupID, targetID string, targetPort int32, timeout time.Duration) (*types.TargetSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.TargetStatusInitial),
		Target:                    enum.Slice(types.TargetStatusHealthy, types.TargetStatusUnhealthy, types.TargetStatusUnused, types.TargetStatusUnavailable),
		Refresh:                   statusTarget(ctx, conn, targetGroupID, targetID, targetPort),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TargetSummary); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ReasonCode)))

		return output, err
	}

	return nil, err
}

func waitTargetGroupAttachmentDeleted(ctx context.Context, conn *vpclattice.Client, targetGroupID, targetID string, targetPort int32, timeout time.Duration) (*types.TargetSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TargetStatusDraining, types.TargetStatusInitial),
		Target:  []string{},
		Refresh: statusTarget(ctx, conn, targetGroupID, targetID, targetPort),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TargetSummary); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ReasonCode)))

		return output, err
	}

	return nil, err
}

func flattenTargetSummary(apiObject *types.TargetSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	if v := apiObject.Port; v != nil {
		tfMap[names.AttrPort] = aws.ToInt32(v)
	}

	return tfMap
}

func expandTarget(tfMap map[string]any) types.Target {
	apiObject := types.Target{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
		apiObject.Port = aws.Int32(int32(v))
	}

	return apiObject
}
