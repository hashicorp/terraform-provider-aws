// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_ip_routes", name="IP Routes")
// @IdentityAttribute("directory_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="directory_id")
// @Testing(importIgnore="update_security_group_for_directory_controllers")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/directoryservice/types;awstypes;awstypes.IpRouteInfo")
func newIPRoutesResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ipRoutesResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type ipRoutesResource struct {
	framework.ResourceWithModel[ipRoutesResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *ipRoutesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"directory_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					directoryIDValidator,
				},
			},
			"update_security_group_for_directory_controllers": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					// Only consumed by AddIpRoutes; not returned by the API, so changes
					// require the routes to be re-created with the new value.
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"ip_route": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[ipRouteModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cidr_ip": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.IPv4CIDRNetworkAddress(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *ipRoutesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ipRoutesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := fwflex.StringValueFromFramework(ctx, data.DirectoryID)
	var input directoryservice.AddIpRoutesInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AddIpRoutes(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	if err := waitIPRoutesAdded(ctx, conn, directoryID, ipRoutesCIDRs(input.IpRoutes), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	// No attribute is populated by the API, so the plan is the authoritative
	// desired state. Any out-of-band routes are surfaced as drift on the next
	// refresh rather than merged in here (which would produce an inconsistent
	// result for the configured "ip_route" set).
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRoutesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ipRoutesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := fwflex.StringValueFromFramework(ctx, data.DirectoryID)
	output, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRoutesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state ipRoutesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)
	directoryID := fwflex.StringValueFromFramework(ctx, plan.DirectoryID)

	planRoutes, diags := plan.IPRoutes.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}
	stateRoutes, diags := state.IPRoutes.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	// A route is (re-)added when its CIDR is new or its description changed;
	// there is no update API, so a changed description is a remove + add.
	planByCIDR := make(map[string]*ipRouteModel, len(planRoutes))
	for _, v := range planRoutes {
		planByCIDR[v.CidrIP.ValueString()] = v
	}
	stateByCIDR := make(map[string]*ipRouteModel, len(stateRoutes))
	for _, v := range stateRoutes {
		stateByCIDR[v.CidrIP.ValueString()] = v
	}

	var add []awstypes.IpRoute
	var removeCIDRs []string
	for cidr, pr := range planByCIDR {
		if sr, ok := stateByCIDR[cidr]; !ok || !sr.Description.Equal(pr.Description) {
			add = append(add, awstypes.IpRoute{
				CidrIp:      aws.String(cidr),
				Description: fwflex.StringFromFramework(ctx, pr.Description),
			})
		}
	}
	for cidr := range stateByCIDR {
		if pr, ok := planByCIDR[cidr]; !ok || !pr.Description.Equal(stateByCIDR[cidr].Description) {
			removeCIDRs = append(removeCIDRs, cidr)
		}
	}

	if len(removeCIDRs) > 0 {
		_, err := conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     removeCIDRs,
		})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
			return
		}
		if err := waitIPRoutesRemoved(ctx, conn, directoryID, removeCIDRs, r.DeleteTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
			return
		}
	}

	if len(add) > 0 {
		_, err := conn.AddIpRoutes(ctx, &directoryservice.AddIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			IpRoutes:    add,
			UpdateSecurityGroupForDirectoryControllers: plan.UpdateSecurityGroupForDirectoryControllers.ValueBool(),
		})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
			return
		}
		if err := waitIPRoutesAdded(ctx, conn, directoryID, ipRoutesCIDRs(add), r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *ipRoutesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ipRoutesResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)
	directoryID := fwflex.StringValueFromFramework(ctx, data.DirectoryID)

	routes, diags := data.IPRoutes.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	cidrs := make([]string, 0, len(routes))
	for _, v := range routes {
		cidrs = append(cidrs, v.CidrIP.ValueString())
	}

	if len(cidrs) == 0 {
		return
	}

	_, err := conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
		DirectoryId: aws.String(directoryID),
		CidrIps:     cidrs,
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	if err := waitIPRoutesRemoved(ctx, conn, directoryID, cidrs, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}
}

func (r *ipRoutesResource) flatten(ctx context.Context, routes []awstypes.IpRouteInfo, data *ipRoutesResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, routes, &data.IPRoutes)...)
	return diags
}

func ipRoutesCIDRs(routes []awstypes.IpRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, v := range routes {
		cidrs = append(cidrs, aws.ToString(v.CidrIp))
	}
	return cidrs
}

// findIPRoutes returns all IPv4 IP routes for a directory regardless of status.
// It is used by the waiters, which must observe in-progress ("Adding"/"Removing")
// routes. Errors are wrapped with smarterr per the finder contract.
func findIPRoutes(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.IpRouteInfo, error) {
	input := directoryservice.ListIpRoutesInput{
		DirectoryId: aws.String(directoryID),
	}

	var output []awstypes.IpRouteInfo
	paginator := directoryservice.NewListIpRoutesPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.IpRoutesInfo {
			// Only manage IPv4 CIDR routes.
			if v.CidrIp == nil {
				continue
			}
			output = append(output, v)
		}
	}

	return output, nil
}

// findIPRoutesByDirectoryID returns the active IPv4 IP routes for a directory.
// Routes that are being removed (or have been removed) are excluded, and a
// directory with no active routes is treated as not found so the resource is
// removed from state. Used by Read and the list resource.
func findIPRoutesByDirectoryID(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.IpRouteInfo, error) {
	routes, err := findIPRoutes(ctx, conn, directoryID)
	if err != nil {
		return nil, err
	}

	var active []awstypes.IpRouteInfo
	for _, v := range routes {
		if v.IpRouteStatusMsg == awstypes.IpRouteStatusMsgRemoved || v.IpRouteStatusMsg == awstypes.IpRouteStatusMsgRemoving {
			continue
		}
		active = append(active, v)
	}

	if len(active) == 0 {
		return nil, smarterr.NewError(&retry.NotFoundError{})
	}

	return active, nil
}

const (
	ipRoutesStatusAdding   = "Adding"
	ipRoutesStatusAdded    = "Added"
	ipRoutesStatusRemoving = "Removing"
	ipRoutesStatusRemoved  = "Removed"
)

func statusIPRoutesAdded(conn *directoryservice.Client, directoryID string, cidrs []string) retry.StateRefreshFunc {
	want := make(map[string]struct{}, len(cidrs))
	for _, c := range cidrs {
		want[c] = struct{}{}
	}

	return func(ctx context.Context) (any, string, error) {
		routes, err := findIPRoutes(ctx, conn, directoryID)

		if retry.NotFound(err) {
			return []awstypes.IpRouteInfo{}, ipRoutesStatusAdding, nil
		}

		if err != nil {
			return nil, "", err
		}

		got := make(map[string]awstypes.IpRouteStatusMsg, len(routes))
		for _, v := range routes {
			got[aws.ToString(v.CidrIp)] = v.IpRouteStatusMsg
		}

		added := true
		for c := range want {
			s, ok := got[c]
			if !ok {
				added = false
				continue
			}
			switch s {
			case awstypes.IpRouteStatusMsgAddFailed:
				return routes, string(s), fmt.Errorf("IP route %q failed to add", c)
			case awstypes.IpRouteStatusMsgAdded:
			default:
				added = false
			}
		}

		if added {
			return routes, ipRoutesStatusAdded, nil
		}
		return routes, ipRoutesStatusAdding, nil
	}
}

func statusIPRoutesRemoved(conn *directoryservice.Client, directoryID string, cidrs []string) retry.StateRefreshFunc {
	want := make(map[string]struct{}, len(cidrs))
	for _, c := range cidrs {
		want[c] = struct{}{}
	}

	return func(ctx context.Context) (any, string, error) {
		routes, err := findIPRoutes(ctx, conn, directoryID)

		if retry.NotFound(err) {
			return []awstypes.IpRouteInfo{}, ipRoutesStatusRemoved, nil
		}

		if err != nil {
			return nil, "", err
		}

		remaining := false
		for _, v := range routes {
			if _, ok := want[aws.ToString(v.CidrIp)]; !ok {
				continue
			}
			if v.IpRouteStatusMsg == awstypes.IpRouteStatusMsgRemoveFailed {
				return routes, string(v.IpRouteStatusMsg), fmt.Errorf("IP route %q failed to remove", aws.ToString(v.CidrIp))
			}
			// A route still present in any non-Removed state (including
			// "Removing") means removal is not yet complete.
			if v.IpRouteStatusMsg == awstypes.IpRouteStatusMsgRemoved {
				continue
			}
			remaining = true
		}

		if remaining {
			return routes, ipRoutesStatusRemoving, nil
		}
		return routes, ipRoutesStatusRemoved, nil
	}
}

func waitIPRoutesAdded(ctx context.Context, conn *directoryservice.Client, directoryID string, cidrs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ipRoutesStatusAdding},
		Target:  []string{ipRoutesStatusAdded},
		Refresh: statusIPRoutesAdded(conn, directoryID, cidrs),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitIPRoutesRemoved(ctx context.Context, conn *directoryservice.Client, directoryID string, cidrs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ipRoutesStatusRemoving},
		Target:  []string{ipRoutesStatusRemoved},
		Refresh: statusIPRoutesRemoved(conn, directoryID, cidrs),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

type ipRoutesResourceModel struct {
	framework.WithRegionModel
	DirectoryID                                types.String                                 `tfsdk:"directory_id"`
	IPRoutes                                   fwtypes.SetNestedObjectValueOf[ipRouteModel] `tfsdk:"ip_route"`
	Timeouts                                   timeouts.Value                               `tfsdk:"timeouts"`
	UpdateSecurityGroupForDirectoryControllers types.Bool                                   `tfsdk:"update_security_group_for_directory_controllers"`
}

type ipRouteModel struct {
	CidrIP      types.String `tfsdk:"cidr_ip"`
	Description types.String `tfsdk:"description"`
}
