package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestDecodeIamServiceLinkedRoleID(t *testing.T) {
	var testCases = []struct {
		Input        string
		ServiceName  string
		RoleName     string
		CustomSuffix string
		ErrCount     int
	}{
		{
			Input:    "not-arn",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:iam::123456789012:role/not-service-linked-role", //lintignore:AWSAT005
			ErrCount: 1,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling", //lintignore:AWSAT005
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling",
			CustomSuffix: "",
			ErrCount:     0,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling_custom-suffix", //lintignore:AWSAT005
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling_custom-suffix",
			CustomSuffix: "custom-suffix",
			ErrCount:     0,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/dynamodb.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_DynamoDBTable", //lintignore:AWSAT005
			ServiceName:  "dynamodb.application-autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForApplicationAutoScaling_DynamoDBTable",
			CustomSuffix: "DynamoDBTable",
			ErrCount:     0,
		},
	}

	for _, tc := range testCases {
		serviceName, roleName, customSuffix, err := tfiam.DecodeServiceLinkedRoleID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if serviceName != tc.ServiceName {
			t.Fatalf("expected service name %q to be %q", serviceName, tc.ServiceName)
		}
		if roleName != tc.RoleName {
			t.Fatalf("expected role name %q to be %q", roleName, tc.RoleName)
		}
		if customSuffix != tc.CustomSuffix {
			t.Fatalf("expected custom suffix %q to be %q", customSuffix, tc.CustomSuffix)
		}
	}
}

func TestAccIAMServiceLinkedRole_basic(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "elasticbeanstalk.amazonaws.com"
	name := "AWSServiceRoleForElasticBeanstalk"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Remove existing if possible
					conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
					deletionID, err := tfiam.DeleteServiceLinkedRole(conn, name)
					if err != nil {
						t.Fatalf("Error deleting service-linked role %s: %s", name, err)
					}
					if deletionID == "" {
						return
					}

					err = tfiam.DeleteServiceLinkedRoleWaiter(conn, deletionID)
					if err != nil {
						t.Fatalf("Error waiting for role (%s) to be deleted: %s", name, err)
					}
				},
				Config: testAccServiceLinkedRoleConfig(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("role%s%s", path, name)),
					resource.TestCheckResourceAttr(resourceName, "aws_service_name", awsServiceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "path", path),
					resource.TestCheckResourceAttrSet(resourceName, "unique_id"),
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

func TestAccIAMServiceLinkedRole_customSuffix(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	name := fmt.Sprintf("AWSServiceRoleForAutoScaling_%s", customSuffix)
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_CustomSuffix(awsServiceName, customSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("role%s%s", path, name)),
					resource.TestCheckResourceAttr(resourceName, "name", name),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4439
func TestAccIAMServiceLinkedRole_CustomSuffix_diffSuppressFunc(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "custom-resource.application-autoscaling.amazonaws.com"
	name := "AWSServiceRoleForApplicationAutoScaling_CustomResource"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("role/aws-service-role/%s/%s", awsServiceName, name)),
					resource.TestCheckResourceAttr(resourceName, "custom_suffix", "CustomResource"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
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

func TestAccIAMServiceLinkedRole_description(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceLinkedRoleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func testAccCheckServiceLinkedRoleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_service_linked_role" {
			continue
		}

		_, roleName, _, err := tfiam.DecodeServiceLinkedRoleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err = conn.GetRole(params)

		if err == nil {
			return fmt.Errorf("Service-Linked Role still exists: %q", rs.Primary.ID)
		}

		if !tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil

}

func testAccCheckServiceLinkedRoleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		_, roleName, _, err := tfiam.DecodeServiceLinkedRoleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err = conn.GetRole(params)

		if err != nil {
			if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
				return fmt.Errorf("Service-Linked Role doesn't exists: %q", rs.Primary.ID)
			}
			return err
		}

		return nil
	}
}

func testAccServiceLinkedRoleConfig(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
}
`, awsServiceName)
}

func testAccServiceLinkedRoleConfig_CustomSuffix(awsServiceName, customSuffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
  custom_suffix    = "%s"
}
`, awsServiceName, customSuffix)
}

func testAccServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
  custom_suffix    = "%s"
  description      = "%s"
}
`, awsServiceName, customSuffix, description)
}
