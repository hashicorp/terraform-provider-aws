// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

const (
	errCodeDependencyViolation = "DependencyViolation"
)

func failureError(apiObject *awstypes.Failure) error {
	if apiObject == nil {
		return nil
	}

	var err error
	if reason, detail := aws.ToString(apiObject.Reason), aws.ToString(apiObject.Detail); detail == "" {
		err = errors.New(reason)
	} else {
		err = fmt.Errorf("%s: %s", reason, detail)
	}

	return fmt.Errorf("%s: %w", aws.ToString(apiObject.Arn), err)
}

// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/api_failures_messages.html.
const (
	failureReasonMissing = "MISSING"
)
