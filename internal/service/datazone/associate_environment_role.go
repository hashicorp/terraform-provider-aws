// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_associate_environment_role", name="Associate Environment Role")
// @Tags(identifierAttribute="arn")
func newAssociateEnvironmentRoleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &associateEnvironmentRoleResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type associateEnvironmentRoleResource struct {
	framework.ResourceWithModel[associateEnvironmentRoleResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *associateEnvironmentRoleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"domain_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^dzd[-_][a-zA-Z0-9_-]{1,36}$`), "must conform to: ^dzd[-_][a-zA-Z0-9_-]{1,36}$ "),
				},
			},
			"environment_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9_-]{1,36}$"), "must match ^[a-zA-Z0-9_-]{1,36}$"),
				},
			},
			"environment_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *associateEnvironmentRoleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan associateEnvironmentRoleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	var input datazone.AssociateEnvironmentRoleInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	const (
		timeout = 30 * time.Second
	)
	_, err := tfresource.RetryWhenAWSErrCodeContains(ctx, timeout, func(ctx context.Context) (*datazone.AssociateEnvironmentRoleOutput, error) {
		return conn.AssociateEnvironmentRole(ctx, &input)
	}, errCodeAccessDenied)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Associate Environment Role with internal ID (%s)", plan.ID.ValueString()), err.Error())

		return
	}
	plan.ID = types.StringValue(
		fmt.Sprintf(
			"%s|%s|%s",
			plan.DomainIdentifier.ValueString(),
			plan.EnvironmentIdentifier.ValueString(),
			plan.EnvironmentRoleArn.ValueString(),
		),
	)

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *associateEnvironmentRoleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state associateEnvironmentRoleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	_, err := conn.GetEnvironment(ctx, &datazone.GetEnvironmentInput{
		DomainIdentifier: aws.String(state.DomainIdentifier.ValueString()),
		Identifier:       aws.String(state.EnvironmentIdentifier.ValueString()),
	})

	if retry.NotFound(err) {
		state.ID = types.StringNull()
		return
	}
	if state.DomainIdentifier.ValueString() != "" && err != nil {
		response.Diagnostics.AddError("GetEnvironment failed", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *associateEnvironmentRoleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data associateEnvironmentRoleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)
	input := datazone.DisassociateEnvironmentRoleInput{
		DomainIdentifier:      aws.String(data.DomainIdentifier.ValueString()),
		EnvironmentIdentifier: aws.String(data.EnvironmentIdentifier.ValueString()),
		EnvironmentRoleArn:    aws.String(data.EnvironmentRoleArn.ValueString()),
	}
	_, err := conn.DisassociateEnvironmentRole(ctx, &input)

	if isResourceMissing(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Disassociate Environment Role with internal ID (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type associateEnvironmentRoleResourceModel struct {
	framework.WithRegionModel
	ID                    types.String   `tfsdk:"id"`
	DomainIdentifier      types.String   `tfsdk:"domain_identifier"`
	EnvironmentIdentifier types.String   `tfsdk:"environment_identifier"`
	EnvironmentRoleArn    fwtypes.ARN    `tfsdk:"environment_role_arn"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}
