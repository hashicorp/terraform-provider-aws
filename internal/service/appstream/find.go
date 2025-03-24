// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// findImages finds all images from a describe images input
func findImages(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImagesInput) ([]awstypes.Image, error) {
	var output []awstypes.Image

	pages := appstream.NewDescribeImagesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Images...)
	}

	return output, nil
}
