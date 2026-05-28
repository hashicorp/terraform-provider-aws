// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import "testing"

func TestSuppressIdentityProviderProviderDetailsCountDiff(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		providerType string
		oldDetails   map[string]string
		newDetails   map[string]string
		want         bool
	}{
		"google defaults removed": {
			providerType: "Google",
			oldDetails: map[string]string{
				"attributes_url":                "https://people.googleapis.com/v1/people/me?personFields=",
				"attributes_url_add_attributes": "true",
				"authorize_scopes":              "email profile openid",
				"authorize_url":                 "https://accounts.google.com/o/oauth2/v2/auth",
				"client_id":                     "test-url.apps.googleusercontent.com",
				"client_secret":                 "client_secret",
				"oidc_issuer":                   "https://accounts.google.com",
				"token_request_method":          "POST",
				"token_url":                     "https://www.googleapis.com/oauth2/v4/token",
			},
			newDetails: map[string]string{
				"authorize_scopes": "email profile openid",
				"client_id":        "test-url.apps.googleusercontent.com",
				"client_secret":    "client_secret",
			},
			want: true,
		},
		"configured value changes": {
			providerType: "Google",
			oldDetails: map[string]string{
				"attributes_url":   "https://people.googleapis.com/v1/people/me?personFields=",
				"authorize_scopes": "email profile openid",
				"client_id":        "test-url.apps.googleusercontent.com",
				"client_secret":    "client_secret",
			},
			newDetails: map[string]string{
				"authorize_scopes": "email profile openid",
				"client_id":        "new-client-id-url.apps.googleusercontent.com",
				"client_secret":    "client_secret",
			},
			want: false,
		},
		"custom google detail removed": {
			providerType: "Google",
			oldDetails: map[string]string{
				"attributes_url": "https://example.com/attributes",
				"client_id":      "test-url.apps.googleusercontent.com",
				"client_secret":  "client_secret",
			},
			newDetails: map[string]string{
				"client_id":     "test-url.apps.googleusercontent.com",
				"client_secret": "client_secret",
			},
			want: false,
		},
		"saml returned details removed with metadata file configured": {
			providerType: "SAML",
			oldDetails: map[string]string{
				"ActiveEncryptionCertificate": "certificate",
				"MetadataFile":                "<xml>",
				"SSORedirectBindingURI":       "https://example.com/sso",
			},
			newDetails: map[string]string{
				"MetadataFile": "<xml>",
			},
			want: true,
		},
		"saml redirect removed without metadata file configured": {
			providerType: "SAML",
			oldDetails: map[string]string{
				"ActiveEncryptionCertificate": "certificate",
				"SSORedirectBindingURI":       "https://example.com/sso",
			},
			newDetails: map[string]string{},
			want:       false,
		},
		"saml metadata file removed": {
			providerType: "SAML",
			oldDetails: map[string]string{
				"ActiveEncryptionCertificate": "certificate",
				"MetadataFile":                "<xml>",
				"SSORedirectBindingURI":       "https://example.com/sso",
			},
			newDetails: map[string]string{},
			want:       false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := suppressIdentityProviderProviderDetailsCountDiff(tc.providerType, tc.oldDetails, tc.newDetails)

			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}

func TestIsReadOnlyOrDefaultIdentityProviderDetail(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		providerType      string
		key               string
		oldValue          string
		configuredDetails map[string]string
		want              bool
	}{
		"default google detail": {
			providerType:      "Google",
			key:               "authorize_url",
			oldValue:          "https://accounts.google.com/o/oauth2/v2/auth",
			configuredDetails: map[string]string{},
			want:              true,
		},
		"custom google detail": {
			providerType:      "Google",
			key:               "authorize_url",
			oldValue:          "https://example.com/auth",
			configuredDetails: map[string]string{},
			want:              false,
		},
		"saml active encryption certificate": {
			providerType:      "SAML",
			key:               "ActiveEncryptionCertificate",
			oldValue:          "certificate",
			configuredDetails: map[string]string{},
			want:              true,
		},
		"saml redirect with metadata file": {
			providerType: "SAML",
			key:          "SSORedirectBindingURI",
			oldValue:     "https://example.com/sso",
			configuredDetails: map[string]string{
				"MetadataFile": "<xml>",
			},
			want: true,
		},
		"saml redirect without metadata file": {
			providerType:      "SAML",
			key:               "SSORedirectBindingURI",
			oldValue:          "https://example.com/sso",
			configuredDetails: map[string]string{},
			want:              false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := isReadOnlyOrDefaultIdentityProviderDetail(tc.providerType, tc.key, tc.oldValue, tc.configuredDetails)

			if got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}
