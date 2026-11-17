// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func endpointError(apiObject *lambdaweb.GetWebFunctionEndpointOutput) error {
	return fmt.Errorf("%s: %w", aws.ToString(apiObject.StateReason), errors.Join(regionalEndpointsErrors(apiObject.RegionalEndpoints, regionalEndpointError)...))
}

func endpointUpdateError(apiObject *lambdaweb.GetWebFunctionEndpointOutput) error {
	return fmt.Errorf("%s: %w", aws.ToString(apiObject.UpdateStatusReason), errors.Join(regionalEndpointsErrors(apiObject.RegionalEndpoints, regionalEndpointUpdateError)...))
}

func regionalEndpointsErrors(regionalEndpoints map[string]awstypes.RegionalEndpoint, cb func(awstypes.RegionalEndpoint) error) []error {
	var errs []error
	for _, region := range slices.Sorted(maps.Keys(regionalEndpoints)) {
		err := fmt.Errorf("%s: %w", region, cb(regionalEndpoints[region]))
		errs = append(errs, err)
	}
	return errs
}

func regionalEndpointError(apiObject awstypes.RegionalEndpoint) error {
	return fmt.Errorf("%s: %s", apiObject.State, aws.ToString(apiObject.StateReason))
}

func regionalEndpointUpdateError(apiObject awstypes.RegionalEndpoint) error {
	return fmt.Errorf("%s: %s", apiObject.UpdateStatus, aws.ToString(apiObject.UpdateStatusReason))
}

func revisionError(apiObject *lambdaweb.GetWebFunctionRevisionOutput) error {
	return fmt.Errorf("%s: %w", aws.ToString(apiObject.StateReason), errors.Join(tfslices.ApplyToAll(apiObject.Errors, revisionErrorError)...))
}

func revisionErrorError(apiObject awstypes.RevisionError) error {
	return fmt.Errorf("%s: %s: %s", aws.ToString(apiObject.Attribute), aws.ToString(apiObject.ErrorCode), aws.ToString(apiObject.ErrorMessage))
}
