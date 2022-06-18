package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
)

func TestAccIoTThingPrincipalAttachment_basic(t *testing.T) {
	thingName := sdkacctest.RandomWithPrefix("tf-acc")
	thingName2 := sdkacctest.RandomWithPrefix("tf-acc2")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckThingPrincipalAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingPrincipalAttachmentConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update1(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(thingName2, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update2(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(thingName2, true, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update3(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert", "aws_iot_certificate.cert2"}),
					testAccCheckThingPrincipalAttachmentStatus(thingName2, false, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update4(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists("aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(thingName, true, []string{"aws_iot_certificate.cert2"}),
				),
			},
		},
	})
}

func testAccCheckThingPrincipalAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_principal_attachment" {
			continue
		}

		principal := rs.Primary.Attributes["principal"]
		thing := rs.Primary.Attributes["thing"]

		found, err := tfiot.GetThingPricipalAttachment(conn, thing, principal)

		if err != nil {
			return fmt.Errorf("Error: Failed listing principals for thing (%s): %s", thing, err)
		}

		if !found {
			continue
		}

		return fmt.Errorf("IOT Thing Principal Attachment (%s) still exists", rs.Primary.Attributes["id"])
	}

	return nil
}

func testAccCheckThingPrincipalAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No attachment")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
		thing := rs.Primary.Attributes["thing"]
		principal := rs.Primary.Attributes["principal"]

		found, err := tfiot.GetThingPricipalAttachment(conn, thing, principal)

		if err != nil {
			return fmt.Errorf("Error: Failed listing principals for thing (%s), resource (%s): %s", thing, n, err)
		}

		if !found {
			return fmt.Errorf("Error: Principal (%s) is not attached to thing (%s), resource (%s)", principal, thing, n)
		}

		return nil
	}
}

func testAccCheckThingPrincipalAttachmentStatus(thingName string, exists bool, principals []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		principalARNs := make(map[string]string)

		for _, p := range principals {
			pr, ok := s.RootModule().Resources[p]
			if !ok {
				return fmt.Errorf("Not found: %s", p)
			}
			principalARNs[pr.Primary.Attributes["arn"]] = p
		}

		thing, err := conn.DescribeThing(&iot.DescribeThingInput{
			ThingName: aws.String(thingName),
		})

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
			if exists {
				return fmt.Errorf("Error: Thing (%s) exists, but expected to be removed", thingName)
			} else {
				return nil
			}
		} else if err != nil {
			return fmt.Errorf("Error: cannot describe thing %s: %s", thingName, err)
		} else if !exists {
			return fmt.Errorf("Error: Thing (%s) does not exist, but expected to be", thingName)
		}

		res, err := conn.ListThingPrincipals(&iot.ListThingPrincipalsInput{
			ThingName: aws.String(thingName),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list thing (%s) principals: %s", thingName, err)
		}

		if len(res.Principals) != len(principalARNs) {
			return fmt.Errorf("Error: Thing (%s) has wrong number of principals attached", thing)
		}

		for _, p := range res.Principals {
			if principal, ok := principalARNs[aws.StringValue(p)]; !ok {
				return fmt.Errorf("Error: Principal %s is not attached to thing %s", principal, thingName)
			}
		}

		return nil
	}
}

func testAccThingPrincipalAttachmentConfig_basic(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn
}
`, thingName)
}

func testAccThingPrincipalAttachmentConfig_update1(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing" "thing2" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing     = aws_iot_thing.thing2.name
  principal = aws_iot_certificate.cert.arn
}
`, thingName, thingName2)
}

func testAccThingPrincipalAttachmentConfig_update2(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing" "thing2" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn
}
`, thingName, thingName2)
}

func testAccThingPrincipalAttachmentConfig_update3(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_certificate" "cert2" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert2.arn
}
`, thingName)
}

func testAccThingPrincipalAttachmentConfig_update4(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert2" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = "%s"
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert2.arn
}
`, thingName)
}
