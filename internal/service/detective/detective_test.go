package detective_test

import (
	"os"
	"testing"
)

func TestAccDetective_serial(t *testing.T) {
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
