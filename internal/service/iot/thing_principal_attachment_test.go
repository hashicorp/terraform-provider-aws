// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThingPrincipalAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	thingName := acctest.RandomWithPrefix(t, "tf-acc")
	thingName2 := acctest.RandomWithPrefix(t, "tf-acc2")
	resourceName := "aws_iot_thing_principal_attachment.att"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingPrincipalAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingPrincipalAttachmentConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert"}),
					resource.TestCheckResourceAttr(resourceName, "thing_principal_type", string(awstypes.ThingPrincipalTypeNonExclusiveThing)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update1(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName2, true, []string{"aws_iot_certificate.cert"}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update2(thingName, thingName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName2, true, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update3(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert", "aws_iot_certificate.cert2"}),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName2, false, []string{}),
				),
			},
			{
				Config: testAccThingPrincipalAttachmentConfig_update4(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att2"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert2"}),
				),
			},
		},
	})
}
func TestAccIoTThingPrincipalAttachment_thingPrincipalType(t *testing.T) {
	ctx := acctest.Context(t)
	thingName := acctest.RandomWithPrefix(t, "tf-acc")
	thingName2 := acctest.RandomWithPrefix(t, "tf-acc2")
	resourceName := "aws_iot_thing_principal_attachment.att"
	resourceThingName := "aws_iot_thing.thing"
	resourceCertName := "aws_iot_certificate.cert"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingPrincipalAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingPrincipalAttachmentConfig_thingPrincipalType(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert"}),
					resource.TestCheckResourceAttr(resourceName, "thing_principal_type", string(awstypes.ThingPrincipalTypeExclusiveThing)),
					resource.TestCheckResourceAttrPair(resourceName, "thing", resourceThingName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, resourceCertName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// The first attachment is specified as EXCLUSIVE_THING.
				// Try to attach the same principal to another Thing, which should fail
				// because exclusive principals can only be attached to one thing.
				Config:      testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate1(thingName, thingName2),
				ExpectError: regexache.MustCompile(`InvalidRequestException: Principal already has an exclusive Thing attached to it`),
			},
			{
				// Reset to one attachment with NON_EXCLUSIVE_THING.
				Config: testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate2(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingPrincipalAttachmentExists(ctx, t, "aws_iot_thing_principal_attachment.att"),
					testAccCheckThingPrincipalAttachmentStatus(ctx, t, thingName, true, []string{"aws_iot_certificate.cert"}),
					resource.TestCheckResourceAttr(resourceName, "thing_principal_type", string(awstypes.ThingPrincipalTypeNonExclusiveThing)),
					resource.TestCheckResourceAttrPair(resourceName, "thing", resourceThingName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, resourceCertName, names.AttrARN),
				),
			},
			{
				// Try to attach the same principal to another Thing specifying EXCLUSIVE_THING,
				// which should fail because the principal already has a non-exclusive attachment
				// and exclusive attachments cannot coexist with any other attachments.
				Config:      testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate3(thingName, thingName2),
				ExpectError: regexache.MustCompile(`InvalidRequestException: Principal already has a Thing attached to it`),
			},
		},
	})
}

func testAccCheckThingPrincipalAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_principal_attachment" {
				continue
			}

			_, err := tfiot.FindThingPrincipalAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing"], rs.Primary.Attributes[names.AttrPrincipal])

			if retry.NotFound(err) {
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

func testAccCheckThingPrincipalAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindThingPrincipalAttachmentByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing"], rs.Primary.Attributes[names.AttrPrincipal])

		return err
	}
}

func testAccCheckThingPrincipalAttachmentStatus(ctx context.Context, t *testing.T, thingName string, exists bool, principals []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

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
			return fmt.Errorf("Error: cannot describe thing %s: %w", thingName, err)
		} else if !exists {
			return fmt.Errorf("Error: Thing (%s) does not exist, but expected to be", thingName)
		}

		res, err := conn.ListThingPrincipals(ctx, &iot.ListThingPrincipalsInput{
			ThingName: aws.String(thingName),
		})

		if err != nil {
			return fmt.Errorf("Error: Cannot list thing (%s) principals: %w", thingName, err)
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

func testAccThingPrincipalAttachmentConfig_thingPrincipalType(thingName string) string {
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

  thing_principal_type = "EXCLUSIVE_THING"
}
`, thingName)
}

func testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate1(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = %[1]q
}

resource "aws_iot_thing" "thing2" {
  name = %[2]q
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn

  thing_principal_type = "EXCLUSIVE_THING"
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing     = aws_iot_thing.thing2.name
  principal = aws_iot_certificate.cert.arn
}
`, thingName, thingName2)
}

func testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate2(thingName string) string {
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

  thing_principal_type = "NON_EXCLUSIVE_THING"
}
`, thingName)
}

func testAccThingPrincipalAttachmentConfig_thingPrincipalTypeUpdate3(thingName, thingName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_thing" "thing" {
  name = %[1]q
}

resource "aws_iot_thing" "thing2" {
  name = %[2]q
}

resource "aws_iot_thing_principal_attachment" "att" {
  thing     = aws_iot_thing.thing.name
  principal = aws_iot_certificate.cert.arn
}

resource "aws_iot_thing_principal_attachment" "att2" {
  thing     = aws_iot_thing.thing2.name
  principal = aws_iot_certificate.cert.arn

  thing_principal_type = "EXCLUSIVE_THING"
}
`, thingName, thingName2)
}
