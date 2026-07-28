// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestConnectionsDataSourceSchema_noSyntheticID(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	d, err := tfdirectconnect.NewConnectionsDataSource(ctx)
	if err != nil {
		t.Fatalf("creating data source: %s", err)
	}

	response := datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, &response)

	if response.Diagnostics.HasError() {
		t.Fatalf("reading schema: %v", response.Diagnostics)
	}

	if _, ok := response.Schema.Attributes[names.AttrID]; ok {
		t.Errorf("schema defines a top-level %q attribute; plural data sources must not have a synthetic ID", names.AttrID)
	}

	connections, ok := response.Schema.Attributes["connections"]
	if !ok {
		t.Fatal("schema is missing the `connections` attribute")
	}

	// The public attribute names of each list element, spelled out. The item
	// schema is derived from connectionsDataSourceItemModel, so it can never
	// disagree with the model -- a typo in a nested `tfsdk` tag just renames the
	// attribute, and framework.ValidateModel only walks the top-level model, so
	// nothing else in the repo notices. These are literals rather than names.*
	// constants on purpose: they are the documented wire contract, and every one
	// of them appears in website/docs/d/dx_connections.html.markdown.
	wantNames := []string{
		"arn",
		"aws_device",
		"bandwidth",
		"id",
		"location",
		"name",
		"owner_account_id",
		"partner_name",
		"provider_name",
		"state",
		"tags",
		"vlan_id",
	}

	elementType, ok := connections.GetType().(attr.TypeWithElementType)
	if !ok {
		t.Fatalf("`connections` is a %T, want a list type", connections.GetType())
	}

	object, ok := elementType.ElementType().(attr.TypeWithAttributeTypes)
	if !ok {
		t.Fatalf("`connections` element is a %T, want an object type", elementType.ElementType())
	}

	gotNames := slices.Sorted(maps.Keys(object.AttributeTypes()))

	if !slices.Equal(gotNames, wantNames) {
		t.Errorf("`connections` element attributes = %q, want %q", gotNames, wantNames)
	}
}

func TestConnectionARNInvariant(t *testing.T) {
	t.Parallel()

	testCases := map[string][]awstypes.Connection{
		"single": {
			testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
		},
		"multiple regions and accounts": {
			testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
			testConnection("dxcon-fh6ab999", "eu-west-1", "210987654321"),
			testConnection("dxcon-fg31dyv6", "ap-southeast-2", acctest.Ct12Digit),
		},
	}

	for name, apiObjects := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			elements := testFlattenConnections(ctx, t, apiObjects, endpoints.AwsPartitionID, nil)

			for i, element := range elements {
				attrs := testConnectionAttributes(ctx, t, element)
				id := testAttributeString(ctx, t, attrs, names.AttrID)
				arn := testAttributeString(ctx, t, attrs, names.AttrARN)

				if want := "dxcon/" + id; !strings.HasSuffix(arn, want) {
					t.Errorf("element %d: ARN %q does not end in %q", i, arn, want)
				}
			}
		})
	}
}

func TestFlattenConnectionsCount(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		apiObjects []awstypes.Connection
		want       int
	}{
		"nil": {
			apiObjects: nil,
			want:       0,
		},
		"empty": {
			apiObjects: []awstypes.Connection{},
			want:       0,
		},
		"one": {
			apiObjects: []awstypes.Connection{
				testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
			},
			want: 1,
		},
		// No dedup: the same connection ID reported twice must yield two elements,
		// otherwise a pagination bug that repeats a page would be silently hidden.
		"duplicate IDs are not deduplicated": {
			apiObjects: []awstypes.Connection{
				testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
				testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
			},
			want: 2,
		},
		// Every state is returned unfiltered, including terminal ones. The data
		// source reports what the API reports; filtering `deleted`/`rejected`
		// here would silently hide connections the caller asked for.
		"terminal states are not filtered out": {
			apiObjects: []awstypes.Connection{
				testConnectionWithState("dxcon-ffabc123", awstypes.ConnectionStateDeleted),
				testConnectionWithState("dxcon-fh6ab999", awstypes.ConnectionStateRejected),
				testConnectionWithState("dxcon-fg31dyv6", awstypes.ConnectionStateAvailable),
			},
			want: 3,
		},
		// A large list is flattened element-for-element. Guards against an
		// accumulator that overwrites rather than appends, which a 1- or 2-element
		// case cannot distinguish. AC4 (the pager itself) is covered by
		// TestFindConnectionsPaginates in connection_find_test.go.
		"many": {
			apiObjects: slices.Repeat([]awstypes.Connection{testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit)}, 150),
			want:       150,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			connections, diags := tfdirectconnect.FlattenConnections(ctx, testCase.apiObjects, endpoints.AwsPartitionID, nil)
			if diags.HasError() {
				t.Fatalf("flattening connections: %v", diags)
			}

			// Zero connections must be an empty list, never null: a null list would
			// make `connections` unusable in a `for_each`/`length()` expression.
			if connections.IsNull() {
				t.Fatal("flattened connections is null; want an empty list")
			}

			elements := connections.Elements()

			if got, want := len(elements), testCase.want; got != want {
				t.Fatalf("element count = %d, want %d", got, want)
			}

			// Count alone would still match if every element flattened to zero
			// values, so assert each one carries its identity through. aws_device
			// is included because it is the one field whose Go name differs from
			// its attribute name (AwsDeviceV2 -> aws_device), so a mismatch there
			// is invisible to the schema and silently yields null.
			for i, element := range elements {
				attrs := testConnectionAttributes(ctx, t, element)

				for _, name := range []string{names.AttrID, "aws_device"} {
					if testAttributeString(ctx, t, attrs, name) == "" {
						t.Errorf("element %d: %s is empty", i, name)
					}
				}
			}
		})
	}
}

func TestFlattenConnectionsTagsNeverNull(t *testing.T) {
	t.Parallel()

	untagged := testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit)
	untagged.Tags = nil

	emptyTags := testConnection("dxcon-fh6ab999", "us-west-2", acctest.Ct12Digit)
	emptyTags.Tags = []awstypes.Tag{}

	tagged := testConnection("dxcon-fg31dyv6", "us-west-2", acctest.Ct12Digit)
	tagged.Tags = []awstypes.Tag{
		{Key: aws.String("Env"), Value: aws.String("prod")},
	}

	// Only the AWS-managed tag survives IgnoreAWS() removal as an empty map.
	onlyAWSTags := testConnection("dxcon-fgu6yv82", "us-west-2", acctest.Ct12Digit)
	onlyAWSTags.Tags = []awstypes.Tag{
		{Key: aws.String("aws:cloudformation:stack-name"), Value: aws.String("test")},
	}

	ctx := t.Context()
	apiObjects := []awstypes.Connection{untagged, emptyTags, tagged, onlyAWSTags}
	wantCounts := []int{0, 0, 1, 0}

	elements := testFlattenConnections(ctx, t, apiObjects, endpoints.AwsPartitionID, nil)

	for i, element := range elements {
		attrs := testConnectionAttributes(ctx, t, element)

		tags, ok := attrs[names.AttrTags]
		if !ok {
			t.Fatalf("element %d: missing %q attribute", i, names.AttrTags)
		}

		if tags.IsNull() {
			t.Errorf("element %d: %q is null; want an empty map for an untagged connection", i, names.AttrTags)
			continue
		}

		if got, want := len(testAttributeMapElements(ctx, t, tags)), wantCounts[i]; got != want {
			t.Errorf("element %d: %q has %d entries, want %d", i, names.AttrTags, got, want)
		}
	}
}

func TestConnectionARNCrossAccount(t *testing.T) {
	t.Parallel()

	// A hosted connection is owned by an account and located in a Region that
	// differ from the caller's. The ARN must describe the connection, not the
	// caller.
	testCases := map[string]struct {
		apiObject awstypes.Connection
		partition string
		want      string
	}{
		"same account": {
			apiObject: testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit),
			partition: endpoints.AwsPartitionID,
			want:      "arn:aws:directconnect:us-west-2:123456789012:dxcon/dxcon-ffabc123",
		},
		"hosted connection in another account": {
			apiObject: testConnection("dxcon-fh6ab999", "eu-west-1", "210987654321"),
			partition: endpoints.AwsPartitionID,
			want:      "arn:aws:directconnect:eu-west-1:210987654321:dxcon/dxcon-fh6ab999",
		},
		"non-standard partition": {
			apiObject: testConnection("dxcon-fg31dyv6", "us-gov-west-1", "210987654321"),
			partition: endpoints.AwsUsGovPartitionID,
			want:      "arn:aws-us-gov:directconnect:us-gov-west-1:210987654321:dxcon/dxcon-fg31dyv6",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			elements := testFlattenConnections(ctx, t, []awstypes.Connection{testCase.apiObject}, testCase.partition, nil)

			if len(elements) != 1 {
				t.Fatalf("element count = %d, want 1", len(elements))
			}

			attrs := testConnectionAttributes(ctx, t, elements[0])
			if got, want := testAttributeString(ctx, t, attrs, names.AttrARN), testCase.want; got != want {
				t.Errorf("ARN = %q, want %q", got, want)
			}
		})
	}
}

func TestFlattenConnectionsNilFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// A partial API object (throttled/degraded response, or a field the API
	// simply omits) must not panic the ARN builder, which dereferences
	// ConnectionId, OwnerAccount and Region.
	apiObjects := []awstypes.Connection{
		{},
		{ConnectionId: aws.String("dxcon-ffabc123")},
		{OwnerAccount: aws.String(acctest.Ct12Digit), Region: aws.String("us-west-2")},
	}

	elements := testFlattenConnections(ctx, t, apiObjects, endpoints.AwsPartitionID, nil)

	if got, want := len(elements), len(apiObjects); got != want {
		t.Fatalf("element count = %d, want %d", got, want)
	}

	wantARNs := []string{
		"arn:aws:directconnect:::dxcon/",
		"arn:aws:directconnect:::dxcon/dxcon-ffabc123",
		"arn:aws:directconnect:us-west-2:123456789012:dxcon/",
	}

	for i, element := range elements {
		attrs := testConnectionAttributes(ctx, t, element)

		// The ARN is hand-built, so a nil field becomes an empty ARN segment
		// rather than making the whole attribute null.
		if got, want := testAttributeString(ctx, t, attrs, names.AttrARN), wantARNs[i]; got != want {
			t.Errorf("element %d: ARN = %q, want %q", i, got, want)
		}

		// A nil *string field is left NULL by AutoFlex, not set to "". Asserted
		// explicitly because ValueString() reports "" for null and would hide
		// the difference.
		if name, ok := attrs[names.AttrName]; !ok {
			t.Errorf("element %d: missing %q attribute", i, names.AttrName)
		} else if !name.IsNull() {
			t.Errorf("element %d: name = %s, want null", i, name)
		}
	}
}

func TestFlattenConnectionsTagsIgnoreAWS(t *testing.T) {
	t.Parallel()

	apiObject := testConnection("dxcon-ffabc123", "us-west-2", acctest.Ct12Digit)
	apiObject.Tags = []awstypes.Tag{
		{Key: aws.String("aws:cloudformation:stack-name"), Value: aws.String("test")},
		{Key: aws.String("aws:cloudformation:logical-id"), Value: aws.String("Connection")},
		{Key: aws.String("Env"), Value: aws.String("prod")},
		{Key: aws.String("Ignored"), Value: aws.String("yes")},
		{Key: aws.String("prefix:Ignored"), Value: aws.String("yes")},
	}

	testCases := map[string]struct {
		ignoreTagsConfig *tftags.IgnoreConfig
		want             map[string]string
	}{
		"no ignore config": {
			ignoreTagsConfig: nil,
			want: map[string]string{
				"Env":            "prod",
				"Ignored":        "yes",
				"prefix:Ignored": "yes",
			},
		},
		"ignored keys and key prefixes": {
			ignoreTagsConfig: &tftags.IgnoreConfig{
				Keys:        tftags.New(t.Context(), []string{"Ignored"}),
				KeyPrefixes: tftags.New(t.Context(), []string{"prefix:"}),
			},
			want: map[string]string{
				"Env": "prod",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			elements := testFlattenConnections(ctx, t, []awstypes.Connection{apiObject}, endpoints.AwsPartitionID, testCase.ignoreTagsConfig)

			if len(elements) != 1 {
				t.Fatalf("element count = %d, want 1", len(elements))
			}

			attrs := testConnectionAttributes(ctx, t, elements[0])
			got := testAttributeMapElements(ctx, t, attrs[names.AttrTags])

			if len(got) != len(testCase.want) {
				t.Fatalf("tags = %v, want %v", got, testCase.want)
			}

			for k := range testCase.want {
				if _, ok := got[k]; !ok {
					t.Errorf("tags is missing key %q; got %v", k, got)
				}
			}
		})
	}
}

func testConnection(connectionID, region, ownerAccount string) awstypes.Connection {
	return awstypes.Connection{
		// AwsDeviceV2, not the deprecated AwsDevice: the model reads V2, so a
		// fixture that populates the old field flattens `aws_device` to null and
		// every assertion on it passes vacuously.
		AwsDeviceV2:     aws.String("EqDC2-abcdefgh"),
		Bandwidth:       aws.String("1Gbps"),
		ConnectionId:    aws.String(connectionID),
		ConnectionName:  aws.String("tf-acc-test-" + connectionID),
		ConnectionState: awstypes.ConnectionStateAvailable,
		Location:        aws.String("EqDC2"),
		OwnerAccount:    aws.String(ownerAccount),
		PartnerName:     aws.String("partner"),
		ProviderName:    aws.String("provider"),
		Region:          aws.String(region),
		Vlan:            1234,
	}
}

func testConnectionWithState(connectionID string, state awstypes.ConnectionState) awstypes.Connection {
	apiObject := testConnection(connectionID, "us-west-2", acctest.Ct12Digit)
	apiObject.ConnectionState = state

	return apiObject
}

// testFlattenConnections flattens apiObjects and returns the list elements.
func testFlattenConnections(ctx context.Context, t *testing.T, apiObjects []awstypes.Connection, partition string, ignoreTagsConfig *tftags.IgnoreConfig) []attr.Value {
	t.Helper()

	connections, diags := tfdirectconnect.FlattenConnections(ctx, apiObjects, partition, ignoreTagsConfig)
	if diags.HasError() {
		t.Fatalf("flattening connections: %v", diags)
	}

	return connections.Elements()
}

// testConnectionAttributes returns the attributes of a single flattened connection.
func testConnectionAttributes(ctx context.Context, t *testing.T, v attr.Value) map[string]attr.Value {
	t.Helper()

	valuable, ok := v.(basetypes.ObjectValuable)
	if !ok {
		t.Fatalf("element is a %T, want an object value", v)
	}

	object, diags := valuable.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("converting element to an object value: %v", diags)
	}

	return object.Attributes()
}

func testAttributeString(ctx context.Context, t *testing.T, attrs map[string]attr.Value, name string) string {
	t.Helper()

	v, ok := attrs[name]
	if !ok {
		t.Fatalf("missing %q attribute", name)
	}

	valuable, ok := v.(basetypes.StringValuable)
	if !ok {
		t.Fatalf("attribute %q is a %T, want a string value", name, v)
	}

	s, diags := valuable.ToStringValue(ctx)
	if diags.HasError() {
		t.Fatalf("converting attribute %q to a string value: %v", name, diags)
	}

	return s.ValueString()
}

func testAttributeMapElements(ctx context.Context, t *testing.T, v attr.Value) map[string]attr.Value {
	t.Helper()

	valuable, ok := v.(basetypes.MapValuable)
	if !ok {
		t.Fatalf("attribute is a %T, want a map value", v)
	}

	m, diags := valuable.ToMapValue(ctx)
	if diags.HasError() {
		t.Fatalf("converting attribute to a map value: %v", diags)
	}

	return m.Elements()
}

func TestAccDirectConnectConnectionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_dx_connections.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "connections.#", 2),
					testAccCheckConnectionsWellFormed(dataSourceName),
					// AC3: the per-item attributes match the singular data source
					// for the same connection.
					testAccCheckConnectionsElementAttrPair(dataSourceName, "data.aws_dx_connection.test1",
						names.AttrARN, names.AttrName, "bandwidth", names.AttrLocation, names.AttrOwnerAccountID, "aws_device"),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_dx_connections.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectionsWellFormed(dataSourceName),
					// Regression: fwflex.Flatten leaves tftags.Map null, so the
					// flattener sets tags by hand.
					testAccCheckConnectionsElementTag(dataSourceName, "aws_dx_connection.test", "Env", "prod"),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_dx_connections.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Reads a Region the rest of the suite does not create connections
				// in. The account is shared, so assert on shape rather than on an
				// exact count of zero: an empty list must be empty, not null, and
				// any connection that does exist must still be well-formed.
				Config: testAccConnectionsDataSourceConfig_alternateRegion(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "connections.#"),
					testAccCheckConnectionsWellFormed(dataSourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionsDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_dx_connections.test"
	alternateDataSourceName := "data.aws_dx_connections.alternate"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionsDataSourceConfig_region(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The connection is in the default Region, so it appears there
					// and must not appear in the `region`-overridden read.
					testAccCheckConnectionsElementAttrPair(dataSourceName, "data.aws_dx_connection.test", names.AttrARN),
					testAccCheckConnectionsElementAbsent(alternateDataSourceName, "data.aws_dx_connection.test"),
					testAccCheckConnectionsWellFormed(alternateDataSourceName),
				),
			},
		},
	})
}

// testAccCheckConnectionsWellFormed asserts that every element of the list has a
// populated id, arn, name, bandwidth, location and state, and that its ARN
// identifies that same element.
func testAccCheckConnectionsWellFormed(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attrs, err := testAccConnectionsDataSourceAttributes(s, n)
		if err != nil {
			return err
		}

		count, err := strconv.Atoi(attrs["connections.#"])
		if err != nil {
			return fmt.Errorf("%s: reading connections.#: %w", n, err)
		}

		for i := range count {
			prefix := fmt.Sprintf("connections.%d.", i)

			for _, name := range []string{names.AttrID, names.AttrARN, names.AttrName, "bandwidth", names.AttrLocation, names.AttrState} {
				if attrs[prefix+name] == "" {
					return fmt.Errorf("%s: %s%s is empty", n, prefix, name)
				}
			}

			if arn, want := attrs[prefix+names.AttrARN], "dxcon/"+attrs[prefix+names.AttrID]; !strings.HasSuffix(arn, want) {
				return fmt.Errorf("%s: %s%s = %q, want suffix %q", n, prefix, names.AttrARN, arn, want)
			}
		}

		return nil
	}
}

// testAccCheckConnectionsElementAttrPair finds the element of the list matching
// the connection identified by resourceName and compares the given attributes.
func testAccCheckConnectionsElementAttrPair(n, resourceName string, attrNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prefix, want, err := testAccConnectionsDataSourceElement(s, n, resourceName)
		if err != nil {
			return err
		}

		attrs, err := testAccConnectionsDataSourceAttributes(s, n)
		if err != nil {
			return err
		}

		for _, name := range attrNames {
			if got, want := attrs[prefix+name], want[name]; got != want {
				return fmt.Errorf("%s: %s%s = %q, want %q (from %s)", n, prefix, name, got, want, resourceName)
			}
		}

		return nil
	}
}

// testAccCheckConnectionsElementTag asserts that the element matching the
// connection identified by resourceName carries the given tag.
func testAccCheckConnectionsElementTag(n, resourceName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prefix, _, err := testAccConnectionsDataSourceElement(s, n, resourceName)
		if err != nil {
			return err
		}

		attrs, err := testAccConnectionsDataSourceAttributes(s, n)
		if err != nil {
			return err
		}

		// A null tags map has no `tags.%` element count at all.
		if _, ok := attrs[prefix+"tags.%"]; !ok {
			return fmt.Errorf("%s: %stags is null, want a map", n, prefix)
		}

		if got := attrs[prefix+"tags."+key]; got != value {
			return fmt.Errorf("%s: %stags.%s = %q, want %q", n, prefix, key, got, value)
		}

		return nil
	}
}

// testAccCheckConnectionsElementAbsent asserts that the connection identified by
// resourceName is not in the list.
func testAccCheckConnectionsElementAbsent(n, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, _, err := testAccConnectionsDataSourceElement(s, n, resourceName); err == nil {
			return fmt.Errorf("%s: unexpectedly contains the connection from %s", n, resourceName)
		}

		return nil
	}
}

// testAccConnectionsDataSourceElement returns the `connections.N.` attribute
// prefix of the element whose ID matches resourceName's, along with
// resourceName's attributes.
func testAccConnectionsDataSourceElement(s *terraform.State, n, resourceName string) (string, map[string]string, error) {
	attrs, err := testAccConnectionsDataSourceAttributes(s, n)
	if err != nil {
		return "", nil, err
	}

	want, err := testAccConnectionsDataSourceAttributes(s, resourceName)
	if err != nil {
		return "", nil, err
	}

	count, err := strconv.Atoi(attrs["connections.#"])
	if err != nil {
		return "", nil, fmt.Errorf("%s: reading connections.#: %w", n, err)
	}

	for i := range count {
		prefix := fmt.Sprintf("connections.%d.", i)

		if attrs[prefix+names.AttrID] == want[names.AttrID] {
			return prefix, want, nil
		}
	}

	return "", nil, fmt.Errorf("%s: no element with %s %q", n, names.AttrID, want[names.AttrID])
}

func testAccConnectionsDataSourceAttributes(s *terraform.State, n string) (map[string]string, error) {
	rs, ok := s.RootModule().Resources[n]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", n)
	}

	return rs.Primary.Attributes, nil
}

func testAccConnectionsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test1" {
  name      = "%[1]s-1"
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}

resource "aws_dx_connection" "test2" {
  name      = "%[1]s-2"
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}

data "aws_dx_connection" "test1" {
  name = aws_dx_connection.test1.name
}

data "aws_dx_connections" "test" {
  depends_on = [aws_dx_connection.test1, aws_dx_connection.test2]
}
`, rName)
}

func testAccConnectionsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]

  tags = {
    Env = "prod"
  }
}

data "aws_dx_connections" "test" {
  depends_on = [aws_dx_connection.test]
}
`, rName)
}

func testAccConnectionsDataSourceConfig_alternateRegion() string {
	return fmt.Sprintf(`
data "aws_dx_connections" "test" {
  region = %[1]q
}
`, acctest.AlternateRegion())
}

func testAccConnectionsDataSourceConfig_region(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}

data "aws_dx_connection" "test" {
  name = aws_dx_connection.test.name
}

data "aws_dx_connections" "test" {
  depends_on = [aws_dx_connection.test]
}

data "aws_dx_connections" "alternate" {
  region = %[2]q

  depends_on = [aws_dx_connection.test]
}
`, rName, acctest.AlternateRegion())
}
