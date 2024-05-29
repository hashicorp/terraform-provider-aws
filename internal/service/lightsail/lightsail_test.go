package lightsail_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
)

// serializing tests so that we do not hit the lightsail rate limit for distributions
func TestAccLightsail_serial(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	semaphore := tfsync.GetSemaphore("Lightsail", "AWS_LIGHTSAIL_LIMIT", 6)

	testCases := map[string]map[string]func(*testing.T, tfsync.Semaphore){
		"database": {
			"backupRetentionEnabled":     testAccDatabase_backupRetentionEnabled,
			acctest.CtBasic:              testAccDatabase_basic,
			acctest.CtDisappears:         testAccDatabase_disappears,
			"finalSnapshotName":          testAccDatabase_finalSnapshotName,
			"ha":                         testAccDatabase_ha,
			"masterDatabaseName":         testAccDatabase_masterDatabaseName,
			"masterUsername":             testAccDatabase_masterUsername,
			"masterPassword":             testAccDatabase_masterPassword,
			"preferredBackupWindow":      testAccDatabase_preferredBackupWindow,
			"preferredMaintenanceWindow": testAccDatabase_preferredMaintenanceWindow,
			"publiclyAccessible":         testAccDatabase_publiclyAccessible,
			"relationalDatabaseName":     testAccDatabase_relationalDatabaseName,
			"tags":                       testAccDatabase_tags,
			"keyOnlyTags":                testAccDatabase_keyOnlyTags,
		},
		"domain": {
			acctest.CtBasic:      testAccDomain_basic,
			acctest.CtDisappears: testAccDomain_disappears,
		},
		"domainEntry": {
			"apex":               testAccDomainEntry_apex,
			acctest.CtBasic:      testAccDomainEntry_basic,
			acctest.CtDisappears: testAccDomainEntry_disappears,
			"typeAAAA":           testAccDomainEntry_typeAAAA,
			"underscore":         testAccDomainEntry_underscore,
		},
	}

	acctest.RunLimitedConcurrencyTests2Levels(t, semaphore, testCases)
}

func testAccPreCheckLightsailSynchronize(t *testing.T, semaphore tfsync.Semaphore) {
	tfsync.TestAccPreCheckSyncronize(t, semaphore, "Lightsail")
}
