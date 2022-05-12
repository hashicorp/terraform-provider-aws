package mq_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmq "github.com/hashicorp/terraform-provider-aws/internal/service/mq"
)

func TestValidateBrokerName(t *testing.T) {
	validNames := []string{
		"ValidName",
		"V_-dN01e",
		"0",
		"-",
		"_",
		strings.Repeat("x", 50),
	}
	for _, v := range validNames {
		_, errors := tfmq.ValidateBrokerName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid broker name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Inval:d.~Name",
		"Invalid Name",
		"*",
		"",
		strings.Repeat("x", 51),
	}
	for _, v := range invalidNames {
		_, errors := tfmq.ValidateBrokerName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid broker name", v)
		}
	}
}

func TestBrokerPasswordValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "123456789012",
			ErrCount: 0,
		},
		{
			Value:    "12345678901",
			ErrCount: 1,
		},
		{
			Value:    "1234567890" + strings.Repeat("#", 240),
			ErrCount: 0,
		},
		{
			Value:    "1234567890" + strings.Repeat("#", 241),
			ErrCount: 1,
		},
		{
			Value:    "123" + strings.Repeat("#", 9),
			ErrCount: 0,
		},
		{
			Value:    "12" + strings.Repeat("#", 10),
			ErrCount: 1,
		},
		{
			Value:    "12345678901,",
			ErrCount: 1,
		},
		{
			Value:    "1," + strings.Repeat("#", 9),
			ErrCount: 3,
		},
	}

	for _, tc := range cases {
		_, errors := tfmq.ValidBrokerPassword(tc.Value, "aws_mq_broker_user_password")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected errors %d for %s while returned errors %d", tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func TestDiffUsers(t *testing.T) {
	testCases := []struct {
		OldUsers []interface{}
		NewUsers []interface{}

		Creations []*mq.CreateUserRequest
		Deletions []*mq.DeleteUserInput
		Updates   []*mq.UpdateUserRequest
	}{
		{
			OldUsers: []interface{}{},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
					"groups":         schema.NewSet(schema.HashString, []interface{}{"admin"}),
				},
			},
			Creations: []*mq.CreateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
					Groups:        aws.StringSlice([]string{"admin"}),
				},
			},
			Deletions: []*mq.DeleteUserInput{},
			Updates:   []*mq.UpdateUserRequest{},
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access": true,
					"username":       "first",
					"password":       "TestTest1111",
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
				},
			},
			Creations: []*mq.CreateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
				},
			},
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: []*mq.UpdateUserRequest{},
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access": true,
					"username":       "first",
					"password":       "TestTest1111updated",
				},
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
					"groups":         schema.NewSet(schema.HashString, []interface{}{"admin"}),
				},
			},
			Creations: []*mq.CreateUserRequest{},
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: []*mq.UpdateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
					Groups:        aws.StringSlice([]string{"admin"}),
				},
			},
		},
	}

	for _, tc := range testCases {
		creations, deletions, updates, err := tfmq.DiffBrokerUsers("test", tc.OldUsers, tc.NewUsers)
		if err != nil {
			t.Fatal(err)
		}

		expectedCreations := fmt.Sprintf("%s", tc.Creations)
		creationsString := fmt.Sprintf("%s", creations)
		if creationsString != expectedCreations {
			t.Fatalf("Expected creations: %s\nGiven: %s", expectedCreations, creationsString)
		}

		expectedDeletions := fmt.Sprintf("%s", tc.Deletions)
		deletionsString := fmt.Sprintf("%s", deletions)
		if deletionsString != expectedDeletions {
			t.Fatalf("Expected deletions: %s\nGiven: %s", expectedDeletions, deletionsString)
		}

		expectedUpdates := fmt.Sprintf("%s", tc.Updates)
		updatesString := fmt.Sprintf("%s", updates)
		if updatesString != expectedUpdates {
			t.Fatalf("Expected updates: %s\nGiven: %s", expectedUpdates, updatesString)
		}
	}
}

const (
	testAccBrokerVersionNewer = "5.16.3"  // before changing, check b/c must be valid on GovCloud
	testAccBrokerVersionOlder = "5.15.12" // before changing, check b/c must be valid on GovCloud
	testAccRabbitMQVersion    = "3.8.6"   // before changing, check b/c must be valid on GovCloud
)

func TestAccMQBroker_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "simple"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_throughputOptimized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerEBSConfig(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "ebs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
		},
	})
}

func TestAccMQBroker_AllFields_defaultVPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_mq_broker.test"

	cfgBodyBefore := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>`
	cfgBodyAfter := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_allFieldsDefaultVpc(rName, testAccBrokerVersionNewer, rName, cfgBodyBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "false"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "3",
						"username":       "SecondTest",
						"password":       "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				// Update configuration in-place
				Config: testAccBrokerConfig_allFieldsDefaultVpc(rName, testAccBrokerVersionNewer, rName, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "3"),
				),
			},
			{
				// Replace configuration
				Config: testAccBrokerConfig_allFieldsDefaultVpc(rName, testAccBrokerVersionNewer, rNameUpdated, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccMQBroker_AllFields_customVPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_mq_broker.test"

	cfgBodyBefore := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>`
	cfgBodyAfter := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_allFieldsCustomVpc(rName, testAccBrokerVersionNewer, rName, cfgBodyBefore, "CET"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "true"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "3",
						"username":       "SecondTest",
						"password":       "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				// Update configuration in-place
				Config: testAccBrokerConfig_allFieldsCustomVpc(rName, testAccBrokerVersionNewer, rName, cfgBodyAfter, "GMT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "3"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "GMT"),
				),
			},
			{
				// Replace configuration
				Config: testAccBrokerConfig_allFieldsCustomVpc(rName, testAccBrokerVersionNewer, rNameUpdated, cfgBodyAfter, "GMT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_kmsKeyID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfigEncryptionOptionsKmsKeyId(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_options.0.kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_awsOwnedKeyDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerEncryptionOptionsUseAWSOwnedKeyConfig(rName, testAccBrokerVersionNewer, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_awsOwnedKeyEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerEncryptionOptionsUseAWSOwnedKeyConfig(rName, testAccBrokerVersionNewer, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_Update_users(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_updateUsers1(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "first",
						"password":       "TestTest1111",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			// Adding new user + modify existing
			{
				Config: testAccBrokerConfig_updateUsers2(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "second",
						"password":       "TestTest2222",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "0",
						"username":       "first",
						"password":       "TestTest1111updated",
					}),
				),
			},
			// Deleting user + modify existing
			{
				Config: testAccBrokerConfig_updateUsers3(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "1",
						"username":       "second",
						"password":       "TestTest2222",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "admin"),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfigTags1(rName, testAccBrokerVersionNewer, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccBrokerConfigTags2(rName, testAccBrokerVersionNewer, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBrokerConfigTags1(rName, testAccBrokerVersionNewer, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_securityGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccBrokerConfig_updateSecurityGroups(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
				),
			},
			// Trigger a reboot and ensure the password change was applied
			// User hashcode can be retrieved by calling resourceUserHash
			{
				Config: testAccBrokerConfig_updateUsersSecurityGroups(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"username": "Test",
						"password": "TestTest9999",
					}),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_engineVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig(rName, testAccBrokerVersionOlder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionOlder),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccBrokerEngineVersionUpdateConfig(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccBrokerVersionNewer),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_hostInstanceType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker1, broker2 mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfigInstanceType(rName, testAccBrokerVersionNewer, "mq.t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker1),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
				),
			},
			{
				Config: testAccBrokerConfigInstanceType(rName, testAccBrokerVersionNewer, "mq.t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker2),
					testAccCheckBrokerNotRecreated(&broker1, &broker2),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
				),
			},
		},
	})
}

func TestAccMQBroker_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					acctest.CheckResourceDisappears(acctest.Provider, tfmq.ResourceBroker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfigRabbit(rName, testAccRabbitMQVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccRabbitMQVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^amqps://[a-z0-9-\.]+:5671$`)),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "false"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_logs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfigRabbitLogs(rName, testAccRabbitMQVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccRabbitMQVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^amqps://[a-z0-9-\.]+:5671$`)),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_validationAuditLog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBrokerConfigRabbitAuditLog(rName, testAccRabbitMQVersion, true),
				ExpectError: regexp.MustCompile(`logs.audit: Can not be configured when engine is RabbitMQ`),
			},
			{
				// Special case: allow explicitly setting logs.0.audit to false,
				// though the AWS API does not accept the parameter.
				Config: testAccBrokerConfigRabbitAuditLog(rName, testAccRabbitMQVersion, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_cluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRabbitClusterBrokerConfig(rName, testAccRabbitMQVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "CLUSTER_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", testAccRabbitMQVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "false"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^amqps://[a-z0-9-\.]+:5671$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccMQBroker_ldap(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(mq.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, mq.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_ldap(rName, testAccBrokerVersionNewer, "anyusername"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "ldap"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.0", "my.ldap.server-1.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.1", "my.ldap.server-2.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_base", "role.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_name", "role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_matching", "role.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_subtree", "true"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.service_account_username", "anyusername"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_base", "user.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_role_name", "user.role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_matching", "user.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_subtree", "true"),
				),
			},
		},
	})
}

func testAccCheckBrokerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MQConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mq_broker" {
			continue
		}

		input := &mq.DescribeBrokerInput{
			BrokerId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeBroker(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mq.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MQ Broker to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckBrokerExists(name string, broker *mq.DescribeBrokerResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MQ Broker is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MQConn
		resp, err := conn.DescribeBroker(&mq.DescribeBrokerInput{
			BrokerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error describing MQ Broker: %s", err.Error())
		}

		*broker = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MQConn

	input := &mq.ListBrokersInput{}

	_, err := conn.ListBrokers(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckBrokerNotRecreated(before, after *mq.DescribeBrokerResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.BrokerId), aws.StringValue(after.BrokerId); before != after {
			return fmt.Errorf("MQ Broker (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccBrokerConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = %[2]q
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerEBSConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  storage_type       = "ebs"
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerEngineVersionUpdateConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  apply_immediately  = true
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerConfig_allFieldsDefaultVpc(rName, version, cfgName, cfgBody string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "mq1" {
  name = %[1]q
}

resource "aws_security_group" "mq2" {
  name = "%[1]s-2"
}

resource "aws_mq_configuration" "test" {
  name           = %[3]q
  engine_type    = "ActiveMQ"
  engine_version = %[2]q

  data = <<DATA
%[4]s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = %[1]q

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  storage_type       = "efs"
  host_instance_type = "mq.t2.micro"

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = [aws_security_group.mq1.id, aws_security_group.mq2.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  user {
    username       = "SecondTest"
    password       = "SecondTestTest1234"
    console_access = true
    groups         = ["first", "second", "third"]
  }
}
`, rName, version, cfgName, cfgBody)
}

func testAccBrokerConfig_allFieldsCustomVpc(rName, version, cfgName, cfgBody, tz string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.11.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "private" {
  count             = 2
  cidr_block        = "10.11.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = aws_subnet.private.*.id[count.index]
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "mq1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "mq2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_configuration" "test" {
  name           = %[3]q
  engine_type    = "ActiveMQ"
  engine_version = %[2]q

  data = <<DATA
%[4]s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = %[1]q

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  storage_type       = "efs"
  host_instance_type = "mq.t2.micro"

  logs {
    general = true
    audit   = true
  }

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = %[5]q
  }

  publicly_accessible = true
  security_groups     = [aws_security_group.mq1.id, aws_security_group.mq2.id]
  subnet_ids          = aws_subnet.private[*].id

  user {
    username = "Test"
    password = "TestTest1234"
  }

  user {
    username       = "SecondTest"
    password       = "SecondTestTest1234"
    console_access = true
    groups         = ["first", "second", "third"]
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, version, cfgName, cfgBody, tz)
}

func testAccBrokerConfigEncryptionOptionsKmsKeyId(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  encryption_options {
    kms_key_id        = aws_kms_key.test.arn
    use_aws_owned_key = false
  }

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerEncryptionOptionsUseAWSOwnedKeyConfig(rName, version string, useAwsOwnedKey bool) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  encryption_options {
    use_aws_owned_key = %[3]t
  }

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version, useAwsOwnedKey)
}

func testAccBrokerConfig_updateUsers1(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "first"
    password = "TestTest1111"
  }
}
`, rName, version)
}

func testAccBrokerConfig_updateUsers2(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    console_access = true
    username       = "first"
    password       = "TestTest1111updated"
  }

  user {
    username = "second"
    password = "TestTest2222"
  }
}
`, rName, version)
}

func testAccBrokerConfig_updateUsers3(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "second"
    password = "TestTest2222"
    groups   = ["admin"]
  }
}
`, rName, version)
}

func testAccBrokerConfigTags1(rName, version, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, version, tagKey1, tagValue1)
}

func testAccBrokerConfigTags2(rName, version, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, version, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccBrokerConfig_updateSecurityGroups(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id, aws_security_group.test2.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerConfig_updateUsersSecurityGroups(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test2.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest9999"
  }
}
`, rName, version)
}

func testAccBrokerConfigRabbit(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t3.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerConfigRabbitLogs(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t3.micro"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}

resource "aws_security_group" "test" {
  name = %[1]q
}
`, rName, version)
}

func testAccBrokerConfigRabbitAuditLog(rName, version string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t3.micro"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
    audit   = %[3]t
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}

resource "aws_security_group" "test" {
  name = %[1]q
}
`, rName, version, enabled)
}

func testAccRabbitClusterBrokerConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]
  storage_type       = "ebs"
  deployment_mode    = "CLUSTER_MULTI_AZ"

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerConfig_ldap(rName, version, ldapUsername string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately       = true
  authentication_strategy = "ldap"
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = %[2]q
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }

  ldap_server_metadata {
    hosts                    = ["my.ldap.server-1.com", "my.ldap.server-2.com"]
    role_base                = "role.base"
    role_name                = "role.name"
    role_search_matching     = "role.search.matching"
    role_search_subtree      = true
    service_account_password = "supersecret"
    service_account_username = %[3]q
    user_base                = "user.base"
    user_role_name           = "user.role.name"
    user_search_matching     = "user.search.matching"
    user_search_subtree      = true
  }
}
`, rName, version, ldapUsername)
}

func testAccBrokerConfigInstanceType(rName, version, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  apply_immediately  = true
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = %[3]q
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version, instanceType)
}
