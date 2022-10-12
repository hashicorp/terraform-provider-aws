package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"golang.org/x/exp/slices"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourceIPRanges)
}

// newDataSourceIPRanges instantiates a new DataSource for the aws_ip_ranges data source.
func newDataSourceIPRanges(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceIPRanges{}, nil
}

type dataSourceIPRanges struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceIPRanges) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_ip_ranges"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceIPRanges) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"cidr_blocks": {
				Type:     types.ListType{ElemType: types.StringType},
				Computed: true,
			},
			"create_date": {
				Type:     types.StringType,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"ipv6_cidr_blocks": {
				Type:     types.ListType{ElemType: types.StringType},
				Computed: true,
			},
			"regions": {
				Type:     types.SetType{ElemType: types.StringType},
				Optional: true,
			},
			"services": {
				Type:     types.SetType{ElemType: types.StringType},
				Required: true,
			},
			"sync_token": {
				Type:     types.Int64Type,
				Computed: true,
			},
			"url": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceIPRanges) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceIPRanges) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceIPRangesData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var url string

	if data.URL.IsNull() {
		// Data sources make no use of AttributePlanModifiers to set default values.
		url = "https://ip-ranges.amazonaws.com/ip-ranges.json"
	} else {
		url = data.URL.Value
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

	regions := flex.ExpandFrameworkStringValueSet(ctx, data.Regions)
	services := flex.ExpandFrameworkStringValueSet(ctx, data.Services)
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

	sort.Strings(ipv4Prefixes)

	var ipv6Prefixes []string

	for _, v := range ipRanges.IPv6Prefixes {
		if matchFilter(v.Region, v.Service) {
			ipv6Prefixes = append(ipv6Prefixes, v.Prefix)
		}
	}

	sort.Strings(ipv6Prefixes)

	data.CreateDate = types.String{Value: ipRanges.CreateDate}
	data.ID = types.String{Value: ipRanges.SyncToken}
	data.IPv4CIDRBlocks = flex.FlattenFrameworkStringValueList(ctx, ipv4Prefixes)
	data.IPv6CIDRBlocks = flex.FlattenFrameworkStringValueList(ctx, ipv6Prefixes)
	data.SyncToken = types.Int64{Value: int64(syncToken)}
	data.URL = types.String{Value: url}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceIPRangesData struct {
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
