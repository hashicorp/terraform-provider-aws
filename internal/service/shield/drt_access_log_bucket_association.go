package shield

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/shield/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="DRT Access Log Bucket Association")
func newResourceDRTAccessLogBucketAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDRTAccessLogBucketAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDRTAccessLogBucketAssociation = "DRT Access Log Bucket Association"
)

type resourceDRTAccessLogBucketAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDRTAccessLogBucketAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_drt_access_log_bucket_association"
}

func (r *resourceDRTAccessLogBucketAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // required by hashicorps terraform plugin testing framework
				DeprecationMessage:  "id is only for framework compatability and not used by the provider",
				MarkdownDescription: "The ID of the directory.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"log_bucket": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					// Validate string value length must be at least 3 characters and max 63.
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"role_arn_association_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Read:   true,
			}),
		},
	}
}

func (r *resourceDRTAccessLogBucketAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan resourceDRTAccessLogBucketAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.AssociateDRTLogBucketInput{
		LogBucket: aws.String(plan.LogBucket.ValueString()),
	}
	out, err := conn.AssociateDRTLogBucketWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDRTAccessLogBucketAssociationCreated(ctx, conn, plan.LogBucket.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForCreation, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(plan.LogBucket.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDRTAccessLogBucketAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceDRTAccessLogBucketAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &shield.DescribeDRTAccessInput{}

	out, err := conn.DescribeDRTAccessWithContext(ctx, in)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameDRTAccessLogBucketAssociation, state.LogBucket.String(), err),
			err.Error(),
		)
		return
	}

	associatedLogBucket := getAssociatedLogBucket(state.LogBucket.ValueString(), out.LogBucketList)
	if out != nil && len(out.LogBucketList) > 0 && associatedLogBucket == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameDRTAccessLogBucketAssociation, state.LogBucket.String(), nil),
			errors.New("Log Bucket not in list").Error(),
		)
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Setting ID of state - required by hashicorps terraform plugin testing framework for Import. See issue https://github.com/hashicorp/terraform-plugin-testing/issues/84
		state.ID = types.StringValue(state.LogBucket.ValueString())
	}
	state.LogBucket = flex.StringToFramework(ctx, associatedLogBucket)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func getAssociatedLogBucket(bucket string, bucketList []*string) *string {
	for _, bkt := range bucketList {

		if *bkt == bucket {
			return bkt
		}
	}
	return nil
}

func (r *resourceDRTAccessLogBucketAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceDRTAccessLogBucketAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.LogBucket.Equal(state.LogBucket) {
		in := &shield.AssociateDRTLogBucketInput{
			LogBucket: aws.String(plan.LogBucket.ValueString()),
		}
		out, err := conn.AssociateDRTLogBucketWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDRTAccessLogBucketAssociationUpdated(ctx, conn, plan.LogBucket.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameDRTAccessLogBucketAssociation, plan.LogBucket.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDRTAccessLogBucketAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceDRTAccessLogBucketAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return

	}
	if state.LogBucket.ValueString() == "" {
		return
	}

	in := &shield.DisassociateDRTLogBucketInput{
		LogBucket: aws.String(state.LogBucket.ValueString()),
	}

	_, err := conn.DisassociateDRTLogBucketWithContext(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameDRTAccessLogBucketAssociation, state.LogBucket.String(), err),
			err.Error(),
		)
		return
	}
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDRTAccessLogBucketAssociationDeleted(ctx, conn, state.LogBucket.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForDeletion, ResNameDRTAccessLogBucketAssociation, state.LogBucket.String(), err),
			err.Error(),
		)
		return
	}
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitDRTAccessLogBucketAssociationCreated(ctx context.Context, conn *shield.Shield, bucket string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusDRTAccessLogBucketAssociation(ctx, conn, bucket),
		Timeout:                   timeout,
		NotFoundChecks:            2,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessLogBucketAssociationUpdated(ctx context.Context, conn *shield.Shield, bucket string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusDRTAccessLogBucketAssociation(ctx, conn, bucket),
		Timeout:                   timeout,
		NotFoundChecks:            2,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessLogBucketAssociationDeleted(ctx context.Context, conn *shield.Shield, bucket string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusDRTAccessLogBucketAssociation(ctx, conn, bucket),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDRTAccessLogBucketAssociation(ctx context.Context, conn *shield.Shield, bucket string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeDRTAccessLogBucketAssociation(ctx, conn, bucket)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		if out == nil || out.LogBucketList == nil || len(out.LogBucketList) == 0 {
			return nil, "", nil
		}

		if out != nil && len(out.LogBucketList) > 0 {
			for _, bkt := range out.LogBucketList {
				if aws.ToString(bkt) == bucket {
					return out, statusNormal, nil
				}
			}
			return nil, "", nil
		}

		return out, statusNormal, nil
	}
}

func describeDRTAccessLogBucketAssociation(ctx context.Context, conn *shield.Shield, bucketName string) (*shield.DescribeDRTAccessOutput, error) {
	in := &shield.DescribeDRTAccessInput{}

	out, err := conn.DescribeDRTAccessWithContext(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
	}

	if out == nil || out.LogBucketList == nil || len(out.LogBucketList) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	for _, bucket := range out.LogBucketList {
		if aws.ToString(bucket) == bucketName {
			return out, nil
		}
	}
	return nil, err
}

type resourceDRTAccessLogBucketAssociationData struct {
	ID                   types.String   `tfsdk:"id"`
	RoleArnAssociationID types.String   `tfsdk:"role_arn_association_id"`
	LogBucket            types.String   `tfsdk:"log_bucket"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
}
