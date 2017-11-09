package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppautoscalingScheduledAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(""),
				),
			},
		},
	})
}

func testAccCheckAwsAppautoscalingScheduledActionDestroy(s *terraform.State) error {
	return errors.New("error")
}

func testAccCheckAwsAppautoscalingScheduledActionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		return nil
	}
}

func testAccAppautoscalingScheduledActionConfig() string {
	return fmt.Sprintf(``)
}
