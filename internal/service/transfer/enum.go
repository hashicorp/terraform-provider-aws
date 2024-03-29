// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

const (
	SecurityPolicyName2018_11             = "TransferSecurityPolicy-2018-11"
	SecurityPolicyName2020_06             = "TransferSecurityPolicy-2020-06"
	SecurityPolicyNameFIPS_2020_06        = "TransferSecurityPolicy-FIPS-2020-06"
	SecurityPolicyNameFIPS_2023_05        = "TransferSecurityPolicy-FIPS-2023-05"
	SecurityPolicyNameFIPS_2024_01        = "TransferSecurityPolicy-FIPS-2024-01"
	SecurityPolicyName2022_03             = "TransferSecurityPolicy-2022-03"
	SecurityPolicyName2023_05             = "TransferSecurityPolicy-2023-05"
	SecurityPolicyName2024_01             = "TransferSecurityPolicy-2024-01"
	SecurityPolicyNamePQ_SSH_2023_04      = "TransferSecurityPolicy-PQ-SSH-Experimental-2023-04"
	SecurityPolicyNamePQ_SSH_FIPS_2023_04 = "TransferSecurityPolicy-PQ-SSH-FIPS-Experimental-2023-04"
)

func SecurityPolicyName_Values() []string {
	return []string{
		SecurityPolicyName2018_11,
		SecurityPolicyName2020_06,
		SecurityPolicyNameFIPS_2020_06,
		SecurityPolicyNameFIPS_2023_05,
		SecurityPolicyNameFIPS_2024_01,
		SecurityPolicyName2022_03,
		SecurityPolicyName2023_05,
		SecurityPolicyName2024_01,
		SecurityPolicyNamePQ_SSH_2023_04,
		SecurityPolicyNamePQ_SSH_FIPS_2023_04,
	}
}
