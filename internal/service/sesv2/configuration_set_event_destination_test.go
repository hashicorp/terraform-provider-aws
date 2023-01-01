package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2ConfigurationSetEventDestination_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration_set_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.default_dimension_value", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.cloud_watch_destination.0.dimension_configuration.0.dimension_value_source", "MESSAGE_TAG"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_destination.0.matching_event_types.0", "SEND"),
					resource.TestCheckResourceAttr(resourceName, "event_destination_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSESV2ConfigurationSetEventDestination_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_configuration_set_event_destination.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetEventDestinationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetEventDestinationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationSetEventDestinationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsesv2.ResourceConfigurationSetEventDestination(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationSetEventDestinationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sesv2_configuration_set_event_destination" {
			continue
		}

		_, err := tfsesv2.FindConfigurationSetEventDestinationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			var nfe *types.NotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameConfigurationSetEventDestination, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckConfigurationSetEventDestinationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client()

		_, err := tfsesv2.FindConfigurationSetEventDestinationByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameConfigurationSetEventDestination, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccConfigurationSetEventDestinationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_configuration_set" "test" {
  configuration_set_name = %[1]q
}

resource "aws_sesv2_configuration_set_event_destination" "test" {
  configuration_set_name = aws_sesv2_configuration_set.test.configuration_set_name
  event_destination_name = %[1]q

  event_destination {
	cloud_watch_destination {
	  dimension_configuration {
		default_dimension_value = %[1]q
		dimension_name 			= %[1]q
		dimension_value_source  = "MESSAGE_TAG"
	  }
	}

	matching_event_types = ["SEND"]
  }
}
`, rName)
}
