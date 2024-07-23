// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThingPrincipalAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	thingName := sdkacctest.RandomWithPrefix("tf-acc")
	thingName2 := sdkacctest.RandomWithPrefix("tf-acc2")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingPrincipalAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingPrincipalAttachmentConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update1(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName2, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update2(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName2, true, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update3(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName, true, []string{"aws_iot_certificate.cert", "aws_iot_certificate.cert2"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName2, false, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update4(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, thingName, true, []string{"aws_iot_certificate.cert2"}),
				),
			},
		},
	})
}

func testAccCheckThingPrincipalAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_principal_attachment" {
				continue
			}

			_, err := tfiot.FindThingPrincipalAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing"], rs.Primary.Attributes[names.AttrPrincipal])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Thing Principal Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckThingPrincipalAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindThingPrincipalAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing"], rs.Primary.Attributes[names.AttrPrincipal])

		return err
	}
}

func testAccCheckThingPrincipalAttachmentStatus(ctx context.Context, thingName string, exists bool, principals []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		principalARNs := make(map[string]string)

		for _, p := range principals {
			pr, ok := s.RootModule().Resources[p]
			if !ok {
				return fmt.Errorf("Not found: %s", p)
			}
			principalARNs[pr.Primary.Attributes[names.AttrARN]] = p
		}

		_, err := conn.DescribeThing(ctx, &iot.DescribeThingInput{
			ThingName: aws.String(thingName),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

		res, err := conn.ListThingPrincipals(ctx, &iot.ListThingPrincipalsInput{
			ThingName: aws.String(thingName),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list thing (%s) principals: %s", thingName, err)
		}

		if len(res.Principals) != len(principalARNs) {
			return fmt.Errorf("Error: Thing (%s) has wrong number of principals attached", thingName)
		}

		for _, p := range res.Principals {
			if principal, ok := principalARNs[p]; !ok {
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
