// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package datazone

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_datazone_environment_blueprint_configuration", name="Environment Blueprint Configuration")
func newEnvironmentBlueprintConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &environmentBlueprintConfigurationResource{}

	return r, nil
}

type environmentBlueprintConfigurationResource struct {
	framework.ResourceWithModel[environmentBlueprintConfigurationResourceModel]
}

func (r *environmentBlueprintConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled_regions": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			"environment_blueprint_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"manage_access_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"provisioning_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"regional_parameters": schema.MapAttribute{
				CustomType: fwtypes.MapOfMapOfStringType,
				Optional:   true,
			},
		},
	}
}

func (r *environmentBlueprintConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentBlueprintConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	domainID, environmentBlueprintID := fwflex.StringValueFromFramework(ctx, data.DomainIdentifier), fwflex.StringValueFromFramework(ctx, data.EnvironmentBlueprintIdentifier)
	var input datazone.PutEnvironmentBlueprintConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutEnvironmentBlueprintConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating DataZone Environment Blueprint Configuration (%s/%s)", domainID, environmentBlueprintID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *environmentBlueprintConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data environmentBlueprintConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	domainID, environmentBlueprintID := fwflex.StringValueFromFramework(ctx, data.DomainIdentifier), fwflex.StringValueFromFramework(ctx, data.EnvironmentBlueprintIdentifier)
	output, err := findEnvironmentBlueprintConfigurationByTwoPartKey(ctx, conn, domainID, environmentBlueprintID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading DataZone Environment Blueprint Configuration (%s/%s)", domainID, environmentBlueprintID), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.DomainIdentifier = fwflex.StringToFramework(ctx, output.DomainId)
	data.EnvironmentBlueprintIdentifier = fwflex.StringToFramework(ctx, output.EnvironmentBlueprintId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *environmentBlueprintConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data environmentBlueprintConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	domainID, environmentBlueprintID := fwflex.StringValueFromFramework(ctx, data.DomainIdentifier), fwflex.StringValueFromFramework(ctx, data.EnvironmentBlueprintIdentifier)
	var input datazone.PutEnvironmentBlueprintConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.PutEnvironmentBlueprintConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating DataZone Environment Blueprint Configuration (%s/%s)", domainID, environmentBlueprintID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *environmentBlueprintConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentBlueprintConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	domainID, environmentBlueprintID := fwflex.StringValueFromFramework(ctx, data.DomainIdentifier), fwflex.StringValueFromFramework(ctx, data.EnvironmentBlueprintIdentifier)
	input := datazone.DeleteEnvironmentBlueprintConfigurationInput{
		DomainIdentifier:               aws.String(domainID),
		EnvironmentBlueprintIdentifier: aws.String(environmentBlueprintID),
	}

	_, err := conn.DeleteEnvironmentBlueprintConfiguration(ctx, &input)

	if isResourceMissing(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting DataZone Environment Blueprint Configuration (%s/%s)", domainID, environmentBlueprintID), err.Error())

		return
	}
}

func (r *environmentBlueprintConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		environmentBlueprintConfigurationIDParts     = 2
		environmentBlueprintConfigurationIDSeparator = "/"
	)
	parts := strings.Split(request.ID, environmentBlueprintConfigurationIDSeparator)
	if len(parts) != environmentBlueprintConfigurationIDParts {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected DOMAIN-ID%[2]sENVIRONMENT-BLUEPRINT-ID", request.ID, environmentBlueprintConfigurationIDSeparator)
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("domain_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("environment_blueprint_id"), parts[1])...)
}

func findEnvironmentBlueprintConfigurationByTwoPartKey(ctx context.Context, conn *datazone.Client, domainID, environmentBlueprintID string) (*datazone.GetEnvironmentBlueprintConfigurationOutput, error) {
	input := datazone.GetEnvironmentBlueprintConfigurationInput{
		DomainIdentifier:               aws.String(domainID),
		EnvironmentBlueprintIdentifier: aws.String(environmentBlueprintID),
	}
	output, err := conn.GetEnvironmentBlueprintConfiguration(ctx, &input)

	if isResourceMissing(err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type environmentBlueprintConfigurationResourceModel struct {
	framework.WithRegionModel
	DomainIdentifier               types.String             `tfsdk:"domain_id"`
	EnabledRegions                 fwtypes.ListOfString     `tfsdk:"enabled_regions"`
	EnvironmentBlueprintIdentifier types.String             `tfsdk:"environment_blueprint_id"`
	ManageAccessRoleARN            fwtypes.ARN              `tfsdk:"manage_access_role_arn"`
	ProvisioningRoleARN            fwtypes.ARN              `tfsdk:"provisioning_role_arn"`
	RegionalParameters             fwtypes.MapOfMapOfString `tfsdk:"regional_parameters"`
}
