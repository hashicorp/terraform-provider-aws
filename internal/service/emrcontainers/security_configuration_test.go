// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers_test

import (
	"context"
	"errors"
	"fmt"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emrcontainers/types"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfemrcontainers "github.com/hashicorp/terraform-provider-aws/internal/service/emrcontainers"
)

func TestAccEMRContainersSecurityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfiguration awstypes.SecurityConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emrcontainers_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityconfiguration),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrID,
			},
		},
	})
}

// Disappear test will make sense in the future when (if) delete operation is implemented by AWS

//func TestAccEMRContainersSecurityConfiguration_disappears(t *testing.T) {
//	ctx := acctest.Context(t)
//
//	var securityconfiguration awstypes.SecurityConfiguration
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_emrcontainers_security_configuration.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.EMRContainersServiceID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccSecurityConfigurationConfig_basic(rName),
//				Check: resource.ComposeAggregateTestCheckFunc(
//					testAccCheckSecurityConfigurationExists(ctx, resourceName, &securityconfiguration),
//					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfemrcontainers.ResourceSecurityConfiguration, resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}

func testAccCheckSecurityConfigurationExists(ctx context.Context, name string, securityconfiguration *awstypes.SecurityConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

		resp, err := tfemrcontainers.FindSecurityConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.EMRContainers, create.ErrActionCheckingExistence, tfemrcontainers.ResNameSecurityConfiguration, rs.Primary.ID, err)
		}

		*securityconfiguration = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRContainersClient(ctx)

	input := &emrcontainers.ListSecurityConfigurationsInput{}

	_, err := conn.ListSecurityConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`

resource "aws_emrcontainers_security_configuration" "test" {
  name = %[1]q

  security_configuration_data {
    authorization_configuration {
      lake_formation_configuration {
		authorized_session_tag_value = "EMR on EKS Engine"
		query_engine_role_arn = "arn:aws:iam::123456789012:role/query-engine-role"
		secure_namespace_info {
		  cluster_id = "test"
		  namespace = "default"
		}
	  }
	  encryption_configuration {
	    in_transit_encryption_configuration {
		  tls_certificate_configuration {
		    certificate_provider_type = "PEM"
		    private_certificate_secret_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:tls/certificate/private"
		    public_certificate_secret_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:tls/certificate/public"
		  }
	    }
	  }
	}
  }

  tags = {
    Environment = "test"
  }
}
`, rName)
}
