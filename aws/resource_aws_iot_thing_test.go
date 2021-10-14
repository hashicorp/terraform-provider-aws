package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iot_thing", &resource.Sweeper{
		Name:         "aws_iot_thing",
		F:            testSweepIotThings,
		Dependencies: []string{"aws_iot_thing_principal_attachment"},
	})
}

func testSweepIotThings(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingsInput{}

	err = conn.ListThingsPages(input, func(page *iot.ListThingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thing := range page.Things {
			r := ResourceThing()
			d := r.Data(nil)

			d.SetId(aws.StringValue(thing.ThingName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIotThing_basic(t *testing.T) {
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSIotThingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccAWSIotThing_full(t *testing.T) {
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	typeName := fmt.Sprintf("tf_acc_type_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSIotThingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingConfig_full(thingName, typeName, "42"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Update attribute
				Config: testAccAWSIotThingConfig_full(thingName, typeName, "differentOne"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "differentOne"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{ // Remove thing type association
				Config: testAccAWSIotThingConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingExists(resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, "name", thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func testAccCheckIotThingExists(n string, thing *iot.DescribeThingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn
		params := &iot.DescribeThingInput{
			ThingName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeThing(params)
		if err != nil {
			return err
		}

		*thing = *resp

		return nil
	}
}

func testAccCheckAWSIotThingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing" {
			continue
		}

		params := &iot.DescribeThingInput{
			ThingName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThing(params)
		if err != nil {
			if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoT Thing to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccAWSIotThingConfig_basic(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"
}
`, thingName)
}

func testAccAWSIotThingConfig_full(thingName, typeName, answer string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"

  attributes = {
    One    = "11111"
    Two    = "TwoTwo"
    Answer = "%s"
  }

  thing_type_name = aws_iot_thing_type.test.name
}

resource "aws_iot_thing_type" "test" {
  name = "%s"
}
`, thingName, answer, typeName)
}
