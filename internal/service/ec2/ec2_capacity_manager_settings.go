// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_capacity_manager_settings", name="Capacity Manager Settings")
// @SingletonIdentity
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
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.Enabled = types.BoolValue(output.CapacityManagerStatus == awstypes.CapacityManagerStatusEnabled)
	data.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))

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
		if _, err := conn.UpdateCapacityManagerOrganizationsAccess(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
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
	if _, err := conn.DisableCapacityManager(ctx, &input); err != nil {
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
		if err != nil {
			return err
		}
		data.OrganizationsAccess = types.BoolValue(aws.ToBool(output.OrganizationsAccess))
	}

	return nil
}

type capacityManagerSettingsResourceModel struct {
	framework.WithRegionModel
	Enabled             types.Bool `tfsdk:"enabled"`
	OrganizationsAccess types.Bool `tfsdk:"organizations_access"`
}
