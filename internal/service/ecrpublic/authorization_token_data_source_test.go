package ecrpublic_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRPublicAuthorizationTokenDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecrpublic_authorization_token.repo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
					resource.TestMatchResourceAttr(dataSourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "password"),
				),
			},
		},
	})
}

var testAccAuthorizationTokenDataSourceConfig_basic = `
data "aws_ecrpublic_authorization_token" "repo" {}
`
