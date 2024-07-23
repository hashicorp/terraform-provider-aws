// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmq "github.com/hashicorp/terraform-provider-aws/internal/service/mq"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidateBrokerName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ValidName",
		"V_-dN01e",
		acctest.Ct0,
		"-",
		"_",
		strings.Repeat("x", 50),
	}
	for _, v := range validNames {
		_, errors := tfmq.ValidateBrokerName(v, names.AttrName)
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
		_, errors := tfmq.ValidateBrokerName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid broker name", v)
		}
	}
}

func TestBrokerPasswordValidation(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	testCases := []struct {
		OldUsers []interface{}
		NewUsers []interface{}

		Creations []*mq.CreateUserInput
		Deletions []*mq.DeleteUserInput
		Updates   []*mq.UpdateUserInput
	}{
		{
			OldUsers: []interface{}{},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access":   false,
					names.AttrUsername: "second",
					names.AttrPassword: "TestTest2222",
					"groups":           schema.NewSet(schema.HashString, []interface{}{"admin"}),
					"replication_user": false,
				},
			},
			Creations: []*mq.CreateUserInput{
				{
					BrokerId:        aws.String("test"),
					ConsoleAccess:   aws.Bool(false),
					Username:        aws.String("second"),
					Password:        aws.String("TestTest2222"),
					Groups:          []string{"admin"},
					ReplicationUser: aws.Bool(false),
				},
			},
			Deletions: nil,
			Updates:   nil,
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access":   true,
					names.AttrUsername: "first",
					names.AttrPassword: "TestTest1111",
					"replication_user": false,
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access":   false,
					names.AttrUsername: "second",
					names.AttrPassword: "TestTest2222",
					"replication_user": false,
				},
			},
			Creations: []*mq.CreateUserInput{
				{
					BrokerId:        aws.String("test"),
					ConsoleAccess:   aws.Bool(false),
					Username:        aws.String("second"),
					Password:        aws.String("TestTest2222"),
					ReplicationUser: aws.Bool(false),
				},
			},
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: nil,
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access":   true,
					names.AttrUsername: "first",
					names.AttrPassword: "TestTest1111updated",
					"replication_user": false,
				},
				map[string]interface{}{
					"console_access":   false,
					names.AttrUsername: "second",
					names.AttrPassword: "TestTest2222",
					"replication_user": false,
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access":   false,
					names.AttrUsername: "second",
					names.AttrPassword: "TestTest2222",
					"groups":           schema.NewSet(schema.HashString, []interface{}{"admin"}),
					"replication_user": false,
				},
			},
			Creations: nil,
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: []*mq.UpdateUserInput{
				{
					BrokerId:        aws.String("test"),
					ConsoleAccess:   aws.Bool(false),
					Username:        aws.String("second"),
					Password:        aws.String("TestTest2222"),
					Groups:          []string{"admin"},
					ReplicationUser: aws.Bool(false),
				},
			},
		},
	}

	for _, tc := range testCases {
		creations, deletions, updates, err := tfmq.DiffBrokerUsers("test", tc.OldUsers, tc.NewUsers)
		if err != nil {
			t.Fatal(err)
		}

		var got, want any = creations, tc.Creations
		if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(mq.CreateUserInput{})); diff != "" {
			t.Fatalf("unexpected CreateUserInput diff (+wanted, -got): %s", diff)
		}

		got, want = deletions, tc.Deletions
		if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(mq.DeleteUserInput{})); diff != "" {
			t.Fatalf("unexpected DeleteUserInput diff (+wanted, -got): %s", diff)
		}

		got, want = updates, tc.Updates
		if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(mq.UpdateUserInput{})); diff != "" {
			t.Fatalf("unexpected UpdateUserInput diff (+wanted, -got): %s", diff)
		}
	}
}

const (
	testAccBrokerVersionNewer = "5.17.6"  // before changing, check b/c must be valid on GovCloud
	testAccBrokerVersionOlder = "5.16.7"  // before changing, check b/c must be valid on GovCloud
	testAccRabbitVersion      = "3.11.20" // before changing, check b/c must be valid on GovCloud
)

func TestAccMQBroker_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_basic(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mq", regexache.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "simple"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexache.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "efs"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest1234",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_basic(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmq.ResourceBroker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMQBroker_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_tags1(rName, testAccBrokerVersionNewer, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccBrokerConfig_tags2(rName, testAccBrokerVersionNewer, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBrokerConfig_tags1(rName, testAccBrokerVersionNewer, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccMQBroker_throughputOptimized(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_ebs(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexache.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "ebs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mq", regexache.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
				),
			},
		},
	})
}

func TestAccMQBroker_AllFields_defaultVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
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
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_allFieldsDefaultVPC(rName, testAccBrokerVersionNewer, rName, cfgBodyBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtTrue,
						"groups.#":         acctest.Ct3,
						names.AttrUsername: "SecondTest",
						names.AttrPassword: "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mq", regexache.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				// Update configuration in-place
				Config: testAccBrokerConfig_allFieldsDefaultVPC(rName, testAccBrokerVersionNewer, rName, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct3),
				),
			},
			{
				// Replace configuration
				Config: testAccBrokerConfig_allFieldsDefaultVPC(rName, testAccBrokerVersionNewer, rNameUpdated, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccMQBroker_AllFields_customVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
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
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_allFieldsCustomVPC(rName, testAccBrokerVersionNewer, rName, cfgBodyBefore, "CET"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionNewer),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageType, "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtTrue,
						"groups.#":         acctest.Ct3,
						names.AttrUsername: "SecondTest",
						names.AttrPassword: "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mq", regexache.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexache.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexache.MustCompile(`^ssl://[0-9a-z.-]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexache.MustCompile(`^amqp\+ssl://[0-9a-z.-]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexache.MustCompile(`^stomp\+ssl://[0-9a-z.-]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexache.MustCompile(`^mqtt\+ssl://[0-9a-z.-]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexache.MustCompile(`^wss://[0-9a-z.-]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				// Update configuration in-place
				Config: testAccBrokerConfig_allFieldsCustomVPC(rName, testAccBrokerVersionNewer, rName, cfgBodyAfter, "GMT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "GMT"),
				),
			},
			{
				// Replace configuration
				Config: testAccBrokerConfig_allFieldsCustomVPC(rName, testAccBrokerVersionNewer, rNameUpdated, cfgBodyAfter, "GMT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_encryptionOptionsKMSKeyID(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_options.0.kms_key_id", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_managedKeyDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_encryptionOptionsManagedKey(rName, testAccBrokerVersionNewer, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_EncryptionOptions_managedKeyEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_encryptionOptionsManagedKey(rName, testAccBrokerVersionNewer, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_Update_users(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_updateUsers1(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "first",
						names.AttrPassword: "TestTest1111",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			// Adding new user + modify existing
			{
				Config: testAccBrokerConfig_updateUsers2(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "second",
						names.AttrPassword: "TestTest2222",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtTrue,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "first",
						names.AttrPassword: "TestTest1111updated",
					}),
				),
			},
			// Deleting user + modify existing
			{
				Config: testAccBrokerConfig_updateUsers3(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct1,
						names.AttrUsername: "second",
						names.AttrPassword: "TestTest2222",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "admin"),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_securityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_basic(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccBrokerConfig_updateSecurityGroups(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct2),
				),
			},
			// Trigger a reboot and ensure the password change was applied
			// User hashcode can be retrieved by calling resourceUserHash
			{
				Config: testAccBrokerConfig_updateUsersSecurityGroups(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest9999",
					}),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_engineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_basic(rName, testAccBrokerVersionOlder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionOlder),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccBrokerConfig_engineVersionUpdate(rName, testAccBrokerVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccBrokerVersionNewer),
				),
			},
		},
	})
}

func TestAccMQBroker_Update_hostInstanceType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker1, broker2 mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_instanceType(rName, testAccBrokerVersionNewer, "mq.t2.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker1),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
				),
			},
			{
				Config: testAccBrokerConfig_instanceType(rName, testAccBrokerVersionNewer, "mq.t3.micro"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker2),
					testAccCheckBrokerNotRecreated(&broker1, &broker2),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
				),
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_rabbit(rName, testAccRabbitVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccRabbitVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^amqps://[0-9a-z.-]+:5671$`)),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_config(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_rabbitConfig(rName, testAccRabbitVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexache.MustCompile(`^c-[0-9a-z-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccRabbitVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^amqps://[0-9a-z.-]+:5671$`)),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_logs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_rabbitLogs(rName, testAccRabbitVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccRabbitVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^amqps://[0-9a-z.-]+:5671$`)),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_validationAuditLog(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBrokerConfig_rabbitAuditLog(rName, testAccRabbitVersion, true),
				ExpectError: regexache.MustCompile(`logs.audit: Can not be configured when engine is RabbitMQ`),
			},
			{
				// Special case: allow explicitly setting logs.0.audit to false,
				// though the AWS API does not accept the parameter.
				Config: testAccBrokerConfig_rabbitAuditLog(rName, testAccRabbitVersion, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
				),
			},
		},
	})
}

func TestAccMQBroker_RabbitMQ_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_rabbitCluster(rName, testAccRabbitVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "CLUSTER_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, testAccRabbitVersion),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", ""),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPubliclyAccessible, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", "data.aws_subnets.default", "ids.#"),
					resource.TestCheckResourceAttr(resourceName, "user.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access":   acctest.CtFalse,
						"groups.#":         acctest.Ct0,
						names.AttrUsername: "Test",
						names.AttrPassword: "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "mq", regexache.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexache.MustCompile(`^https://[0-9a-f-]+\.mq.[0-9a-z-]+.amazonaws.com$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexache.MustCompile(`^amqps://[0-9a-z.-]+:5671$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccMQBroker_ldap(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_ldap(rName, testAccBrokerVersionNewer, "anyusername"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "ldap"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.0", "my.ldap.server-1.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.1", "my.ldap.server-2.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_base", "role.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_name", "role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_matching", "role.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_subtree", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.service_account_username", "anyusername"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_base", "user.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_role_name", "user.role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_matching", "user.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_subtree", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccMQBroker_dataReplicationMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var broker mq.DescribeBrokerOutput
	var brokerAlternate mq.DescribeBrokerOutput
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mq_broker.test"
	primaryBrokerResourceName := "aws_mq_broker.primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckPartitionHasService(t, names.MQEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MQServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             testAccCheckBrokerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrokerConfig_dataReplicationMode(rName, testAccBrokerVersionNewer, string(types.DataReplicationModeCrdr)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBrokerExists(ctx, resourceName, &broker),
					testAccCheckBrokerExistsWithProvider(ctx, primaryBrokerResourceName, &brokerAlternate, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers)),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", string(types.DeploymentModeActiveStandbyMultiAz)),
					// data_replication_mode is not returned until after reboot
					resource.TestCheckResourceAttr(resourceName, "data_replication_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_data_replication_mode", string(types.DataReplicationModeCrdr)),
					resource.TestCheckResourceAttrPair(resourceName, "data_replication_primary_broker_arn", primaryBrokerResourceName, names.AttrARN),
				),
			},
			{
				Config:                  testAccBrokerConfig_dataReplicationMode(rName, testAccBrokerVersionNewer, string(types.DataReplicationModeCrdr)),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user", "data_replication_primary_broker_arn"},
			},
			{
				// Preparation for destruction would require multiple configuration changes
				// and applies to unpair brokers. Instead, complete the necessary update, reboot,
				// and delete opreations on the primary cluster out-of-band to ensure remaining
				// resources will be freed for clean up.
				PreConfig: func() {
					// In order to delete, replicated brokers must first be unpaired by setting
					// data replication mode on the primary broker to "NONE".
					testAccUnpairBrokerWithProvider(ctx, t, &brokerAlternate, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers))
					// The primary broker must be deleted before replica broker. The direct
					// dependency in the Terraform configuration would cause this to happen
					// in the opposite order, so delete the primary out of band instead.
					testAccDeleteBrokerWithProvider(ctx, t, &brokerAlternate, acctest.RegionProviderFunc(acctest.AlternateRegion(), &providers))
				},
				Config:             testAccBrokerConfig_dataReplicationMode(rName, testAccBrokerVersionNewer, string(types.DataReplicationModeNone)),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBrokerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MQClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mq_broker" {
				continue
			}

			_, err := tfmq.FindBrokerByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("MQ Broker %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBrokerExists(ctx context.Context, n string, v *mq.DescribeBrokerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MQClient(ctx)

		output, err := tfmq.FindBrokerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBrokerExistsWithProvider(ctx context.Context, n string, v *mq.DescribeBrokerOutput, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := providerF().Meta().(*conns.AWSClient).MQClient(ctx)

		output, err := tfmq.FindBrokerByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MQClient(ctx)

	input := &mq.ListBrokersInput{}

	_, err := conn.ListBrokers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccUnpairBrokerWithProvider(ctx context.Context, t *testing.T, broker *mq.DescribeBrokerOutput, providerF func() *schema.Provider) {
	brokerID := aws.ToString(broker.BrokerId)
	deadline := tfresource.NewDeadline(30 * time.Minute)
	conn := providerF().Meta().(*conns.AWSClient).MQClient(ctx)

	_, err := conn.UpdateBroker(ctx, &mq.UpdateBrokerInput{
		BrokerId:            aws.String(brokerID),
		DataReplicationMode: types.DataReplicationModeNone,
	})
	if err != nil {
		t.Fatalf("updating broker (%s): %s", brokerID, err)
	}

	_, err = conn.RebootBroker(ctx, &mq.RebootBrokerInput{BrokerId: aws.String(brokerID)})
	if err != nil {
		t.Fatalf("rebooting broker (%s): %s", brokerID, err)
	}

	_, err = tfmq.WaitBrokerRebooted(ctx, conn, brokerID, deadline.Remaining())
	if err != nil {
		t.Fatalf("waiting for broker (%s) reboot: %s", brokerID, err)
	}
}

func testAccDeleteBrokerWithProvider(ctx context.Context, t *testing.T, broker *mq.DescribeBrokerOutput, providerF func() *schema.Provider) {
	brokerID := aws.ToString(broker.BrokerId)
	deadline := tfresource.NewDeadline(30 * time.Minute)
	conn := providerF().Meta().(*conns.AWSClient).MQClient(ctx)

	_, err := conn.DeleteBroker(ctx, &mq.DeleteBrokerInput{BrokerId: aws.String(brokerID)})
	if err != nil {
		t.Fatalf("deleting broker (%s): %s", brokerID, err)
	}

	_, err = tfmq.WaitBrokerDeleted(ctx, conn, brokerID, deadline.Remaining())
	if err != nil {
		t.Fatalf("waiting for broker (%s) deletion: %s", brokerID, err)
	}
}

func testAccCheckBrokerNotRecreated(before, after *mq.DescribeBrokerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.BrokerId), aws.ToString(after.BrokerId); before != after {
			return fmt.Errorf("MQ Broker (%s/%s) recreated", before, after)
		}

		return nil
	}
}

func testAccBrokerConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_ebs(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_engineVersionUpdate(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_allFieldsDefaultVPC(rName, version, cfgName, cfgBody string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"

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

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = false
  security_groups     = aws_security_group.test[*].id

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

func testAccBrokerConfig_baseCustomVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count = 2

  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccBrokerConfig_allFieldsCustomVPC(rName, version, cfgName, cfgBody, tz string) string {
	return acctest.ConfigCompose(testAccBrokerConfig_baseCustomVPC(rName), fmt.Sprintf(`
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
  security_groups     = aws_security_group.test[*].id
  subnet_ids          = aws_subnet.test[*].id

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
`, rName, version, cfgName, cfgBody, tz))
}

func testAccBrokerConfig_encryptionOptionsKMSKeyID(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_encryptionOptionsManagedKey(rName, version string, useAwsOwnedKey bool) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_tags1(rName, version, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_tags2(rName, version, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name = "%[1]s-1"

  tags = {
    Name = %[1]q
  }
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name = "%[1]s-1"

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_rabbit(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_rabbitConfig(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_configuration" "test" {
  description    = "TfAccTest MQ Configuration"
  name           = %[1]q
  engine_type    = "RabbitMQ"
  engine_version = %[2]q

  data = <<DATA
  # Default RabbitMQ delivery acknowledgement timeout is 30 minutes
  consumer_timeout = 1800000
  
  DATA
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.t3.micro"
  security_groups    = [aws_security_group.test.id]

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}

func testAccBrokerConfig_rabbitLogs(rName, version string) string {
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

  tags = {
    Name = %[1]q
  }
}
`, rName, version)
}

func testAccBrokerConfig_rabbitAuditLog(rName, version string, enabled bool) string {
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

  tags = {
    Name = %[1]q
  }
}
`, rName, version, enabled)
}

func testAccBrokerConfig_rabbitCluster(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_vpc" "default" {
  default = true
}
`, rName, version)
}

func testAccBrokerConfig_ldap(rName, version, ldapUsername string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

func testAccBrokerConfig_instanceType(rName, version, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
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

// testAccBrokerConfig_dataReplicationMode creates a primary and replica broker
// in different regions, linking the former using the data replication arguments
func testAccBrokerConfig_dataReplicationMode(rName, version, dataReplicationMode string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_security_group" "primary" {
  provider = awsalternate

  name = "%[1]s-primary"

  tags = {
    Name = "%[1]s-primary"
  }
}

resource "aws_mq_broker" "primary" {
  provider = awsalternate

  apply_immediately  = true
  broker_name        = "%[1]s-primary"
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.primary.id]
  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
  user {
    username         = "Test-ReplicationUser"
    password         = "TestTest1234"
    replication_user = true
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = %[2]q
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]
  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"

  data_replication_mode               = %[3]q
  data_replication_primary_broker_arn = aws_mq_broker.primary.arn

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
  user {
    username         = "Test-ReplicationUser"
    password         = "TestTest1234"
    replication_user = true
  }
}
`, rName, version, dataReplicationMode))
}
