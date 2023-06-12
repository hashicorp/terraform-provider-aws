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

func TestAccRekognitionCollection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection rekognition.DescribeCollectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName, &collection),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rekognition", regexp.MustCompile(`collection/.+$`)),
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

func TestAccRekognitionCollection_extend_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection rekognition.DescribeCollectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_extend_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccCollectionConfig_extend_tags2(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCollectionConfig_extend_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRekognitionCollection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection rekognition.DescribeCollectionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccPreCheckCollection(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, resourceName, &collection),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrekognition.ResourceCollection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCollectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_collection" {
				continue
			}

			input := &rekognition.DescribeCollectionInput{
				CollectionId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeCollection(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameCollection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCollectionExists(ctx context.Context, name string, collection *rekognition.DescribeCollectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()
		resp, err := conn.DescribeCollection(ctx, &rekognition.DescribeCollectionInput{
			CollectionId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.Rekognition, create.ErrActionCheckingExistence, tfrekognition.ResNameCollection, rs.Primary.ID, err)
		}

		*collection = *resp

		return nil
	}
}

func testAccPreCheckCollection(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient()

	input := &rekognition.ListCollectionsInput{}
	_, err := conn.ListCollections(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCollectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
	collection_id = %[1]q
}
`, rName)
}

func testAccCollectionConfig_extend_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
	collection_id = %[1]q
	tags = {
		%[2]q = %[3]q
	}
}
`, rName, tagKey1, tagValue1)
}

func testAccCollectionConfig_extend_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
	collection_id = %[1]q
	tags = {
		%[2]q = %[3]q
		%[4]q = %[5]q
	}
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
