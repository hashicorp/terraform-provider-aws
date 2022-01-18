package detective_test

import (
	"os"
	"testing"
)

func TestAccDetective_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Graph": {
			"basic":      testAccDetectiveGraph_basic,
			"disappears": testAccDetectiveGraph_disappears,
			"tags":       testAccDetectiveGraph_tags,
		},
		"InvitationAccepter": {
			"basic": testAccDetectiveInvitationAccepter_basic,
		},
		"Member": {
			"basic":     testAccDetectiveMember_basic,
			"disappear": testAccDetectiveMember_disappears,
			"message":   testAccDetectiveMember_message,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccMemberFromEnv(t *testing.T, isAlternate bool) string {
	var email string
	if isAlternate {
		email = os.Getenv("AWS_DETECTIVE_ALTERNATE_ACCOUNT_EMAIL")
		if email == "" {
			t.Skip(
				"Environment variable AWS_DETECTIVE_ALTERNATE_ACCOUNT_EMAIL is not set. " +
					"To properly test inviting Detective member account must be provided.")
		}
		return email
	}

	email = os.Getenv("AWS_DETECTIVE_ACCOUNT_EMAIL")
	if email == "" {
		t.Skip(
			"Environment variable AWS_DETECTIVE_ACCOUNT_EMAIL is not set. " +
				"To properly test inviting Detective member account must be provided")
	}
	return email
}
