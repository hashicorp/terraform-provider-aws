package ssmincidents_test

import (
	// goimports -w <file> fixes these imports.
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	tfssmincidents "github.com/hashicorp/terraform-provider-aws/internal/service/ssmincidents"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestUpdateRegionIsValidUnitTest(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    map[string]string
		Expected bool
	}{
		{
			TestName: "valid empty",
			Input: map[string]string{
				"reg1": "",
				"reg2": "",
				"reg3": "",
			},
			Expected: true,
		},
		{
			TestName: "single item",
			Input: map[string]string{
				"reg1": "someKey",
			},
			Expected: true,
		},
		{
			TestName: "valid with keys",
			Input: map[string]string{
				"reg1": "key1",
				"reg2": "key2",
				"reg3": "key3",
			},
			Expected: true,
		},
		{
			TestName: "invalid mix",
			Input: map[string]string{
				"reg1": "",
				"reg2": "",
				"reg3": "key",
			},
			Expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := tfssmincidents.UpdateRegionsIsValid(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
func TestAccSSMIncidentsReplicationSet_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := "us-west-2"
	region2 := "ap-southeast-2"
	rKey := sdkacctest.RandString(26)
	rVal := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basic(region1, "", region2, "", rKey, rVal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						region1: "",
						region2: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tags.*", map[string]string{
						rKey: rVal,
					}),

					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssmincidents", regexp.MustCompile(`replicationset:+.`)),
				),
			},
		},
	})
}

func TestAccSSMIncidentsReplicationSet_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := "us-west-2"
	iniRegion2 := "ap-southeast-2"
	updRegion2 := "eu-west-2"
	rKey1 := sdkacctest.RandString(26)
	rVal1Ini := sdkacctest.RandString(26)
	rVal1Updated := sdkacctest.RandString(26)
	rKey2 := sdkacctest.RandString(26)
	rVal2 := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basic(region1, "", iniRegion2, "", rKey1, rVal1Ini),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						region1:    "",
						iniRegion2: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tags.*", map[string]string{
						rKey1: rVal1Ini,
					}),

					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssmincidents", regexp.MustCompile(`replicationset:+.`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_basic(region1, "", updRegion2, "", rKey1, rVal1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						region1:    "",
						updRegion2: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tags.*", map[string]string{
						rKey1: rVal1Updated,
					}),

					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssmincidents", regexp.MustCompile(`replicationset:+.`)),
				),
			},
			{
				Config: testAccReplicationSetConfig_basic(region1, "", updRegion2, "", rKey2, rVal2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regions.*", map[string]string{
						region1:    "",
						updRegion2: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tags.*", map[string]string{
						rKey2: rVal2,
					}),

					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ssmincidents", regexp.MustCompile(`replicationset:+.`)),
				),
			},
		},
	})
}

func TestAccSSMIncidentsReplicationSet_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_ssmincidents_replication_set.test"
	region1 := "us-west-2"
	region2 := "ap-southeast-2"
	rKey := sdkacctest.RandString(26)
	rVal := sdkacctest.RandString(26)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(names.SSMIncidentsEndpointID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetConfig_basic(region1, "", region2, "", rKey, rVal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfssmincidents.ResourceReplicationSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TODO:
// single region change just CMK
// single region change region
// multiple region error when updating to mix of cmk and not cmk
// multiple region valid when just changing cmk

func testAccCheckReplicationSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssmincidents_replication_set" {
			continue
		}

		_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, rs.Primary.ID,
				errors.New("expected resource not found error, received an unexpected error"))
		}

		return create.Error(names.SSMIncidents, create.ErrActionCheckingDestroyed, tfssmincidents.ResNameReplicationSet, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckReplicationSetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMIncidentsClient
		ctx := context.Background()

		_, err := tfssmincidents.FindReplicationSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SSMIncidents, create.ErrActionCheckingExistence, tfssmincidents.ResNameReplicationSet, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccReplicationSetConfig_basic(region1, region1key, region2, region2key, tagKey, tagVal string) string {
	return fmt.Sprintf(`

resource "aws_ssmincidents_replication_set" "test" {
  regions {
	%[1]q = %[2]q
	%[3]q = %[4]q
  }

  tags {
	%[5]q = %[6]q
  }
}
`, trim(region1), region1key, trim(region2), region2key, trim(tagKey), tagVal)
}

func trim(s string) string {
	return strings.Trim(s, "\"")
}
