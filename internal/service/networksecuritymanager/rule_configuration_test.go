// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"testing"
)

func TestRuleConfigurationStringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		prior     string
		new       string
		want      bool
		wantError bool
	}{
		"identical": {
			prior: `{"DefaultAction":{"Allow":{}}}`,
			new:   `{"DefaultAction":{"Allow":{}}}`,
			want:  true,
		},
		"whitespace and key order": {
			prior: `{"VisibilityConfig":{"MetricName":"test","SampledRequestsEnabled":true}}`,
			new:   "{\n  \"VisibilityConfig\" : {\n    \"SampledRequestsEnabled\" : true,\n    \"MetricName\" : \"test\"\n  }\n}\n",
			want:  true,
		},
		// The service stores every number in a configuration as a string and
		// returns it that way, so the two spellings are the same document.
		"number and its decimal string": {
			prior: `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"300"}}}`,
			new:   `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`,
			want:  true,
		},
		"exponent notation and the returned string": {
			prior: `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"300"}}}`,
			new:   `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":3e2}}}`,
			want:  true,
		},
		"trailing zero and the returned string": {
			prior: `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"300"}}}`,
			new:   `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300.0}}}`,
			want:  true,
		},
		"nested integer and the returned string": {
			prior: `{"DefaultAction":{"Block":{"CustomResponse":{"ResponseCode":"403"}}}}`,
			new:   `{"DefaultAction":{"Block":{"CustomResponse":{"ResponseCode":403}}}}`,
			want:  true,
		},
		"fractional number and its decimal string": {
			prior: `{"RateBasedStatement":{"Rate":"0.5"}}`,
			new:   `{"RateBasedStatement":{"Rate":0.5}}`,
			want:  true,
		},
		// A fractional value keeps its literal spelling, so a different
		// spelling of the same quantity is reported as a change rather than
		// silently collapsed.
		"fractional number, different spelling": {
			prior: `{"RateBasedStatement":{"Rate":"0.50"}}`,
			new:   `{"RateBasedStatement":{"Rate":0.5}}`,
			want:  false,
		},
		// Too large for an exact float64: the literal must survive intact.
		"integer beyond float64 precision": {
			prior: `{"Limit":"12345678901234567890"}`,
			new:   `{"Limit":12345678901234567890}`,
			want:  true,
		},
		"integer beyond float64 precision, different value": {
			prior: `{"Limit":"12345678901234567890"}`,
			new:   `{"Limit":12345678901234567891}`,
			want:  false,
		},
		"different numbers": {
			prior: `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":"300"}}}`,
			new:   `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":301}}}`,
			want:  false,
		},
		"number and a non-numeric string": {
			prior: `{"TokenDomains":["example.com"]}`,
			new:   `{"TokenDomains":[300]}`,
			want:  false,
		},
		"different documents": {
			prior: `{"DefaultAction":{"Allow":{}}}`,
			new:   `{"DefaultAction":{"Block":{}}}`,
			want:  false,
		},
		"list order": {
			prior: `{"TokenDomains":["example.com","example.net"]}`,
			new:   `{"TokenDomains":["example.net","example.com"]}`,
			want:  false,
		},
		"invalid JSON": {
			prior:     `{"DefaultAction":{"Allow":{}}}`,
			new:       `{"DefaultAction":`,
			wantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			got, diags := ruleConfigurationValue(testCase.prior).StringSemanticEquals(ctx, ruleConfigurationValue(testCase.new))

			if diags.HasError() != testCase.wantError {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if testCase.wantError {
				return
			}
			if got != testCase.want {
				t.Errorf("StringSemanticEquals(%s, %s) = %t, want %t", testCase.prior, testCase.new, got, testCase.want)
			}
		})
	}
}
