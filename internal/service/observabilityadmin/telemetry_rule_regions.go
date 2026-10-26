// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
)

func normalizeTelemetryRuleRegionSelection(rule *awstypes.TelemetryRule) {
	if rule != nil && aws.ToBool(rule.AllRegions) {
		rule.Regions = nil
	}
}
