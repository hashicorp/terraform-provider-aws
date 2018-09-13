package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIotThingPrincipalAttachment_basic(t *testing.T) {
	thingName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingPrincipalAttachmentDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingPrincipalAttachmentConfig(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
				),
			},
		},
	})
}

func testAccCheckAWSIotThingPrincipalAttachmentDestroy_basic(s *terraform.State) error {
	return nil
}

func testAccCheckAWSIotThingPrincipalAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No attachment")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		thing := rs.Primary.Attributes["thing"]
		principal := rs.Primary.Attributes["principal"]

		out, err := conn.ListThingPrincipals(&iot.ListThingPrincipalsInput{
			ThingName: aws.String(thing),
		})

		if err != nil {
			return fmt.Errorf("Error: Failed to get principals for thing %s (%s)", thing, n)
		}

		if len(out.Principals) != 1 {
			return fmt.Errorf("Error: Thing (%s) has wrong number of principals attached on initial creation", thing)
		}

		if principal != aws.StringValue(out.Principals[0]) {
			return fmt.Errorf("Error: Thing (%s) has wrong principal, expected %s, got %s", thing, principal, aws.StringValue(out.Principals[0]))
		}

		return nil
	}
}

func testAccAWSIotThingPrincipalAttachmentConfig(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr = "${file("test-fixtures/iot-csr.pem")}"
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing = "${aws_iot_thing.thing.name}"
  principal = "${aws_iot_certificate.cert.arn}"
}
`, thingName)
}
