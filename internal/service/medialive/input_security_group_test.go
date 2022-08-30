package medialive_test

//
//import (
//	"context"
//	"fmt"
//	"regexp"
//	"strings"
//	"testing"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/medialive"
//	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
//	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
//	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	tfmedialive "github.com/hashicorp/terraform-provider-aws/internal/service/medialive"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//
//func TestInputSecurityGroupExampleUnitTest(t *testing.T) {
//	testCases := []struct {
//		TestName string
//		Input    string
//		Expected string
//		Error    bool
//	}{
//		{
//			TestName: "empty",
//			Input:    "",
//			Expected: "",
//			Error:    true,
//		},
//		{
//			TestName: "descriptive name",
//			Input:    "some input",
//			Expected: "some output",
//			Error:    false,
//		},
//		{
//			TestName: "another descriptive name",
//			Input:    "more input",
//			Expected: "more output",
//			Error:    false,
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(testCase.TestName, func(t *testing.T) {
//			got, err := tfmedialive.FunctionFromResource(testCase.Input)
//
//			if err != nil && !testCase.Error {
//				t.Errorf("got error (%s), expected no error", err)
//			}
//
//			if err == nil && testCase.Error {
//				t.Errorf("got (%s) and no error, expected error", got)
//			}
//
//			if got != testCase.Expected {
//				t.Errorf("got %s, expected %s", got, testCase.Expected)
//			}
//		})
//	}
//}
//
//func TestAccMediaLiveInputSecurityGroup_basic(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var inputsecuritygroup medialive.DescribeInputSecurityGroupResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_medialive_input_security_group.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(t)
//			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
//			testAccPreCheck(t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckInputSecurityGroupDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccInputSecurityGroupConfig_basic(rName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckInputSecurityGroupExists(resourceName, &inputsecuritygroup),
//					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
//					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
//					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
//						"console_access": "false",
//						"groups.#":       "0",
//						"username":       "Test",
//						"password":       "TestTest1234",
//					}),
//					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "medialive", regexp.MustCompile(`inputsecuritygroup:+.`)),
//				),
//			},
//			{
//				ResourceName:            resourceName,
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
//			},
//		},
//	})
//}
//
//func TestAccMediaLiveInputSecurityGroup_disappears(t *testing.T) {
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var inputsecuritygroup medialive.DescribeInputSecurityGroupResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_medialive_input_security_group.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(t)
//			acctest.PreCheckPartitionHasService(names.MediaLiveEndpointID, t)
//			testAccPreCheck(t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.MediaLiveEndpointID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckInputSecurityGroupDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccInputSecurityGroupConfig_basic(rName, testAccInputSecurityGroupVersionNewer),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckInputSecurityGroupExists(resourceName, &inputsecuritygroup),
//					acctest.CheckResourceDisappears(acctest.Provider, tfmedialive.ResourceInputSecurityGroup(), resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}
//
//func testAccCheckInputSecurityGroupDestroy(s *terraform.State) error {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
//	ctx := context.Background()
//
//	for _, rs := range s.RootModule().Resources {
//		if rs.Type != "aws_medialive_input_security_group" {
//			continue
//		}
//
//		input := &medialive.DescribeInputSecurityGroupInput{
//			InputSecurityGroupId: aws.String(rs.Primary.ID),
//		}
//		_, err := conn.DescribeInputSecurityGroup(ctx, &medialive.DescribeInputSecurityGroupInput{
//			InputSecurityGroupId: aws.String(rs.Primary.ID),
//		})
//		if err != nil {
//			var nfe *types.ResourceNotFoundException
//			if errors.As(err, &nfe) {
//				return nil
//			}
//			return err
//		}
//
//		return create.Error(names.MediaLive, create.ErrActionCheckingDestroyed, tfmedialive.ResNameInputSecurityGroup, rs.Primary.ID, errors.New("not destroyed"))
//	}
//
//	return nil
//}
//
//func testAccCheckInputSecurityGroupExists(name string, inputsecuritygroup *medialive.DescribeInputSecurityGroupResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[name]
//		if !ok {
//			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, name, errors.New("not found"))
//		}
//
//		if rs.Primary.ID == "" {
//			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, name, errors.New("not set"))
//		}
//
//		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
//		ctx := context.Background()
//		resp, err := conn.DescribeInputSecurityGroup(ctx, &medialive.DescribeInputSecurityGroupInput{
//			InputSecurityGroupId: aws.String(rs.Primary.ID),
//		})
//
//		if err != nil {
//			return create.Error(names.MediaLive, create.ErrActionCheckingExistence, tfmedialive.ResNameInputSecurityGroup, rs.Primary.ID, err)
//		}
//
//		*inputsecuritygroup = *resp
//
//		return nil
//	}
//}
//
//func testAccPreCheck(t *testing.T) {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaLiveConn
//	ctx := context.Background()
//
//	input := &medialive.ListInputSecurityGroupsInput{}
//	_, err := conn.ListInputSecurityGroups(ctx, input)
//
//	if acctest.PreCheckSkipError(err) {
//		t.Skipf("skipping acceptance testing: %s", err)
//	}
//
//	if err != nil {
//		t.Fatalf("unexpected PreCheck error: %s", err)
//	}
//}
//
//func testAccCheckInputSecurityGroupNotRecreated(before, after *medialive.DescribeInputSecurityGroupResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		if before, after := aws.StringValue(before.InputSecurityGroupId), aws.StringValue(after.InputSecurityGroupId); before != after {
//			return create.Error(names.MediaLive, create.ErrActionCheckingNotRecreated, tfmedialive.ResNameInputSecurityGroup, aws.StringValue(before.InputSecurityGroupId), errors.New("recreated"))
//		}
//
//		return nil
//	}
//}
//
//func testAccInputSecurityGroupConfig_basic(rName, version string) string {
//	return fmt.Sprintf(`
//resource "aws_security_group" "test" {
//  name = %[1]q
//}
//
//resource "aws_medialive_input_security_group" "test" {
//  input_security_group_name             = %[1]q
//  engine_type             = "ActiveMediaLive"
//  engine_version          = %[2]q
//  host_instance_type      = "medialive.t2.micro"
//  security_groups         = [aws_security_group.test.id]
//  authentication_strategy = "simple"
//  storage_type            = "efs"
//
//  logs {
//    general = true
//  }
//
//  user {
//    username = "Test"
//    password = "TestTest1234"
//  }
//}
//`, rName, version)
//}
