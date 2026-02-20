// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_networkmanager_prefix_list_association", name="Prefix List Association")
// @IdentityAttribute("core_network_id")
// @IdentityAttribute("prefix_list_arn")
// @ImportIDHandler("prefixListAssociationImportID")
// @Testing(hasNoPreExistingResource=true)
func newPrefixListAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &prefixListAssociationResource{}, nil
}

const (
	prefixListAssociationAvailableTimeout = 20 * time.Minute
)

type prefixListAssociationResource struct {
	framework.ResourceWithModel[prefixListAssociationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

type prefixListAssociationResourceModel struct {
	CoreNetworkID   types.String `tfsdk:"core_network_id"`
	PrefixListAlias types.String `tfsdk:"prefix_list_alias"`
	PrefixListARN   fwtypes.ARN  `tfsdk:"prefix_list_arn"`
}

func (r *prefixListAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"core_network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix_list_alias": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix_list_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *prefixListAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan prefixListAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID := fwflex.StringValueFromFramework(ctx, plan.CoreNetworkID)
	prefixListARN := fwflex.StringValueFromFramework(ctx, plan.PrefixListARN)
	input := networkmanager.CreateCoreNetworkPrefixListAssociationInput{
		ClientToken:     aws.String(sdkid.UniqueId()),
		CoreNetworkId:   aws.String(coreNetworkID),
		PrefixListAlias: fwflex.StringFromFramework(ctx, plan.PrefixListAlias),
		PrefixListArn:   aws.String(prefixListARN),
	}
	_, err := conn.CreateCoreNetworkPrefixListAssociation(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating Network Manager Prefix List Association (%s/%s)", coreNetworkID, prefixListARN),
			err.Error(),
		)
		return
	}

	if err := waitCoreNetworkAvailableAfterPrefixListChange(ctx, conn, coreNetworkID, prefixListAssociationAvailableTimeout); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("waiting for Network Manager Core Network (%s) to become available after creating prefix list association", coreNetworkID),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *prefixListAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state prefixListAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID := fwflex.StringValueFromFramework(ctx, state.CoreNetworkID)
	prefixListARN := fwflex.StringValueFromFramework(ctx, state.PrefixListARN)
	output, err := findPrefixListAssociationByTwoPartKey(ctx, conn, coreNetworkID, prefixListARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading Network Manager Prefix List Association (%s/%s)", coreNetworkID, prefixListARN),
			err.Error(),
		)
		return
	}

	state.PrefixListAlias = fwflex.StringToFramework(ctx, output.PrefixListAlias)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *prefixListAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state prefixListAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkManagerClient(ctx)

	coreNetworkID := fwflex.StringValueFromFramework(ctx, state.CoreNetworkID)
	prefixListARN := fwflex.StringValueFromFramework(ctx, state.PrefixListARN)
	input := networkmanager.DeleteCoreNetworkPrefixListAssociationInput{
		CoreNetworkId: aws.String(coreNetworkID),
		PrefixListArn: aws.String(prefixListARN),
	}
	_, err := conn.DeleteCoreNetworkPrefixListAssociation(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting Network Manager Prefix List Association (%s/%s)", coreNetworkID, prefixListARN),
			err.Error(),
		)
		return
	}

	if err := waitCoreNetworkAvailableAfterPrefixListChange(ctx, conn, coreNetworkID, prefixListAssociationAvailableTimeout); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("waiting for Network Manager Core Network (%s) to become available after deleting prefix list association", coreNetworkID),
			err.Error(),
		)
		return
	}
}

var (
	_ inttypes.ImportIDParser = prefixListAssociationImportID{}
)

type prefixListAssociationImportID struct{}

const (
	prefixListAssociationIDParts = 2
)

func (prefixListAssociationImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, prefixListAssociationIDParts, true)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"core_network_id": parts[0],
		"prefix_list_arn": parts[1],
	}

	return id, result, nil
}

// Finder functions.

func findPrefixListAssociationByTwoPartKey(ctx context.Context, conn *networkmanager.Client, coreNetworkID, prefixListARN string) (*awstypes.PrefixListAssociation, error) {
	input := networkmanager.ListCoreNetworkPrefixListAssociationsInput{
		CoreNetworkId: aws.String(coreNetworkID),
		PrefixListArn: aws.String(prefixListARN),
	}
	output, err := findPrefixListAssociation(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findPrefixListAssociation(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListCoreNetworkPrefixListAssociationsInput) (*awstypes.PrefixListAssociation, error) {
	output, err := findPrefixListAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrefixListAssociations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListCoreNetworkPrefixListAssociationsInput) ([]awstypes.PrefixListAssociation, error) {
	var output []awstypes.PrefixListAssociation

	pages := networkmanager.NewListCoreNetworkPrefixListAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PrefixListAssociations...)
	}

	return output, nil
}

// Waiter for core network to become available after prefix list association changes.

func waitCoreNetworkAvailableAfterPrefixListChange(ctx context.Context, conn *networkmanager.Client, coreNetworkID string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CoreNetworkStateUpdating, coreNetworkStatePending),
		Target:  enum.Slice(awstypes.CoreNetworkStateAvailable),
		Timeout: timeout,
		Refresh: statusCoreNetworkState(conn, coreNetworkID),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
