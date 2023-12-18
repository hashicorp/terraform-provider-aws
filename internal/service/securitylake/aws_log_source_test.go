package securitylake_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityLakeAwsLogSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// var awslogsource securitylake.CreateAwsLogSourceOutput
	// rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// resourceName := "aws_securitylake_aws_log_source.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckDataLakeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLogSourceConfig_basic(),
				Check:  resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

// func TestAccSecurityLakeAwsLogSource_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var awslogsource securitylake.DescribeAwsLogSourceResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_securitylake_aws_log_source.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.SecurityLakeEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeEndpointID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckAwsLogSourceDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAwsLogSourceConfig_basic(rName, testAccAwsLogSourceVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAwsLogSourceExists(ctx, resourceName, &awslogsource),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceAwsLogSource = newResourceAwsLogSource
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceAwsLogSource, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccCheckAwsLogSourceDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_securitylake_aws_log_source" {
// 				continue
// 			}

// 			input := &securitylake.DescribeAwsLogSourceInput{
// 				AwsLogSourceId: aws.String(rs.Primary.ID),
// 			}
// 			_, err := conn.DescribeAwsLogSource(ctx, &securitylake.DescribeAwsLogSourceInput{
// 				AwsLogSourceId: aws.String(rs.Primary.ID),
// 			})
// 			if errs.IsA[*types.ResourceNotFoundException](err){
// 				return nil
// 			}
// 			if err != nil {
// 				return nil
// 			}

// 			return create.Error(names.SecurityLake, create.ErrActionCheckingDestroyed, tfsecuritylake.ResNameAwsLogSource, rs.Primary.ID, errors.New("not destroyed"))
// 		}

// 		return nil
// 	}
// }

// func testAccCheckAwsLogSourceExists(ctx context.Context, name string, awslogsource *securitylake.DescribeAwsLogSourceResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[name]
// 		if !ok {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, name, errors.New("not found"))
// 		}

// 		if rs.Primary.ID == "" {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, name, errors.New("not set"))
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)
// 		resp, err := conn.DescribeAwsLogSource(ctx, &securitylake.DescribeAwsLogSourceInput{
// 			AwsLogSourceId: aws.String(rs.Primary.ID),
// 		})

// 		if err != nil {
// 			return create.Error(names.SecurityLake, create.ErrActionCheckingExistence, tfsecuritylake.ResNameAwsLogSource, rs.Primary.ID, err)
// 		}

// 		*awslogsource = *resp

// 		return nil
// 	}
// }

func testAccAwsLogSourceConfig_basic() string {
	return fmt.Sprintf(`

resource "aws_securitylake_aws_log_source" "example" {
	sources {
		regions         = "eu-west-2"
		source_name    = "ROUTE53"
		source_version = "latest"
	}
}
`)
}
