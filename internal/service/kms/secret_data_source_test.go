// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSSecretDataSource_removed(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecretDataSourceConfig_basic,
				ExpectError: regexache.MustCompile(tfkms.SecretRemovedMessage),
			},
		},
	})
}

const testAccSecretDataSourceConfig_basic = `
data "aws_kms_secret" "testing" {
  secret {
    name    = "secret_name"
    payload = "data-source-removed"

    context = {
      name = "value"
    }
  }
}
`
