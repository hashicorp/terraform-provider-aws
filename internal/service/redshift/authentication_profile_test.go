package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftAuthenticationProfile_basic(t *testing.T) {
	resourceName := "aws_redshift_authentication_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthenticationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_profile_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_profile_name", rName),
				),
			},
		},
	})
}

func TestAccRedshiftAuthenticationProfile_disappears(t *testing.T) {
	resourceName := "aws_redshift_authentication_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAuthenticationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceAuthenticationProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthenticationProfileDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_authentication_profile" {
			continue
		}

		_, err := tfredshift.FindAuthenticationProfileByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Authentication Profile %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAuthenticationProfileExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Authentication Profile ID is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, err := tfredshift.FindAuthenticationProfileByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAuthenticationProfileConfig_basic(rName, id string) string {
	return fmt.Sprintf(`
resource "aws_redshift_authentication_profile" "test" {
  authentication_profile_name = %[1]q
  authentication_profile_content = jsonencode(
    {
      AllowDBUserOverride = "1"
      Client_ID           = "ExampleClientID"
      App_ID              = %[2]q
    }
  )
}
`, rName, id)
}
