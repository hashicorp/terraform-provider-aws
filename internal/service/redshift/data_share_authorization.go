// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_redshift_data_share_authorization", name="Data Share Authorization")
func newDataShareAuthorizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataShareAuthorizationResource{}, nil
}

const (
	dataShareAuthorizationIDPartCount = 2
)

type dataShareAuthorizationResource struct {
	framework.ResourceWithModel[dataShareAuthorizationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *dataShareAuthorizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allow_writes": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"consumer_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_share_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"managed_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"producer_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *dataShareAuthorizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dataShareAuthorizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().RedshiftClient(ctx)

	dataShareARN, consumerIdentifier := fwflex.StringValueFromFramework(ctx, plan.DataShareARN), fwflex.StringValueFromFramework(ctx, plan.ConsumerIdentifier)
	parts := []string{
		dataShareARN,
		consumerIdentifier,
	}
	id, err := intflex.FlattenResourceId(parts, dataShareAuthorizationIDPartCount, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewCreatingResourceIDErrorDiagnostic(err))
		return
	}

	in := redshift.AuthorizeDataShareInput{
		AllowWrites:        fwflex.BoolFromFramework(ctx, plan.AllowWrites),
		DataShareArn:       aws.String(dataShareARN),
		ConsumerIdentifier: aws.String(consumerIdentifier),
	}
	out, err := conn.AuthorizeDataShare(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("creating Redshift Data Share Authorization (%s)", id), err.Error())
		return
	}

	// Set values for unknowns.
	plan.ID = fwflex.StringValueToFramework(ctx, id)
	plan.ManagedBy = fwflex.StringToFramework(ctx, out.ManagedBy)
	plan.ProducerARN = fwflex.StringToFrameworkARN(ctx, out.ProducerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *dataShareAuthorizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataShareAuthorizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	parts, err := intflex.ExpandResourceId(id, dataShareAuthorizationIDPartCount, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	dataShareARN, consumerID := parts[0], parts[1]
	out, err := findDataShareAuthorizationByTwoPartKey(ctx, conn, dataShareARN, consumerID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading Redshift Data Share Authorization (%s)", id), err.Error())
		return
	}

	// Set attributes for import.
	state.ConsumerIdentifier = fwflex.StringValueToFramework(ctx, consumerID)
	state.DataShareARN = fwtypes.ARNValue(dataShareARN)
	state.ManagedBy = fwflex.StringToFramework(ctx, out.ManagedBy)
	state.ProducerARN = fwflex.StringToFrameworkARN(ctx, out.ProducerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataShareAuthorizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dataShareAuthorizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	parts, err := intflex.ExpandResourceId(id, dataShareAuthorizationIDPartCount, false)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	dataShareARN, consumerID := parts[0], parts[1]
	in := redshift.DeauthorizeDataShareInput{
		ConsumerIdentifier: aws.String(consumerID),
		DataShareArn:       aws.String(dataShareARN),
	}
	_, err = conn.DeauthorizeDataShare(ctx, &in)
	if errs.IsA[*awstypes.ResourceNotFoundFault](err) ||
		errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "because the ARN doesn't exist.") ||
		errs.IsAErrorMessageContains[*awstypes.InvalidDataShareFault](err, "because you have already removed authorization from the data share.") {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting Redshift Data Share Authorization (%s)", id), err.Error())
		return
	}
}

type dataShareAuthorizationResourceModel struct {
	framework.WithRegionModel
	AllowWrites        types.Bool   `tfsdk:"allow_writes"`
	ConsumerIdentifier types.String `tfsdk:"consumer_identifier"`
	DataShareARN       fwtypes.ARN  `tfsdk:"data_share_arn"`
	ID                 types.String `tfsdk:"id"`
	ManagedBy          types.String `tfsdk:"managed_by"`
	ProducerARN        fwtypes.ARN  `tfsdk:"producer_arn"`
}
