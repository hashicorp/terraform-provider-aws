// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iam_organizations_features", name="Organizations Features")
func newOrganizationsFeaturesResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &organizationsFeaturesResource{}

	return r, nil
}

type organizationsFeaturesResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *organizationsFeaturesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled_features": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.FeatureType]](ctx),
				Required:    true,
				ElementType: fwtypes.StringEnumType[awstypes.FeatureType](),
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *organizationsFeaturesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data organizationsFeaturesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	var enabledFeatures []awstypes.FeatureType
	response.Diagnostics.Append(fwflex.Expand(ctx, data.EnabledFeatures, &enabledFeatures)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := updateOrganizationFeatures(ctx, conn, enabledFeatures, []awstypes.FeatureType{}); err != nil {
		response.Diagnostics.AddError("creating IAM Organizations Features", err.Error())

		return
	}

	output, err := findOrganizationsFeatures(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading IAM Organizations Features", err.Error())

		return
	}

	// Set values for unknowns.
	data.OrganizationID = fwflex.StringToFramework(ctx, output.OrganizationId)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *organizationsFeaturesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data organizationsFeaturesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	output, err := findOrganizationsFeatures(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading IAM Organizations Features (%s)", data.OrganizationID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.EnabledFeatures, &data.EnabledFeatures)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *organizationsFeaturesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new organizationsFeaturesResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	var oldFeatures, newFeatures []awstypes.FeatureType
	response.Diagnostics.Append(fwflex.Expand(ctx, old.EnabledFeatures, &oldFeatures)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(fwflex.Expand(ctx, new.EnabledFeatures, &newFeatures)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	if err := updateOrganizationFeatures(ctx, conn, newFeatures, oldFeatures); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating IAM Organizations Features (%s)", new.OrganizationID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *organizationsFeaturesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data organizationsFeaturesResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	var enabledFeatures []awstypes.FeatureType
	response.Diagnostics.Append(fwflex.Expand(ctx, data.EnabledFeatures, &enabledFeatures)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := updateOrganizationFeatures(ctx, conn, []awstypes.FeatureType{}, enabledFeatures); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting IAM Organizations Features (%s)", data.OrganizationID.ValueString()), err.Error())

		return
	}
}

type organizationsFeaturesResourceModel struct {
	EnabledFeatures fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.FeatureType]] `tfsdk:"enabled_features"`
	OrganizationID  types.String                                                 `tfsdk:"id"`
}

func findOrganizationsFeatures(ctx context.Context, conn *iam.Client) (*iam.ListOrganizationsFeaturesOutput, error) {
	input := &iam.ListOrganizationsFeaturesInput{}

	output, err := conn.ListOrganizationsFeatures(ctx, input)

	if errs.IsA[*awstypes.OrganizationNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.EnabledFeatures) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func updateOrganizationFeatures(ctx context.Context, conn *iam.Client, new, old []awstypes.FeatureType) error {
	toEnable := itypes.Set[awstypes.FeatureType](new).Difference(old)
	toDisable := itypes.Set[awstypes.FeatureType](old).Difference(new)

	if slices.Contains(toEnable, awstypes.FeatureTypeRootCredentialsManagement) {
		input := &iam.EnableOrganizationsRootCredentialsManagementInput{}

		_, err := conn.EnableOrganizationsRootCredentialsManagement(ctx, input)

		if err != nil {
			return fmt.Errorf("enabling Organizations root credentials management: %w", err)
		}
	}

	if slices.Contains(toEnable, awstypes.FeatureTypeRootSessions) {
		input := &iam.EnableOrganizationsRootSessionsInput{}

		_, err := conn.EnableOrganizationsRootSessions(ctx, input)

		if err != nil {
			return fmt.Errorf("enabling Organizations root sessions: %w", err)
		}
	}

	if slices.Contains(toDisable, awstypes.FeatureTypeRootCredentialsManagement) {
		input := &iam.DisableOrganizationsRootCredentialsManagementInput{}

		_, err := conn.DisableOrganizationsRootCredentialsManagement(ctx, input)

		if err != nil {
			return fmt.Errorf("disabling Organizations root credentials management: %w", err)
		}
	}

	if slices.Contains(toDisable, awstypes.FeatureTypeRootSessions) {
		input := &iam.DisableOrganizationsRootSessionsInput{}

		_, err := conn.DisableOrganizationsRootSessions(ctx, input)

		if err != nil {
			return fmt.Errorf("disabling Organizations root sessions: %w", err)
		}
	}

	return nil
}
