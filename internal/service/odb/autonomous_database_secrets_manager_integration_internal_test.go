// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/identity"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/importer"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const testAutonomousDatabaseAccountID = "123456789012" // nosemgrep:ci.literal-12Digit-string-test-constant -- same-package resource tests cannot import acctest without a provider import cycle

const testAutonomousDatabaseServiceRoleARN = "arn:aws:iam::123456789012:role/ADBSSecretManagerServiceRole-123456789012" //lintignore:AWSAT005

func TestAutonomousDatabaseSecretsManagerIntegrationCreateState(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		failure        string
		regionOverride string
	}{
		"success":                               {},
		"success with region override":          {regionOverride: "us-west-2"},
		"initialization failed":                 {failure: "initialize"},
		"status API failed":                     {failure: "status API"},
		"provisioning failed":                   {failure: "provisioning"},
		"waiter timed out":                      {failure: names.AttrTimeout},
		"waiter timed out with region override": {failure: names.AttrTimeout, regionOverride: "us-west-2"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				wantRegion := "us-east-1"
				if tc.regionOverride != "" {
					wantRegion = tc.regionOverride
					ctx = conns.NewResourceContext(ctx, "", "", "aws_odb_autonomous_database_secrets_manager_integration", wantRegion)
				}
				initializations, polls := 0, 0
				client := new(conns.AWSClient)
				client.SetHTTPClient(ctx, &http.Client{Transport: autonomousDatabaseIntegrationTransport(func(req *http.Request) (*http.Response, error) {
					code := http.StatusOK
					body := `{}`
					if !strings.Contains(req.URL.Host, wantRegion) {
						t.Errorf("request endpoint %s does not use region %s", req.URL.Host, wantRegion)
					}
					switch req.Header.Get("X-Amz-Target") {
					case "Odb.InitializeService":
						initializations++
						if tc.failure == "initialize" {
							code = http.StatusBadRequest
							body = `{"__type":"ValidationException","message":"initialization failed"}`
						}
					case "Odb.GetOciOnboardingStatus":
						polls++
						status := odbtypes.OciIamRoleStatusAvailable
						switch tc.failure {
						case "status API":
							code = http.StatusBadRequest
							body = `{"__type":"ValidationException","message":"status lookup failed"}`
						case "provisioning":
							status = odbtypes.OciIamRoleStatusProvisionFailed
						case names.AttrTimeout:
							status = odbtypes.OciIamRoleStatusProvisioning
						}
						if code == http.StatusOK {
							body = fmt.Sprintf(`{"autonomousDatabaseOciIntegrationIamRoles":[{"awsIntegration":"SecretsManager","iamRoleArn":%q,"status":%q,"statusReason":"service role status"}]}`, testAutonomousDatabaseServiceRoleARN, status)
						}
					default:
						t.Fatalf("unexpected operation %q", req.Header.Get("X-Amz-Target"))
					}
					return &http.Response{StatusCode: code, Header: http.Header{"Content-Type": {"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
				})})
				client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.ODB: &servicePackage{}})
				config := conns.Config{AccessKey: "test", SecretKey: "test", Region: "us-east-1", SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{}}
				client, diags := config.ConfigureProvider(ctx, client)
				if diags.HasError() {
					t.Fatal(diags)
				}
				rawResource, err := newResourceAutonomousDatabaseSecretsManagerIntegration(ctx)
				if err != nil {
					t.Fatal(err)
				}
				r := rawResource.(*resourceAutonomousDatabaseSecretsManagerIntegration)
				r.SetDefaultCreateTimeout(time.Second)
				r.Configure(ctx, resource.ConfigureRequest{ProviderData: client}, &resource.ConfigureResponse{})
				var schemaResponse resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
				schemaResponse.Schema.Attributes[names.AttrRegion] = resourceattribute.Region()
				terraformType := schemaResponse.Schema.Type().TerraformType(ctx)
				planValues := map[string]tftypes.Value{}
				for name, attributeType := range terraformType.(tftypes.Object).AttributeTypes {
					planValues[name] = tftypes.NewValue(attributeType, nil)
				}
				planValues[names.AttrRegion] = tftypes.NewValue(tftypes.String, wantRegion)
				request := resource.CreateRequest{Plan: tfsdk.Plan{Raw: tftypes.NewValue(terraformType, planValues), Schema: schemaResponse.Schema}}
				response := resource.CreateResponse{State: tfsdk.State{Raw: tftypes.NewValue(terraformType, nil), Schema: schemaResponse.Schema}}
				r.Create(ctx, request, &response)
				if got, want := response.Diagnostics.HasError(), tc.failure != ""; got != want {
					t.Fatalf("diagnostics.HasError() = %t, want %t: %v", got, want, response.Diagnostics)
				}
				if initializations != 1 {
					t.Errorf("initialization calls = %d, want 1", initializations)
				}
				if tc.failure == "initialize" {
					if !response.State.Raw.IsNull() || polls != 0 {
						t.Fatalf("failed initialization must leave null state and no polls; state = %v, polls = %d", response.State.Raw, polls)
					}
					return
				}
				if response.State.Raw.IsNull() {
					t.Fatal("successful initialization must preserve state even when the waiter fails")
				}
				for _, name := range []string{names.AttrID, names.AttrRegion} {
					var got types.String
					if diags := response.State.GetAttribute(ctx, path.Root(name), &got); diags.HasError() {
						t.Fatal(diags)
					}
					if got.ValueString() != wantRegion {
						t.Errorf("state %s = %q, want %q", name, got.ValueString(), wantRegion)
					}
				}
				if tc.failure == "" {
					var status fwtypes.StringEnum[odbtypes.OciIamRoleStatus]
					if diags := response.State.GetAttribute(ctx, path.Root(names.AttrStatus), &status); diags.HasError() {
						t.Fatal(diags)
					}
					if status.ValueString() != string(odbtypes.OciIamRoleStatusAvailable) {
						t.Errorf("state status = %q, want AVAILABLE", status.ValueString())
					}
				}
			})
		})
	}
}

type autonomousDatabaseIntegrationTransport func(*http.Request) (*http.Response, error)

func (f autonomousDatabaseIntegrationTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAutonomousDatabaseSecretsManagerIntegrationImport(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id          string
		identity    map[string]string
		stateRegion string
		wantRegion  string
		wantError   string
	}{
		"regional ID":                           {id: "us-east-1", wantRegion: "us-east-1"},
		"regional ID override":                  {id: "us-west-2", stateRegion: "us-west-2", wantRegion: "us-west-2"},
		"mismatched regional ID":                {id: "us-west-2", stateRegion: "us-east-1", wantError: "does not match"},
		"unreleased literal is not regional ID": {id: "secrets-manager", stateRegion: "us-east-1", wantError: "does not match"},
		// nosemgrep:ci.semgrep.acctest.naming.attr-names-as-test-names -- identity attribute keys, not test names
		"identity explicit":        {identity: map[string]string{names.AttrAccountID: testAutonomousDatabaseAccountID, names.AttrRegion: "us-east-1"}, wantRegion: "us-east-1"},
		"identity defaults":        {identity: map[string]string{}, wantRegion: "us-east-1"},
		"identity region override": {identity: map[string]string{names.AttrRegion: "us-west-2"}, wantRegion: "us-west-2"},
		"identity account only":    {identity: map[string]string{names.AttrAccountID: testAutonomousDatabaseAccountID}, wantRegion: "us-east-1"},
		"identity wrong account":   {identity: map[string]string{names.AttrAccountID: "987654321098"}, wantError: "cannot be used to import resources from account"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := importer.Context(t.Context(), autonomousDatabaseIntegrationImportClient{})
			var registration *inttypes.ServicePackageFrameworkResource
			for _, candidate := range (&servicePackage{}).FrameworkResources(ctx) {
				if candidate.TypeName == "aws_odb_autonomous_database_secrets_manager_integration" {
					registration = candidate
					break
				}
			}
			if registration == nil || !registration.Identity.IsSingleton || registration.Identity.IsGlobalResource {
				t.Fatal("expected registered regional singleton identity")
			}
			r, err := registration.Factory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			identityResource, ok := r.(framework.ImportByIdentityer)
			if !ok {
				t.Fatal("resource does not support identity import")
			}
			identityResource.SetIdentitySpec(registration.Identity)
			identityResource.SetImportSpec(registration.Import)
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			schemaResponse.Schema.Attributes[names.AttrRegion] = resourceattribute.Region()
			identitySchema := identity.NewIdentitySchema(registration.Identity)
			identityValues := map[string]tftypes.Value{}
			for name := range identitySchema.Attributes {
				value := tftypes.NewValue(tftypes.String, nil)
				if v, ok := tc.identity[name]; ok {
					value = tftypes.NewValue(tftypes.String, v)
				}
				identityValues[name] = value
			}
			identityValue := &tfsdk.ResourceIdentity{
				Raw:    tftypes.NewValue(identitySchema.Type().TerraformType(ctx), identityValues),
				Schema: &identitySchema,
			}
			response := resource.ImportStateResponse{
				State: tfsdk.State{
					Raw:    tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil),
					Schema: schemaResponse.Schema,
				},
				Identity: identityValue,
			}
			if tc.stateRegion != "" {
				if diags := response.State.SetAttribute(ctx, path.Root(names.AttrRegion), tc.stateRegion); diags.HasError() {
					t.Fatalf("setting state region: %v", diags)
				}
			}
			r.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: tc.id, Identity: identityValue}, &response)
			if tc.wantError != "" {
				if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics[0].Detail(), tc.wantError) {
					t.Fatalf("expected error containing %q, got %v", tc.wantError, response.Diagnostics)
				}
				return
			}
			if response.Diagnostics.HasError() {
				t.Fatalf("import failed: %v", response.Diagnostics)
			}
			for _, name := range []string{names.AttrID, names.AttrRegion} {
				var value types.String
				if diags := response.State.GetAttribute(ctx, path.Root(name), &value); diags.HasError() {
					t.Fatal(diags)
				}
				if value.ValueString() != tc.wantRegion {
					t.Errorf("state %s = %q, want %q", name, value.ValueString(), tc.wantRegion)
				}
			}
			// nosemgrep:ci.semgrep.acctest.naming.attr-names-as-test-names -- identity attribute keys, not test names
			for name, want := range map[string]string{names.AttrAccountID: testAutonomousDatabaseAccountID, names.AttrRegion: tc.wantRegion} {
				var value types.String
				if diags := response.Identity.GetAttribute(ctx, path.Root(name), &value); diags.HasError() {
					t.Fatal(diags)
				}
				if value.ValueString() != want {
					t.Errorf("identity %s = %q, want %q", name, value.ValueString(), want)
				}
			}
		})
	}
}

type autonomousDatabaseIntegrationImportClient struct{}

func (autonomousDatabaseIntegrationImportClient) AccountID(context.Context) string {
	return testAutonomousDatabaseAccountID
}

func (autonomousDatabaseIntegrationImportClient) Region(context.Context) string {
	return "us-east-1"
}

func TestAutonomousDatabaseSecretsManagerIntegrationSchema(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	rawResource, err := newResourceAutonomousDatabaseSecretsManagerIntegration(ctx)
	if err != nil {
		t.Fatalf("creating resource: %s", err)
	}

	response := resource.SchemaResponse{}
	rawResource.Schema(ctx, resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("creating resource schema: %v", response.Diagnostics)
	}
	if diagnostics := response.Schema.ValidateImplementation(ctx); diagnostics.HasError() {
		t.Fatalf("validating resource schema: %v", diagnostics)
	}
}

func TestAutonomousDatabaseSecretsManagerIntegrationWaiters(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		deleting   bool
		statuses   []odbtypes.OciIamRoleStatus
		wantState  odbtypes.OciIamRoleStatus
		wantError  string
		omitReason bool
	}{
		"create immediately available": {
			statuses: []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusAvailable},
		},
		"create eventually visible": {
			statuses: []odbtypes.OciIamRoleStatus{"", "", odbtypes.OciIamRoleStatusProvisioning, odbtypes.OciIamRoleStatusAvailable},
		},
		"create provisioning": {
			statuses: []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusProvisioning, odbtypes.OciIamRoleStatusAvailable},
		},
		"delete eventually terminating": {
			deleting: true,
			statuses: []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusAvailable, odbtypes.OciIamRoleStatusAvailable, odbtypes.OciIamRoleStatusTerminating, ""},
		},
		"delete while provisioning": {
			deleting: true,
			statuses: []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusProvisioning, odbtypes.OciIamRoleStatusTerminating, ""},
		},
		"delete already absent": {
			deleting: true,
			statuses: []odbtypes.OciIamRoleStatus{""},
		},
		"create provision failed": {
			statuses:  []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusProvisionFailed},
			wantState: odbtypes.OciIamRoleStatusProvisionFailed,
			wantError: "service role permissions are missing",
		},
		"create terminate failed": {
			statuses:  []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusTerminateFailed},
			wantState: odbtypes.OciIamRoleStatusTerminateFailed,
			wantError: "service role permissions are missing",
		},
		"delete terminate failed": {
			deleting:  true,
			statuses:  []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusTerminateFailed},
			wantState: odbtypes.OciIamRoleStatusTerminateFailed,
			wantError: "service role permissions are missing",
		},
		"delete provision failed": {
			deleting:  true,
			statuses:  []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusProvisionFailed},
			wantState: odbtypes.OciIamRoleStatusProvisionFailed,
			wantError: "service role permissions are missing",
		},
		"create failed without reason": {
			statuses:   []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusProvisionFailed},
			wantState:  odbtypes.OciIamRoleStatusProvisionFailed,
			wantError:  "unexpected state",
			omitReason: true,
		},
		"delete failed without reason": {
			deleting:   true,
			statuses:   []odbtypes.OciIamRoleStatus{odbtypes.OciIamRoleStatusTerminateFailed},
			wantState:  odbtypes.OciIamRoleStatusTerminateFailed,
			wantError:  "unexpected state",
			omitReason: true,
		},
		"create unexpected state": {
			statuses:  []odbtypes.OciIamRoleStatus{"UNKNOWN"},
			wantState: "UNKNOWN",
			wantError: "unexpected state",
		},
		"delete unexpected state": {
			deleting:  true,
			statuses:  []odbtypes.OciIamRoleStatus{"UNKNOWN"},
			wantState: "UNKNOWN",
			wantError: "unexpected state",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				conn := odb.New(odb.Options{
					Region:      "us-east-1",
					Credentials: aws.AnonymousCredentials{},
					HTTPClient: smithyhttp.ClientDoFunc(func(req *http.Request) (*http.Response, error) {
						if calls >= len(tc.statuses) {
							t.Fatalf("unexpected poll %d", calls+1)
						}
						status := tc.statuses[calls]
						calls++
						body := `{}`
						if status != "" {
							body = fmt.Sprintf(`{"autonomousDatabaseOciIntegrationIamRoles":[{"awsIntegration":"SecretsManager","iamRoleArn":%q,"status":%q,"statusReason":"service role permissions are missing"}]}`, testAutonomousDatabaseServiceRoleARN, status)
							if tc.omitReason {
								body = strings.Replace(body, `,"statusReason":"service role permissions are missing"`, "", 1)
							}
						}
						return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
					}),
				})

				var err error
				if tc.deleting {
					err = waitAutonomousDatabaseSecretsManagerIntegrationDeleted(t.Context(), conn, time.Minute)
				} else {
					var role *odbtypes.OciIamRole
					role, err = waitAutonomousDatabaseSecretsManagerIntegrationCreated(t.Context(), conn, time.Minute)
					if tc.wantError == "" && (role == nil || role.Status != odbtypes.OciIamRoleStatusAvailable) {
						t.Errorf("expected available role, got %#v", role)
					}
				}
				if tc.wantError != "" {
					var unexpected *retry.UnexpectedStateError
					if !errors.As(err, &unexpected) || unexpected.State != string(tc.wantState) {
						t.Fatalf("expected unexpected state %q, got %v", tc.wantState, err)
					}
					if !strings.Contains(err.Error(), tc.wantError) {
						t.Errorf("expected error containing %q, got %v", tc.wantError, err)
					}
				} else if err != nil {
					t.Fatalf("unexpected waiter error: %v", err)
				}
				if calls != len(tc.statuses) {
					t.Errorf("polls = %d, want %d", calls, len(tc.statuses))
				}
			})
		})
	}
}

func TestAutonomousDatabaseSecretsManagerIntegrationWaiterErrors(t *testing.T) {
	t.Parallel()

	for _, deleting := range []bool{false, true} {
		for _, failure := range []string{"API error", "transport error", "cancellation", names.AttrTimeout, "not found limit"} {
			if deleting && failure == "not found limit" {
				continue
			}
			t.Run(fmt.Sprintf("deleting=%t/%s", deleting, failure), func(t *testing.T) {
				t.Parallel()
				synctest.Test(t, func(t *testing.T) {
					ctx, cancel := context.WithCancel(t.Context())
					defer cancel()
					transportErr := errors.New("test transport failure")
					calls := 0
					conn := odb.New(odb.Options{
						Region:      "us-east-1",
						Credentials: aws.AnonymousCredentials{},
						Retryer:     aws.NopRetryer{},
						HTTPClient: smithyhttp.ClientDoFunc(func(req *http.Request) (*http.Response, error) {
							calls++
							statusCode := http.StatusOK
							body := `{"autonomousDatabaseOciIntegrationIamRoles":[{"awsIntegration":"SecretsManager","status":"PROVISIONING"}]}`
							switch failure {
							case "API error":
								statusCode = http.StatusBadRequest
								body = `{"__type":"ValidationException","message":"test API error"}`
							case "transport error":
								return nil, transportErr
							case "cancellation":
								cancel()
								return nil, req.Context().Err()
							case "not found limit":
								body = `{}`
							}
							return &http.Response{StatusCode: statusCode, Header: http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
						}),
					})
					timeout := 10 * time.Minute
					if failure == names.AttrTimeout {
						timeout = time.Second
					}
					var err error
					if deleting {
						err = waitAutonomousDatabaseSecretsManagerIntegrationDeleted(ctx, conn, timeout)
					} else {
						_, err = waitAutonomousDatabaseSecretsManagerIntegrationCreated(ctx, conn, timeout)
					}
					switch failure {
					case "API error":
						var apiError smithy.APIError
						if !errors.As(err, &apiError) || apiError.ErrorCode() != "ValidationException" || apiError.ErrorMessage() != "test API error" {
							t.Fatalf("expected API error preserved, got %v", err)
						}
					case "transport error":
						if !errors.Is(err, transportErr) {
							t.Fatalf("expected transport error preserved, got %v", err)
						}
					case "cancellation":
						if !errors.Is(err, context.Canceled) {
							t.Fatalf("expected cancellation, got %v", err)
						}
					case names.AttrTimeout:
						var timeoutError *retry.TimeoutError
						if !errors.As(err, &timeoutError) || timeoutError.LastState != string(odbtypes.OciIamRoleStatusProvisioning) {
							t.Fatalf("expected timeout in PROVISIONING, got %v", err)
						}
					case "not found limit":
						if !retry.NotFound(err) || calls != 21 {
							t.Fatalf("expected not found after 21 polls, got %v after %d polls", err, calls)
						}
					}
				})
			})
		}
	}
}

func TestAutonomousDatabaseSecretsManagerIntegrationStatus(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		body       string
		wantStatus string
	}{
		"missing roles":          {body: `{}`},
		"empty roles":            {body: `{"autonomousDatabaseOciIntegrationIamRoles":[]}`},
		"other integration only": {body: `{"autonomousDatabaseOciIntegrationIamRoles":[{"awsIntegration":"KmsTde","status":"AVAILABLE"}]}`},
		"selects matching integration": {
			body:       `{"autonomousDatabaseOciIntegrationIamRoles":[{"awsIntegration":"KmsTde","status":"PROVISIONING"},{"awsIntegration":"SecretsManager","status":"AVAILABLE"}]}`,
			wantStatus: string(odbtypes.OciIamRoleStatusAvailable),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			conn := odb.New(odb.Options{
				Region:      "us-east-1",
				Credentials: aws.AnonymousCredentials{},
				HTTPClient: smithyhttp.ClientDoFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: req}, nil
				}),
			})
			role, status, err := statusAutonomousDatabaseSecretsManagerIntegration(conn)(t.Context())
			if err != nil || status != tc.wantStatus {
				t.Fatalf("status = %q, error = %v; want %q", status, err, tc.wantStatus)
			}
			if tc.wantStatus == "" && role != nil {
				t.Fatalf("missing integration must return nil for waiter not-found polling, got %#v", role)
			}
			if tc.wantStatus != "" && role == nil {
				t.Fatal("expected matching integration")
			}
		})
	}
}

func TestFlattenAutonomousDatabaseSecretsManagerIntegration(t *testing.T) {
	t.Parallel()

	model := autonomousDatabaseSecretsManagerIntegrationResourceModel{}
	role := &odbtypes.OciIamRole{
		IamRoleArn:   aws.String(testAutonomousDatabaseServiceRoleARN),
		Status:       odbtypes.OciIamRoleStatusAvailable,
		StatusReason: aws.String("available"),
	}

	flattenAutonomousDatabaseSecretsManagerIntegration(role, &model)

	if got, want := model.RoleARN.ValueString(), aws.ToString(role.IamRoleArn); got != want {
		t.Fatalf("role_arn = %q, want %q", got, want)
	}
	if got, want := model.Status.ValueString(), string(odbtypes.OciIamRoleStatusAvailable); got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
	if got, want := model.StatusReason.ValueString(), aws.ToString(role.StatusReason); got != want {
		t.Fatalf("status_reason = %q, want %q", got, want)
	}
}
