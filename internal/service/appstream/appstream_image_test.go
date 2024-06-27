// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

	input := &appstream.DescribeImagesInput{
		// should I put something here?

	}

	_, err := conn.DescribeImages(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

}

func testAccCheckAppstreamImageDestroy(ctx context.Context) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx) // probably right ?
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_image_builder" {
				continue
			}
			_, err := tfappstream.FindImageBuilderByName(ctx, conn, rs.Primary.ID) // should it be this

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Appstream %s still exists", rs.Primary.ID) // also wrong

		}
		return nil
	}

}
