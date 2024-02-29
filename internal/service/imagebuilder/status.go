// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// statusImage fetches the Image and its Status
func statusImage(ctx context.Context, conn *imagebuilder.Client, imageBuildVersionArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(imageBuildVersionArn),
		}

		output, err := conn.GetImage(ctx, input)

		if err != nil {
			return nil, string(awstypes.ImageStatusPending), err
		}

		if output == nil || output.Image == nil || output.Image.State == nil {
			return nil, string(awstypes.ImageStatusPending), nil
		}

		status := aws.ToString(output.Image.State.Status)

		if status == string(awstypes.ImageStatusFailed) {
			return output.Image, status, fmt.Errorf("%s", aws.ToString(output.Image.State.Reason))
		}

		return output.Image, status, nil
	}
}
