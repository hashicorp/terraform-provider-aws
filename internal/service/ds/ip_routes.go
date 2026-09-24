// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
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
					// Only consumed by AddIpRoutes and never returned by the API.
					// Use the configured-only variant so an omitted/defaulted value
					// (e.g. after import) does not force replacement, while an
					// explicit change still does.
					boolplanmodifier.RequiresReplaceIfConfigured(),
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
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv4CIDRNetworkAddress(),
							},
						},
						"cidr_ipv6": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								fwvalidators.IPv6CIDRNetworkAddress(),
							},
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
					},
					Validators: []validator.Object{
						tfobjectvalidator.ExactlyOneOfChildren(
							path.MatchRelative().AtName("cidr_ip"),
							path.MatchRelative().AtName("cidr_ipv6"),
						),
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

func (r *ipRoutesResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data ipRoutesResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.IPRoutes.IsNull() || data.IPRoutes.IsUnknown() {
		return
	}

	routes, diags := data.IPRoutes.ToSlice(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// A set only deduplicates whole objects, so two blocks with the same CIDR
	// but different descriptions would both be kept. Because routes are keyed by
	// CIDR internally, that is ambiguous; require each route's CIDR (IPv4 or
	// IPv6) to be unique.
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		cidr := route.routeKey()
		if cidr == "" {
			continue
		}
		if _, ok := seen[cidr]; ok {
			response.Diagnostics.AddAttributeError(
				path.Root("ip_route"),
				"Duplicate CIDR",
				fmt.Sprintf("CIDR %q is configured in more than one ip_route block; each cidr_ip/cidr_ipv6 must be unique.", cidr),
			)
			return
		}
		seen[cidr] = struct{}{}
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
		planByCIDR[v.routeKey()] = v
	}
	stateByCIDR := make(map[string]*ipRouteModel, len(stateRoutes))
	for _, v := range stateRoutes {
		stateByCIDR[v.routeKey()] = v
	}

	var add []awstypes.IpRoute
	for key, pr := range planByCIDR {
		if sr, ok := stateByCIDR[key]; !ok || !sr.Description.Equal(pr.Description) {
			var route awstypes.IpRoute
			smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, pr, &route))
			if response.Diagnostics.HasError() {
				return
			}
			add = append(add, route)
		}
	}

	var removeV4, removeV6 []string
	for key, sr := range stateByCIDR {
		if pr, ok := planByCIDR[key]; !ok || !pr.Description.Equal(sr.Description) {
			if sr.isIPv6() {
				removeV6 = append(removeV6, key)
			} else {
				removeV4 = append(removeV4, key)
			}
		}
	}

	if len(removeV4) > 0 || len(removeV6) > 0 {
		_, err := conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     removeV4,
			CidrIpv6s:   removeV6,
		})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
			return
		}
		removeKeys := slices.Concat(removeV4, removeV6)
		if err := waitIPRoutesRemoved(ctx, conn, directoryID, removeKeys, r.DeleteTimeout(ctx, plan.Timeouts)); err != nil {
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

	var removeV4, removeV6 []string
	for _, v := range routes {
		if v.isIPv6() {
			removeV6 = append(removeV6, v.CidrIPv6.ValueString())
		} else {
			removeV4 = append(removeV4, v.CidrIP.ValueString())
		}
	}

	if len(removeV4) == 0 && len(removeV6) == 0 {
		return
	}

	_, err := conn.RemoveIpRoutes(ctx, &directoryservice.RemoveIpRoutesInput{
		DirectoryId: aws.String(directoryID),
		CidrIps:     removeV4,
		CidrIpv6s:   removeV6,
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	removeKeys := slices.Concat(removeV4, removeV6)
	if err := waitIPRoutesRemoved(ctx, conn, directoryID, removeKeys, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}
}

func (r *ipRoutesResource) flatten(ctx context.Context, routes []awstypes.IpRouteInfo, data *ipRoutesResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, routes, &data.IPRoutes)...)
	return diags
}

// ipRoutesCIDRs returns the identifying CIDR (IPv4 or IPv6) of each route.
func ipRoutesCIDRs(routes []awstypes.IpRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, v := range routes {
		cidrs = append(cidrs, ipRouteKey(v))
	}
	return cidrs
}

func ipRouteKey(v awstypes.IpRoute) string {
	if v.CidrIp != nil {
		return aws.ToString(v.CidrIp)
	}
	return aws.ToString(v.CidrIpv6)
}

func ipRouteInfoKey(v awstypes.IpRouteInfo) string {
	if v.CidrIp != nil {
		return aws.ToString(v.CidrIp)
	}
	return aws.ToString(v.CidrIpv6)
}

// findIPRoutes returns all IP routes (IPv4 and IPv6) for a directory regardless
// of status. It is used by the waiters, which must observe in-progress
// ("Adding"/"Removing") routes. Errors are wrapped with smarterr per the finder
// contract.
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
			// Skip anything without an identifying CIDR.
			if v.CidrIp == nil && v.CidrIpv6 == nil {
				continue
			}
			output = append(output, v)
		}
	}

	return output, nil
}

// findIPRoutesByDirectoryID returns the active IP routes (IPv4 and IPv6) for a
// directory. Routes that are being removed (or have been removed) are excluded,
// and a directory with no active routes is treated as not found so the resource
// is removed from state. Used by Read and the list resource.
func findIPRoutesByDirectoryID(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.IpRouteInfo, error) {
	routes, err := findIPRoutes(ctx, conn, directoryID)
	if err != nil {
		return nil, err
	}

	var active []awstypes.IpRouteInfo
	for _, v := range routes {
		switch v.IpRouteStatusMsg {
		case awstypes.IpRouteStatusMsgRemoved, awstypes.IpRouteStatusMsgRemoving:
			// Being removed or already gone.
			continue
		case awstypes.IpRouteStatusMsgAddFailed:
			// A failed addition was never successfully added, so it must not
			// appear in state; leaving it out preserves the configuration diff
			// so the next apply retries it.
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
			got[ipRouteInfoKey(v)] = v.IpRouteStatusMsg
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
			if _, ok := want[ipRouteInfoKey(v)]; !ok {
				continue
			}
			if v.IpRouteStatusMsg == awstypes.IpRouteStatusMsgRemoveFailed {
				return routes, string(v.IpRouteStatusMsg), fmt.Errorf("IP route %q failed to remove", ipRouteInfoKey(v))
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
	CidrIPv6    types.String `tfsdk:"cidr_ipv6"`
	Description types.String `tfsdk:"description"`
}

// routeKey returns the CIDR that identifies a route model (IPv4 or IPv6).
func (m ipRouteModel) routeKey() string {
	if !m.CidrIP.IsNull() && !m.CidrIP.IsUnknown() && m.CidrIP.ValueString() != "" {
		return m.CidrIP.ValueString()
	}
	return m.CidrIPv6.ValueString()
}

func (m ipRouteModel) isIPv6() bool {
	return m.CidrIP.IsNull() || m.CidrIP.ValueString() == ""
}
