// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
)

func testAccCheckVirtualInterfaceExists(ctx context.Context, t *testing.T, n string, v *awstypes.VirtualInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindVirtualInterfaceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVirtualInterfaceDestroy(ctx context.Context, t *testing.T, s *terraform.State, typ string) error { // nosemgrep:ci.semgrep.acctest.naming.destroy-check-signature
	conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != typ {
			continue
		}

		_, err := tfdirectconnect.FindVirtualInterfaceByID(ctx, conn, rs.Primary.ID)

		if retry.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Virtual Interface (%s) %s still exists", typ, rs.Primary.ID)
	}

	return nil
}
