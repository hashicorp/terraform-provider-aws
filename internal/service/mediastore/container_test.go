package mediastore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMediaStoreContainer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_media_store_container.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediastore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig_basic(sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
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

func TestAccMediaStoreContainer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(5)
	resourceName := "aws_media_store_container.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediastore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerConfig_tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf_mediastore_%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccContainerConfig_tags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf_mediastore_%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckContainerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_store_container" {
				continue
			}

			input := &mediastore.DescribeContainerInput{
				ContainerName: aws.String(rs.Primary.ID),
			}

			resp, err := conn.DescribeContainerWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
					return nil
				}
				return err
			}

			if *resp.Container.Status != mediastore.ContainerStatusDeleting {
				return fmt.Errorf("MediaStore Container (%s) not deleted", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccCheckContainerExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn()

		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeContainerWithContext(ctx, input)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn()

	input := &mediastore.ListContainersInput{}

	_, err := conn.ListContainersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccContainerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}
`, rName)
}

func testAccContainerConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%[1]s"

  tags = {
    Name = "tf_mediastore_%[1]s"

    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
