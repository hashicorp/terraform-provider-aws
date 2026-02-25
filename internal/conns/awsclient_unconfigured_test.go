// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	organizationssvc "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAWSClientOrganizationsClient_Unconfigured(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := t.Context()
	c := &conns.AWSClient{}
	c.SetServicePackages(ctx, map[string]conns.ServicePackage{
		names.Organizations: organizationssvc.ServicePackage(ctx),
	})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()

	client := c.OrganizationsClient(ctx)
	if client == nil {
		t.Fatal("expected non-nil organizations client")
	}

	// Ensure default client caching does not panic when the map is lazily initialized.
	client = c.OrganizationsClient(ctx)
	if client == nil {
		t.Fatal("expected non-nil organizations client from cache")
	}
}
