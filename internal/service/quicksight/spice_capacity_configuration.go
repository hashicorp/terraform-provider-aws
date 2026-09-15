// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_spice_capacity_configuration", name="SPICE Capacity Configuration")
func newSPICECapacityConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &spiceCapacityConfigurationResource{}

	return r, nil
}

type spiceCapacityConfigurationResource struct {
	framework.ResourceWithModel[spiceCapacityConfigurationResourceModel]
}

func (r *spiceCapacityConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			"purchase_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PurchaseMode](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.PurchaseModeManual)),
			},
		},
	}
}

func (r *spiceCapacityConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data spiceCapacityConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	var input quicksight.UpdateSPICECapacityConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateSPICECapacityConfiguration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating QuickSight SPICE Capacity Configuration (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

// Read is a no-op. QuickSight provides no API to describe the SPICE capacity
// configuration (purchase mode), so the configured value is retained in state
// and drift cannot be detected.
func (r *spiceCapacityConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data spiceCapacityConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *spiceCapacityConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data spiceCapacityConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	var input quicksight.UpdateSPICECapacityConfigurationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateSPICECapacityConfiguration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating QuickSight SPICE Capacity Configuration (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Delete resets the purchase mode to MANUAL, the QuickSight default. There is no
// API to delete the SPICE capacity configuration itself.
func (r *spiceCapacityConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data spiceCapacityConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	input := quicksight.UpdateSPICECapacityConfigurationInput{
		AwsAccountId: aws.String(accountID),
		PurchaseMode: awstypes.PurchaseModeManual,
	}
	_, err := conn.UpdateSPICECapacityConfiguration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting QuickSight SPICE Capacity Configuration (%s)", accountID), err.Error())

		return
	}
}

func (r *spiceCapacityConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrAWSAccountID), request, response)
}

type spiceCapacityConfigurationResourceModel struct {
	framework.WithRegionModel
	AWSAccountID types.String                              `tfsdk:"aws_account_id"`
	PurchaseMode fwtypes.StringEnum[awstypes.PurchaseMode] `tfsdk:"purchase_mode"`
}
