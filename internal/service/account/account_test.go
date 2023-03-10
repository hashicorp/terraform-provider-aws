package account_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAccount_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AlternateContact": {
			"basic":      testAccAlternateContact_basic,
			"disappears": testAccAlternateContact_disappears,
			"AccountID":  testAccAlternateContact_accountID,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
