// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncLocationHDFS_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationHDFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHDFSConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "SIMPLE"),
					resource.TestCheckResourceAttr(resourceName, "block_size", "134217728"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_keytab"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_keytab_base64"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_krb5_conf"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_krb5_conf_base64"),
					resource.TestCheckResourceAttr(resourceName, "kerberos_principal", ""),
					resource.TestCheckResourceAttr(resourceName, "kms_key_provider_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "name_node.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "name_node.*", map[string]string{
						names.AttrPort: "80",
					}),
					resource.TestCheckResourceAttr(resourceName, "qop_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "replication_factor", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "simple_user", rName),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^hdfs://.+/`)),
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

func TestAccDataSyncLocationHDFS_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationHDFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHDFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationHDFS(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationHDFS_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationHDFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHDFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLocationHDFSConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLocationHDFSConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccDataSyncLocationHDFS_kerberos(t *testing.T) {
	ctx := acctest.Context(t)
	var v datasync.DescribeLocationHdfsOutput
	resourceName := "aws_datasync_location_hdfs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	principal := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationHDFSDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationHDFSConfig_kerberos(rName, principal),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationHDFSExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication_type", "KERBEROS"),
					resource.TestCheckResourceAttr(resourceName, "block_size", "134217728"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_keytab"),
					resource.TestCheckResourceAttrSet(resourceName, "kerberos_keytab_base64"),
					resource.TestCheckResourceAttrSet(resourceName, "kerberos_krb5_conf"),
					resource.TestCheckNoResourceAttr(resourceName, "kerberos_krb5_conf_base64"),
					resource.TestCheckResourceAttr(resourceName, "kerberos_principal", principal),
					resource.TestCheckResourceAttr(resourceName, "kms_key_provider_uri", ""),
					resource.TestCheckResourceAttr(resourceName, "name_node.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "name_node.*", map[string]string{
						names.AttrPort: "80",
					}),
					resource.TestCheckResourceAttr(resourceName, "qop_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replication_factor", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrURI, regexache.MustCompile(`^hdfs://.+/`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"kerberos_keytab_base64",
					"kerberos_krb5_conf",
				},
			},
		},
	})
}

func testAccCheckLocationHDFSDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_hdfs" {
				continue
			}

			_, err := tfdatasync.FindLocationHDFSByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location HDFS %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationHDFSExists(ctx context.Context, n string, v *datasync.DescribeLocationHdfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindLocationHDFSByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationHDFSConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAgentAgentConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationHDFSConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationHDFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }
}
`, rName))
}

func testAccLocationHDFSConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationHDFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccLocationHDFSConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationHDFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "SIMPLE"
  simple_user         = %[1]q

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}

func testAccLocationHDFSConfig_kerberos(rName, principal string) string {
	return acctest.ConfigCompose(testAccLocationHDFSConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_hdfs" "test" {
  agent_arns          = [aws_datasync_agent.test.arn]
  authentication_type = "KERBEROS"

  name_node {
    hostname = aws_instance.test.private_dns
    port     = 80
  }

  kerberos_principal     = %[1]q
  kerberos_keytab_base64 = filebase64("test-fixtures/keytab.krb")
  kerberos_krb5_conf     = file("test-fixtures/krb5.conf")
}
`, principal))
}
