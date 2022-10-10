package cloudfront

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidPublicKeyName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(129, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validPublicKeyName(tc.Value, "aws_cloudfront_public_key")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the CloudFront PublicKey Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidPublicKeyNamePrefix(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(128, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validPublicKeyNamePrefix(tc.Value, "aws_cloudfront_public_key")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the CloudFront PublicKey Name to trigger a validation error for %q", tc.Value)
		}
	}
}
