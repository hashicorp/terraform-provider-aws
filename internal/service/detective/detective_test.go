package detective_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDetective_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Graph": {
			"basic":      testAccGraph_basic,
			"disappears": testAccGraph_disappears,
			"tags":       testAccGraph_tags,
		},
		"InvitationAccepter": {
			"basic": testAccInvitationAccepter_basic,
		},
		"Member": {
			"basic":     testAccMember_basic,
			"disappear": testAccMember_disappears,
			"message":   testAccMember_message,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccMemberFromEnv(t *testing.T) string {
	email := os.Getenv("AWS_DETECTIVE_MEMBER_EMAIL")
	if email == "" {
		t.Skip(
			"Environment variable AWS_DETECTIVE_MEMBER_EMAIL is not set. " +
				"To properly test inviting Detective member accounts, " +
				"a valid email associated with the alternate AWS acceptance " +
				"test account must be provided.")
	}
	return email
}
