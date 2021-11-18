package s3_test

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3Bucket_Basic_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	hostedZoneID, _ := tfs3.HostedZoneIDForRegion(region)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hosted_zone_id", hostedZoneID),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckNoResourceAttr(resourceName, "website_endpoint"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "arn", "s3", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					testAccCheckS3BucketDomainName(resourceName, "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr(resourceName, "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

// Support for common Terraform 0.11 pattern
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7868
func TestAccS3Bucket_Basic_emptyString(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketBucketEmptyStringConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^terraform-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Tags_withNoSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Verify update from 0 tags.
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_withSystemTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	var stackID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:  acctest.Providers,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckBucketDestroy,
			func(s *terraform.State) error {
				// Tear down CF stack.
				conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

				requestToken := resource.UniqueId()
				req := &cloudformation.DeleteStackInput{
					StackName:          aws.String(stackID),
					ClientRequestToken: aws.String(requestToken),
				}

				log.Printf("[DEBUG] Deleting CloudFormation stack: %s", req)
				if _, err := conn.DeleteStack(req); err != nil {
					return fmt.Errorf("error deleting CloudFormation stack: %w", err)
				}

				if _, err := tfcloudformation.WaitStackDeleted(conn, stackID, requestToken, 10*time.Minute); err != nil {
					return fmt.Errorf("Error waiting for CloudFormation stack deletion: %s", err)
				}

				return nil
			},
		),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckDestroyBucket(resourceName),
					testAccCheckBucketCreateViaCloudFormation(bucketName, &stackID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_withTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_withUpdatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
			{
				Config: testAccBucketConfig_withNoTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckBucketTagKeys(resourceName, "aws:cloudformation:stack-name", "aws:cloudformation:stack-id", "aws:cloudformation:logical-id"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_ignoreTags(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_withNoTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketUpdateTags(resourceName, nil, map[string]string{"ignorekey1": "ignorevalue1"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckBucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
					}),
				),
			},
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigIgnoreTagsKeyPrefixes1("ignorekey"),
					testAccBucketConfig_withTags(bucketName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
					testAccCheckBucketCheckTags(resourceName, map[string]string{
						"ignorekey1": "ignorevalue1",
						"Key1":       "AAA",
						"Key2":       "BBB",
						"Key3":       "CCC",
					}),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_s3_bucket.bucket1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiBucketWithTagsConfig(rInt),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_namePrefix(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_generatedName(t *testing.T) {
	resourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "bucket_prefix"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_acceleration(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithAccelerationConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", "Enabled"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithoutAccelerationConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acceleration_status", "Suspended"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_requestPayer(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPayerBucketOwnerConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", "BucketOwner"),
					testAccCheckRequestPayer(resourceName, "BucketOwner"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketRequestPayerRequesterConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "request_payer", "Requester"),
					testAccCheckRequestPayer(resourceName, "Requester"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_policy(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	partition := acctest.Partition()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithPolicyConfig(bucketName, partition),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketPolicy(resourceName, testAccBucketPolicy(bucketName, partition)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"acl",
					"force_destroy",
					"grant",
					// NOTE: Prior to Terraform AWS Provider 3.0, this attribute did not import correctly either.
					//       The Read function does not require GetBucketPolicy, if the argument is not configured.
					//       Rather than introduce that breaking change as well with 3.0, instead we leave the
					//       current Read behavior and note this will be deprecated in a later 3.x release along
					//       with other inline policy attributes across the provider.
					"policy",
				},
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketPolicy(resourceName, ""),
				),
			},
			{
				Config: testAccBucketWithEmptyPolicyConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketPolicy(resourceName, ""),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_updateACL(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithACLConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccBucketWithACLUpdateConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "private"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_updateGrant(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithGrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "2",
						"type":          "CanonicalUser",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "FULL_CONTROL"),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "WRITE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithGrantsUpdateConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "1",
						"type":          "CanonicalUser",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "READ"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "grant.*", map[string]string{
						"permissions.#": "1",
						"type":          "Group",
						"uri":           "http://acs.amazonaws.com/groups/s3/LogDelivery",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "grant.*.permissions.*", "READ_ACP"),
				),
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_aclToGrant(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithACLConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
				),
			},
			{
				Config: testAccBucketWithGrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					// check removed ACLs
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_grantToACL(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithGrantsConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
				),
			},
			{
				Config: testAccBucketWithACLConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl", "public-read"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "0"),
					// check removed grants
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_simple(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "index.html", "", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccBucketWebsiteWithErrorConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_redirect(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteWithRedirectConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "", "", "", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccBucketWebsiteWithHTTPSRedirectConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "", "", "https", "hashicorp.com?my=query"),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "", "", "", ""),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccS3Bucket_Web_routingRules(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteWithRoutingRulesConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(
						resourceName, "index.html", "error.html", "", ""),
					testAccCheckBucketWebsiteRoutingRules(
						resourceName,
						[]*s3.RoutingRule{
							{
								Condition: &s3.Condition{
									KeyPrefixEquals: aws.String("docs/"),
								},
								Redirect: &s3.Redirect{
									ReplaceKeyPrefixWith: aws.String("documents/"),
								},
							},
						},
					),
					testAccCheckS3BucketWebsiteEndpoint(resourceName, "website_endpoint", bucketName, region),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "grant"},
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketWebsite(resourceName, "", "", "", ""),
					testAccCheckBucketWebsiteRoutingRules(resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, "website_endpoint", ""),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_enableDefaultEncryptionWhenTypical(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketEnableDefaultEncryption(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexp.MustCompile("^arn")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_enableDefaultEncryptionWhenAES256IsUsed(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketEnableDefaultEncryptionWithAES256(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_disableDefaultEncryptionWhenDefaultEncryptionIsEnabled(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketEnableDefaultEncryptionWithDefaultKey(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketDisableDefaultEncryption(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_keyEnabled(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketKeyEnabled(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.sse_algorithm", "aws:kms"),
					resource.TestMatchResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.apply_server_side_encryption_by_default.0.kms_master_key_id", regexp.MustCompile("^arn")),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption_configuration.0.rule.0.bucket_key_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

// Test TestAccS3Bucket_Basic_shouldFailNotFound is designed to fail with a "plan
// not empty" error in Terraform, to check against regresssions.
// See https://github.com/hashicorp/terraform/pull/2925
func TestAccS3Bucket_Basic_shouldFailNotFound(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDestroyedConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckDestroyBucket(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Bucket_Manage_versioning(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketVersioning(resourceName, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithVersioningConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketVersioning(resourceName, s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config: testAccBucketWithDisableVersioningConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketVersioning(resourceName, s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_corsUpdate(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	updateBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
			_, err := conn.PutBucketCors(&s3.PutBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
				CORSConfiguration: &s3.CORSConfiguration{
					CORSRules: []*s3.CORSRule{
						{
							AllowedHeaders: []*string{aws.String("*")},
							AllowedMethods: []*string{aws.String("GET")},
							AllowedOrigins: []*string{aws.String("https://www.example.com")},
						},
					},
				},
			})
			if err != nil && !tfawserr.ErrMessageContains(err, "NoSuchCORSConfiguration", "") {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithCORSConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
					updateBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithCORSConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.example.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
			},
		},
	})
}

func TestAccS3Bucket_Security_corsDelete(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	deleteBucketCors := func(n string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources[n]
			if !ok {
				return fmt.Errorf("Not found: %s", n)
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
			_, err := conn.DeleteBucketCors(&s3.DeleteBucketCorsInput{
				Bucket: aws.String(rs.Primary.ID),
			})
			if err != nil && !tfawserr.ErrMessageContains(err, "NoSuchCORSConfiguration", "") {
				return err
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithCORSConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					deleteBucketCors(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3Bucket_Security_corsEmptyOrigin(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithCORSEmptyOriginConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketCors(
						resourceName,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Security_logging(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithLoggingConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketLogging(resourceName, "aws_s3_bucket.log_bucket", "log/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleBasic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithLifecycleConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "60",
						"storage_class": "INTELLIGENT_TIERING",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "90",
						"storage_class": "ONEZONE_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "120",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.transition.*", map[string]string{
						"date":          "",
						"days":          "210",
						"storage_class": "DEEP_ARCHIVE",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.date", "2016-01-12"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.expiration.0.expired_object_delete_marker", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.transition.*", map[string]string{
						"days": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.id", "id4"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.prefix", "path4/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.3.tags.terraform", "hashicorp"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.id", "id5"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.4.tags.terraform", "hashicorp"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.4.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.id", "id6"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.5.tags.tagKey", "tagValue"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.5.transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketWithVersioningLifecycleConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "30",
						"storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.0.noncurrent_version_transition.*", map[string]string{
						"days":          "60",
						"storage_class": "GLACIER",
					}),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.id", "id2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.prefix", "path2/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.1.noncurrent_version_expiration.0.days", "365"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.id", "id3"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.2.prefix", "path3/"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lifecycle_rule.2.noncurrent_version_transition.*", map[string]string{
						"days":          "0",
						"storage_class": "GLACIER",
					}),
				),
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_lifecycleExpireMarkerOnly(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWithLifecycleExpireMarkerConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.id", "id1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.prefix", "path1/"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.days", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.date", ""),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_rule.0.expiration.0.expired_object_delete_marker", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketConfig_Basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11420
func TestAccS3Bucket_Manage_lifecycleRuleExpirationEmptyBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleRuleExpirationEmptyConfigurationBlockConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15138
func TestAccS3Bucket_Manage_lifecycleRuleAbortIncompleteMultipartUploadDaysNoExpiration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLifecycleRuleAbortIncompleteMultipartUploadDaysConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	partition := acctest.Partition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "0"),
				),
			},
			{
				Config:                  testAccBucketReplicationConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketReplicationWithConfigurationConfig(rInt, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithConfigurationConfig(rInt, "GLACIER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassGlacier),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithSseKMSEncryptedObjectsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccS3Bucket_Replication_multipleDestinationsEmptyFilter(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithMultipleDestinationsEmptyFilterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination3", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccBucketReplicationWithMultipleDestinationsEmptyFilterConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_multipleDestinationsNonEmptyFilter(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithMultipleDestinationsNonEmptyFilterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination3", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule3",
						"priority":                    "3",
						"status":                      "Disabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix3",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key3":          "Value3",
						"destination.#":               "1",
						"destination.0.storage_class": "ONEZONE_IA",
					}),
				),
			},
			{
				Config:                  testAccBucketReplicationWithMultipleDestinationsNonEmptyFilterConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_twoDestination(t *testing.T) {
	// This tests 2 destinations since GovCloud and possibly other non-standard partitions allow a max of 2
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        testAccErrorCheckSkipS3(t),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithMultipleDestinationsTwoDestinationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination2", acctest.RegionProviderFunc(alternateRegion, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule1",
						"priority":                    "1",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.prefix":             "prefix1",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "replication_configuration.0.rules.*", map[string]string{
						"id":                          "rule2",
						"priority":                    "2",
						"status":                      "Enabled",
						"filter.#":                    "1",
						"filter.0.tags.%":             "1",
						"filter.0.tags.Key2":          "Value2",
						"destination.#":               "1",
						"destination.0.storage_class": "STANDARD_IA",
					}),
				),
			},
			{
				Config:                  testAccBucketReplicationWithMultipleDestinationsTwoDestinationConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_ruleDestinationAccessControlTranslation(t *testing.T) {
	rInt := sdkacctest.RandInt()
	region := acctest.Region()
	partition := acctest.Partition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithAccessControlTranslationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:                  testAccBucketReplicationWithAccessControlTranslationConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccBucketReplicationWithSseKMSEncryptedObjectsAndAccessControlTranslationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									EncryptionConfiguration: &s3.EncryptionConfiguration{
										ReplicaKmsKeyID: aws.String("${aws_kms_key.replica.arn}"),
									},
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								SourceSelectionCriteria: &s3.SourceSelectionCriteria{
									SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
										Status: aws.String(s3.SseKmsEncryptedObjectsStatusEnabled),
									},
								},
							},
						},
					),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12480
func TestAccS3Bucket_Replication_ruleDestinationAddAccessControlTranslation(t *testing.T) {
	rInt := sdkacctest.RandInt()
	region := acctest.Region()
	partition := acctest.Partition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationConfigurationRulesDestinationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config:                  testAccBucketReplicationWithAccessControlTranslationConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl", "versioning"},
			},
			{
				Config: testAccBucketReplicationWithAccessControlTranslationConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Account:      aws.String("${data.aws_caller_identity.current.account_id}"),
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
									AccessControlTranslation: &s3.AccessControlTranslation{
										Owner: aws.String("Destination"),
									},
								},
								Prefix: aws.String("foo"),
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
		},
	})
}

// StorageClass issue: https://github.com/hashicorp/terraform/issues/10909
func TestAccS3Bucket_Replication_withoutStorageClass(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithoutStorageClassConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:                  testAccBucketReplicationWithoutStorageClassConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_expectVersioningValidationError(t *testing.T) {
	rInt := sdkacctest.RandInt()

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketReplicationNoVersioningConfig(rInt),
				ExpectError: regexp.MustCompile(`versioning must be enabled to allow S3 bucket replication`),
			},
		},
	})
}

// Prefix issue: https://github.com/hashicorp/terraform-provider-aws/issues/6340
func TestAccS3Bucket_Replication_withoutPrefix(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithoutPrefixConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
				),
			},
			{
				Config:                  testAccBucketReplicationWithoutPrefixConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_schemaV2(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	partition := acctest.Partition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithV2ConfigurationDeleteMarkerReplicationDisabledConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithV2ConfigurationNoTagsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusEnabled),
								},
							},
						},
					),
				),
			},
			{
				Config:                  testAccBucketReplicationWithV2ConfigurationNoTagsConfig(rInt),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketReplicationWithV2ConfigurationOnlyOneTagConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
										},
									},
								},
								Priority: aws.Int64(42),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithV2ConfigurationPrefixAndTagsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String("foo"),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
										},
									},
								},
								Priority: aws.Int64(41),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithV2ConfigurationMultipleTagsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("foobar"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									And: &s3.ReplicationRuleAndOperator{
										Prefix: aws.String(""),
										Tags: []*s3.Tag{
											{
												Key:   aws.String("ReplicateMe"),
												Value: aws.String("Yes"),
											},
											{
												Key:   aws.String("AnotherTag"),
												Value: aws.String("OK"),
											},
											{
												Key:   aws.String("Foo"),
												Value: aws.String("Bar"),
											},
										},
									},
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
							},
						},
					),
				),
			},
		},
	})
}

func TestAccS3Bucket_Replication_schemaV2SameRegion(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationResourceName := "aws_s3_bucket.destination"
	rNameDestination := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketSameRegionReplicationWithV2ConfigurationNoTagsConfig(rName, rNameDestination),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "replication_configuration.0.role", "iam", fmt.Sprintf("role/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExists(destinationResourceName),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("testid"),
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::%s", acctest.Partition(), rNameDestination)),
									StorageClass: aws.String(s3.ObjectStorageClassStandard),
								},
								Status: aws.String(s3.ReplicationRuleStatusEnabled),
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("testprefix"),
								},
								Priority: aws.Int64(0),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusEnabled),
								},
							},
						},
					),
				),
			},
			{
				Config:            testAccBucketSameRegionReplicationWithV2ConfigurationNoTagsConfig(rName, rNameDestination),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy", "acl"},
			},
		},
	})
}

func TestAccS3Bucket_Replication_RTC_valid(t *testing.T) {
	rInt := sdkacctest.RandInt()
	alternateRegion := acctest.AlternateRegion()
	region := acctest.Region()
	partition := acctest.Partition()
	iamRoleResourceName := "aws_iam_role.role"
	resourceName := "aws_s3_bucket.bucket"

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckBucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketReplicationWithReplicationConfigurationWithRTCConfig(rInt, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("rtc"),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
									Metrics: &s3.Metrics{
										EventThreshold: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.MetricsStatusEnabled),
									},
									ReplicationTime: &s3.ReplicationTime{
										Time: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.ReplicationTimeStatusEnabled),
									},
								},
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String(""),
								},
								Status:   aws.String(s3.ReplicationRuleStatusEnabled),
								Priority: aws.Int64(0),
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithReplicationConfigurationWithRTCNoMinutesConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("rtc-no-minutes"),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
									Metrics: &s3.Metrics{
										EventThreshold: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.MetricsStatusEnabled),
									},
									ReplicationTime: &s3.ReplicationTime{
										Time: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.ReplicationTimeStatusEnabled),
									},
								},
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String(""),
								},
								Status:   aws.String(s3.ReplicationRuleStatusEnabled),
								Priority: aws.Int64(0),
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithReplicationConfigurationWithRTCNoStatusConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("rtc-no-status"),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
									Metrics: &s3.Metrics{
										EventThreshold: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.MetricsStatusEnabled),
									},
									ReplicationTime: &s3.ReplicationTime{
										Time: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.ReplicationTimeStatusEnabled),
									},
								},
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								Status:   aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
			{
				Config: testAccBucketReplicationWithReplicationConfigurationWithRTCNoConfigConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExistsWithProvider(resourceName, acctest.RegionProviderFunc(region, &providers)),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_configuration.0.role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.0.replication_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_configuration.0.rules.0.destination.0.metrics.#", "1"),
					testAccCheckBucketExistsWithProvider("aws_s3_bucket.destination", acctest.RegionProviderFunc(alternateRegion, &providers)),
					testAccCheckBucketReplicationRules(
						resourceName,
						[]*s3.ReplicationRule{
							{
								ID: aws.String("rtc-no-config"),
								DeleteMarkerReplication: &s3.DeleteMarkerReplication{
									Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
								},
								Destination: &s3.Destination{
									Bucket:       aws.String(fmt.Sprintf("arn:%s:s3:::tf-test-bucket-destination-%d", partition, rInt)),
									StorageClass: aws.String(s3.StorageClassStandard),
									Metrics: &s3.Metrics{
										EventThreshold: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.MetricsStatusEnabled),
									},
									ReplicationTime: &s3.ReplicationTime{
										Time: &s3.ReplicationTimeValue{
											Minutes: aws.Int64(15),
										},
										Status: aws.String(s3.ReplicationTimeStatusEnabled),
									},
								},
								Filter: &s3.ReplicationRuleFilter{
									Prefix: aws.String("foo"),
								},
								Priority: aws.Int64(0),
								Status:   aws.String(s3.ReplicationRuleStatusEnabled),
							},
						},
					),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "aws_s3_bucket.arbitrary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockEnabledNoDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "acl"},
			},
			{
				Config: testAccBucketObjectLockEnabledWithDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.object_lock_enabled", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.mode", "COMPLIANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_configuration.0.rule.0.default_retention.0.days", "3"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroy(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

// By default, the AWS Go SDK cleans up URIs by removing extra slashes
// when the service API requests use the URI as part of making a request.
// While the aws_s3_bucket_object resource automatically cleans the key
// to not contain these extra slashes, out-of-band handling and other AWS
// services may create keys with extra slashes (empty "directory" prefixes).
func TestAccS3Bucket_Basic_forceDestroyWithEmptyPrefixes(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "/extraleadingslash.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithObjectLockEnabled(t *testing.T) {
	resourceName := "aws_s3_bucket.bucket"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroyWithObjectLockEnabled(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjectsWithLegalHold(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestBucketName(t *testing.T) {
	validDnsNames := []string{
		"foobar",
		"foo.bar",
		"foo.bar.baz",
		"1234",
		"foo-bar",
		strings.Repeat("x", 63),
	}

	for _, v := range validDnsNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsWest2RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidDnsNames := []string{
		"foo..bar",
		"Foo.Bar",
		"192.168.0.1",
		"127.0.0.1",
		".foo",
		"bar.",
		"foo_bar",
		strings.Repeat("x", 64),
	}

	for _, v := range invalidDnsNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsWest2RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}

	validEastNames := []string{
		"foobar",
		"foo_bar",
		"127.0.0.1",
		"foo..bar",
		"foo_bar_baz",
		"foo.bar.baz",
		"Foo.Bar",
		strings.Repeat("x", 255),
	}

	for _, v := range validEastNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsEast1RegionID); err != nil {
			t.Fatalf("%q should be a valid S3 bucket name", v)
		}
	}

	invalidEastNames := []string{
		"foo;bar",
		strings.Repeat("x", 256),
	}

	for _, v := range invalidEastNames {
		if err := tfs3.ValidBucketName(v, endpoints.UsEast1RegionID); err == nil {
			t.Fatalf("%q should not be a valid S3 bucket name", v)
		}
	}
}

func TestBucketRegionalDomainName(t *testing.T) {
	const bucket = "bucket-name"

	var testCases = []struct {
		ExpectedErrCount int
		ExpectedOutput   string
		Region           string
	}{
		{
			Region:           "",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           "custom",
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.custom.amazonaws.com",
		},
		{
			Region:           endpoints.UsEast1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + ".s3.amazonaws.com",
		},
		{
			Region:           endpoints.UsWest2RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsWest2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           endpoints.UsGovWest1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.%s", endpoints.UsGovWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			Region:           endpoints.CnNorth1RegionID,
			ExpectedErrCount: 0,
			ExpectedOutput:   bucket + fmt.Sprintf(".s3.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
		},
	}

	for _, tc := range testCases {
		output, err := tfs3.BucketRegionalDomainName(bucket, tc.Region)
		if tc.ExpectedErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Region, err)
		}
		if tc.ExpectedErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Region)
		}
		if output != tc.ExpectedOutput {
			t.Fatalf("expected %q, received %q", tc.ExpectedOutput, output)
		}
	}
}

func TestWebsiteEndpoint(t *testing.T) {
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/WebsiteEndpoints.html
	testCases := []struct {
		TestingClient      *conns.AWSClient
		LocationConstraint string
		Expected           string
	}{
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsEast1RegionID,
			},
			LocationConstraint: "",
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsWest2RegionID,
			},
			LocationConstraint: endpoints.UsWest2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsWest1RegionID,
			},
			LocationConstraint: endpoints.UsWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuWest1RegionID,
			},
			LocationConstraint: endpoints.EuWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.EuWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuWest3RegionID,
			},
			LocationConstraint: endpoints.EuWest3RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuWest3RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.EuCentral1RegionID,
			},
			LocationConstraint: endpoints.EuCentral1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.EuCentral1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSouth1RegionID,
			},
			LocationConstraint: endpoints.ApSouth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApSouth1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSoutheast1RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApNortheast1RegionID,
			},
			LocationConstraint: endpoints.ApNortheast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApNortheast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApSoutheast2RegionID,
			},
			LocationConstraint: endpoints.ApSoutheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.ApSoutheast2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.ApNortheast2RegionID,
			},
			LocationConstraint: endpoints.ApNortheast2RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.ApNortheast2RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.SaEast1RegionID,
			},
			LocationConstraint: endpoints.SaEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.SaEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsGovEast1RegionID,
			},
			LocationConstraint: endpoints.UsGovEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.%s", endpoints.UsGovEast1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com",
				Region:    endpoints.UsGovWest1RegionID,
			},
			LocationConstraint: endpoints.UsGovWest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website-%s.%s", endpoints.UsGovWest1RegionID, acctest.PartitionDNSSuffix()),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "c2s.ic.gov",
				Region:    endpoints.UsIsoEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsoEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.c2s.ic.gov", endpoints.UsIsoEast1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "sc2s.sgov.gov",
				Region:    endpoints.UsIsobEast1RegionID,
			},
			LocationConstraint: endpoints.UsIsobEast1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.sc2s.sgov.gov", endpoints.UsIsobEast1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com.cn",
				Region:    endpoints.CnNorthwest1RegionID,
			},
			LocationConstraint: endpoints.CnNorthwest1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorthwest1RegionID),
		},
		{
			TestingClient: &conns.AWSClient{
				DNSSuffix: "amazonaws.com.cn",
				Region:    endpoints.CnNorth1RegionID,
			},
			LocationConstraint: endpoints.CnNorth1RegionID,
			Expected:           fmt.Sprintf("bucket-name.s3-website.%s.amazonaws.com.cn", endpoints.CnNorth1RegionID),
		},
	}

	for _, testCase := range testCases {
		got := tfs3.WebsiteEndpoint(testCase.TestingClient, "bucket-name", testCase.LocationConstraint)
		if got.Endpoint != testCase.Expected {
			t.Errorf("WebsiteEndpointUrl(\"bucket-name\", %q) => %q, want %q", testCase.LocationConstraint, got.Endpoint, testCase.Expected)
		}
	}
}

// testAccErrorCheckSkipS3 skips tests that have error messages indicating unsupported features
func testAccErrorCheckSkipS3(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Number of distinct destination bucket ARNs cannot exceed",
	)
}

func testAccCheckBucketDestroy(s *terraform.State) error {
	return testAccCheckBucketDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckBucketDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket" {
			continue
		}

		input := &s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		// Retry for S3 eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.HeadBucket(input)

			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") || tfawserr.ErrMessageContains(err, "NotFound", "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("AWS S3 Bucket still exists: %s", rs.Primary.ID))
		})

		if tfresource.TimedOut(err) {
			_, err = conn.HeadBucket(input)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckBucketExists(n string) resource.TestCheckFunc {
	return testAccCheckBucketExistsWithProvider(n, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckBucketExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*conns.AWSClient).S3Conn
		_, err := conn.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			return err
		}
		return nil

	}
}

func testAccCheckDestroyBucket(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		_, err := conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error destroying Bucket (%s) in testAccCheckDestroyBucket: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

func testAccCheckBucketAddObjects(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ConnURICleaningDisabled

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

func testAccCheckBucketAddObjectsWithLegalHold(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket:                    aws.String(rs.Primary.ID),
				Key:                       aws.String(key),
				ObjectLockLegalHoldStatus: aws.String(s3.ObjectLockLegalHoldStatusOn),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

// Create an S3 bucket via a CF stack so that it has system tags.
func testAccCheckBucketCreateViaCloudFormation(n string, stackID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn
		stackName := sdkacctest.RandomWithPrefix("tf-acc-test-s3tags")
		templateBody := fmt.Sprintf(`{
  "Resources": {
    "TfTestBucket": {
      "Type": "AWS::S3::Bucket",
      "Properties": {
        "BucketName": "%s"
      }
    }
  }
}`, n)

		requestToken := resource.UniqueId()
		req := &cloudformation.CreateStackInput{
			StackName:          aws.String(stackName),
			TemplateBody:       aws.String(templateBody),
			ClientRequestToken: aws.String(requestToken),
		}

		log.Printf("[DEBUG] Creating CloudFormation stack: %s", req)
		resp, err := conn.CreateStack(req)
		if err != nil {
			return fmt.Errorf("error creating CloudFormation stack: %w", err)
		}

		stack, err := tfcloudformation.WaitStackCreated(conn, aws.StringValue(resp.StackId), requestToken, 10*time.Minute)
		if err != nil {
			return fmt.Errorf("Error waiting for CloudFormation stack creation: %w", err)
		}
		status := aws.StringValue(stack.StackStatus)
		if status != cloudformation.StackStatusCreateComplete {
			return fmt.Errorf("Invalid CloudFormation stack creation status: %s", status)
		}

		*stackID = aws.StringValue(resp.StackId)
		return nil
	}
}

func testAccCheckBucketTagKeys(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		got, err := tfs3.BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		for _, want := range keys {
			ok := false
			for _, key := range got.Keys() {
				if want == key {
					ok = true
					break
				}
			}
			if !ok {
				return fmt.Errorf("Key %s not found in bucket's tag set", want)
			}
		}

		return nil
	}
}

func testAccCheckBucketPolicy(n string, policy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if policy == "" {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucketPolicy" {
				// expected
				return nil
			}
			if err == nil {
				return fmt.Errorf("Expected no policy, got: %#v", *out.Policy)
			} else {
				return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
			}
		}
		if err != nil {
			return fmt.Errorf("GetBucketPolicy error: %v, expected %s", err, policy)
		}

		if v := out.Policy; v == nil {
			if policy != "" {
				return fmt.Errorf("bad policy, found nil, expected: %s", policy)
			}
		} else {
			expected := make(map[string]interface{})
			if err := json.Unmarshal([]byte(policy), &expected); err != nil {
				return err
			}
			actual := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*v), &actual); err != nil {
				return err
			}

			if !reflect.DeepEqual(expected, actual) {
				return fmt.Errorf("bad policy, expected: %#v, got %#v", expected, actual)
			}
		}

		return nil
	}
}

func testAccCheckBucketWebsite(n string, indexDoc string, errorDoc string, redirectProtocol string, redirectTo string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if indexDoc == "" {
				// If we want to assert that the website is not there, than
				// this error is expected
				return nil
			} else {
				return fmt.Errorf("S3BucketWebsite error: %v", err)
			}
		}

		if v := out.IndexDocument; v == nil {
			if indexDoc != "" {
				return fmt.Errorf("bad index doc, found nil, expected: %s", indexDoc)
			}
		} else {
			if *v.Suffix != indexDoc {
				return fmt.Errorf("bad index doc, expected: %s, got %#v", indexDoc, out.IndexDocument)
			}
		}

		if v := out.ErrorDocument; v == nil {
			if errorDoc != "" {
				return fmt.Errorf("bad error doc, found nil, expected: %s", errorDoc)
			}
		} else {
			if *v.Key != errorDoc {
				return fmt.Errorf("bad error doc, expected: %s, got %#v", errorDoc, out.ErrorDocument)
			}
		}

		if v := out.RedirectAllRequestsTo; v == nil {
			if redirectTo != "" {
				return fmt.Errorf("bad redirect to, found nil, expected: %s", redirectTo)
			}
		} else {
			if *v.HostName != redirectTo {
				return fmt.Errorf("bad redirect to, expected: %s, got %#v", redirectTo, out.RedirectAllRequestsTo)
			}
			if redirectProtocol != "" && v.Protocol != nil && *v.Protocol != redirectProtocol {
				return fmt.Errorf("bad redirect protocol to, expected: %s, got %#v", redirectProtocol, out.RedirectAllRequestsTo)
			}
		}

		return nil
	}
}

func testAccCheckBucketWebsiteRoutingRules(n string, routingRules []*s3.RoutingRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if routingRules == nil {
				return nil
			}
			return fmt.Errorf("GetBucketWebsite error: %v", err)
		}

		if !reflect.DeepEqual(out.RoutingRules, routingRules) {
			return fmt.Errorf("bad routing rule, expected: %v, got %v", routingRules, out.RoutingRules)
		}

		return nil
	}
}

func testAccCheckBucketVersioning(n string, versioningStatus string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketVersioning error: %v", err)
		}

		if v := out.Status; v == nil {
			if versioningStatus != "" {
				return fmt.Errorf("bad error versioning status, found nil, expected: %s", versioningStatus)
			}
		} else {
			if *v != versioningStatus {
				return fmt.Errorf("bad error versioning status, expected: %s, got %s", versioningStatus, *v)
			}
		}

		return nil
	}
}

func testAccCheckBucketCors(n string, corsRules []*s3.CORSRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() != "NoSuchCORSConfiguration" {
				return fmt.Errorf("GetBucketCors error: %v", err)
			}
		}

		if !reflect.DeepEqual(out.CORSRules, corsRules) {
			return fmt.Errorf("bad error cors rule, expected: %v, got %v", corsRules, out.CORSRules)
		}

		return nil
	}
}

func testAccCheckRequestPayer(n, expectedPayer string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketRequestPayment(&s3.GetBucketRequestPaymentInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketRequestPayment error: %v", err)
		}

		if *out.Payer != expectedPayer {
			return fmt.Errorf("bad error request payer type, expected: %v, got %v",
				expectedPayer, out.Payer)
		}

		return nil
	}
}

func testAccCheckBucketLogging(n, b, p string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		out, err := conn.GetBucketLogging(&s3.GetBucketLoggingInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketLogging error: %v", err)
		}

		if out.LoggingEnabled == nil {
			return fmt.Errorf("logging not enabled for bucket: %s", rs.Primary.ID)
		}

		tb := s.RootModule().Resources[b]

		if v := out.LoggingEnabled.TargetBucket; v == nil {
			if tb.Primary.ID != "" {
				return fmt.Errorf("bad target bucket, found nil, expected: %s", tb.Primary.ID)
			}
		} else {
			if *v != tb.Primary.ID {
				return fmt.Errorf("bad target bucket, expected: %s, got %s", tb.Primary.ID, *v)
			}
		}

		if v := out.LoggingEnabled.TargetPrefix; v == nil {
			if p != "" {
				return fmt.Errorf("bad target prefix, found nil, expected: %s", p)
			}
		} else {
			if *v != p {
				return fmt.Errorf("bad target prefix, expected: %s, got %s", p, *v)
			}
		}

		return nil
	}
}

func testAccCheckBucketReplicationRules(n string, rules []*s3.ReplicationRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		for _, rule := range rules {
			if dest := rule.Destination; dest != nil {
				if account := dest.Account; account != nil && strings.HasPrefix(aws.StringValue(dest.Account), "${") {
					resourceReference := strings.Replace(aws.StringValue(dest.Account), "${", "", 1)
					resourceReference = strings.Replace(resourceReference, "}", "", 1)
					resourceReferenceParts := strings.Split(resourceReference, ".")
					resourceAttribute := resourceReferenceParts[len(resourceReferenceParts)-1]
					resourceName := strings.Join(resourceReferenceParts[:len(resourceReferenceParts)-1], ".")
					value := s.RootModule().Resources[resourceName].Primary.Attributes[resourceAttribute]
					dest.Account = aws.String(value)
				}
				if ec := dest.EncryptionConfiguration; ec != nil {
					if ec.ReplicaKmsKeyID != nil {
						key_arn := s.RootModule().Resources["aws_kms_key.replica"].Primary.Attributes["arn"]
						ec.ReplicaKmsKeyID = aws.String(strings.Replace(*ec.ReplicaKmsKeyID, "${aws_kms_key.replica.arn}", key_arn, -1))
					}
				}
			}
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		out, err := conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("S3 bucket not found")
			}
			if rules == nil {
				return nil
			}
			return fmt.Errorf("GetReplicationConfiguration error: %v", err)
		}

		for _, rule := range out.ReplicationConfiguration.Rules {
			// Sort filter tags by key.
			if filter := rule.Filter; filter != nil {
				if and := filter.And; and != nil {
					if tags := and.Tags; tags != nil {
						sort.Slice(tags, func(i, j int) bool { return *tags[i].Key < *tags[j].Key })
					}
				}
			}
		}
		if !reflect.DeepEqual(out.ReplicationConfiguration.Rules, rules) {
			return fmt.Errorf("bad replication rules, expected: %v, got %v", rules, out.ReplicationConfiguration.Rules)
		}

		return nil
	}
}

func testAccCheckS3BucketDomainName(resourceName string, attributeName string, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedValue := acctest.Provider.Meta().(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.s3", bucketName))

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccBucketRegionalDomainName(bucket, region string) string {
	regionalEndpoint, err := tfs3.BucketRegionalDomainName(bucket, region)
	if err != nil {
		return fmt.Sprintf("Regional endpoint not found for bucket %s", bucket)
	}
	return regionalEndpoint
}

func testAccCheckS3BucketWebsiteEndpoint(resourceName string, attributeName string, bucketName string, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		website := tfs3.WebsiteEndpoint(acctest.Provider.Meta().(*conns.AWSClient), bucketName, region)
		expectedValue := website.Endpoint

		return resource.TestCheckResourceAttr(resourceName, attributeName, expectedValue)(s)
	}
}

func testAccCheckBucketUpdateTags(n string, oldTags, newTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		return tfs3.BucketUpdateTags(conn, rs.Primary.Attributes["bucket"], oldTags, newTags)
	}
}

func testAccCheckBucketCheckTags(n string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		got, err := tfs3.BucketListTags(conn, rs.Primary.Attributes["bucket"])
		if err != nil {
			return err
		}

		want := tftags.New(expectedTags)
		if !reflect.DeepEqual(want, got) {
			return fmt.Errorf("Incorrect tags, want: %v got: %v", want, got)
		}

		return nil
	}
}

func testAccBucketPolicy(bucketName, partition string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:%[1]s:s3:::%[2]s/*"
    }
  ]
}`, partition, bucketName)
}

func testAccBucketConfig_Basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketConfig_withNoTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = false
}
`, bucketName)
}

func testAccBucketConfig_withTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = false

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, bucketName)
}

func testAccBucketConfig_withUpdatedTags(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = false

  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }
}
`, bucketName)
}

func testAccMultiBucketWithTagsConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket1" {
  bucket        = "tf-test-bucket-1-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-1-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket2" {
  bucket        = "tf-test-bucket-2-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-2-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket3" {
  bucket        = "tf-test-bucket-3-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-3-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket4" {
  bucket        = "tf-test-bucket-4-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-4-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket5" {
  bucket        = "tf-test-bucket-5-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-5-%[1]d"
    Environment = "%[1]d"
  }
}

resource "aws_s3_bucket" "bucket6" {
  bucket        = "tf-test-bucket-6-%[1]d"
  acl           = "private"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-6-%[1]d"
    Environment = "%[1]d"
  }
}
`, randInt)
}

func testAccBucketWebsiteConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
  }
}
`, bucketName)
}

func testAccBucketWebsiteWithErrorConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}
`, bucketName)
}

func testAccBucketWebsiteWithRedirectConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    redirect_all_requests_to = "hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccBucketWebsiteWithHTTPSRedirectConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    redirect_all_requests_to = "https://hashicorp.com?my=query"
  }
}
`, bucketName)
}

func testAccBucketWebsiteWithRoutingRulesConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[
  {
    "Condition": {
      "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
      "ReplaceKeyPrefixWith": "documents/"
    }
  }
]
EOF

  }
}
`, bucketName)
}

func testAccBucketWithAccelerationConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = %[1]q
  acceleration_status = "Enabled"
}
`, bucketName)
}

func testAccBucketWithoutAccelerationConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket              = %[1]q
  acceleration_status = "Suspended"
}
`, bucketName)
}

func testAccBucketRequestPayerBucketOwnerConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  request_payer = "BucketOwner"
}
`, bucketName)
}

func testAccBucketRequestPayerRequesterConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = %[1]q
  request_payer = "Requester"
}
`, bucketName)
}

func testAccBucketWithPolicyConfig(bucketName, partition string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
  policy = %[2]s
}
`, bucketName, strconv.Quote(testAccBucketPolicy(bucketName, partition)))
}

func testAccBucketDestroyedConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
}
`, bucketName)
}

func testAccBucketEnableDefaultEncryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "arbitrary" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}
`, bucketName)
}

func testAccBucketKeyEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "arbitrary" {
  description             = "KMS Key for Bucket %[1]s"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.arbitrary.arn
        sse_algorithm     = "aws:kms"
      }
      bucket_key_enabled = true
    }
  }
}
`, bucketName)
}

func testAccBucketEnableDefaultEncryptionWithAES256(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
`, bucketName)
}

func testAccBucketEnableDefaultEncryptionWithDefaultKey(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "aws:kms"
      }
    }
  }
}
`, bucketName)
}

func testAccBucketDisableDefaultEncryption(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccBucketWithEmptyPolicyConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
  policy = ""
}
`, bucketName)
}

func testAccBucketWithVersioningConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = true
  }
}
`, bucketName)
}

func testAccBucketWithDisableVersioningConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  versioning {
    enabled = false
  }
}
`, bucketName)
}

func testAccBucketWithCORSConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccBucketWithCORSEmptyOriginConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = [""]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, bucketName)
}

func testAccBucketWithACLConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"
}
`, bucketName)
}

func testAccBucketWithACLUpdateConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"
}
`, bucketName)
}

func testAccBucketWithGrantsConfig(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["FULL_CONTROL", "WRITE"]
  }
}
`, bucketName)
}

func testAccBucketWithGrantsUpdateConfig(bucketName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    type        = "CanonicalUser"
    permissions = ["READ"]
  }

  grant {
    type        = "Group"
    permissions = ["READ_ACP"]
    uri         = "http://acs.amazonaws.com/groups/s3/LogDelivery"
  }
}
`, bucketName)
}

func testAccBucketWithLoggingConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  logging {
    target_bucket = aws_s3_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
`, bucketName)
}

func testAccBucketWithLifecycleConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 60
      storage_class = "INTELLIGENT_TIERING"
    }

    transition {
      days          = 90
      storage_class = "ONEZONE_IA"
    }

    transition {
      days          = 120
      storage_class = "GLACIER"
    }

    transition {
      days          = 210
      storage_class = "DEEP_ARCHIVE"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id4"
    prefix  = "path4/"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    expiration {
      date = "2016-01-12"
    }
  }

  lifecycle_rule {
    id      = "id5"
    enabled = true

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id6"
    enabled = true

    tags = {
      "tagKey" = "tagValue"
    }

    transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
}
`, bucketName)
}

func testAccBucketWithLifecycleExpireMarkerConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      expired_object_delete_marker = "true"
    }
  }
}
`, bucketName)
}

func testAccBucketWithVersioningLifecycleConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  versioning {
    enabled = false
  }

  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    noncurrent_version_expiration {
      days = 365
    }

    noncurrent_version_transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    noncurrent_version_transition {
      days          = 60
      storage_class = "GLACIER"
    }
  }

  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = false

    noncurrent_version_expiration {
      days = 365
    }
  }

  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = true

    noncurrent_version_transition {
      days          = 0
      storage_class = "GLACIER"
    }
  }
}
`, bucketName)
}

func testAccBucketLifecycleRuleExpirationEmptyConfigurationBlockConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    enabled = true
    id      = "id1"

    expiration {}
  }
}
`, rName)
}

func testAccBucketLifecycleRuleAbortIncompleteMultipartUploadDaysConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  lifecycle_rule {
    abort_incomplete_multipart_upload_days = 7
    enabled                                = true
    id                                     = "id1"
  }
}
`, rName)
}

func testAccBucketReplicationBasicConfig(randInt int) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "role" {
  name = "tf-iam-role-replication-%[1]d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination-%[1]d"

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccBucketReplicationConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccBucketReplicationWithConfigurationConfig(randInt int, storageClass string) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "%[2]s"
      }
    }
  }
}
`, randInt, storageClass)
}

func testAccBucketReplicationWithReplicationConfigurationWithRTCConfig(randInt int, minutes int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-rtc-%[1]d"
  acl    = "private"
  versioning {
    enabled = true
  }
  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc"
      status = "Enabled"
      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {
          status  = "Enabled"
          minutes = %[2]d
        }
        replication_time {
          status  = "Enabled"
          minutes = %[2]d
        }
      }
    }
  }
}
`, randInt, minutes))
}

func testAccBucketReplicationWithReplicationConfigurationWithRTCNoMinutesConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-rtc-no-minutes-%[1]d"
  acl    = "private"
  versioning {
    enabled = true
  }
  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-minutes"
      status = "Enabled"
      filter {
        tags = {}
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {
          status = "Enabled"
        }
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithReplicationConfigurationWithRTCNoStatusConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-rtc-no-minutes-%[1]d"
  acl    = "private"
  versioning {
    enabled = true
  }
  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-status"
      status = "Enabled"
      filter {
        prefix = "foo"
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {
          minutes = 15
        }
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithReplicationConfigurationWithRTCNoConfigConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-rtc-no-config-%[1]d"
  acl    = "private"
  versioning {
    enabled = true
  }
  replication_configuration {
    role = aws_iam_role.role.arn
    rules {
      id     = "rtc-no-config"
      status = "Enabled"
      filter {
        prefix = "foo"
      }
      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
        metrics {}
        replication_time {}
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithMultipleDestinationsEmptyFilterConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {}

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithMultipleDestinationsNonEmptyFilterConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination3" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination3-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }

    rules {
      id       = "rule3"
      priority = 3
      status   = "Disabled"

      filter {
        prefix = "prefix3"

        tags = {
          Key3 = "Value3"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination3.arn
        storage_class = "ONEZONE_IA"
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithMultipleDestinationsTwoDestinationConfig(randInt int) string {
	return acctest.ConfigCompose(
		testAccBucketReplicationBasicConfig(randInt),
		fmt.Sprintf(`
resource "aws_s3_bucket" "destination2" {
  provider = "awsalternate"
  bucket   = "tf-test-bucket-destination2-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id       = "rule1"
      priority = 1
      status   = "Enabled"

      filter {
        prefix = "prefix1"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }

    rules {
      id       = "rule2"
      priority = 2
      status   = "Enabled"

      filter {
        tags = {
          Key2 = "Value2"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination2.arn
        storage_class = "STANDARD_IA"
      }
    }
  }
}
`, randInt))
}

func testAccBucketReplicationWithSseKMSEncryptedObjectsConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithAccessControlTranslationConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"

        access_control_translation {
          owner = "Destination"
        }
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationConfigurationRulesDestinationConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "bucket" {
  acl    = "private"
  bucket = "tf-test-bucket-%[1]d"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id    = data.aws_caller_identity.current.account_id
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }

  versioning {
    enabled = true
  }
}
`, randInt)
}

func testAccBucketReplicationWithSseKMSEncryptedObjectsAndAccessControlTranslationConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "replica" {
  provider                = "awsalternate"
  description             = "TF Acceptance Test S3 repl KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        account_id         = data.aws_caller_identity.current.account_id
        bucket             = aws_s3_bucket.destination.arn
        storage_class      = "STANDARD"
        replica_kms_key_id = aws_kms_key.replica.arn

        access_control_translation {
          owner = "Destination"
        }
      }

      source_selection_criteria {
        sse_kms_encrypted_objects {
          enabled = true
        }
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithoutStorageClassConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket = aws_s3_bucket.destination.arn
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithoutPrefixConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationNoVersioningConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketSameRegionReplicationWithV2ConfigurationNoTagsConfig(rName, rNameDestination string) string {
	return acctest.ConfigCompose(testAccBucketReplicationConfig_iamPolicy(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.test.arn

    rules {
      id     = "testid"
      status = "Enabled"

      filter {
        prefix = "testprefix"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}

resource "aws_s3_bucket" "destination" {
  bucket = %[2]q

  versioning {
    enabled = true
  }
}
`, rName, rNameDestination))
}

func testAccBucketReplicationWithV2ConfigurationDeleteMarkerReplicationDisabledConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithV2ConfigurationNoTagsConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        prefix = "foo"
      }

      delete_marker_replication_status = "Enabled"

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithV2ConfigurationOnlyOneTagConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 42

      filter {
        tags = {
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithV2ConfigurationPrefixAndTagsConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      priority = 41

      filter {
        prefix = "foo"

        tags = {
          AnotherTag  = "OK"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketReplicationWithV2ConfigurationMultipleTagsConfig(randInt int) string {
	return testAccBucketReplicationBasicConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  versioning {
    enabled = true
  }

  replication_configuration {
    role = aws_iam_role.role.arn

    rules {
      id     = "foobar"
      status = "Enabled"

      filter {
        tags = {
          AnotherTag  = "OK"
          Foo         = "Bar"
          ReplicateMe = "Yes"
        }
      }

      destination {
        bucket        = aws_s3_bucket.destination.arn
        storage_class = "STANDARD"
      }
    }
  }
}
`, randInt)
}

func testAccBucketObjectLockEnabledNoDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketObjectLockEnabledWithDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "arbitrary" {
  bucket = %[1]q

  object_lock_configuration {
    object_lock_enabled = "Enabled"

    rule {
      default_retention {
        mode = "COMPLIANCE"
        days = 3
      }
    }
  }
}
`, bucketName)
}

func testAccBucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  acl           = "private"
  force_destroy = true
}
`, bucketName)
}

func testAccBucketConfig_forceDestroyWithObjectLockEnabled(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket        = "%s"
  acl           = "private"
  force_destroy = true

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}
`, bucketName)
}

func testAccBucketReplicationConfig_iamPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}
`, rName)
}

const testAccBucketBucketEmptyStringConfig = `
resource "aws_s3_bucket" "test" {
  bucket = ""
}
`

const testAccBucketConfig_namePrefix = `
resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`

const testAccBucketConfig_generatedName = `
resource "aws_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`
