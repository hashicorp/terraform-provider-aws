package rekognition_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionStreamProcessor_connectedHome_base(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "input.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "input.0.kinesis_video_stream.0.arn", "kinesisvideo", regexp.MustCompile(`stream/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "notification_channel.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "notification_channel.0.sns_topic_arn", "sns", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "settings.0.connected_home.0.labels.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.connected_home.0.labels.0", "ALL"),
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

func TestAccRekognitionStreamProcessor_faceSearch_base(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_faceSearch_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "input.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "input.0.kinesis_video_stream.0.arn", "kinesisvideo", regexp.MustCompile(`stream/.+$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "output.0.kinesis_data_stream.0.arn", "kinesis", regexp.MustCompile(`stream/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "settings.0.face_search.#", "1"),
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

func TestAccRekognitionStreamProcessor_connectedHome_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceStreamProcessor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_faceSearch_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_faceSearch_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceStreamProcessor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_extend_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_tags2(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_faceSearch_extend_faceMatchThreshold(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	threshold := 80.0

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_faceSearch_extend_faceMatchThreshold(rName, threshold),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "settings.0.face_search.0.face_match_threshold", "80"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
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

func TestAccRekognitionStreamProcessor_connectedHome_extend_minConfidence(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	minConfidence := 80.0
	minConfidenceUpdate := 90.0

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_minConfidence(rName, minConfidence),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.connected_home.0.min_confidence", "80"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_minConfidence(rName, minConfidenceUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v2),
					testAccCheckStreamProcessorNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "settings.0.connected_home.0.min_confidence", "90"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_extend_dataSharedPreference(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var v1, v2 rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_dataSharedPreference(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "data_sharing_preference.0.opt_in", "true"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_dataSharedPreference(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v2),
					testAccCheckStreamProcessorNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "data_sharing_preference.0.opt_in", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_extend_kmsKeyId(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_kmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`streamprocessor/.+$`)),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "kms_key_id", "aws_kms_key.test", "key_id"),
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

func TestAccRekognitionStreamProcessor_connectedHome_extend_regionsOfInterest_boundingBox_create(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	boundingBox := map[string]float64{
		"height": 0.2930403,
		"left":   0.3922065,
		"top":    0.1556776,
		"width":  0.2930403,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connected_home_extend_regionsOfInterest_boundingBox(rName, boundingBox),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", "1"),
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

func TestAccRekognitionStreamProcessor_connectedHome_extend_regionsOfInterest_boundingBox_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	boundingBox := map[string]float64{
		"height": 0.2930403,
		"left":   0.3922065,
		"top":    0.1556776,
		"width":  0.2930403,
	}
	boundingUpdate := map[string]float64{
		"height": 0.3930403,
		"left":   0.4922065,
		"top":    0.2556776,
		"width":  0.3930403,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connected_home_extend_regionsOfInterest_boundingBox(rName, boundingBox),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamProcessorConfig_connected_home_extend_regionsOfInterest_boundingBox(rName, boundingUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v2),
					testAccCheckStreamProcessorNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.0.height", "0.3930403"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.0.left", "0.4922065"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.0.top", "0.2556776"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.bounding_box.0.width", "0.3930403"),
				),
			},
		},
	})
}

func TestAccRekognitionStreamProcessor_connectedHome_extend_regionsOfInterest_polygon_create(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var streamprocessor rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	point1 := map[string]float64{
		"x": 0.2930403,
		"y": 0.3922065,
	}
	point2 := map[string]float64{
		"x": 0.5102923,
		"y": 0.7810281,
	}
	point3 := map[string]float64{
		"x": 0.1209321,
		"y": 0.2903921,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_regionsOfInterest_polygon_points3(rName, point1, point2, point3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &streamprocessor),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.0.x", "0.2930403"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.0.y", "0.3922065"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.1.x", "0.5102923"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.1.y", "0.7810281"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.2.x", "0.1209321"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.2.y", "0.2903921"),
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

func TestAccRekognitionStreamProcessor_connectedHome_extend_regionsOfInterest_polygon_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 rekognition.DescribeStreamProcessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_stream_processor.test"
	point1 := map[string]float64{
		"x": 0.2930403,
		"y": 0.3922065,
	}
	point2 := map[string]float64{
		"x": 0.5102923,
		"y": 0.7810281,
	}
	point3 := map[string]float64{
		"x": 0.1209321,
		"y": 0.2903921,
	}
	point4 := map[string]float64{
		"x": 0.2209321,
		"y": 0.3903921,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckStreamProcessor(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStreamProcessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_regionsOfInterest_polygon_points3(rName, point1, point2, point3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStreamProcessorConfig_connectedHome_extend_regionsOfInterest_polygon_points4(rName, point1, point2, point3, point4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStreamProcessorExists(ctx, resourceName, &v2),
					testAccCheckStreamProcessorNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.3.x", "0.2209321"),
					resource.TestCheckResourceAttr(resourceName, "regions_of_interest.0.polygon.0.point.3.y", "0.3903921"),
				),
			},
		},
	})
}

func testAccCheckStreamProcessorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_stream_processor" {
				continue
			}
			_, err := conn.DescribeStreamProcessor(ctx, &rekognition.DescribeStreamProcessorInput{
				Name: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}
			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameStreamProcessor, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckStreamProcessorExists(ctx context.Context, name string, streamprocessor *rekognition.DescribeStreamProcessorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()
		resp, err := conn.DescribeStreamProcessor(ctx, &rekognition.DescribeStreamProcessorInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameStreamProcessor, rs.Primary.ID, err)
		}

		*streamprocessor = *resp

		return nil
	}
}

func testAccPreCheckStreamProcessor(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()

	input := &rekognition.ListStreamProcessorsInput{}
	_, err := conn.ListStreamProcessors(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckStreamProcessorNotRecreated(before, after *rekognition.DescribeStreamProcessorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Name), aws.ToString(after.Name); before != after {
			return create.Error(names.Rekognition, create.ErrActionCheckingNotRecreated, tfrekognition.ResNameStreamProcessor, before, errors.New("recreated"))
		}
		return nil
	}
}

func testAccStreamProcessorConfig_connectedHome_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = %[1]q
  path = "/service-role/"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
        {
            "Effect" = "Allow",
            "Principal" = {
                "Service" = [
                    "rekognition.amazonaws.com",
                ]
            },
            "Action" = "sts:AssumeRole"
        }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRekognitionServiceRole"
}

resource "aws_iam_role_policy" "test" {
	name = %[1]q
	role = aws_iam_role.test.id
	policy = jsonencode({
		"Version" = "2012-10-17",
		"Statement" = [
			{
				"Effect" = "Allow",
				"Action" = [
					"s3:PutObject"
				],
				"Resource" = [
					"${aws_s3_bucket.test.arn}/*"
				]
			}
		]
	})
}

resource "aws_kinesis_video_stream" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}	

resource "aws_sns_topic" "test" {
	name = format("%%s-%%s", "AmazonRekognition", %[1]q)
}

`, rName)
}

func testAccStreamProcessorConfig_faceSearch_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/service-role/"
  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
        {
            "Effect" = "Allow",
            "Principal" = {
                "Service" = [
                    "rekognition.amazonaws.com",
                ]
            },
            "Action" = "sts:AssumeRole"
        }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRekognitionServiceRole"
}

resource "aws_kinesis_video_stream" "test" {
  name = %[1]q
}

resource "aws_kinesis_stream" "test" {
  name             = %[1]q
  shard_count      = 1
  stream_mode_details {
    stream_mode = "PROVISIONED"
  }
}

resource "aws_rekognition_collection" "test" {
  collection_id = %[1]q
}
`, rName)
}

func testAccStreamProcessorConfig_connectedHome_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
	kinesis_video_stream {
	  arn = aws_kinesis_video_stream.test.arn
	}
  }	
  output {
	s3_destination {
	  bucket = aws_s3_bucket.test.id
	}
  }
  notification_channel {
	sns_topic_arn = aws_sns_topic.test.arn
  }
  role_arn = aws_iam_role.test.arn
  settings {
	connected_home {
	  labels = ["ALL"]
	}
  }
}
`, rName))
}

func testAccStreamProcessorConfig_faceSearch_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_faceSearch_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name = %[1]q
  input {
	kinesis_video_stream {
	  arn = aws_kinesis_video_stream.test.arn
	}
  }	
  output {
	kinesis_data_stream {
	  arn = aws_kinesis_stream.test.arn
	}
  }
  role_arn = aws_iam_role.test.arn
  settings {
	face_search {
	  collection_id = aws_rekognition_collection.test.collection_id
	}
  }
}
`, rName))
}

func testAccStreamProcessorConfig_faceSearch_extend_faceMatchThreshold(rName string, threshold float64) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_faceSearch_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name = %[1]q
  input {
	kinesis_video_stream {
	  arn = aws_kinesis_video_stream.test.arn
	}
  }	
  output {
	kinesis_data_stream {
	  arn = aws_kinesis_stream.test.arn
	}
  }
  role_arn = aws_iam_role.test.arn
  settings {
	face_search {
	  collection_id = aws_rekognition_collection.test.collection_id
	  face_match_threshold = %[2]g
	}
  }
}
`, rName, threshold))
}

func testAccStreamProcessorConfig_connectedHome_extend_minConfidence(rName string, minConf float64) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
	data_sharing_preference {
		opt_in = true
	}
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
			min_confidence = %[2]g
		}
	}
}
`, rName, minConf))
}

func testAccStreamProcessorConfig_connectedHome_extend_dataSharedPreference(rName string, optIn bool) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
	data_sharing_preference {
		opt_in = %[2]t
	}
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
}
`, rName, optIn))
}

func testAccStreamProcessorConfig_connectedHome_extend_kmsKeyId(rName string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
	description = %[1]q
}
resource "aws_rekognition_stream_processor" "test" {
	name             = %[1]q
	data_sharing_preference {
		opt_in = true
	}
	kms_key_id = aws_kms_key.test.key_id
	input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
	}	
	output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
	}
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
	role_arn 				  = aws_iam_role.test.arn
	settings {
		connected_home {
			labels = ["ALL"]
		}
	}
}
`, rName))
}

func testAccStreamProcessorConfig_connected_home_extend_regionsOfInterest_boundingBox(rName string, boundingBox map[string]float64) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
	regions_of_interest {
		bounding_box {
			height = %[2]g
			left = %[3]g
			top = %[4]g
			width = %[5]g
		}
	}
}
`, rName, boundingBox["height"], boundingBox["left"], boundingBox["top"], boundingBox["width"]))
}

func testAccStreamProcessorConfig_connectedHome_extend_regionsOfInterest_polygon_points3(rName string, point1, point2, point3 map[string]float64) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
	regions_of_interest {
		polygon {
			point {
				 x = %[2]g
				 y = %[3]g
			}
			point {
				 x = %[4]g
				 y = %[5]g
			}
			point {
				 x = %[6]g
				 y = %[7]g
			}
		}
	}
}
`, rName, point1["x"], point1["y"], point2["x"], point2["y"], point3["x"], point3["y"]))
}

func testAccStreamProcessorConfig_connectedHome_extend_regionsOfInterest_polygon_points4(rName string, point1, point2, point3, point4 map[string]float64) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
	regions_of_interest {
		polygon {
			point {
				 x = %[2]g
				 y = %[3]g
			}
			point {
				 x = %[4]g
				 y = %[5]g
			}
			point {
				 x = %[6]g
				 y = %[7]g
			}
			point {
				 x = %[8]g
				 y = %[9]g
			}
		}
	}
}
`, rName, point1["x"], point1["y"], point2["x"], point2["y"], point3["x"], point3["y"], point4["x"], point4["y"]))
}

func testAccStreamProcessorConfig_connectedHome_extend_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
	tags = {
		%[2]q = %[3]q
	}
}
`, rName, tagKey1, tagValue1))
}

func testAccStreamProcessorConfig_connectedHome_extend_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccStreamProcessorConfig_connectedHome_base(rName),
		fmt.Sprintf(`
resource "aws_rekognition_stream_processor" "test" {
  name             = %[1]q
  input {
		kinesis_video_stream {
			arn = aws_kinesis_video_stream.test.arn
		}
  }	
  output {
		s3_destination {
			bucket = aws_s3_bucket.test.id
		}
  }
	notification_channel {
		sns_topic_arn = aws_sns_topic.test.arn
	}
  role_arn 				  = aws_iam_role.test.arn
  settings {
		connected_home {
			labels = ["ALL"]
		}
	}
	tags = {
		%[2]q = %[3]q
		%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
