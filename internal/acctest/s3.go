// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func S3BucketHasTag(ctx context.Context, bucketName, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		tags, err := tfs3.BucketListTags(ctx, conn, bucketName)
		if err != nil {
			return err
		}

		for k, v := range tags {
			if k == key {
				if v.ValueString() == value {
					return nil
				} else {
					return fmt.Errorf("expected tag %q value to be %s, got %s", key, value, v.ValueString())
				}
			}
		}

		return fmt.Errorf("expected tag %q not found", key)
	}
}
