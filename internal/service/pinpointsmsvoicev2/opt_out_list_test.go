package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"

	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
)

func TestAccPinpointSMSVoiceV2OptOutList_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var optOutList pinpointsmsvoicev2.DescribeOptOutListsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(pinpointsmsvoicev2.EndpointsID, t)
			testAccPreCheckOptOutList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, pinpointsmsvoicev2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOptOutListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptOutListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptOutListExists(resourceName, &optOutList),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccPinpointSMSVoiceV2OptOutList_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var optOutList pinpointsmsvoicev2.DescribeOptOutListsOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(pinpointsmsvoicev2.EndpointsID, t)
			testAccPreCheckOptOutList(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, pinpointsmsvoicev2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOptOutListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOptOutListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOptOutListExists(resourceName, &optOutList),
					acctest.CheckResourceDisappears(acctest.Provider, tfpinpointsmsvoicev2.ResourceOptOutList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOptOutListDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpointsmsvoicev2_opt_out_list" {
			continue
		}

		input := &pinpointsmsvoicev2.DescribeOptOutListsInput{
			OptOutListNames: aws.StringSlice([]string{rs.Primary.ID}),
		}

		_, err := conn.DescribeOptOutLists(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("expected PinpointSMSVoiceV2 OptOutList to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckOptOutListExists(name string, optOutList *pinpointsmsvoicev2.DescribeOptOutListsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no PinpointSMSVoiceV2 OptOutList is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn
		resp, err := conn.DescribeOptOutLists(&pinpointsmsvoicev2.DescribeOptOutListsInput{
			OptOutListNames: aws.StringSlice([]string{rs.Primary.ID}),
		})

		if err != nil {
			return fmt.Errorf("error describing PinpointSMSVoiceV2 OptOutList: %s", err.Error())
		}

		*optOutList = *resp

		return nil
	}
}

func testAccPreCheckOptOutList(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn

	input := &pinpointsmsvoicev2.DescribeOptOutListsInput{}

	_, err := conn.DescribeOptOutLists(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOptOutListConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName)
}
