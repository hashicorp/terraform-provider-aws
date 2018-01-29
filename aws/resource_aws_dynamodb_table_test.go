package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestDiffDynamoDbGSI(t *testing.T) {
	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates []*dynamodb.GlobalSecondaryIndexUpdate
	}{
		{ // No-op
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{},
		},

		{ // Creation
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					"name":            "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att2"),
								KeyType:       aws.String("HASH"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(11),
						},
						Projection: &dynamodb.Projection{
							ProjectionType: aws.String("ALL"),
						},
					},
				},
			},
		},

		{ // Deletion
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					"name":            "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
					},
				},
			},
		},

		{ // Update
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  20,
					"read_capacity":   30,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(20),
							ReadCapacityUnits:  aws.Int64(30),
						},
					},
				},
			},
		},

		{ // Update of non-capacity attributes
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "KEYS_ONLY",
					"non_key_attributes": []interface{}{"RandomAttribute"},
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       aws.String("HASH"),
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       aws.String("RANGE"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(10),
							ReadCapacityUnits:  aws.Int64(10),
						},
						Projection: &dynamodb.Projection{
							ProjectionType:   aws.String("KEYS_ONLY"),
							NonKeyAttributes: aws.StringSlice([]string{"RandomAttribute"}),
						},
					},
				},
			},
		},

		{ // Update of all attributes
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     12,
					"read_capacity":      12,
					"projection_type":    "INCLUDE",
					"non_key_attributes": []interface{}{"RandomAttribute"},
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       aws.String("HASH"),
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       aws.String("RANGE"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(12),
						},
						Projection: &dynamodb.Projection{
							ProjectionType:   aws.String("INCLUDE"),
							NonKeyAttributes: aws.StringSlice([]string{"RandomAttribute"}),
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		ops, err := diffDynamoDbGSI(tc.Old, tc.New)
		if err != nil {
			t.Fatal(err)
		}

		// Convert to strings to avoid dealing with pointers
		opsS := fmt.Sprintf("%v", ops)
		opsExpectedS := fmt.Sprintf("%v", tc.ExpectedUpdates)

		if opsS != opsExpectedS {
			t.Fatalf("Case #%d: Given:\n%s\n\nExpected:\n%s",
				i, opsS, opsExpectedS)
		}
	}
}

func TestAccAWSDynamoDbTable_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigAddSecondaryGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableWasUpdated("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_streamSpecification(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigStreamSpecification(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
					resource.TestCheckResourceAttr(
						"aws_dynamodb_table.basic-dynamodb-table", "stream_enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_dynamodb_table.basic-dynamodb-table", "stream_view_type", "KEYS_ONLY"),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_arn"),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_label"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_tags(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
					resource.TestCheckResourceAttr(
						"aws_dynamodb_table.basic-dynamodb-table", "tags.%", "3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/13243
func TestAccAWSDynamoDbTable_gsiUpdateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.write_capacity", "10"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedCapacity(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1458161653.read_capacity", "20"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1458161653.write_capacity", "20"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2102366979.read_capacity", "20"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2102366979.write_capacity", "20"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.4183493016.read_capacity", "20"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.4183493016.write_capacity", "20"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_gsiUpdateOtherAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2147693858.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.hash_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.800193359.write_capacity", "10"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.non_key_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.write_capacity", "10"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/566
func TestAccAWSDynamoDbTable_gsiUpdateNonKeyAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.non_key_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.536018608.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.write_capacity", "10"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedNonKeyAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2842459794.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.non_key_attributes.#", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.non_key_attributes.1", "AnotherAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3566650036.write_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.read_capacity", "10"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.68661177.write_capacity", "10"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_ttl(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigAddTimeToLive(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableTimeToLiveWasUpdated("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
		},
	})
}
func testAccCheckDynamoDbTableTimeToLiveWasUpdated(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		params := &dynamodb.DescribeTimeToLiveInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTimeToLive(params)

		if err != nil {
			return fmt.Errorf("[ERROR] Problem describing time to live for table '%s': %s", rs.Primary.ID, err)
		}

		ttlDescription := resp.TimeToLiveDescription

		log.Printf("[DEBUG] Checking on table %s", rs.Primary.ID)

		if *ttlDescription.TimeToLiveStatus != dynamodb.TimeToLiveStatusEnabled {
			return fmt.Errorf("TimeToLiveStatus %s, not ENABLED!", *ttlDescription.TimeToLiveStatus)
		}

		if *ttlDescription.AttributeName != "TestTTL" {
			return fmt.Errorf("AttributeName was %s, not TestTTL!", *ttlDescription.AttributeName)
		}

		return nil
	}
}

func TestResourceAWSDynamoDbTableStreamViewType_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "KEYS-ONLY",
			ErrCount: 1,
		},
		{
			Value:    "RANDOM-STRING",
			ErrCount: 1,
		},
		{
			Value:    "KEYS_ONLY",
			ErrCount: 0,
		},
		{
			Value:    "NEW_AND_OLD_IMAGES",
			ErrCount: 0,
		},
		{
			Value:    "NEW_IMAGE",
			ErrCount: 0,
		},
		{
			Value:    "OLD_IMAGE",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateStreamViewType(tc.Value, "aws_dynamodb_table_stream_view_type")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DynamoDB stream_view_type to trigger a validation error")
		}
	}
}

func testAccCheckAWSDynamoDbTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB table %s exists", rs.Primary.ID)
		// Check if queue exists by checking for its attributes
		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTable(params)
		if err == nil {
			return fmt.Errorf("DynamoDB table %s still exists. Failing!", rs.Primary.ID)
		}

		// Verify the error is what we want
		if dbErr, ok := err.(awserr.Error); ok && dbErr.Code() == "ResourceNotFoundException" {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckInitialAWSDynamoDbTableExists(n string, table *dynamodb.DescribeTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTable(params)

		if err != nil {
			return fmt.Errorf("[ERROR] Problem describing table '%s': %s", rs.Primary.ID, err)
		}

		*table = *resp

		return nil
	}
}

func testAccCheckInitialAWSDynamoDbTableConf(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTable(params)

		if err != nil {
			return fmt.Errorf("[ERROR] Problem describing table '%s': %s", rs.Primary.ID, err)
		}

		table := resp.Table

		log.Printf("[DEBUG] Checking on table %s", rs.Primary.ID)

		if *table.ProvisionedThroughput.WriteCapacityUnits != 20 {
			return fmt.Errorf("Provisioned write capacity was %d, not 20!", table.ProvisionedThroughput.WriteCapacityUnits)
		}

		if *table.ProvisionedThroughput.ReadCapacityUnits != 10 {
			return fmt.Errorf("Provisioned read capacity was %d, not 10!", table.ProvisionedThroughput.ReadCapacityUnits)
		}

		attrCount := len(table.AttributeDefinitions)
		gsiCount := len(table.GlobalSecondaryIndexes)
		lsiCount := len(table.LocalSecondaryIndexes)

		if attrCount != 4 {
			return fmt.Errorf("There were %d attributes, not 4 like there should have been!", attrCount)
		}

		if gsiCount != 1 {
			return fmt.Errorf("There were %d GSIs, not 1 like there should have been!", gsiCount)
		}

		if lsiCount != 1 {
			return fmt.Errorf("There were %d LSIs, not 1 like there should have been!", lsiCount)
		}

		attrmap := dynamoDbAttributesToMap(&table.AttributeDefinitions)
		if attrmap["TestTableHashKey"] != "S" {
			return fmt.Errorf("Test table hash key was of type %s instead of S!", attrmap["TestTableHashKey"])
		}
		if attrmap["TestTableRangeKey"] != "S" {
			return fmt.Errorf("Test table range key was of type %s instead of S!", attrmap["TestTableRangeKey"])
		}
		if attrmap["TestLSIRangeKey"] != "N" {
			return fmt.Errorf("Test table LSI range key was of type %s instead of N!", attrmap["TestLSIRangeKey"])
		}
		if attrmap["TestGSIRangeKey"] != "S" {
			return fmt.Errorf("Test table GSI range key was of type %s instead of S!", attrmap["TestGSIRangeKey"])
		}

		return nil
	}
}

func testAccCheckDynamoDbTableWasUpdated(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeTable(params)
		table := resp.Table

		if err != nil {
			return err
		}

		attrCount := len(table.AttributeDefinitions)
		gsiCount := len(table.GlobalSecondaryIndexes)
		lsiCount := len(table.LocalSecondaryIndexes)

		if attrCount != 4 {
			return fmt.Errorf("There were %d attributes, not 4 like there should have been!", attrCount)
		}

		if gsiCount != 1 {
			return fmt.Errorf("There were %d GSIs, not 1 like there should have been!", gsiCount)
		}

		if lsiCount != 1 {
			return fmt.Errorf("There were %d LSIs, not 1 like there should have been!", lsiCount)
		}

		if dynamoDbGetGSIIndex(&table.GlobalSecondaryIndexes, "ReplacementTestTableGSI") == -1 {
			return fmt.Errorf("Could not find GSI named 'ReplacementTestTableGSI' in the table!")
		}

		if dynamoDbGetGSIIndex(&table.GlobalSecondaryIndexes, "InitialTestTableGSI") != -1 {
			return fmt.Errorf("Should have removed 'InitialTestTableGSI' but it still exists!")
		}

		attrmap := dynamoDbAttributesToMap(&table.AttributeDefinitions)
		if attrmap["TestTableHashKey"] != "S" {
			return fmt.Errorf("Test table hash key was of type %s instead of S!", attrmap["TestTableHashKey"])
		}
		if attrmap["TestTableRangeKey"] != "S" {
			return fmt.Errorf("Test table range key was of type %s instead of S!", attrmap["TestTableRangeKey"])
		}
		if attrmap["TestLSIRangeKey"] != "N" {
			return fmt.Errorf("Test table LSI range key was of type %s instead of N!", attrmap["TestLSIRangeKey"])
		}
		if attrmap["ReplacementGSIRangeKey"] != "N" {
			return fmt.Errorf("Test table replacement GSI range key was of type %s instead of N!", attrmap["ReplacementGSIRangeKey"])
		}

		return nil
	}
}

func dynamoDbGetGSIIndex(gsiList *[]*dynamodb.GlobalSecondaryIndexDescription, target string) int {
	for idx, gsiObject := range *gsiList {
		if *gsiObject.IndexName == target {
			return idx
		}
	}

	return -1
}

func dynamoDbAttributesToMap(attributes *[]*dynamodb.AttributeDefinition) map[string]string {
	attrmap := make(map[string]string)

	for _, attrdef := range *attributes {
		attrmap[*attrdef.AttributeName] = *attrdef.AttributeType
	}

	return attrmap
}

func testAccAWSDynamoDbConfigInitialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 10
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigAddSecondaryGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 20
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "ReplacementGSIRangeKey"
    type = "N"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "ReplacementTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "ReplacementGSIRangeKey"
    write_capacity = 5
    read_capacity = 5
    projection_type = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}`, rName)
}

func testAccAWSDynamoDbConfigStreamSpecification() string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "TerraformTestStreamTable-%d"
  read_capacity = 10
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
  stream_enabled = true
  stream_view_type = "KEYS_ONLY"
}
`, acctest.RandInt())
}

func testAccAWSDynamoDbConfigTags() string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "TerraformTestTable-%d"
  read_capacity = 10
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }

  tags {
    Name = "terraform-test-table-%d"
    AccTest = "yes"
    Testing = "absolutely"
  }
}
`, acctest.RandInt(), acctest.RandInt())
}

func testAccAWSDynamoDbConfigGsiUpdate(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 10
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedCapacity(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 20
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 10
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    range_key       = "att4"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "INCLUDE"
    non_key_attributes = ["RandomAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedNonKeyAttributes(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 10
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    range_key       = "att4"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "INCLUDE"
    non_key_attributes = ["RandomAttribute", "AnotherAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigAddTimeToLive(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 10
  write_capacity = 20
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "TestTTL"
    enabled = true
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}
