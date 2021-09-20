package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_appconfig_hosted_configuration_version", &resource.Sweeper{
		Name: "aws_appconfig_hosted_configuration_version",
		F:    testSweepAppConfigHostedConfigurationVersions,
	})
}

func testSweepAppConfigHostedConfigurationVersions(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).appconfigconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPages(input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			appId := aws.StringValue(item.Id)

			profilesInput := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: item.Id,
			}

			err := conn.ListConfigurationProfilesPages(profilesInput, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, item := range page.Items {
					if item == nil {
						continue
					}

					profId := aws.StringValue(item.Id)

					versionInput := &appconfig.ListHostedConfigurationVersionsInput{
						ApplicationId:          aws.String(appId),
						ConfigurationProfileId: aws.String(profId),
					}

					err := conn.ListHostedConfigurationVersionsPages(versionInput, func(page *appconfig.ListHostedConfigurationVersionsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, item := range page.Items {
							if item == nil {
								continue
							}

							id := fmt.Sprintf("%s/%s/%d", appId, profId, aws.Int64Value(item.VersionNumber))

							log.Printf("[INFO] Deleting AppConfig Hosted Configuration Version (%s)", id)
							r := resourceAwsAppconfigHostedConfigurationVersion()
							d := r.Data(nil)
							d.SetId(id)

							sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
						}

						return !lastPage
					})

					if err != nil {
						errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Hosted Configuration Versions for Application (%s) and Configuration Profile (%s): %w", appId, profId, err))
					}
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Configuration Profiles for Application (%s): %w", appId, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Hosted Configuration Versions for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Hosted Configuration Versions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSAppConfigHostedConfigurationVersion_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}/configurationprofile/[a-z0-9]{4,7}/hostedconfigurationversion/[0-9]+`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_appconfig_application.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", "aws_appconfig_configuration_profile.test", "configuration_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "content", "{\"foo\":\"bar\"}"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
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

func TestAccAWSAppConfigHostedConfigurationVersion_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAppconfigHostedConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppConfigHostedConfigurationVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_hosted_configuration_version" {
			continue
		}

		appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		}

		output, err := conn.GetHostedConfigurationVersion(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Hosted Configuration Version (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetHostedConfigurationVersion(&appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		})

		if err != nil {
			return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Hosted Configuration Version (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSAppConfigHostedConfigurationVersion(rName string) string {
	return acctest.ConfigCompose(
		testAccAWSAppConfigConfigurationProfileConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %q
}
`, rName))
}
