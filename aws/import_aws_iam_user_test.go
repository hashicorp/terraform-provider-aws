package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSUser_importBasic(t *testing.T) {
	resourceName := "aws_iam_user.user"

	n := fmt.Sprintf("test-user-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSUserConfig(n, "/"),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_importWithPolicy(t *testing.T) {
	resourceName := "aws_iam_user.user"

	rInt := acctest.RandInt()

	checkFn := func(s []*terraform.InstanceState) error {
		// Expect 2: user + policy
		if len(s) != 2 {
			return fmt.Errorf("expected 2 states: %#v", s)
		}

		// TODO: Is this order guaranteed?
		policyState, userState := s[0], s[1]

		expectedUserId := fmt.Sprintf("test_user_%d", rInt)
		expectedPolicyId := fmt.Sprintf("test_user_%d:foo_policy_%d", rInt, rInt)

		if userState.ID != expectedUserId {
			return fmt.Errorf("expected user of ID %s, %s received",
				expectedUserId, userState.ID)
		}

		if policyState.ID != expectedPolicyId {
			return fmt.Errorf("expected policy of ID %s, %s received",
				expectedPolicyId, policyState.ID)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSUserDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccIAMUserPolicyConfig(rInt),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}
