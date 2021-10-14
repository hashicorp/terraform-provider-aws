package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/workspaces/lister"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_ip_group", &resource.Sweeper{
		Name: "aws_workspaces_ip_group",
		F:    testSweepWorkspacesIpGroups,
	})
}

func testSweepWorkspacesIpGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).workspacesconn
	input := &workspaces.DescribeIpGroupsInput{}
	sweepResources := make([]*testSweepResource, 0)

	err = lister.DescribeIpGroupsPages(conn, input, func(page *workspaces.DescribeIpGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ipGroup := range page.Result {
			r := resourceAwsWorkspacesIpGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(ipGroup.GroupId))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WorkSpaces Ip Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WorkSpaces Ip Groups (%s): %w", region, err)
	}

	err = testSweepResourceOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Ip Groups (%s): %w", region, err)
	}

	return nil
}

func testAccAwsWorkspacesIpGroup_basic(t *testing.T) {
	var v workspaces.IpGroup
	ipGroupName := sdkacctest.RandomWithPrefix("tf-acc-test")
	ipGroupNewName := sdkacctest.RandomWithPrefix("tf-acc-test-upd")
	ipGroupDescription := fmt.Sprintf("Terraform Acceptance Test %s", strings.Title(sdkacctest.RandString(20)))
	resourceName := "aws_workspaces_ip_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesIpGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWorkspacesIpGroupConfigA(ipGroupName, ipGroupDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ipGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ipGroupDescription),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsWorkspacesIpGroupConfigB(ipGroupNewName, ipGroupDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ipGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ipGroupDescription),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
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

func testAccAwsWorkspacesIpGroup_tags(t *testing.T) {
	var v workspaces.IpGroup
	resourceName := "aws_workspaces_ip_group.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesIpGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWorkspacesIpGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsWorkspacesIpGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsWorkspacesIpGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAwsWorkspacesIpGroup_disappears(t *testing.T) {
	var v workspaces.IpGroup
	ipGroupName := sdkacctest.RandomWithPrefix("tf-acc-test")
	ipGroupDescription := fmt.Sprintf("Terraform Acceptance Test %s", strings.Title(sdkacctest.RandString(20)))
	resourceName := "aws_workspaces_ip_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesIpGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWorkspacesIpGroupConfigA(ipGroupName, ipGroupDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsWorkspacesIpGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsWorkspacesIpGroup_MultipleDirectories(t *testing.T) {
	var v workspaces.IpGroup
	var d1, d2 workspaces.WorkspaceDirectory

	ipGroupName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domain := acctest.RandomDomainName()

	resourceName := "aws_workspaces_ip_group.test"
	directoryResourceName1 := "aws_workspaces_directory.test1"
	directoryResourceName2 := "aws_workspaces_directory.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, workspaces.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesIpGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWorkspacesIpGroupConfigMultipleDirectories(ipGroupName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					testAccCheckAwsWorkspacesDirectoryExists(directoryResourceName1, &d1),
					resource.TestCheckTypeSetElemAttrPair(directoryResourceName1, "ip_group_ids.*", "aws_workspaces_ip_group.test", "id"),
					testAccCheckAwsWorkspacesDirectoryExists(directoryResourceName2, &d2),
					resource.TestCheckTypeSetElemAttrPair(directoryResourceName2, "ip_group_ids.*", "aws_workspaces_ip_group.test", "id"),
				),
			},
		},
	})
}

func testAccCheckAwsWorkspacesIpGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_ip_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeIpGroups(&workspaces.DescribeIpGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("error Describing Workspaces IP Group: %w", err)
		}

		// Return nil if the IP Group is already destroyed (does not exist)
		if len(resp.Result) == 0 {
			return nil
		}

		if *resp.Result[0].GroupId == rs.Primary.ID {
			return fmt.Errorf("Workspaces IP Group %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesIpGroupExists(n string, v *workspaces.IpGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Workpsaces IP Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeIpGroups(&workspaces.DescribeIpGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if *resp.Result[0].GroupId == rs.Primary.ID {
			*v = *resp.Result[0]
			return nil
		}

		return fmt.Errorf("Workspaces IP Group (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWorkspacesIpGroupConfigA(name, description string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name        = %[1]q
  description = %[2]q

  rules {
    source = "10.0.0.0/16"
  }

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }
}
`, name, description)
}

func testAccAwsWorkspacesIpGroupConfigB(name, description string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name        = %[1]q
  description = %[2]q

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }
}
`, name, description)
}

func testAccAwsWorkspacesIpGroupConfigTags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name = %[1]q

  rules {
    source = "10.0.0.0/16"
  }

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAwsWorkspacesIpGroupConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name = %[1]q

  rules {
    source = "10.0.0.0/16"
  }

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsWorkspacesIpGroupConfigMultipleDirectories(name, domain string) string {
	return acctest.ConfigCompose(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(name, domain),
		fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name = %[1]q
}

resource "aws_workspaces_directory" "test1" {
  directory_id = aws_directory_service_directory.main.id

  ip_group_ids = [
    aws_workspaces_ip_group.test.id
  ]
}

resource "aws_workspaces_directory" "test2" {
  directory_id = aws_directory_service_directory.main.id

  ip_group_ids = [
    aws_workspaces_ip_group.test.id
  ]
}
  `, name))
}
