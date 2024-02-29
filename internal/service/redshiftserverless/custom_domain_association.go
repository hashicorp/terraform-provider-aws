// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Custom Domain Association")
func newResourceCustomDomainAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCustomDomainAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCustomDomainAssociation = "Custom Domain Association"
)

type resourceCustomDomainAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCustomDomainAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_redshiftserverless_custom_domain_association"
}

func (r *resourceCustomDomainAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"workgroup_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_domain_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
				},
			},
			"custom_domain_certificate_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"custom_domain_certificate_expiry_time": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceCustomDomainAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan resourceCustomDomainAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshiftserverless.CreateCustomDomainAssociationInput{
		CustomDomainCertificateArn: aws.String(plan.CustomDomainCertificateArn.ValueString()),
		CustomDomainName:           aws.String(plan.CustomDomainName.ValueString()),
		WorkgroupName:              aws.String(plan.WorkgroupName.ValueString()),
	}

	out, err := conn.CreateCustomDomainAssociation(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.RedshiftServerless,
				create.ErrActionCreating,
				ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", plan.WorkgroupName.String(), plan.CustomDomainName.String()),
				err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CustomDomainCertificateExpiryTime == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.RedshiftServerless,
				create.ErrActionCreating,
				ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", plan.WorkgroupName.String(), plan.CustomDomainName.String()),
				nil),
			errors.New("empty output").Error(),
		)
		return
	}

	formattedTime := out.CustomDomainCertificateExpiryTime.Format(time.RFC3339)
	plan.CustomDomainCertificateExpiryTime = flex.StringToFramework(ctx, &formattedTime)
	plan.ID = types.StringValue(fmt.Sprintf("%s,%s", plan.WorkgroupName, plan.CustomDomainName))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceCustomDomainAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state resourceCustomDomainAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCustomDomainAssociationByWorkgroupNameAndDomainName(ctx, conn, state.WorkgroupName.ValueString(), state.CustomDomainName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.RedshiftServerless,
				create.ErrActionSetting,
				ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", state.WorkgroupName.String(), state.CustomDomainName.String()),
				err),
			err.Error(),
		)
		return
	}

	state.CustomDomainCertificateArn = flex.StringToFrameworkARN(ctx, out.CustomDomainCertificateArn)
	formattedTime := out.CustomDomainCertificateExpiryTime.Format(time.RFC3339)
	state.CustomDomainCertificateExpiryTime = flex.StringToFramework(ctx, &formattedTime)
	state.CustomDomainName = flex.StringToFramework(ctx, out.CustomDomainName)
	state.WorkgroupName = flex.StringToFramework(ctx, out.WorkgroupName)
	state.ID = types.StringValue(fmt.Sprintf("%s,%s", state.WorkgroupName, state.CustomDomainName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCustomDomainAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var plan, state resourceCustomDomainAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.CustomDomainCertificateArn.Equal(state.CustomDomainCertificateArn) {
		in := &redshiftserverless.UpdateCustomDomainAssociationInput{
			CustomDomainCertificateArn: aws.String(plan.CustomDomainCertificateArn.ValueString()),
			CustomDomainName:           aws.String(plan.CustomDomainName.ValueString()),
			WorkgroupName:              aws.String(plan.WorkgroupName.ValueString()),
		}

		out, err := conn.UpdateCustomDomainAssociation(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(
					names.RedshiftServerless,
					create.ErrActionUpdating,
					ResNameCustomDomainAssociation,
					fmt.Sprintf("%s,%s", plan.WorkgroupName.String(), plan.CustomDomainName.String()), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.CustomDomainCertificateExpiryTime == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(
					names.RedshiftServerless,
					create.ErrActionUpdating,
					ResNameCustomDomainAssociation,
					fmt.Sprintf("%s,%s", plan.WorkgroupName.String(), plan.CustomDomainName.String()),
					nil),
				errors.New("empty output").Error(),
			)
			return
		}

		formattedTime := out.CustomDomainCertificateExpiryTime.Format(time.RFC3339)
		plan.CustomDomainCertificateExpiryTime = flex.StringToFramework(ctx, &formattedTime)
		plan.ID = types.StringValue(fmt.Sprintf("%s,%s", plan.WorkgroupName, plan.CustomDomainName))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCustomDomainAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RedshiftServerlessClient(ctx)

	var state resourceCustomDomainAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &redshiftserverless.DeleteCustomDomainAssociationInput{
		CustomDomainName: aws.String(state.CustomDomainName.ValueString()),
		WorkgroupName:    aws.String(state.WorkgroupName.ValueString()),
	}

	_, err := conn.DeleteCustomDomainAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.RedshiftServerless,
				create.ErrActionDeleting,
				ResNameCustomDomainAssociation,
				fmt.Sprintf("%s,%s", state.WorkgroupName.String(), state.CustomDomainName.String()),
				err),
			err.Error(),
		)
		return
	}
}

func (r *resourceCustomDomainAssociation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: workgroup_name,custom_domain_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workgroup_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("custom_domain_name"), idParts[1])...)
}

func findCustomDomainAssociationByWorkgroupNameAndDomainName(ctx context.Context, conn *redshiftserverless.Client, workgroupName string, customDomainName string) (*redshiftserverless.GetCustomDomainAssociationOutput, error) {
	in := &redshiftserverless.GetCustomDomainAssociationInput{
		WorkgroupName:    aws.String(workgroupName),
		CustomDomainName: aws.String(customDomainName),
	}

	out, err := conn.GetCustomDomainAssociation(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil || out.CustomDomainCertificateExpiryTime == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceCustomDomainAssociationData struct {
	ID                                types.String `tfsdk:"id"`
	WorkgroupName                     types.String `tfsdk:"workgroup_name"`
	CustomDomainName                  types.String `tfsdk:"custom_domain_name"`
	CustomDomainCertificateArn        fwtypes.ARN  `tfsdk:"custom_domain_certificate_arn"`
	CustomDomainCertificateExpiryTime types.String `tfsdk:"custom_domain_certificate_expiry_time"`
}
