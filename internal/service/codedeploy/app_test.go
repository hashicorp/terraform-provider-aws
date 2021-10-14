package codedeploy_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodedeploy "github.com/hashicorp/terraform-provider-aws/internal/service/codedeploy"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codedeploy_app", &resource.Sweeper{
		Name: "aws_codedeploy_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).CodeDeployConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &codedeploy.ListApplicationsInput{}

	err = conn.ListApplicationsPages(input, func(page *codedeploy.ListApplicationsOutput, lastPage bool) bool {
		for _, app := range page.Applications {
			if app == nil {
				continue
			}

			appName := aws.StringValue(app)
			r := tfcodedeploy.ResourceApp()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", "xxxx", appName))
			d.Set("name", appName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing CodeDeploy Applications for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CodeDeploy Applications for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CodeDeploy Applications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccCodeDeployApp_basic(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codedeploy", fmt.Sprintf(`application:%s`, rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "linked_to_github", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
			// Import by ID
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by name
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeDeployApp_computePlatform(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppComputePlatformConfig(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
				),
			},
			{
				Config: testAccAppComputePlatformConfig(rName, "Server"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application2),
					testAccCheckAppRecreated(&application1, &application2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
				),
			},
		},
	})
}

func TestAccCodeDeployApp_ComputePlatform_ecs(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppComputePlatformConfig(rName, "ECS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "ECS"),
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

func TestAccCodeDeployApp_ComputePlatform_lambda(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppComputePlatformConfig(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
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

func TestAccCodeDeployApp_name(t *testing.T) {
	var application1, application2 codedeploy.ApplicationInfo
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppNameConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAppNameConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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

func TestAccCodeDeployApp_tags(t *testing.T) {
	var application codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
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
				Config: testAccAppTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCodeDeployApp_disappears(t *testing.T) {
	var application1 codedeploy.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, codedeploy.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(resourceName, &application1),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodedeploy.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeDeployConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codedeploy_app" {
			continue
		}

		_, err := conn.GetApplication(&codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		})

		if tfawserr.ErrMessageContains(err, codedeploy.ErrCodeApplicationDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAppExists(name string, application *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeDeployConn

		input := &codedeploy.GetApplicationInput{
			ApplicationName: aws.String(rs.Primary.Attributes["name"]),
		}

		output, err := conn.GetApplication(input)

		if err != nil {
			return err
		}

		if output == nil || output.Application == nil {
			return fmt.Errorf("error reading CodeDeploy Application (%s): empty response", rs.Primary.ID)
		}

		*application = *output.Application

		return nil
	}
}

func testAccCheckAppRecreated(i, j *codedeploy.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreateTime).Equal(aws.TimeValue(j.CreateTime)) {
			return errors.New("CodeDeploy Application was not recreated")
		}

		return nil
	}
}

func testAccAppComputePlatformConfig(rName string, computePlatform string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  compute_platform = %q
  name             = %q
}
`, computePlatform, rName)
}

func testAccAppNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %q
}
`, rName)
}

func testAccAppTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
