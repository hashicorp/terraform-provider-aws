// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmsap_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmsap"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfssmsap "github.com/hashicorp/terraform-provider-aws/internal/service/ssmsap"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// // intricate, they should be unit tested.

// func TestApplicationExampleUnitTest(t *testing.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		TestName string
// 		Input    string
// 		Expected string
// 		Error    bool
// 	}{
// 		{
// 			TestName: "empty",
// 			Input:    "",
// 			Expected: "",
// 			Error:    true,
// 		},
// 		{
// 			TestName: "descriptive name",
// 			Input:    "some input",
// 			Expected: "some output",
// 			Error:    false,
// 		},
// 		{
// 			TestName: "another descriptive name",
// 			Input:    "more input",
// 			Expected: "more output",
// 			Error:    false,
// 		},
// 	}

// 	for _, testCase := range testCases {
// 		testCase := testCase
// 		t.Run(testCase.TestName, func(t *testing.T) {
// 			t.Parallel()
// 			got, err := tfssmsap.FunctionFromResource(testCase.Input)

// 			if err != nil && !testCase.Error {
// 				t.Errorf("got error (%s), expected no error", err)
// 			}

// 			if err == nil && testCase.Error {
// 				t.Errorf("got (%s) and no error, expected error", got)
// 			}

// 			if got != testCase.Expected {
// 				t.Errorf("got %s, expected %s", got, testCase.Expected)
// 			}
// 		})
// 	}
// }

func TestAccSSMSAPApplication_basicHANASingle(t *testing.T) {
	ctx := acctest.Context(t)

	// if testing.Short() {
	// 	t.Skip("skipping long-running test in short mode")
	// }

	setEnvVars_TODO_REPLACE()

	var application ssmsap.GetApplicationOutput
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_ssmsap_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.SSMSAPEndpointID) #TODO
			testAccPreCheck(ctx, t)
			sapEnvironmentPreparedPreCheck(t, applicationTypeHANA)
		},
		ErrorCheck:               acctest.ErrorCheck(t, "ssm-sap"), // TODO! names.SSMSAPEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicHANASingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "id", rName),
					resource.TestCheckResourceAttr(resourceName, "application_type", "HANA"),
					resource.TestCheckResourceAttr(resourceName, "sap_instance_number", "00"),
					resource.TestCheckResourceAttr(resourceName, "sap_system_id", "HDB"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "credentials.*", map[string]string{
						"database_name":   "SYSTEMDB",
						"credential_type": "ADMIN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "credentials.*", map[string]string{
						"database_name":   "HDB",
						"credential_type": "ADMIN",
					}),

					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssm-sap", regexp.MustCompile(`HANA/`+rName)),
				),
			},
			{
				Config:      testAccApplicationConfig_incorrectPassword(rName),
				ExpectError: regexp.MustCompile(`.unexpected state 'FAILED'.`),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccSSMSAPApplication_basicHanaHA(t *testing.T) {
	t.Skip("HANA in HA mode is not yet supported")
}

func TestAccSSMSAPApplication_basicSapAbapSingle(t *testing.T) {
	t.Skip("SAP ABAP in single mode is not yet supported")
}

func TestAccSSMSAPApplication_basicSapAbapHA(t *testing.T) {
	t.Skip("SAP ABAP in HA mode is not yet supported")
}

func TestAccSSMSAPApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := strings.Replace(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "", -1)
	resourceName := "aws_ssmsap_application.test"

	var application ssmsap.GetApplicationOutput

	setEnvVars_TODO_REPLACE()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "ssm-sap"), // TODO! names.SSMSAPEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basicHANASingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssmsap.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func setEnvVars_TODO_REPLACE() {
	//TODO: remove...
	os.Setenv("TF_ACC", "1")
	os.Setenv("TF_LOG", "WARN")
	os.Setenv("GOFLAGS", "-mod=readonly")
	os.Setenv("AWS_PROFILE", "aws_provider_profile")
	os.Setenv("AWS_DEFAULT_REGION", "eu-central-1")
	os.Setenv("AWS_ALTERNATE_PROFILE", "aws_alternate_profile")
	os.Setenv("AWS_ALTERNATE_REGION", "us-east-1")
	os.Setenv("AWS_THIRD_REGION", "us-east-2")
	os.Setenv("ACM_CERTIFICATE_ROOT_DOMAIN", "terraform-provider-aws-acctest-acm.com")

	os.Setenv("SAP_HANA_INSTANCE_ID", "i-0dcfaf8486058f01a")
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMSAPClient(ctx)
	input := &ssmsap.ListApplicationsInput{}
	_, err := conn.ListApplications(ctx, input)
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckSkipError(err error) bool {
	if err == nil {
		return false
	}
	return true
	//panic("unimplemented") //TODO
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMSAPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmsap_application" {
				continue
			}

			_, _, err := tfssmsap.FindApplicationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.SSMSAP, create.ErrActionCheckingDestroyed, tfssmsap.ResNameApplication, rs.Primary.ID, err)
			}

			return create.Error(names.SSMSAP, create.ErrActionCheckingDestroyed, tfssmsap.ResNameApplication, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, name string, application *ssmsap.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMSAP, create.ErrActionCheckingExistence, tfssmsap.ResNameApplication, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMSAP, create.ErrActionCheckingExistence, tfssmsap.ResNameApplication, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMSAPClient(ctx)
		resp, err := conn.GetApplication(ctx, &ssmsap.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.SSMSAP, create.ErrActionCheckingExistence, tfssmsap.ResNameApplication, rs.Primary.ID, err)
		}

		*application = *resp

		return nil
	}
}

const applicationTypeHANA = "HANA"
const applicationTypeNetWeaver = "NetWeaver"

func sapEnvironmentPreparedPreCheck(t *testing.T, application_type string) {

	switch application_type {
	case applicationTypeHANA:
		if v := os.Getenv("SAP_HANA_INSTANCE_ID"); v == "" {
			t.Fatal("SAP_HANA_INSTANCE_ID must be set for acceptance tests")
		}
	case applicationTypeNetWeaver:
		if v := os.Getenv("SAP_NETWEAVER_INSTANCE_ID"); v == "" {
			t.Fatal("SAP_NETWEAVER_INSTANCE_ID must be set for acceptance tests")
		}
	default:
		t.Fatal("Unknown application type")
	}
}

func testAccApplicationConfig_basicHANASingle(rName string) string {

	instanceId := os.Getenv("SAP_HANA_INSTANCE_ID")

	return fmt.Sprintf(`

	  resource "aws_ssmsap_application" "test" {
		id      = %[1]q
		application_type    = "HANA"
		instances           = [data.aws_instance.this.id]
		sap_instance_number = "00"
		sap_system_id       = "HDB"
	  
		# tags = {
		#   key = "value"
		# }
	  
		credentials {
		  database_name   = "SYSTEMDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.this.id
		}
	  
		credentials {
		  database_name   = "HDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.this.id
		}
	  
		depends_on = [aws_ec2_tag.ssmsapmanaged]
	  }
	  
	  data "aws_instance" "this" {
		instance_id = %[2]q
	  }
	  
	  resource "aws_ec2_tag" "ssmsapmanaged" {
		resource_id = data.aws_instance.this.id
		key         = "SSMForSAPManaged"
		value       = "true"
	  }
	  
	  # Password123@
	  
	  resource "aws_secretsmanager_secret" "this" {
		name = %[1]q
	  }
	  
	  resource "aws_secretsmanager_secret_version" "this" {
		secret_string = "{\"password\":\"Password123@\", \"username\":\"SYSTEM\"}"
		secret_id     = aws_secretsmanager_secret.this.id
	  }
	  
	  
	  
	  
	  
`, rName, instanceId)
}

func testAccApplicationConfig_incorrectPassword(rName string) string {

	instanceId := os.Getenv("SAP_HANA_INSTANCE_ID")

	return fmt.Sprintf(`

	  resource "aws_ssmsap_application" "test" {
		id      = %[1]q
		application_type    = "HANA"
		instances           = [data.aws_instance.this.id]
		sap_instance_number = "00"
		sap_system_id       = "HDB"
	  
		# tags = {
		#   key = "value"
		# }
	  
		credentials {
		  database_name   = "SYSTEMDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.this.id
		}
	  
		credentials {
		  database_name   = "HDB"
		  credential_type = "ADMIN"
		  secret_id       = aws_secretsmanager_secret.this.id
		}
	  
		depends_on = [aws_ec2_tag.ssmsapmanaged]
	  }
	  
	  data "aws_instance" "this" {
		instance_id = %[2]q
	  }
	  
	  resource "aws_ec2_tag" "ssmsapmanaged" {
		resource_id = data.aws_instance.this.id
		key         = "SSMForSAPManaged"
		value       = "true"
	  }
	  
	  # Password123@
	  
	  resource "aws_secretsmanager_secret" "this" {
		name = %[1]q
	  }
	  
	  resource "aws_secretsmanager_secret_version" "this" {
		secret_string = "{\"password\":\"wr0ng_Password@\", \"username\":\"SYSTEM\"}"
		secret_id     = aws_secretsmanager_secret.this.id
	  }
	  
`, strings.Replace(rName, "-", "", -1), instanceId)
}
