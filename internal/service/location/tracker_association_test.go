package location_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestTrackerAssociationParseId(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected tflocation.TrackerAssociationId
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: tflocation.TrackerAssociationId{},
			Error:    true,
		},
		{
			TestName: "no pipe",
			Input:    "trackerNameConsumerArn",
			Expected: tflocation.TrackerAssociationId{},
			Error:    true,
		},
		{
			TestName: "valid",
			Input:    "trackerName|consumerArn",
			Expected: tflocation.TrackerAssociationId{
				TrackerName: "trackerName",
				ConsumerArn: "consumerArn",
			},
			Error: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tflocation.TrackerAssociationParseId(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestAccLocationTrackerAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "consumer_arn", "aws_location_geofence_collection.test", "collection_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tracker_name", "aws_location_tracker.test", "tracker_name"),
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

func TestAccLocationTrackerAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_tracker_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrackerAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflocation.ResourceTrackerAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTrackerAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_location_tracker_association" {
			continue
		}

		trackerAssociationId, err := tflocation.TrackerAssociationParseId(rs.Primary.ID)

		if err != nil {
			return names.Error(names.Location, names.ErrActionCheckingDestroyed, tflocation.ResNameTrackerAssociation, rs.Primary.ID, err)
		}

		err = tflocation.FindTrackerAssociationByTrackerNameAndConsumerArn(context.TODO(), conn, trackerAssociationId.TrackerName, trackerAssociationId.ConsumerArn)

		if err != nil {
			if tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
				return nil
			}
			return err
		}

		return names.Error(names.Location, names.ErrActionCheckingDestroyed, tflocation.ResNameTrackerAssociation, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckTrackerAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return names.Error(names.Location, names.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return names.Error(names.Location, names.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, errors.New("not set"))
		}

		trackerAssociationId, err := tflocation.TrackerAssociationParseId(rs.Primary.ID)

		if err != nil {
			return names.Error(names.Location, names.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, name, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

		err = tflocation.FindTrackerAssociationByTrackerNameAndConsumerArn(context.TODO(), conn, trackerAssociationId.TrackerName, trackerAssociationId.ConsumerArn)

		if err != nil {
			return names.Error(names.Location, names.ErrActionCheckingExistence, tflocation.ResNameTrackerAssociation, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccTrackerAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}

resource "aws_location_tracker_association" "test" {
  consumer_arn = aws_location_geofence_collection.test.collection_arn
  tracker_name = aws_location_tracker.test.tracker_name
}
`, rName)
}
