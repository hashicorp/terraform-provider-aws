package ssmincidents_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// only one replication set resource can be active at once, so we must have serialised tests
func TestAccSSMIncidentsReplicationSet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Replication Set Resource Tests": {
			"basic":            testReplicationSet_basic,
			"updateDefaultKey": testReplicationSet_updateRegionsWithoutCMK,
			"updateCMK":        testReplicationSet_updateRegionsWithCMK,
			"updateTags":       testReplicationSet_updateTags,
			"updateEmptyTags":  testReplicationSet_updateEmptyTags,
			"disappears":       testReplicationSet_disappears,
		},
		"Replication Set Data Source Tests": {
			"basic": testReplicationSetDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
