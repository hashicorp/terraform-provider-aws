// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ip_ranges", name="IP Ranges")
func newIPRangesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &ipRangesDataSource{}

	return d, nil
}

type ipRangesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *ipRangesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cidr_blocks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"create_date": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_blocks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"regions": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"services": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"sync_token": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrURL: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *ipRangesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data ipRangesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var url string
	if data.URL.IsNull() {
		// Data sources make no use of AttributePlanModifiers to set default values.
		url = "https://ip-ranges.amazonaws.com/ip-ranges.json"
	} else {
		url = data.URL.ValueString()
	}

	bytes, err := readAll(ctx, url)
	if err != nil {
		response.Diagnostics.AddError("downloading IP ranges", err.Error())

		return
	}

	ipRanges := new(ipRanges)
	if err := json.Unmarshal(bytes, ipRanges); err != nil {
		response.Diagnostics.AddError("parsing JSON", err.Error())

		return
	}

	syncToken, err := strconv.Atoi(ipRanges.SyncToken)
	if err != nil {
		response.Diagnostics.AddError("parsing SyncToken", err.Error())

		return
	}

	regions := tfslices.ApplyToAll(fwflex.ExpandFrameworkStringValueSet(ctx, data.Regions), strings.ToLower)
	services := tfslices.ApplyToAll(fwflex.ExpandFrameworkStringValueSet(ctx, data.Services), strings.ToLower)
	matchFilter := func(region, service string) bool {
		matchRegion := len(regions) == 0 || slices.Contains(regions, strings.ToLower(region))
		matchService := slices.Contains(services, strings.ToLower(service))

		return matchRegion && matchService
	}

	var ipv4Prefixes []string
	for _, v := range ipRanges.IPv4Prefixes {
		if matchFilter(v.Region, v.Service) {
			ipv4Prefixes = append(ipv4Prefixes, v.Prefix)
		}
	}
	slices.Sort(ipv4Prefixes)

	var ipv6Prefixes []string
	for _, v := range ipRanges.IPv6Prefixes {
		if matchFilter(v.Region, v.Service) {
			ipv6Prefixes = append(ipv6Prefixes, v.Prefix)
		}
	}
	slices.Sort(ipv6Prefixes)

	data.CreateDate = fwflex.StringValueToFrameworkLegacy(ctx, ipRanges.CreateDate)
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, ipRanges.SyncToken)
	data.IPv4CIDRBlocks = fwflex.FlattenFrameworkStringValueListLegacy(ctx, ipv4Prefixes)
	data.IPv6CIDRBlocks = fwflex.FlattenFrameworkStringValueListLegacy(ctx, ipv6Prefixes)
	data.SyncToken = types.Int64Value(int64(syncToken))
	data.URL = fwflex.StringValueToFrameworkLegacy(ctx, url)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type ipRangesDataSourceModel struct {
	CreateDate     types.String `tfsdk:"create_date"`
	ID             types.String `tfsdk:"id"`
	IPv4CIDRBlocks types.List   `tfsdk:"cidr_blocks"`
	IPv6CIDRBlocks types.List   `tfsdk:"ipv6_cidr_blocks"`
	Regions        types.Set    `tfsdk:"regions"`
	Services       types.Set    `tfsdk:"services"`
	SyncToken      types.Int64  `tfsdk:"sync_token"`
	URL            types.String `tfsdk:"url"`
}

func readAll(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := cleanhttp.DefaultClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET (%s): %w", url, err)
	}
	defer response.Body.Close()

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body (%s): %w", url, err)
	}

	return bytes, nil
}

type ipRanges struct {
	CreateDate   string
	IPv4Prefixes []ipv4Prefix `json:"prefixes"`
	IPv6Prefixes []ipv6Prefix `json:"ipv6_prefixes"`
	SyncToken    string
}

type ipv4Prefix struct {
	Prefix  string `json:"ip_prefix"`
	Region  string
	Service string
}

type ipv6Prefix struct {
	Prefix  string `json:"ipv6_prefix"`
	Region  string
	Service string
}
