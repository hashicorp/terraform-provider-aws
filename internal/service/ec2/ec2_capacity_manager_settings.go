// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_capacity_manager_settings", name="Capacity Manager Settings")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(generator=false)
func newCapacityManagerSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &capacityManagerSettingsResource{}, nil
}

type capacityManagerSettingsResource struct {
	framework.ResourceWithModel[capacityManagerSettingsResourceModel]
	framework.WithImportByIdentity
}

func (r *capacityManagerSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			"organizations_access": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func (r *capacityManagerSettingsResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data capacityManagerSettingsResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.Enabled.IsUnknown() || data.OrganizationsAccess.IsUnknown() {
		return
	}

	if !data.Enabled.ValueBool() && data.OrganizationsAccess.ValueBool() {
		response.Diagnostics.AddAttributeError(
			path.Root("organizations_access"),
			"Invalid Attribute Combination",
			"organizations_access cannot be true when enabled is false",
		)
	}
}

func (r *capacityManagerSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data capacityManagerSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if err := updateCapacityManagerSettings(ctx, conn, &data); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *capacityManagerSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data capacityManagerSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findCapacityManagerAttributes(ctx, conn)
	switch {
	case retry.NotFound(err):
		// Capacity Manager reports an error when it is disabled.
		data.Enabled = types.BoolValue(false)
		data.OrganizationsAccess = types.BoolValue(false)
	case err != nil:
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	default:
		data.Enabled = types.BoolValue(output.CapacityManagerStatus == awstypes.CapacityManagerStatusEnabled)
		data.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *capacityManagerSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old capacityManagerSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	// Organizations access can only change independently while Capacity Manager stays enabled; disabling resets it to false server-side.
	if new.Enabled.Equal(old.Enabled) && new.Enabled.ValueBool() && !new.OrganizationsAccess.Equal(old.OrganizationsAccess) {
		input := ec2.UpdateCapacityManagerOrganizationsAccessInput{
			OrganizationsAccess: new.OrganizationsAccess.ValueBoolPointer(),
		}
		output, err := conn.UpdateCapacityManagerOrganizationsAccess(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
		new.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))
	} else if !new.Enabled.Equal(old.Enabled) {
		if err := updateCapacityManagerSettings(ctx, conn, &new); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *capacityManagerSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	// Removing the resource disables EC2 Capacity Manager.
	var input ec2.DisableCapacityManagerInput
	if _, err := conn.DisableCapacityManager(ctx, &input); err != nil && !tfawserr.ErrCodeEquals(err, errCodeCapacityManagerDisabled) {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
}

// updateCapacityManagerSettings enables or disables EC2 Capacity Manager to match data.
func updateCapacityManagerSettings(ctx context.Context, conn *ec2.Client, data *capacityManagerSettingsResourceModel) error {
	if data.Enabled.ValueBool() {
		input := ec2.EnableCapacityManagerInput{
			OrganizationsAccess: data.OrganizationsAccess.ValueBoolPointer(),
		}
		output, err := conn.EnableCapacityManager(ctx, &input)
		if err != nil {
			return err
		}
		data.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))
	} else {
		var input ec2.DisableCapacityManagerInput
		output, err := conn.DisableCapacityManager(ctx, &input)
		switch {
		case tfawserr.ErrCodeEquals(err, errCodeCapacityManagerDisabled):
			// Already disabled.
			data.OrganizationsAccess = types.BoolValue(false)
		case err != nil:
			return err
		default:
			data.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))
		}
	}

	return nil
}

type capacityManagerSettingsResourceModel struct {
	framework.WithRegionModel
	Enabled             types.Bool   `tfsdk:"enabled"`
	ID                  types.String `tfsdk:"id"`
	OrganizationsAccess types.Bool   `tfsdk:"organizations_access"`
}
