package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityPolicy_basic(t *testing.T) {
	var securitypolicy opensearchserverless.GetSecurityPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.OpenSearchServerlessEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, "type", "encryption"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccSecurityPolicyImportStateIdFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy"},
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var securitypolicy opensearchserverless.GetSecurityPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.OpenSearchServerlessEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(resourceName, &securitypolicy),
					acctest.CheckResourceDisappears(acctest.Provider, tfopensearchserverless.ResourceSecurityPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opensearchserverless_security_policy" {
			continue
		}

		_, err := conn.GetSecurityPolicy(ctx, &opensearchserverless.GetSecurityPolicyInput{
			Name: aws.String(rs.Primary.ID),
			Type: types.SecurityPolicyTypeEncryption,
		})
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameSecurityPolicy, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckSecurityPolicyExists(name string, securitypolicy *opensearchserverless.GetSecurityPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient()
		ctx := context.Background()
		resp, err := conn.GetSecurityPolicy(ctx, &opensearchserverless.GetSecurityPolicyInput{
			Name: aws.String(rs.Primary.ID),
			Type: types.SecurityPolicyTypeEncryption,
		})

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, rs.Primary.ID, err)
		}

		*securitypolicy = *resp

		return nil
	}
}

func testAccSecurityPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["id"], rs.Primary.Attributes["type"]), nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient()
	ctx := context.Background()

	input := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeEncryption,
	}
	_, err := conn.ListSecurityPolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityPolicyConfig_basic(rName string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name   = %[1]q
  type   = "encryption"
  policy = <<-EOT
  {
	  "Rules": [
		  {
		  	"Resource": [
		  		%[2]q
		  	],
		  	"ResourceType": "collection"
		  }
	  ],
	  "AWSOwnedKey": true
  }
  EOT
}
`, rName, collection)
}
