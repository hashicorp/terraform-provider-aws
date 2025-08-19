// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vcr

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// InteractionNotFoundRetryableFunc is a retryable function to augment retry behavior for AWS service clients
// when VCR testing is enabled
var InteractionNotFoundRetryableFunc = func(err error) aws.Ternary {
	if errs.IsA[*url.Error](err) && strings.Contains(err.Error(), "requested interaction not found") {
		return aws.FalseTernary
	}
	return aws.UnknownTernary // Delegate to configured Retryer.
}
