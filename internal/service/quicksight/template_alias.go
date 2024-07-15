// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Template Alias")
func newResourceTemplateAlias(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTemplateAlias{}, nil
}

const (
	ResNameTemplateAlias = "Template Alias"
)

type resourceTemplateAlias struct {
	framework.ResourceWithConfigure
}

func (r *resourceTemplateAlias) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_quicksight_template_alias"
}

func (r *resourceTemplateAlias) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alias_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"template_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_version_number": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

func (r *resourceTemplateAlias) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan resourceTemplateAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(
		createTemplateAliasID(plan.AWSAccountID.ValueString(), plan.TemplateID.ValueString(), plan.AliasName.ValueString()))

	in := &quicksight.CreateTemplateAliasInput{
		AliasName:             aws.String(plan.AliasName.ValueString()),
		AwsAccountId:          aws.String(plan.AWSAccountID.ValueString()),
		TemplateId:            aws.String(plan.TemplateID.ValueString()),
		TemplateVersionNumber: aws.Int64(plan.TemplateVersionNumber.ValueInt64()),
	}

	out, err := conn.CreateTemplateAliasWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameTemplateAlias, plan.AliasName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.TemplateAlias == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameTemplateAlias, plan.AliasName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.TemplateAlias.Arn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTemplateAlias) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceTemplateAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindTemplateAliasByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AliasName = flex.StringToFramework(ctx, out.AliasName)
	state.TemplateVersionNumber = flex.Int64ToFramework(ctx, out.TemplateVersionNumber)

	// To support import, parse the ID for the component keys and set
	// individual values in state
	awsAccountID, templateID, _, err := ParseTemplateAliasID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.TemplateID = flex.StringValueToFramework(ctx, templateID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTemplateAlias) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var plan, state resourceTemplateAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TemplateVersionNumber.Equal(state.TemplateVersionNumber) {
		in := &quicksight.UpdateTemplateAliasInput{
			AliasName:             aws.String(plan.AliasName.ValueString()),
			AwsAccountId:          aws.String(plan.AWSAccountID.ValueString()),
			TemplateId:            aws.String(plan.TemplateID.ValueString()),
			TemplateVersionNumber: aws.Int64(plan.TemplateVersionNumber.ValueInt64()),
		}

		out, err := conn.UpdateTemplateAliasWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameTemplateAlias, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.TemplateAlias == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, ResNameTemplateAlias, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.TemplateAlias.Arn)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTemplateAlias) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightConn(ctx)

	var state resourceTemplateAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &quicksight.DeleteTemplateAliasInput{
		AliasName:    aws.String(state.AliasName.ValueString()),
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		TemplateId:   aws.String(state.TemplateID.ValueString()),
	}

	_, err := conn.DeleteTemplateAliasWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTemplateAlias) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindTemplateAliasByID(ctx context.Context, conn *quicksight.QuickSight, id string) (*quicksight.TemplateAlias, error) {
	awsAccountID, templateID, aliasName, err := ParseTemplateAliasID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeTemplateAliasInput{
		AliasName:    aws.String(aliasName),
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	}

	out, err := conn.DescribeTemplateAliasWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.TemplateAlias == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.TemplateAlias, nil
}

func ParseTemplateAliasID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,TEMPLATE_ID,ALIAS_NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func createTemplateAliasID(awsAccountID, templateID, aliasName string) string {
	return strings.Join([]string{awsAccountID, templateID, aliasName}, ",")
}

type resourceTemplateAliasData struct {
	AliasName             types.String `tfsdk:"alias_name"`
	ARN                   types.String `tfsdk:"arn"`
	AWSAccountID          types.String `tfsdk:"aws_account_id"`
	ID                    types.String `tfsdk:"id"`
	TemplateID            types.String `tfsdk:"template_id"`
	TemplateVersionNumber types.Int64  `tfsdk:"template_version_number"`
}
