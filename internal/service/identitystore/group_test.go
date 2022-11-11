package identitystore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroup_basic(t *testing.T) {
	var group identitystore.DescribeGroupOutput
	resourceName := "aws_identitystore_group.test"
	displayName := "Acceptance Test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName),
					resource.TestCheckResourceAttrSet(resourceName, "identity_store_id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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

func TestAccIdentityStoreGroup_disappears(t *testing.T) {
	var group identitystore.DescribeGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.IdentityStoreEndpointID, t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &group),
					acctest.CheckResourceDisappears(acctest.Provider, tfidentitystore.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_identitystore_group" {
			continue
		}

		_, err := conn.DescribeGroup(ctx, &identitystore.DescribeGroupInput{
			GroupId:         aws.String(rs.Primary.Attributes["group_id"]),
			IdentityStoreId: aws.String(rs.Primary.Attributes["identity_store_id"]),
		})
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.IdentityStore, create.ErrActionCheckingDestroyed, tfidentitystore.ResNameGroup, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckGroupExists(name string, group *identitystore.DescribeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IdentityStoreClient
		ctx := context.Background()
		resp, err := conn.DescribeGroup(ctx, &identitystore.DescribeGroupInput{
			GroupId:         aws.String(rs.Primary.Attributes["group_id"]),
			IdentityStoreId: aws.String(rs.Primary.Attributes["identity_store_id"]),
		})

		if err != nil {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, rs.Primary.ID, err)
		}

		*group = *resp

		return nil
	}
}

func testAccGroupConfig_basic(displayName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Example description"
}
`, displayName)
}
