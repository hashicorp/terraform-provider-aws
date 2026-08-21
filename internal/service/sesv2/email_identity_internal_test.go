// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

func TestFlattenDKIMAttributes_signingHostedZone(t *testing.T) {
	t.Parallel()

	output := flattenDKIMAttributes(&types.DkimAttributes{
		SigningHostedZone: aws.String("dkim.example.com"),
	})

	// SigningHostedZone is exposed through the Terraform signing_hosted_zone attribute.
	if got, want := output["signing_hosted_zone"], "dkim.example.com"; got != want {
		t.Errorf("expected signing_hosted_zone %q, got %q", want, got)
	}
}
