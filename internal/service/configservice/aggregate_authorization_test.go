package configservice_test

// func TestAccConfigServiceAggregateAuthorization_basic(t *testing.T) {
// 	rString := sdkacctest.RandStringFromCharSet(12, "0123456789")
// 	resourceName := "aws_config_aggregate_authorization.example"
// 	dataSourceName := "data.aws_region.current"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { acctest.PreCheck(t) },
// 		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
// 		ProviderFactories:acctest.ProviderFactories,
// 		CheckDestroy: testAccCheckAggregateAuthorizationDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAggregateAuthorizationConfig_basic(rString),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "account_id", rString),
// 					resource.TestCheckResourceAttrPair(resourceName, "region", dataSourceName, "name"),
// 					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf(`aggregation-authorization/%s/%s$`, rString, acctest.Region()))),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 		},
// 	})
// }

// func TestAccConfigServiceAggregateAuthorization_tags(t *testing.T) {
// 	rString := sdkacctest.RandStringFromCharSet(12, "0123456789")
// 	resourceName := "aws_config_aggregate_authorization.example"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { acctest.PreCheck(t) },
// 		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
// 		ProviderFactories:acctest.ProviderFactories,
// 		CheckDestroy: testAccCheckAggregateAuthorizationDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAggregateAuthorizationConfig_tags(rString, "foo", "bar", "fizz", "buzz"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.Name", rString),
// 					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
// 				),
// 			},
// 			{
// 				Config: testAccAggregateAuthorizationConfig_tags(rString, "foo", "bar2", "fizz2", "buzz2"),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.Name", rString),
// 					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
// 					resource.TestCheckResourceAttr(resourceName, "tags.fizz2", "buzz2"),
// 				),
// 			},
// 			{
// 				ResourceName:      resourceName,
// 				ImportState:       true,
// 				ImportStateVerify: true,
// 			},
// 			{
// 				Config: testAccAggregateAuthorizationConfig_basic(rString),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccCheckAggregateAuthorizationDestroy(s *terraform.State) error {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "aws_config_aggregate_authorization" {
// 			continue
// 		}

// 		accountId, region, err := tfconfig.AggregateAuthorizationParseID(rs.Primary.ID)
// 		if err != nil {
// 			return err
// 		}

// 		aggregateAuthorizations, err := tfconfig.DescribeAggregateAuthorizations(conn)

// 		if err != nil {
// 			return err
// 		}

// 		for _, auth := range aggregateAuthorizations {
// 			if accountId == aws.StringValue(auth.AuthorizedAccountId) && region == aws.StringValue(auth.AuthorizedAwsRegion) {
// 				return fmt.Errorf("Config aggregate authorization still exists: %s", rs.Primary.ID)
// 			}
// 		}
// 	}

// 	return nil
// }

// func testAccAggregateAuthorizationConfig_basic(rString string) string {
// 	return fmt.Sprintf(`
// data "aws_region" "current" {}

// resource "aws_config_aggregate_authorization" "example" {
//   account_id = %[1]q
//   region     = data.aws_region.current.name
// }
// `, rString)
// }

// func testAccAggregateAuthorizationConfig_tags(rString, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
// 	return fmt.Sprintf(`
// data "aws_region" "current" {}

// resource "aws_config_aggregate_authorization" "example" {
//   account_id = %[1]q
//   region     = data.aws_region.current.name

//   tags = {
//     Name = %[1]q

//     %[2]s = %[3]q
//     %[4]s = %[5]q
//   }
// }
// `, rString, tagKey1, tagValue1, tagKey2, tagValue2)
// }
