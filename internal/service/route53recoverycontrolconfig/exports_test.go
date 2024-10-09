// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

// Exports for use in tests only.
var (
	ResourceCluster        = resourceCluster
	ResourceControlPanel   = resourceControlPanel
	ResourceRoutingControl = resourceRoutingControl
	ResourceSafetyRule     = resourceSafetyRule

	FindClusterByARN        = findClusterByARN
	FindControlPanelByARN   = findControlPanelByARN
	FindRoutingControlByARN = findRoutingControlByARN
	FindSafetyRuleByARN     = findSafetyRuleByARN
)
