// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
)

func TestValidSecretName(t *testing.T) {
	t.Parallel()

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
			Value:    acctest.RandStringFromCharSet(t, 513, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := tfsecretsmanager.ValidSecretName(tc.Value, "aws_secretsmanager_secret")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Secretsmanager Secret Name to not trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidSecretNamePrefix(t *testing.T) {
	t.Parallel()

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
			Value:    acctest.RandStringFromCharSet(t, 512, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := tfsecretsmanager.ValidSecretNamePrefix(tc.Value, "aws_secretsmanager_secret")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Secretsmanager Secret Name to not trigger a validation error for %q", tc.Value)
		}
	}
}
