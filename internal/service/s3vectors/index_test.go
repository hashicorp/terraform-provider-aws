// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3vectors_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3vectors "github.com/hashicorp/terraform-provider-aws/internal/service/s3vectors"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3VectorsIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Index
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("data_type"), tfknownvalue.StringExact(awstypes.DataTypeFloat32)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dimension"), knownvalue.Int32Exact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("distance_metric"), tfknownvalue.StringExact(awstypes.DistanceMetricEuclidean)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("index_arn"), tfknownvalue.RegionalARNRegexp("s3vectors", regexache.MustCompile(`bucket/.+/index/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("index_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("vector_bucket_name"), knownvalue.NotNull()),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						"index_arn": tfknownvalue.RegionalARNRegexp("s3vectors", regexache.MustCompile(`bucket/.+/index/.+`)),
					}),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "index_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "index_arn",
			},
		},
	})
}

func TestAccS3VectorsIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Index
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3vectors.ResourceIndex, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3VectorsIndex_withVector(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Index
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3vectors_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3VectorsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &v),
					testAccCheckIndexAddVector(ctx, resourceName, acctest.CtKey1, []float32{1.0, 2.0}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3VectorsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3vectors_index" {
				continue
			}

			_, err := tfs3vectors.FindIndexByARN(ctx, conn, rs.Primary.Attributes["index_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Vectors Index %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexExists(ctx context.Context, n string, v *awstypes.Index) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3VectorsClient(ctx)

		output, err := tfs3vectors.FindIndexByARN(ctx, conn, rs.Primary.Attributes["index_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIndexAddVector(ctx context.Context, n string, key string, value []float32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3VectorsClient(ctx)

		metadata, err := tfsmithy.DocumentFromJSONString(fmt.Sprintf(`{"id": %[1]q}`, key), document.NewLazyDocument)
		if err != nil {
			return err
		}

		input := s3vectors.PutVectorsInput{
			IndexArn: aws.String(rs.Primary.Attributes["index_arn"]),
			Vectors: []awstypes.PutInputVector{{
				Key:      aws.String(key),
				Data:     &awstypes.VectorDataMemberFloat32{Value: value},
				Metadata: metadata,
			}},
		}
		_, err = conn.PutVectors(ctx, &input)

		return err
	}
}

func testAccIndexConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = "%[1]s-bucket"
}

resource "aws_s3vectors_index" "test" {
  index_name         = %[1]q
  vector_bucket_name = aws_s3vectors_vector_bucket.test.vector_bucket_name

  data_type       = "float32"
  dimension       = 2
  distance_metric = "euclidean"
}
`, rName)
}
