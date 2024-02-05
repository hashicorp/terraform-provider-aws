// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"
	var application awstypes.ApplicationSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
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

func TestAccApplication_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"
	var application awstypes.ApplicationSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionOld := "MicroFocus M2 Application"
	descriptionNew := "MicroFocus M2 Application Updated"
	resourceName := "aws_m2_application.test"
	var application awstypes.ApplicationSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_update(rName, descriptionOld),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionOld),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_update(rName, descriptionNew),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, "environment_id"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionNew),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_type"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
				),
			},
		},
	})
}

func TestAccApplicatioon_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_application.test"
	var application awstypes.ApplicationSummary

	tags1 := `
  tags = {
    key1 = "value1"
  }
`
	tags2 := `
  tags = {
    key1 = "value1"
    key2 = "value2"
  }
`
	tags3 := `
  tags = {
    key2 = "value2"
  }
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.M2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckApplicationDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags(rName, tags1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccApplicationConfig_tags(rName, tags2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags(rName, tags3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckApplicationExists(ctx context.Context, resourceName string, v *awstypes.ApplicationSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no M2 Application ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)
		out, err := tfm2.FindAppByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("retrieving M2 Application (%s): %w", rs.Primary.ID, err)
		}

		v = out

		return nil
	}
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_application" {
				continue
			}

			_, err := tfm2.FindAppByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("M2 Application (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.id
  key          = "test.json"
  content_type = "application/json"
  content = <<JSON
{
	"template-version": "2.0",
	"source-locations": [
	  {
		"source-id": "s3-source",
		"source-type": "s3",
		"properties": {
		  "s3-bucket": "my-bankdemo-bucket",
		  "s3-key-prefix": "v1"
		}
	  }
	],
	"definition": {
	  "listeners": [
		{
		  "port": 6000,
		  "type": "tn3270"
		}
	  ],
	  "batch-settings": {
		"initiators": [
		  {
			"classes": ["A","B"],
			"description": "initiator_AB...."
		  },
		  {
			"classes": ["C","D"],
			"description": "initiator_CD...."
		  }
		],
		"jcl-file-location": "${aws_s3_bucket.test.id}/jcl"
	  },
	  "cics-settings": {
		"binary-file-location": "${aws_s3_bucket.test.id}/transaction",
		"csd-file-location": "${aws_s3_bucket.test.id}/RDEF",
		"system-initialization-table": "BNKCICV"
	  },
	  "xa-resources": [
		{
		  "name": "XASQL",
		  "module": "${aws_s3_bucket.test.id}/xa/ESPGSQLXA64.so"
		}
	  ]
	}
} 
JSON
}
resource "aws_m2_application" "test" {
  definition {
	s3_location = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  }
  engine_type   = "microfocus"
  name          = %[1]q

}
`, rName)
}

func testAccApplicationConfig_full(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  description     = "Test-1"
  engine_type     = "microfocus"
  engine_version = "8.0.10"
  high_availability_config {
	desired_capacity = 1
  }
  instance_type   = "M2.m5.large"
  kms_key_id      = aws_kms_key.test.arn
  name            = %[1]q
  security_group_ids   = [aws_security_group.test.id]
  subnet_ids           = aws_subnet.test[*].id
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  tags = {
	  Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description = "tf-test-cmk-kms-key-id"
}
  
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
  
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
	cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName))
}

func testAccApplicationConfig_update(rName string, desc string) string {
	return fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  engine_type   = "microfocus"
  description   = %[2]q
  instance_type = "M2.m5.large"
  name          = %[1]q
}
`, rName, desc)
}

func testAccApplicationConfig_tags(rName, tags string) string {
	return fmt.Sprintf(`
resource "aws_m2_environment" "test" {
  engine_type   = "microfocus"
  instance_type = "M2.m5.large"
  name          = %[1]q

%[2]s
}
`, rName, tags)
}
