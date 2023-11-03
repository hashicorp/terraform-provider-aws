// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"maps"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/aws-sdk-go-base/v2/mockdata"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// TestSharedConfigFileParsing prevents regression in shared config file parsing
// * https://github.com/aws/aws-sdk-go-v2/issues/2349: indented keys
func TestSharedConfigFileParsing(t *testing.T) { //nolint:paralleltest
	testcases := map[string]struct {
		Config                  map[string]any
		SharedConfigurationFile string
		Check                   func(t *testing.T, meta *conns.AWSClient)
	}{
		"leading whitespace": {
			// Do not "fix" indentation!
			SharedConfigurationFile: `
	[default]
	region = us-west-2
	`, //lintignore:AWSAT003
			Check: func(t *testing.T, meta *conns.AWSClient) {
				//lintignore:AWSAT003
				if a, e := meta.Region, "us-west-2"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},
	}

	for name, tc := range testcases { //nolint:paralleltest
		tc := tc

		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			servicemocks.InitSessionTestEnv(t)

			config := map[string]any{
				"access_key":                  servicemocks.MockStaticAccessKey,
				"secret_key":                  servicemocks.MockStaticSecretKey,
				"skip_credentials_validation": true,
				"skip_requesting_account_id":  true,
			}

			if tc.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(tc.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				config["shared_config_files"] = []any{file.Name()}
			}

			rc := terraformsdk.NewResourceConfigRaw(config)

			p, err := New(ctx)
			if err != nil {
				t.Fatal(err)
			}

			var diags diag.Diagnostics
			diags = append(diags, p.Validate(rc)...)
			if diags.HasError() {
				t.Fatalf("validating: %s", sdkdiag.DiagnosticsString(diags))
			}

			diags = append(diags, p.Configure(ctx, rc)...)

			// The provider always returns a warning if there is no account ID
			var expected diag.Diagnostics
			expected = append(expected,
				errs.NewWarningDiagnostic(
					"AWS account ID not found for provider",
					"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications.",
				),
			)

			if diff := cmp.Diff(diags, expected, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			meta := p.Meta().(*conns.AWSClient)

			tc.Check(t, meta)
		})
	}
}

type DiagsValidator func(*testing.T, diag.Diagnostics)

func ExpectNoDiags(t *testing.T, diags diag.Diagnostics) {
	expectDiagsCount(t, diags, 0)
}

func expectDiagsCount(t *testing.T, diags diag.Diagnostics, c int) {
	if l := len(diags); l != c {
		t.Fatalf("Diagnostics: expected %d element, got %d\n%#v", c, l, diags)
	}
}

func TestProviderConfig_Authentication_SSO(t *testing.T) { //nolint:paralleltest
	const ssoSessionName = "test-sso-session"

	testCases := map[string]struct {
		config                     map[string]any
		SharedConfigurationFile    string
		SetSharedConfigurationFile bool
		ExpectedCredentialsValue   aws.Credentials
		ExpectedDiags              diag.Diagnostics
		MockStsEndpoints           []*servicemocks.MockEndpoint
	}{
		"shared configuration file": {
			config: map[string]any{},
			//lintignore:AWSAT003
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_session = %s
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1

[sso-session test-sso-session]
sso_region = us-east-1
sso_start_url = https://d-123456789a.awsapps.com/start
sso_registration_scopes = sso:account:access
`, ssoSessionName),
			SetSharedConfigurationFile: true,
			ExpectedCredentialsValue:   mockdata.MockSsoCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
	}

	for name, tc := range testCases { //nolint:paralleltest
		tc := tc

		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			ctx := context.TODO()

			// Populate required fields
			config := map[string]any{
				"region":                      "us-west-2", //lintignore:AWSAT003
				"skip_credentials_validation": true,
				"skip_requesting_account_id":  true,
			}

			// The provider always returns a warning if there is no account ID
			expectedDiags := tc.ExpectedDiags
			expectedDiags = append(expectedDiags,
				errs.NewWarningDiagnostic(
					"AWS account ID not found for provider",
					"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications.",
				),
			)

			err := servicemocks.SsoTestSetup(t, ssoSessionName)
			if err != nil {
				t.Fatalf("setup: %s", err)
			}

			endpoints := []any{}

			closeSso, ssoEndpoint := servicemocks.SsoCredentialsApiMock()
			defer closeSso()
			endpoints = append(endpoints, map[string]any{
				"sso": ssoEndpoint,
			})

			ts := servicemocks.MockAwsApiServer("STS", tc.MockStsEndpoints)
			defer ts.Close()
			endpoints = append(endpoints, map[string]any{
				"sts": ts.URL,
			})

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			if tc.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(tc.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				tc.config["shared_config_files"] = []any{file.Name()}
			}

			tc.config["skip_credentials_validation"] = true
			tc.config["skip_requesting_account_id"] = true

			tc.config["endpoints"] = endpoints

			maps.Copy(config, tc.config)

			p, err := New(ctx)
			if err != nil {
				t.Fatal(err)
			}

			var diags diag.Diagnostics

			rc := terraformsdk.NewResourceConfigRaw(config)

			diags = append(diags, p.Validate(rc)...)
			if diags.HasError() {
				t.Fatalf("validating: %s", sdkdiag.DiagnosticsString(diags))
			}

			diags = append(diags, p.Configure(ctx, rc)...)

			if diff := cmp.Diff(diags, expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			meta := p.Meta().(*conns.AWSClient)

			credentials, err := meta.CredentialsProvider().Retrieve(ctx)
			if err != nil {
				t.Fatalf("Error when requesting credentials: %s", err)
			}

			if diff := cmp.Diff(credentials, tc.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}

func TestProviderConfig_Authentication_LegacySSO(t *testing.T) { //nolint:paralleltest
	const ssoStartUrl = "https://d-123456789a.awsapps.com/start"

	testCases := map[string]struct {
		config                     map[string]any
		SharedConfigurationFile    string
		SetSharedConfigurationFile bool
		ExpectedCredentialsValue   aws.Credentials
		ExpectedDiags              diag.Diagnostics
		MockStsEndpoints           []*servicemocks.MockEndpoint
	}{
		"shared configuration file": {
			config: map[string]any{},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_start_url = %s
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1
`, ssoStartUrl),
			SetSharedConfigurationFile: true,
			ExpectedCredentialsValue:   mockdata.MockSsoCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
	}

	for name, tc := range testCases { //nolint:paralleltest
		tc := tc

		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			ctx := context.TODO()

			// Populate required fields
			config := map[string]any{
				"region":                      "us-west-2", //lintignore:AWSAT003
				"skip_credentials_validation": true,
				"skip_requesting_account_id":  true,
			}

			// The provider always returns a warning if there is no account ID
			expectedDiags := tc.ExpectedDiags
			expectedDiags = append(expectedDiags,
				errs.NewWarningDiagnostic(
					"AWS account ID not found for provider",
					"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications.",
				),
			)

			err := servicemocks.SsoTestSetup(t, ssoStartUrl)
			if err != nil {
				t.Fatalf("setup: %s", err)
			}

			endpoints := []any{}

			closeSso, ssoEndpoint := servicemocks.SsoCredentialsApiMock()
			defer closeSso()
			endpoints = append(endpoints, map[string]any{
				"sso": ssoEndpoint,
			})

			ts := servicemocks.MockAwsApiServer("STS", tc.MockStsEndpoints)
			defer ts.Close()
			endpoints = append(endpoints, map[string]any{
				"sts": ts.URL,
			})

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			if tc.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(tc.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				tc.config["shared_config_files"] = []any{file.Name()}
			}

			tc.config["skip_credentials_validation"] = true
			tc.config["skip_requesting_account_id"] = true

			tc.config["endpoints"] = endpoints

			maps.Copy(config, tc.config)

			p, err := New(ctx)
			if err != nil {
				t.Fatal(err)
			}

			var diags diag.Diagnostics

			rc := terraformsdk.NewResourceConfigRaw(config)

			diags = append(diags, p.Validate(rc)...)
			if diags.HasError() {
				t.Fatalf("validating: %s", sdkdiag.DiagnosticsString(diags))
			}

			diags = append(diags, p.Configure(ctx, rc)...)

			if diff := cmp.Diff(diags, expectedDiags, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			meta := p.Meta().(*conns.AWSClient)

			credentials, err := meta.CredentialsProvider().Retrieve(ctx)
			if err != nil {
				t.Fatalf("Error when requesting credentials: %s", err)
			}

			if diff := cmp.Diff(credentials, tc.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}
