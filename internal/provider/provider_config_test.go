// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/google/go-cmp/cmp"
	configtesting "github.com/hashicorp/aws-sdk-go-base/v2/configtesting"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"golang.org/x/exp/maps"
)

// TestSharedConfigFileParsing prevents regression in shared config file parsing
// * https://github.com/aws/aws-sdk-go-v2/issues/2349: indented keys
// * https://github.com/aws/aws-sdk-go-v2/issues/2363: leading whitespace
// * https://github.com/aws/aws-sdk-go-v2/issues/2369: trailing `#` in, e.g. SSO Start URLs
func TestSharedConfigFileParsing(t *testing.T) { //nolint:paralleltest
	testcases := map[string]struct {
		Config                  map[string]any
		SharedConfigurationFile string
		Check                   func(t *testing.T, meta *conns.AWSClient)
	}{
		"leading newline": {
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

		"leading whitespace": {
			// Do not "fix" indentation!
			SharedConfigurationFile: `	[default]
	region = us-west-2
	`, //lintignore:AWSAT003
			Check: func(t *testing.T, meta *conns.AWSClient) {
				//lintignore:AWSAT003
				if a, e := meta.Region, "us-west-2"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"leading newline and whitespace": {
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

		"named profile after leading newline and whitespace": {
			Config: map[string]any{
				"profile": "test",
			},
			// Do not "fix" indentation!
			SharedConfigurationFile: `
[default]
region = us-west-2

	[profile test]
	region = us-east-1
			`, //lintignore:AWSAT003
			Check: func(t *testing.T, meta *conns.AWSClient) {
				//lintignore:AWSAT003
				if a, e := meta.Region, "us-east-1"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"named profile": {
			Config: map[string]any{
				"profile": "test",
			},
			SharedConfigurationFile: `
[default]
region = us-west-2

[profile test]
region = us-east-1
`, //lintignore:AWSAT003
			Check: func(t *testing.T, meta *conns.AWSClient) {
				//lintignore:AWSAT003
				if a, e := meta.Region, "us-east-1"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"trailing hash": {
			SharedConfigurationFile: `
[default]
region = us-west-2
sso_start_url = https://d-123456789a.awsapps.com/start#
`, //lintignore:AWSAT003
			Check: func(t *testing.T, meta *conns.AWSClient) {
				awsConfig := meta.AwsConfig()
				var ssoStartUrl string
				for _, source := range awsConfig.ConfigSources {
					if shared, ok := source.(config.SharedConfig); ok {
						ssoStartUrl = shared.SSOStartURL
					}
				}
				if a, e := ssoStartUrl, "https://d-123456789a.awsapps.com/start#"; a != e {
					t.Errorf("expected sso_start_url %q, got %q", e, a)
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

			maps.Copy(config, tc.Config)

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

			if diags.HasError() {
				t.FailNow()
			}

			meta := p.Meta().(*conns.AWSClient)

			tc.Check(t, meta)
		})
	}
}

type DiagsValidator func(*testing.T, diag.Diagnostics)

var _ configtesting.TestDriver = &testDriver{}

type testDriver struct {
	mode configtesting.TestMode
}

func (t *testDriver) Init(mode configtesting.TestMode) {
	t.mode = mode
}

func (t testDriver) TestCase() configtesting.TestCaseDriver {
	return &testCaseDriver{
		mode: t.mode,
	}
}

var _ configtesting.TestCaseDriver = &testCaseDriver{}

type testCaseDriver struct {
	mode   configtesting.TestMode
	config configurer
}

func (d *testCaseDriver) Configuration(fs []configtesting.ConfigFunc) configtesting.Configurer {
	conf := d.configuration()
	for _, f := range fs {
		f(conf)
	}
	return conf
}

func (d *testCaseDriver) configuration() *configurer {
	if d.config == nil {
		d.config = make(configurer, 0)
	}
	return &d.config
}

func (d *testCaseDriver) Setup(t *testing.T) {
	ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
		servicemocks.MockStsGetCallerIdentityValidEndpoint,
	})
	t.Cleanup(func() {
		ts.Close()
	})
	d.config.AddEndpoint("sts", ts.URL)
}

func (d testCaseDriver) Apply(ctx context.Context, t *testing.T) (context.Context, configtesting.Thing) {
	t.Helper()

	// Populate required fields
	d.config.SetRegion("us-west-2") // lintignore:AWSAT003,
	if d.mode == configtesting.TestModeLocal {
		d.config.SetSkipCredsValidation(true)
		d.config.SetSkipRequestingAccountId(true)
	}

	config := map[string]any(d.config)

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
	if d.mode == configtesting.TestModeLocal {
		expected = append(expected,
			errs.NewWarningDiagnostic(
				"AWS account ID not found for provider",
				"See https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id for implications.",
			),
		)
	}

	if diff := cmp.Diff(diags, expected, cmp.Comparer(sdkdiag.Comparer)); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	meta := p.Meta().(*conns.AWSClient)

	return ctx, thing{meta.CredentialsProvider()}
}

var _ configtesting.Configurer = &configurer{}

type configurer map[string]any

func (c configurer) AddEndpoint(k, v string) {
	if endpoints, ok := c["endpoints"]; ok {
		l := endpoints.([]any)
		m := l[0].(map[string]any)
		m[k] = v
	} else {
		c["endpoints"] = []any{
			map[string]any{
				k: v,
			},
		}
	}
}

func (c configurer) AddSharedConfigFile(f string) {
	x := c["shared_config_files"]
	if x == nil {
		c["shared_config_files"] = []any{f}
	} else {
		files := x.([]any)
		files = append(files, f)
		c["shared_config_files"] = files
	}
}

func (c configurer) SetAccessKey(s string) {
	c["access_key"] = s
}

func (c configurer) SetSecretKey(s string) {
	c["secret_key"] = s
}

func (c configurer) SetProfile(s string) {
	c["profile"] = s
}

func (c configurer) SetRegion(s string) {
	c["region"] = s
}

func (c configurer) SetUseFIPSEndpoint(b bool) {
	c["use_fips_endpoint"] = b
}

func (c configurer) SetSkipCredsValidation(b bool) {
	c["skip_credentials_validation"] = b
}

func (c configurer) SetSkipRequestingAccountId(b bool) {
	c["skip_requesting_account_id"] = b
}

var _ configtesting.Thing = thing{}

type thing struct {
	aws.CredentialsProvider
}

func (t thing) GetCredentials() aws.CredentialsProvider {
	return t
}

func (t thing) GetRegion() string {
	panic("not implemented") // lintignore:R009
}

func TestProviderConfig_Authentication_SSO(t *testing.T) { //nolint:paralleltest
	configtesting.SSO(t, &testDriver{})
}

func TestProviderConfig_Authentication_LegacySSO(t *testing.T) { //nolint:paralleltest
	configtesting.LegacySSO(t, &testDriver{})
}

// func TestAccProvider_ProxyConfig(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	var provider *schema.Provider

// 	testcases := map[string]struct {
// 		config               map[string]any
// 		environmentVariables map[string]string
// 		expectedDiags        diag.Diagnostics
// 		urls                 []proxyCase
// 	}{
// 		"no config": {
// 			config: map[string]any{},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"http_proxy empty string": {
// 			config: map[string]any{
// 				"http_proxy": "",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"http_proxy config": {
// 			config: map[string]any{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			expectedDiags: diag.Diagnostics{
// 				diag.Diagnostic{
// 					Severity: diag.Warning,
// 					Summary:  "Missing HTTPS Proxy",
// 					Detail: fmt.Sprintf(
// 						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
// 							"To specify no proxy for HTTPS, set the HTTPS to an empty string",
// 						"http://http-proxy.test:1234"),
// 				},
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"https_proxy config": {
// 			config: map[string]any{
// 				"https_proxy": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config https_proxy config": {
// 			config: map[string]any{
// 				"http_proxy":  "http://http-proxy.test:1234",
// 				"https_proxy": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config https_proxy config empty string": {
// 			config: map[string]any{
// 				"http_proxy":  "http://http-proxy.test:1234",
// 				"https_proxy": "",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"https_proxy config http_proxy config empty string": {
// 			config: map[string]any{
// 				"http_proxy":  "",
// 				"https_proxy": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config https_proxy config no_proxy config": {
// 			config: map[string]any{
// 				"http_proxy":  "http://http-proxy.test:1234",
// 				"https_proxy": "http://https-proxy.test:1234",
// 				"no_proxy":    "dont-proxy.test",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "http://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"HTTP_PROXY envvar": {
// 			config: map[string]any{},
// 			environmentVariables: map[string]string{
// 				"HTTP_PROXY": "http://http-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"http_proxy envvar": {
// 			config: map[string]any{},
// 			environmentVariables: map[string]string{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"HTTPS_PROXY envvar": {
// 			config: map[string]any{},
// 			environmentVariables: map[string]string{
// 				"HTTPS_PROXY": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"https_proxy envvar": {
// 			config: map[string]any{},
// 			environmentVariables: map[string]string{
// 				"https_proxy": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config HTTPS_PROXY envvar": {
// 			config: map[string]any{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"HTTPS_PROXY": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config https_proxy envvar": {
// 			config: map[string]any{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"https_proxy": "http://https-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"http_proxy config NO_PROXY envvar": {
// 			config: map[string]any{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"NO_PROXY": "dont-proxy.test",
// 			},
// 			expectedDiags: diag.Diagnostics{
// 				diag.Diagnostic{
// 					Severity: diag.Warning,
// 					Summary:  "Missing HTTPS Proxy",
// 					Detail: fmt.Sprintf(
// 						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
// 							"To specify no proxy for HTTPS, set the HTTPS to an empty string",
// 						"http://http-proxy.test:1234"),
// 				},
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "http://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"http_proxy config no_proxy envvar": {
// 			config: map[string]any{
// 				"http_proxy": "http://http-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"no_proxy": "dont-proxy.test",
// 			},
// 			expectedDiags: diag.Diagnostics{
// 				diag.Diagnostic{
// 					Severity: diag.Warning,
// 					Summary:  "Missing HTTPS Proxy",
// 					Detail: fmt.Sprintf(
// 						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
// 							"To specify no proxy for HTTPS, set the HTTPS to an empty string",
// 						"http://http-proxy.test:1234"),
// 				},
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "http://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"HTTP_PROXY envvar HTTPS_PROXY envvar NO_PROXY envvar": {
// 			config: map[string]any{},
// 			environmentVariables: map[string]string{
// 				"HTTP_PROXY":  "http://http-proxy.test:1234",
// 				"HTTPS_PROXY": "http://https-proxy.test:1234",
// 				"NO_PROXY":    "dont-proxy.test",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://http-proxy.test:1234",
// 				},
// 				{
// 					url:           "http://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://https-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://dont-proxy.test",
// 					expectedProxy: "",
// 				},
// 			},
// 		},

// 		"http_proxy config overrides HTTP_PROXY envvar": {
// 			config: map[string]any{
// 				"http_proxy": "http://config-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"HTTP_PROXY": "http://envvar-proxy.test:1234",
// 			},
// 			expectedDiags: diag.Diagnostics{
// 				diag.Diagnostic{
// 					Severity: diag.Warning,
// 					Summary:  "Missing HTTPS Proxy",
// 					Detail: fmt.Sprintf(
// 						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
// 							"To specify no proxy for HTTPS, set the HTTPS to an empty string",
// 						"http://config-proxy.test:1234"),
// 				},
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "http://config-proxy.test:1234",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://config-proxy.test:1234",
// 				},
// 			},
// 		},

// 		"https_proxy config overrides HTTPS_PROXY envvar": {
// 			config: map[string]any{
// 				"https_proxy": "http://config-proxy.test:1234",
// 			},
// 			environmentVariables: map[string]string{
// 				"HTTPS_PROXY": "http://envvar-proxy.test:1234",
// 			},
// 			urls: []proxyCase{
// 				{
// 					url:           "http://example.com",
// 					expectedProxy: "",
// 				},
// 				{
// 					url:           "https://example.com",
// 					expectedProxy: "http://config-proxy.test:1234",
// 				},
// 			},
// 		},
// 	}
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
// 		ErrorCheck:               acctest.ErrorCheck(t),
// 		ProtoV5ProviderFactories: testAccProtoV5ProviderFactoriesInternal(ctx, t, &provider),
// 		CheckDestroy:             nil,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccProviderConfig_proxyConfig(),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckProviderProxyConfig(ctx, t, &provider, map[string]string{}),
// 				),
// 			},
// 		},
// 	})
// }

// type proxyCase struct {
// 	url           string
// 	expectedProxy string
// }

// func testAccCheckProviderProxyConfig(ctx context.Context, t *testing.T, p **schema.Provider, urls []proxyCase) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if p == nil || *p == nil || (*p).Meta() == nil || (*p).Meta().(*conns.AWSClient) == nil {
// 			return fmt.Errorf("provider not initialized")
// 		}

// 		providerClient := (*p).Meta().(*conns.AWSClient)
// 		client := providerClient.AwsConfig().HTTPClient
// 		bClient, ok := client.(*awshttp.BuildableClient)
// 		if !ok {
// 			t.Fatalf("expected awshttp.BuildableClient, got %T", client)
// 		}
// 		transport := bClient.GetTransport()
// 		proxyF := transport.Proxy

// 		for _, url := range tc.urls {
// 			req, _ := http.NewRequest("GET", url.url, nil)
// 			pUrl, err := proxyF(req)
// 			if err != nil {
// 				t.Fatalf("unexpected error: %s", err)
// 			}
// 			if url.expectedProxy != "" {
// 				if pUrl == nil {
// 					t.Errorf("expected proxy for %q, got none", url.url)
// 				} else if pUrl.String() != url.expectedProxy {
// 					t.Errorf("expected proxy %q for %q, got %q", url.expectedProxy, url.url, pUrl.String())
// 				}
// 			} else {
// 				if pUrl != nil {
// 					t.Errorf("expected no proxy for %q, got %q", url.url, pUrl.String())
// 				}
// 			}
// 		}
// 	}
// }

// func testAccProviderConfig_proxyConfig() string {
// 	//lintignore:AT004
// 	return acctest.ConfigCompose(
// 		testAccProviderConfig_base, fmt.Sprintf(`
// provider "aws" {
//   default_tags {}

//   skip_credentials_validation = true
//   skip_metadata_api_check     = true
//   skip_requesting_account_id  = true
// }
// `))
// }
