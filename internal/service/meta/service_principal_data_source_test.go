// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMetaServicePrincipalDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service_principal.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServicePrincipalDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, "s3."+acctest.Region()+".amazonaws.com"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "s3.amazonaws.com"),
					resource.TestCheckResourceAttr(dataSourceName, "suffix", "amazonaws.com"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "s3"),
				),
			},
		},
	})
}

func TestAccMetaServicePrincipalDataSource_MissingService(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServicePrincipalDataSourceConfig_empty,
				ExpectError: regexache.MustCompile(`The argument "service_name" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccMetaServicePrincipalDataSource_ByRegion(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_service_principal.test"
	regions := []string{"us-east-1", "cn-north-1", "us-gov-east-1", "us-iso-east-1", "us-isob-east-1", "eu-isoe-west-1"} //lintignore:AWSAT003

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			t.Parallel()
			acctest.Test(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccServicePrincipalDataSourceConfig_withRegion("s3", region),
						Check: resource.ComposeTestCheckFunc(
							//lintignore:AWSR001
							resource.TestCheckResourceAttr(dataSourceName, names.AttrID, fmt.Sprintf("s3.%s.amazonaws.com", region)),
							resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "s3.amazonaws.com"),
							resource.TestCheckResourceAttr(dataSourceName, "suffix", "amazonaws.com"),
							resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, region),
						),
					},
				},
			})
		})
	}
}

func TestAccMetaServicePrincipalDataSource_UniqueForServiceInRegion(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_service_principal.test"

	type spnTestCase struct {
		Service string
		Region  string
		Suffix  string
		ID      string
		SPN     string
	}

	var testCases []spnTestCase

	var regionUniqueServices = []struct {
		Region   string
		Suffix   string
		Services []string
	}{
		{
			Region:   "us-iso-east-1", //lintignore:AWSAT003
			Suffix:   "c2s.ic.gov",
			Services: []string{"cloudhsm", "config", "logs", "workspaces"},
		},
		{
			Region:   "us-isob-east-1", //lintignore:AWSAT003
			Suffix:   "sc2s.sgov.gov",
			Services: []string{"dms", "logs"},
		},
		{
			Region:   "cn-north-1", //lintignore:AWSAT003
			Suffix:   "amazonaws.com.cn",
			Services: []string{"codedeploy", "elasticmapreduce", "logs"},
		},
	}

	for _, region := range regionUniqueServices {
		for _, service := range region.Services {
			testCases = append(testCases, spnTestCase{
				Service: service,
				Region:  region.Region,
				Suffix:  region.Suffix,
				ID:      fmt.Sprintf("%s.%s.%s", service, region.Region, region.Suffix),
				SPN:     fmt.Sprintf("%s.%s", service, region.Suffix),
			})
		}
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s/%s", testCase.Region, testCase.Service), func(t *testing.T) {
			t.Parallel()
			acctest.Test(ctx, t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: testAccServicePrincipalDataSourceConfig_withRegion(testCase.Service, testCase.Region),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, names.AttrID, testCase.ID),
							resource.TestCheckResourceAttr(dataSourceName, names.AttrName, testCase.SPN),
							resource.TestCheckResourceAttr(dataSourceName, "suffix", testCase.Suffix),
							resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, testCase.Region),
						),
					},
				},
			})
		})
	}
}

const testAccServicePrincipalDataSourceConfig_empty = `
data "aws_service_principal" "test" {}
`

const testAccServicePrincipalDataSourceConfig_basic = `
data "aws_service_principal" "test" {
  service_name = "s3"
}
`

func testAccServicePrincipalDataSourceConfig_withRegion(service string, region string) string {
	return fmt.Sprintf(`
data "aws_service_principal" "test" {
  region       = %[1]q
  service_name = %[2]q
}
`, region, service)
}
