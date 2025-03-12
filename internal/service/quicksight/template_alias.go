// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_template_alias", name="Template Alias")
func newTemplateAliasResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &templateAliasResource{}, nil
}

const (
	resNameTemplateAlias = "Template Alias"
)

type templateAliasResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *templateAliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r *templateAliasResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceTemplateAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, templateID, aliasName := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.TemplateID), flex.StringValueFromFramework(ctx, plan.AliasName)
	in := &quicksight.CreateTemplateAliasInput{
		AliasName:             aws.String(aliasName),
		AwsAccountId:          aws.String(awsAccountID),
		TemplateId:            aws.String(templateID),
		TemplateVersionNumber: plan.TemplateVersionNumber.ValueInt64Pointer(),
	}

	out, err := conn.CreateTemplateAlias(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameTemplateAlias, plan.AliasName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.TemplateAlias == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameTemplateAlias, plan.AliasName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = types.StringValue(templateAliasCreateResourceID(awsAccountID, templateID, aliasName))
	plan.ARN = flex.StringToFramework(ctx, out.TemplateAlias.Arn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *templateAliasResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceTemplateAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, templateID, aliasName, err := templateAliasParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	out, err := findTemplateAliasByThreePartKey(ctx, conn, awsAccountID, templateID, aliasName)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, resNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AliasName = flex.StringToFramework(ctx, out.AliasName)
	state.TemplateVersionNumber = flex.Int64ToFramework(ctx, out.TemplateVersionNumber)
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.TemplateID = flex.StringValueToFramework(ctx, templateID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateAliasResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan, state resourceTemplateAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, templateID, aliasName, err := templateAliasParseResourceID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameTemplateAlias, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	if !plan.TemplateVersionNumber.Equal(state.TemplateVersionNumber) {
		in := &quicksight.UpdateTemplateAliasInput{
			AliasName:             aws.String(aliasName),
			AwsAccountId:          aws.String(awsAccountID),
			TemplateId:            aws.String(templateID),
			TemplateVersionNumber: plan.TemplateVersionNumber.ValueInt64Pointer(),
		}

		out, err := conn.UpdateTemplateAlias(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameTemplateAlias, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.TemplateAlias == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QuickSight, create.ErrActionUpdating, resNameTemplateAlias, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.TemplateAlias.Arn)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateAliasResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceTemplateAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, templateID, aliasName, err := templateAliasParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteTemplateAlias(ctx, &quicksight.DeleteTemplateAliasInput{
		AliasName:    aws.String(aliasName),
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameTemplateAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findTemplateAliasByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, templateID, aliasName string) (*awstypes.TemplateAlias, error) {
	input := &quicksight.DescribeTemplateAliasInput{
		AliasName:    aws.String(aliasName),
		AwsAccountId: aws.String(awsAccountID),
		TemplateId:   aws.String(templateID),
	}

	return findTemplateAlias(ctx, conn, input)
}

func findTemplateAlias(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeTemplateAliasInput) (*awstypes.TemplateAlias, error) {
	output, err := conn.DescribeTemplateAlias(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TemplateAlias == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TemplateAlias, nil
}

const templateAliasResourceIDSeparator = ","

func templateAliasCreateResourceID(awsAccountID, templateID, aliasName string) string {
	parts := []string{awsAccountID, templateID, aliasName}
	id := strings.Join(parts, templateAliasResourceIDSeparator)

	return id
}

func templateAliasParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, templateAliasResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sTEMPLATE_ID%[2]sALIAS_NAME", id, templateAliasResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

type resourceTemplateAliasData struct {
	AliasName             types.String `tfsdk:"alias_name"`
	ARN                   types.String `tfsdk:"arn"`
	AWSAccountID          types.String `tfsdk:"aws_account_id"`
	ID                    types.String `tfsdk:"id"`
	TemplateID            types.String `tfsdk:"template_id"`
	TemplateVersionNumber types.Int64  `tfsdk:"template_version_number"`
}
