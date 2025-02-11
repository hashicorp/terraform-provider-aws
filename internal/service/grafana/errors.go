// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func updatesError(apiObjects []awstypes.UpdateError) error {
	errs := tfslices.ApplyToAll(apiObjects, func(v awstypes.UpdateError) error {
		return updateError(&v)
	})

	return errors.Join(errs...)
}

func updateError(apiObject *awstypes.UpdateError) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%d: %s", aws.ToInt32(apiObject.Code), aws.ToString(apiObject.Message))
}
