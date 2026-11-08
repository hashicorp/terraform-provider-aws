// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	providerfw "github.com/hashicorp/terraform-provider-aws/internal/provider/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestVPCBlockPublicAccessOptions_failedCreateRegion(t *testing.T) {
	t.Parallel()

	for _, region := range []string{"us-west-2", "us-east-1"} {
		t.Run(region, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			failWait := true
			var modifications []string
			r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
				if err := req.ParseForm(); err != nil {
					return nil, err
				}
				selected := strings.Split(req.URL.Hostname(), ".")[1]
				action := req.Form.Get("Action")
				if action == "ModifyVpcBlockPublicAccessOptions" {
					modifications = append(modifications, selected+":"+req.Form.Get("InternetGatewayBlockMode"))
				}
				if selected != region {
					t.Errorf("unsafe %s request to %s for instance in %s", action, selected, region)
				}
				state := "update-complete"
				if failWait && action == "DescribeVpcBlockPublicAccessOptions" {
					state = "test-waiter-failure"
				}
				body := fmt.Sprintf(`<%sResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/"><vpcBlockPublicAccessOptions><awsAccountId>123456789012</awsAccountId><awsRegion>%s</awsRegion><internetGatewayBlockMode>block-bidirectional</internetGatewayBlockMode><state>%s</state></vpcBlockPublicAccessOptions></%sResponse>`, action, selected, state, action)
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/xml"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			plan := testVPCBPAState(t, s, map[string]types.String{
				names.AttrRegion:              types.StringValue(region),
				names.AttrID:                  types.StringUnknown(),
				names.AttrAWSAccountID:        types.StringUnknown(),
				"aws_region":                  types.StringUnknown(),
				"internet_gateway_block_mode": types.StringValue("block-bidirectional"),
			})
			created := resource.CreateResponse{State: testVPCBPAState(t, s, nil)}
			r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Raw: plan.Raw, Schema: s}, ProviderMeta: testVPCBPAMeta(ctx)}, &created)
			if !created.Diagnostics.HasError() || !strings.Contains(fmt.Sprint(created.Diagnostics), "test-waiter-failure") {
				t.Fatalf("expected original waiter error, got %s", created.Diagnostics)
			}
			if created.State.Raw.IsNull() {
				t.Fatal("failed Create must return non-null state for Terraform to taint")
			}
			if !created.State.Raw.IsFullyKnown() {
				t.Fatal("failed Create state must not contain planned unknowns")
			}
			for _, attribute := range []string{names.AttrID, names.AttrRegion} {
				var got types.String
				if diags := created.State.GetAttribute(ctx, path.Root(attribute), &got); diags.HasError() {
					t.Fatal(diags)
				}
				if got.ValueString() != region {
					t.Errorf("failed Create %s = %s, want %q", attribute, got, region)
				}
			}
			failWait = false
			serialized, err := tfprotov6.NewDynamicValue(s.Type().TerraformType(ctx), created.State.Raw)
			if err != nil {
				t.Fatal(err)
			}
			created.State.Raw, err = serialized.Unmarshal(s.Type().TerraformType(ctx))
			if err != nil {
				t.Fatal(err)
			}
			// A replacement with refresh disabled must also use the partial state's region.
			directDelete := resource.DeleteResponse{State: created.State}
			r.Delete(ctx, resource.DeleteRequest{State: created.State, ProviderMeta: testVPCBPAMeta(ctx)}, &directDelete)
			if directDelete.Diagnostics.HasError() {
				t.Fatal(directDelete.Diagnostics)
			}
			read := resource.ReadResponse{State: created.State}
			r.Read(ctx, resource.ReadRequest{State: created.State, ProviderMeta: testVPCBPAMeta(ctx)}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			deleted := resource.DeleteResponse{State: read.State}
			r.Delete(ctx, resource.DeleteRequest{State: read.State, ProviderMeta: testVPCBPAMeta(ctx)}, &deleted)
			if deleted.Diagnostics.HasError() {
				t.Fatal(deleted.Diagnostics)
			}
			if got, want := strings.Join(modifications, ","), region+":block-bidirectional,"+region+":off,"+region+":off"; got != want {
				t.Errorf("modify requests = %q, want %q", got, want)
			}
		})
	}
}

func TestVPCBlockPublicAccessOptions_unsafeRegion(t *testing.T) {
	t.Parallel()

	for name, attributes := range map[string]map[string]types.String{
		"partial state":     {names.AttrID: types.StringValue("us-west-2")},
		"mismatched region": {names.AttrID: types.StringValue("us-west-2"), names.AttrRegion: types.StringValue("us-east-1")},
		"empty region":      {names.AttrID: types.StringValue("us-west-2"), names.AttrRegion: types.StringValue("")},
		"unknown region":    {names.AttrID: types.StringValue("us-west-2"), names.AttrRegion: types.StringUnknown()},
		"missing ID":        {names.AttrRegion: types.StringValue("us-west-2")},
		"unknown ID":        {names.AttrID: types.StringUnknown(), names.AttrRegion: types.StringValue("us-west-2")},
		"null state":        nil,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			calls := 0
			r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
				calls++
				return nil, fmt.Errorf("unexpected API request to %s", req.URL.Hostname())
			})
			state := testVPCBPAState(t, s, attributes)
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state, ProviderMeta: testVPCBPAMeta(ctx)}, &read)
			if !read.Diagnostics.HasError() || !read.State.Raw.Equal(state.Raw) {
				t.Errorf("unsafe Read must fail without changing state: %s", read.Diagnostics)
			}
			if !state.Raw.IsNull() && !strings.Contains(fmt.Sprint(read.Diagnostics), "refusing to access") {
				t.Errorf("expected regional safety diagnostic, got %s", read.Diagnostics)
			}
			update := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Raw: state.Raw, Schema: s}, ProviderMeta: testVPCBPAMeta(ctx)}, &update)
			if !update.Diagnostics.HasError() {
				t.Error("unsafe Update must fail")
			}
			deleted := resource.DeleteResponse{State: state}
			r.Delete(ctx, resource.DeleteRequest{State: state, ProviderMeta: testVPCBPAMeta(ctx)}, &deleted)
			if !deleted.Diagnostics.HasError() {
				t.Error("unsafe Delete must fail")
			}
			if calls != 0 {
				t.Errorf("unsafe state sent %d API requests, want zero", calls)
			}
		})
	}
}

func TestVPCBlockPublicAccessOptions_lifecycle(t *testing.T) {
	t.Parallel()

	for _, region := range []string{"us-west-2", "us-east-1"} {
		t.Run(region, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			var modifications []string
			r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
				if err := req.ParseForm(); err != nil {
					return nil, err
				}
				selected := strings.Split(req.URL.Hostname(), ".")[1]
				if selected != region {
					t.Errorf("request to %s for instance in %s", selected, region)
				}
				if req.Form.Get("Action") == "ModifyVpcBlockPublicAccessOptions" {
					modifications = append(modifications, req.Form.Get("InternetGatewayBlockMode"))
				}
				return testVPCBPAResponse(req.Form.Get("Action"), selected), nil
			})
			attributes := map[string]types.String{"internet_gateway_block_mode": types.StringValue("block-bidirectional")}
			if region != "us-east-1" {
				attributes[names.AttrRegion] = types.StringValue(region)
			}
			config := testVPCBPAState(t, s, attributes)
			plan := resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: config.Raw, Schema: s}}
			r.(resource.ResourceWithModifyPlan).ModifyPlan(ctx, resource.ModifyPlanRequest{
				Config: tfsdk.Config{Raw: config.Raw, Schema: s}, Plan: plan.Plan,
				State: testVPCBPAState(t, s, nil), ProviderMeta: testVPCBPAMeta(ctx),
			}, &plan)
			if plan.Diagnostics.HasError() {
				t.Fatal(plan.Diagnostics)
			}
			created := resource.CreateResponse{State: testVPCBPAState(t, s, nil)}
			r.Create(ctx, resource.CreateRequest{Plan: plan.Plan, ProviderMeta: testVPCBPAMeta(ctx)}, &created)
			if created.Diagnostics.HasError() {
				t.Fatal(created.Diagnostics)
			}
			read := resource.ReadResponse{State: created.State}
			r.Read(ctx, resource.ReadRequest{State: created.State, ProviderMeta: testVPCBPAMeta(ctx)}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			updatePlan := tfsdk.Plan{Raw: read.State.Raw, Schema: s}
			if diags := updatePlan.SetAttribute(ctx, path.Root("internet_gateway_block_mode"), "block-ingress"); diags.HasError() {
				t.Fatal(diags)
			}
			updated := resource.UpdateResponse{State: read.State}
			r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: updatePlan, ProviderMeta: testVPCBPAMeta(ctx)}, &updated)
			if updated.Diagnostics.HasError() {
				t.Fatal(updated.Diagnostics)
			}
			deleted := resource.DeleteResponse{State: updated.State}
			r.Delete(ctx, resource.DeleteRequest{State: updated.State, ProviderMeta: testVPCBPAMeta(ctx)}, &deleted)
			if deleted.Diagnostics.HasError() {
				t.Fatal(deleted.Diagnostics)
			}
			if got := strings.Join(modifications, ","); got != "block-bidirectional,block-ingress,off" {
				t.Errorf("modify modes = %q", got)
			}
		})
	}
}

func TestVPCBlockPublicAccessOptions_createAPIError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("test modify failure")
	})
	plan := testVPCBPAState(t, s, map[string]types.String{names.AttrRegion: types.StringValue("us-west-2"), "internet_gateway_block_mode": types.StringValue("block-bidirectional")})
	response := resource.CreateResponse{State: testVPCBPAState(t, s, nil)}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Raw: plan.Raw, Schema: s}, ProviderMeta: testVPCBPAMeta(ctx)}, &response)
	if !response.Diagnostics.HasError() || !response.State.Raw.IsNull() {
		t.Fatalf("API failure must preserve null state and error: %s", response.Diagnostics)
	}
}

func TestVPCBlockPublicAccessOptions_importAndLegacyState(t *testing.T) {
	t.Parallel()

	for _, importID := range []string{"us-east-1", "us-west-2@us-west-2", "legacy"} {
		t.Run(importID, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
				if err := req.ParseForm(); err != nil {
					return nil, err
				}
				return testVPCBPAResponse(req.Form.Get("Action"), strings.Split(req.URL.Hostname(), ".")[1]), nil
			})
			state := testVPCBPAState(t, s, nil)
			if importID == "legacy" {
				state = testVPCBPAState(t, s, map[string]types.String{names.AttrID: types.StringValue("us-east-1")})
			} else {
				imported := resource.ImportStateResponse{State: state}
				r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: importID}, &imported)
				if imported.Diagnostics.HasError() {
					t.Fatal(imported.Diagnostics)
				}
				state = imported.State
			}
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state, ProviderMeta: testVPCBPAMeta(ctx)}, &read)
			if read.Diagnostics.HasError() {
				t.Fatal(read.Diagnostics)
			}
			var id, region types.String
			if diags := read.State.GetAttribute(ctx, path.Root(names.AttrID), &id); diags.HasError() {
				t.Fatal(diags)
			}
			if diags := read.State.GetAttribute(ctx, path.Root(names.AttrRegion), &region); diags.HasError() {
				t.Fatal(diags)
			}
			if id.IsNull() || !id.Equal(region) {
				t.Fatalf("ID %s and region %s must agree", id, region)
			}
		})
	}
}

func TestVPCBlockPublicAccessOptions_readRemoved(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/xml"}}, Body: io.NopCloser(strings.NewReader(`<DescribeVpcBlockPublicAccessOptionsResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/"/>`))}, nil
	})
	state := testVPCBPAState(t, s, map[string]types.String{names.AttrID: types.StringValue("us-west-2"), names.AttrRegion: types.StringValue("us-west-2")})
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state, ProviderMeta: testVPCBPAMeta(ctx)}, &response)
	if response.Diagnostics.HasError() || !response.State.Raw.IsNull() {
		t.Fatalf("Read must preserve resource removal: %s", response.Diagnostics)
	}
}

func TestVPCBlockPublicAccessOptions_regionalIsolation(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	var expectedRegion string
	r, s := testVPCBPAResource(t, func(req *http.Request) (*http.Response, error) {
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		selected := strings.Split(req.URL.Hostname(), ".")[1]
		if selected != expectedRegion {
			t.Errorf("request to %s for instance in %s", selected, expectedRegion)
		}
		return testVPCBPAResponse(req.Form.Get("Action"), selected), nil
	})
	states := map[string]tfsdk.State{}
	for _, region := range []string{"us-west-2", "us-east-1"} {
		expectedRegion = region
		plan := testVPCBPAState(t, s, map[string]types.String{names.AttrRegion: types.StringValue(region), "internet_gateway_block_mode": types.StringValue("block-bidirectional")})
		response := resource.CreateResponse{State: testVPCBPAState(t, s, nil)}
		r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Raw: plan.Raw, Schema: s}, ProviderMeta: testVPCBPAMeta(ctx)}, &response)
		if response.Diagnostics.HasError() {
			t.Fatal(response.Diagnostics)
		}
		states[region] = response.State
	}
	// Both regional clients now exist in the same provider's client cache.
	for _, region := range []string{"us-west-2", "us-east-1"} {
		expectedRegion = region
		response := resource.DeleteResponse{State: states[region]}
		r.Delete(ctx, resource.DeleteRequest{State: states[region], ProviderMeta: testVPCBPAMeta(ctx)}, &response)
		if response.Diagnostics.HasError() {
			t.Fatal(response.Diagnostics)
		}
	}
}

func testVPCBPAResponse(action, region string) *http.Response {
	body := fmt.Sprintf(`<%sResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/"><vpcBlockPublicAccessOptions><awsAccountId>123456789012</awsAccountId><awsRegion>%s</awsRegion><internetGatewayBlockMode>block-bidirectional</internetGatewayBlockMode><state>update-complete</state></vpcBlockPublicAccessOptions></%sResponse>`, action, region, action)
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/xml"}}, Body: io.NopCloser(strings.NewReader(body))}
}

type vpcBPATransport func(*http.Request) (*http.Response, error)

// RoundTrip handles mock requests without network access.
func (f vpcBPATransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type vpcBPAPrimary struct{ client *conns.AWSClient }

// Meta supplies the isolated client to the Framework provider.
func (p vpcBPAPrimary) Meta() any { return p.client }

func testVPCBPAResource(t *testing.T, transport vpcBPATransport) (resource.Resource, schema.Schema) {
	t.Helper()
	ctx := t.Context()
	client := new(conns.AWSClient)
	client.SetHTTPClient(ctx, &http.Client{Transport: transport})
	client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.EC2: &servicePackage{}})
	configFile := filepath.Join(t.TempDir(), "aws-config")
	if err := os.WriteFile(configFile, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := conns.Config{AccessKey: "test", SecretKey: "test", Region: "us-east-1", SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{configFile}, SharedCredentialsFiles: []string{configFile}}
	client, diags := cfg.ConfigureProvider(ctx, client)
	if diags.HasError() {
		t.Fatal(diags)
	}
	p, err := providerfw.NewProvider(ctx, vpcBPAPrimary{client})
	if err != nil {
		t.Fatal(err)
	}
	for _, factory := range p.Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{}, &metadata)
		if metadata.TypeName != "aws_vpc_block_public_access_options" {
			continue
		}
		var configured resource.ConfigureResponse
		r.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: client}, &configured)
		if configured.Diagnostics.HasError() {
			t.Fatal(configured.Diagnostics)
		}
		var response resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &response)
		if response.Diagnostics.HasError() {
			t.Fatal(response.Diagnostics)
		}
		return r, response.Schema
	}
	t.Fatal("resource not registered")
	return nil, schema.Schema{}
}

func testVPCBPAState(t *testing.T, s schema.Schema, attributes map[string]types.String) tfsdk.State {
	t.Helper()
	ctx := t.Context()
	state := tfsdk.State{Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil), Schema: s}
	for name, value := range attributes {
		if diags := state.SetAttribute(ctx, path.Root(name), value); diags.HasError() {
			t.Fatal(diags)
		}
	}
	return state
}

func testVPCBPAMeta(ctx context.Context) tfsdk.Config {
	s := schema.Schema{Attributes: map[string]schema.Attribute{"user_agent": schema.ListAttribute{ElementType: types.StringType, Optional: true}}}
	return tfsdk.Config{Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil), Schema: s}
}
