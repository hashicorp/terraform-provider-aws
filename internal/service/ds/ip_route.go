// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_directory_service_ip_route", name="IP Route")
// @IdentityAttribute("directory_id")
// @IdentityAttribute("cidr_ip", optional="true", testNotNull="true")
// @IdentityAttribute("cidr_ipv6", optional="true")
// @ImportIDHandler("ipRouteImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(domainTfVar="domain")
// @Testing(importStateIdAttributes="directory_id;cidr_ip", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(plannableImportAction="NoOp")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckDirectoryService")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/directoryservice/types;awstypes;awstypes.IpRouteInfo")
func newIPRouteResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ipRouteResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type ipRouteResource struct {
	framework.ResourceWithModel[ipRouteResourceModel]
	framework.WithImportByIdentity
	framework.WithNoUpdate
	framework.WithTimeouts
}

func (r *ipRouteResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cidr_ip": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.IPv4CIDRNetworkAddress(),
				},
			},
			"cidr_ipv6": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.IPv6CIDRNetworkAddress(),
				},
			},
			// There is no API to update a route, so every argument forces replacement.
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
					boolplanmodifier.RequiresReplace(),
				},
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

func (r *ipRouteResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cidr_ip"),
			path.MatchRoot("cidr_ipv6"),
		),
	}
}

func (r *ipRouteResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ipRouteResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID, cidr := fwflex.StringValueFromFramework(ctx, data.DirectoryID), data.routeKey()
	id := directoryID + intflex.ResourceIdSeparator + cidr
	input := directoryservice.AddIpRoutesInput{
		DirectoryId: aws.String(directoryID),
		IpRoutes: []awstypes.IpRoute{{
			CidrIp:      fwflex.StringFromFramework(ctx, data.CidrIP),
			CidrIpv6:    fwflex.StringFromFramework(ctx, data.CidrIPv6),
			Description: fwflex.StringFromFramework(ctx, data.Description),
		}},
		UpdateSecurityGroupForDirectoryControllers: fwflex.BoolValueFromFramework(ctx, data.UpdateSecurityGroupForDirectoryControllers),
	}

	timeout := r.CreateTimeout(ctx, data.Timeouts)
	if err := addIPRoutes(ctx, conn, &input, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if err := waitIPRoutesAdded(ctx, conn, directoryID, []string{cidr}, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRouteResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ipRouteResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// The flag is never returned by the API. Seed the schema default on import
	// so the next plan does not see null -> false as a replacement.
	if data.UpdateSecurityGroupForDirectoryControllers.IsNull() {
		data.UpdateSecurityGroupForDirectoryControllers = types.BoolValue(false)
	}

	conn := r.Meta().DSClient(ctx)

	directoryID, cidr := fwflex.StringValueFromFramework(ctx, data.DirectoryID), data.routeKey()
	id := directoryID + intflex.ResourceIdSeparator + cidr
	output, err := findIPRouteByTwoPartKey(ctx, conn, directoryID, cidr)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, id)
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	r.flatten(ctx, output, &data)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *ipRouteResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ipRouteResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSClient(ctx)

	directoryID, cidr := fwflex.StringValueFromFramework(ctx, data.DirectoryID), data.routeKey()
	id := directoryID + intflex.ResourceIdSeparator + cidr
	input := directoryservice.RemoveIpRoutesInput{
		DirectoryId: aws.String(directoryID),
	}
	if data.isIPv6() {
		input.CidrIpv6s = []string{cidr}
	} else {
		input.CidrIps = []string{cidr}
	}

	timeout := r.DeleteTimeout(ctx, data.Timeouts)
	err := removeIPRoutes(ctx, conn, &input, timeout)

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) || errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if err := waitIPRoutesRemoved(ctx, conn, directoryID, []string{cidr}, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

// flatten copies an API route into the model. A configured CIDR is kept when
// it is equivalent to the AWS-canonicalized value (for example "2001:0db8::/64"
// vs. "2001:db8::/64") so that spelling differences do not force replacement.
func (r *ipRouteResource) flatten(ctx context.Context, route *awstypes.IpRouteInfo, data *ipRouteResourceModel) {
	if v := aws.ToString(route.CidrIp); v != "" {
		if inttypes.CanonicalCIDRBlock(data.CidrIP.ValueString()) != inttypes.CanonicalCIDRBlock(v) {
			data.CidrIP = types.StringValue(v)
		}
		data.CidrIPv6 = types.StringNull()
	} else {
		if inttypes.CanonicalCIDRBlock(data.CidrIPv6.ValueString()) != inttypes.CanonicalCIDRBlock(aws.ToString(route.CidrIpv6)) {
			data.CidrIPv6 = fwflex.StringToFramework(ctx, route.CidrIpv6)
		}
		data.CidrIP = types.StringNull()
	}
	data.Description = fwflex.StringToFramework(ctx, route.Description)
}

type ipRouteResourceModel struct {
	framework.WithRegionModel
	CidrIP                                     types.String   `tfsdk:"cidr_ip"`
	CidrIPv6                                   types.String   `tfsdk:"cidr_ipv6"`
	Description                                types.String   `tfsdk:"description"`
	DirectoryID                                types.String   `tfsdk:"directory_id"`
	Timeouts                                   timeouts.Value `tfsdk:"timeouts"`
	UpdateSecurityGroupForDirectoryControllers types.Bool     `tfsdk:"update_security_group_for_directory_controllers"`
}

func (m ipRouteResourceModel) routeKey() string {
	if m.isIPv6() {
		return inttypes.CanonicalCIDRBlock(m.CidrIPv6.ValueString())
	}
	return inttypes.CanonicalCIDRBlock(m.CidrIP.ValueString())
}

func (m ipRouteResourceModel) isIPv6() bool {
	return m.CidrIP.ValueString() == ""
}

var _ inttypes.ImportIDParser = ipRouteImportID{}

type ipRouteImportID struct{}

func (ipRouteImportID) Parse(id string) (string, map[string]any, error) {
	directoryID, cidr, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || directoryID == "" || cidr == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <directory-id>%s<cidr>", id, intflex.ResourceIdSeparator)
	}

	attr := "cidr_ip"
	if strings.Contains(cidr, ":") {
		attr = "cidr_ipv6"
	}

	return id, map[string]any{
		"directory_id": directoryID,
		attr:           cidr,
	}, nil
}

// Directory Service serializes updates per directory and rejects a concurrent
// one, e.g. when several aws_directory_service_ip_route resources target the
// same directory.
const errMessageUpdateInProgress = "update is already in progress"

func addIPRoutes(ctx context.Context, conn *directoryservice.Client, input *directoryservice.AddIpRoutesInput, timeout time.Duration) error {
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ClientException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.AddIpRoutes(ctx, input)
	}, errMessageUpdateInProgress)

	return smarterr.NewError(err)
}

func removeIPRoutes(ctx context.Context, conn *directoryservice.Client, input *directoryservice.RemoveIpRoutesInput, timeout time.Duration) error {
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ClientException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.RemoveIpRoutes(ctx, input)
	}, errMessageUpdateInProgress)

	return smarterr.NewError(err)
}

func ipRouteInfoKey(v awstypes.IpRouteInfo) string {
	if v.CidrIp != nil {
		return inttypes.CanonicalCIDRBlock(aws.ToString(v.CidrIp))
	}
	return inttypes.CanonicalCIDRBlock(aws.ToString(v.CidrIpv6))
}

// findIPRoutes returns all IP routes (IPv4 and IPv6) for a directory regardless
// of status. The waiters use it because they must observe in-progress
// ("Adding"/"Removing") routes.
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
			if v.CidrIp == nil && v.CidrIpv6 == nil {
				continue
			}
			output = append(output, v)
		}
	}

	return output, nil
}

// findIPRoutesByDirectoryID returns the active IP routes (IPv4 and IPv6) for a
// directory. The result may be empty. Routes being removed, already removed, or
// that failed to add are excluded: a failed addition was never applied, so
// leaving it out preserves the configuration diff and the next apply retries it.
func findIPRoutesByDirectoryID(ctx context.Context, conn *directoryservice.Client, directoryID string) ([]awstypes.IpRouteInfo, error) {
	routes, err := findIPRoutes(ctx, conn, directoryID)
	if err != nil {
		return nil, err
	}

	var active []awstypes.IpRouteInfo
	for _, v := range routes {
		switch v.IpRouteStatusMsg {
		case awstypes.IpRouteStatusMsgRemoved, awstypes.IpRouteStatusMsgRemoving, awstypes.IpRouteStatusMsgAddFailed:
			continue
		}
		active = append(active, v)
	}

	return active, nil
}

// findIPRouteByTwoPartKey returns the active IP route identified by its
// canonical CIDR (IPv4 or IPv6).
func findIPRouteByTwoPartKey(ctx context.Context, conn *directoryservice.Client, directoryID, cidr string) (*awstypes.IpRouteInfo, error) {
	routes, err := findIPRoutesByDirectoryID(ctx, conn, directoryID)
	if err != nil {
		return nil, err
	}

	for _, v := range routes {
		if ipRouteInfoKey(v) == cidr {
			return &v, nil
		}
	}

	return nil, smarterr.NewError(&retry.NotFoundError{})
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

		// StateChangeConf treats a nil refresh result as "not found" regardless
		// of the returned state, so never return a nil slice with a valid state.
		if routes == nil {
			routes = []awstypes.IpRouteInfo{}
		}

		got := make(map[string]awstypes.IpRouteInfo, len(routes))
		for _, v := range routes {
			got[ipRouteInfoKey(v)] = v
		}

		added := true
		for c := range want {
			v, ok := got[c]
			if !ok {
				added = false
				continue
			}
			switch v.IpRouteStatusMsg {
			case awstypes.IpRouteStatusMsgAddFailed:
				return routes, string(v.IpRouteStatusMsg), smarterr.NewError(fmt.Errorf("IP route %q failed to add: %s", c, aws.ToString(v.IpRouteStatusReason)))
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

		// StateChangeConf treats a nil refresh result as "not found" regardless
		// of the returned state, so never return a nil slice with a valid state.
		if routes == nil {
			routes = []awstypes.IpRouteInfo{}
		}

		remaining := false
		for _, v := range routes {
			if _, ok := want[ipRouteInfoKey(v)]; !ok {
				continue
			}
			switch v.IpRouteStatusMsg {
			case awstypes.IpRouteStatusMsgRemoveFailed:
				return routes, string(v.IpRouteStatusMsg), smarterr.NewError(fmt.Errorf("IP route %q failed to remove: %s", ipRouteInfoKey(v), aws.ToString(v.IpRouteStatusReason)))
			case awstypes.IpRouteStatusMsgRemoved:
			default:
				remaining = true
			}
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

	return smarterr.NewError(err)
}

func waitIPRoutesRemoved(ctx context.Context, conn *directoryservice.Client, directoryID string, cidrs []string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ipRoutesStatusRemoving},
		Target:  []string{ipRoutesStatusRemoved},
		Refresh: statusIPRoutesRemoved(conn, directoryID, cidrs),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return smarterr.NewError(err)
}
