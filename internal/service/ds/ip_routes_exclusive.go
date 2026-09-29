// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_ip_routes_exclusive", name="IP Routes Exclusive")
// @IdentityAttribute("directory_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(domainTfVar="domain")
// @Testing(checkDestroyNoop=true)
// @Testing(importStateIdAttribute="directory_id")
// @Testing(importIgnore="update_security_group_for_directory_controllers")
// @Testing(plannableImportAction="NoOp")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckDirectoryService")
func newIPRoutesExclusiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ipRoutesExclusiveResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

type ipRoutesExclusiveResource struct {
	framework.ResourceWithModel[ipRoutesExclusiveResourceModel]
	framework.WithImportByIdentity
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *ipRoutesExclusiveResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			// Consumed only by AddIpRoutes and never returned by the API, so it
			// applies only to routes this resource adds.
			"update_security_group_for_directory_controllers": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
		Blocks: map[string]schema.Block{
			"ip_route": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[ipRouteModel](ctx, fwtypes.WithSemanticEqualityFunc(ipRoutesSemanticEquals)),
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
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *ipRoutesExclusiveResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data ipRoutesExclusiveResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	if data.IPRoutes.IsNull() || data.IPRoutes.IsUnknown() {
		return
	}

	routes, diags := data.IPRoutes.ToSlice(ctx)
	smerr.AddEnrich(ctx, &response.Diagnostics, diags)
	if response.Diagnostics.HasError() {
		return
	}

	// A set only deduplicates whole objects, so two blocks with the same CIDR
	// but different descriptions would both be kept. Routes are keyed by CIDR,
	// so require each CIDR to be unique.
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		// Re-checked once known.
		if route.CidrIP.IsUnknown() || route.CidrIPv6.IsUnknown() {
			continue
		}

		// An object-level validator cannot read sibling values inside a set
		// nested block, so exactly-one-of is enforced here.
		hasV4, hasV6 := !route.CidrIP.IsNull(), !route.CidrIPv6.IsNull()
		if hasV4 == hasV6 {
			smerr.AddOne(ctx, &response.Diagnostics, diag.NewAttributeErrorDiagnostic(
				path.Root("ip_route"),
				"Invalid Attribute Combination",
				"Each ip_route block must configure exactly one of cidr_ip or cidr_ipv6.",
			))
			return
		}

		cidr := route.routeKey()
		if _, ok := seen[cidr]; ok {
			smerr.AddOne(ctx, &response.Diagnostics, diag.NewAttributeErrorDiagnostic(
				path.Root("ip_route"),
				"Duplicate CIDR",
				fmt.Sprintf("CIDR %q is configured in more than one ip_route block; each cidr_ip/cidr_ipv6 must be unique.", cidr),
			))
			return
		}
		seen[cidr] = struct{}{}
	}
}

func (r *ipRoutesExclusiveResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ipRoutesExclusiveResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.sync(ctx, data, r.CreateTimeout(ctx, data.Timeouts)))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRoutesExclusiveResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ipRoutesExclusiveResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// The flag is never returned by the API. Seed the schema default on import
	// so the value is concrete.
	if data.UpdateSecurityGroupForDirectoryControllers.IsNull() {
		data.UpdateSecurityGroupForDirectoryControllers = types.BoolValue(false)
	}

	conn := r.Meta().DSClient(ctx)

	directoryID := fwflex.StringValueFromFramework(ctx, data.DirectoryID)
	routes, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)

	// ListIpRoutes can report EntityDoesNotExistException for an existing
	// directory, so only a missing directory removes the resource.
	if retry.NotFound(err) {
		_, err = findDirectoryByID(ctx, conn, directoryID)
		if retry.NotFound(err) {
			smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, directoryID)
			response.State.RemoveResource(ctx)
			return
		}
		routes = nil
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, directoryID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, routes, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRoutesExclusiveResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data ipRoutesExclusiveResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.sync(ctx, data, r.UpdateTimeout(ctx, data.Timeouts)))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

// sync reconciles the directory's routes with the plan. It diffs against the
// live routes rather than prior state so that routes added outside this
// resource (including at create time) are removed.
func (r *ipRoutesExclusiveResource) sync(ctx context.Context, plan ipRoutesExclusiveResourceModel, timeout time.Duration) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := r.Meta().DSClient(ctx)
	directoryID := fwflex.StringValueFromFramework(ctx, plan.DirectoryID)

	planRoutes, d := plan.IPRoutes.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	current, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)
	if retry.NotFound(err) {
		// Distinguish "no routes" from "no directory"; the latter must fail.
		if _, err = findDirectoryByID(ctx, conn, directoryID); err == nil {
			current = nil
		}
	}
	if err != nil {
		smerr.AddError(ctx, &diags, err, smerr.ID, directoryID)
		return diags
	}

	// There is no update API, so a changed description is a remove + add.
	planByCIDR := make(map[string]*ipRouteModel, len(planRoutes))
	for _, v := range planRoutes {
		planByCIDR[v.routeKey()] = v
	}
	currentByCIDR := make(map[string]awstypes.IpRouteInfo, len(current))
	for _, v := range current {
		currentByCIDR[ipRouteInfoKey(v)] = v
	}

	var add []awstypes.IpRoute
	for cidr, pr := range planByCIDR {
		if cr, ok := currentByCIDR[cidr]; ok && aws.ToString(cr.Description) == pr.Description.ValueString() {
			continue
		}
		var route awstypes.IpRoute
		diags.Append(fwflex.Expand(ctx, pr, &route)...)
		if diags.HasError() {
			return diags
		}
		add = append(add, route)
	}

	var removeV4, removeV6 []string
	for cidr, cr := range currentByCIDR {
		if pr, ok := planByCIDR[cidr]; ok && aws.ToString(cr.Description) == pr.Description.ValueString() {
			continue
		}
		if cr.CidrIp != nil {
			removeV4 = append(removeV4, aws.ToString(cr.CidrIp))
		} else {
			removeV6 = append(removeV6, aws.ToString(cr.CidrIpv6))
		}
	}

	if len(removeV4) > 0 || len(removeV6) > 0 {
		input := directoryservice.RemoveIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			CidrIps:     removeV4,
			CidrIpv6s:   removeV6,
		}
		if err := removeIPRoutes(ctx, conn, &input, timeout); err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, directoryID)
			return diags
		}

		var keys []string
		for _, v := range slices.Concat(removeV4, removeV6) {
			keys = append(keys, inttypes.CanonicalCIDRBlock(v))
		}
		if err := waitIPRoutesRemoved(ctx, conn, directoryID, keys, timeout); err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, directoryID)
			return diags
		}
	}

	if len(add) > 0 {
		input := directoryservice.AddIpRoutesInput{
			DirectoryId: aws.String(directoryID),
			IpRoutes:    add,
			UpdateSecurityGroupForDirectoryControllers: fwflex.BoolValueFromFramework(ctx, plan.UpdateSecurityGroupForDirectoryControllers),
		}
		if err := addIPRoutes(ctx, conn, &input, timeout); err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, directoryID)
			return diags
		}

		if err := waitIPRoutesAdded(ctx, conn, directoryID, ipRoutesCIDRs(add), timeout); err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, directoryID)
			return diags
		}
	}

	return diags
}

func (r *ipRoutesExclusiveResource) flatten(ctx context.Context, routes []awstypes.IpRouteInfo, data *ipRoutesExclusiveResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Blocks are never null in configuration; an empty result must be an empty
	// set, not null, to match a configuration with no ip_route blocks.
	if len(routes) == 0 {
		data.IPRoutes = fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*ipRouteModel{})
		return diags
	}

	diags.Append(fwflex.Flatten(ctx, routes, &data.IPRoutes)...)
	return diags
}

// ipRoutesCIDRs returns the canonical CIDR (IPv4 or IPv6) of each route.
func ipRoutesCIDRs(routes []awstypes.IpRoute) []string {
	cidrs := make([]string, 0, len(routes))
	for _, v := range routes {
		cidrs = append(cidrs, ipRouteKey(v))
	}
	return cidrs
}

func ipRouteKey(v awstypes.IpRoute) string {
	if v.CidrIp != nil {
		return inttypes.CanonicalCIDRBlock(aws.ToString(v.CidrIp))
	}
	return inttypes.CanonicalCIDRBlock(aws.ToString(v.CidrIpv6))
}

type ipRoutesExclusiveResourceModel struct {
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

// routeKey returns the canonical CIDR identifying the route, so equivalent IPv6
// spellings (e.g. "2001:db8::/64" and "2001:0db8::/64") compare equal.
func (m ipRouteModel) routeKey() string {
	if m.CidrIP.ValueString() != "" {
		return inttypes.CanonicalCIDRBlock(m.CidrIP.ValueString())
	}
	return inttypes.CanonicalCIDRBlock(m.CidrIPv6.ValueString())
}

// ipRoutesSemanticEquals treats two "ip_route" sets as equal when they contain
// the same routes keyed by canonical CIDR and description. Without it, a
// configured "2001:0db8::/64" vs. the AWS-canonicalized "2001:db8::/64" would be
// a different set element and show as persistent drift.
func ipRoutesSemanticEquals(ctx context.Context, a, b fwtypes.NestedCollectionValue[ipRouteModel]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if a.Equal(b) {
		return true, diags
	}

	aSlice, d := a.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}
	bSlice, d := b.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	if len(aSlice) != len(bSlice) {
		return false, diags
	}

	aByKey := make(map[string]*ipRouteModel, len(aSlice))
	for _, r := range aSlice {
		aByKey[r.routeKey()] = r
	}

	for _, r := range bSlice {
		if other, ok := aByKey[r.routeKey()]; !ok || !other.Description.Equal(r.Description) {
			return false, diags
		}
	}

	return true, diags
}
