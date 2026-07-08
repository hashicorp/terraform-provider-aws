// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import "strconv"

func normalizeCommitmentValue(v string) string {
	if v == "" {
		return ""
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return v
	}

	return strconv.FormatFloat(f, 'f', -1, 64)
}
