// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
)

func TestExpandOracleSettings_source_trimSpaceInChar(t *testing.T) {
	t.Parallel()

	tfMap := map[string]any{
		"trim_space_in_char": false,
	}

	got := expandOracleSettings(tfMap, awstypes.ReplicationEndpointTypeValueSource)

	if got.TrimSpaceInChar == nil {
		t.Fatal("expected TrimSpaceInChar to be set, got nil")
	}
	if *got.TrimSpaceInChar != false {
		t.Errorf("expected TrimSpaceInChar = false, got %v", *got.TrimSpaceInChar)
	}
}
