// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfguardduty "github.com/hashicorp/terraform-provider-aws/internal/service/guardduty"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccDetector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexache.MustCompile("detector/.+$")),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "SIX_HOURS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_disable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "false"),
				),
			},
			{
				Config: testAccDetectorConfig_enable,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable", "true"),
				),
			},
			{
				Config: testAccDetectorConfig_findingPublishingFrequency,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", "FIFTEEN_MINUTES"),
				),
			},
		},
	})
}

func testAccDetector_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_tags1("key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDetectorConfig_tags1("key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccDetector_datasources_s3logs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesS3Logs(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_datasources_kubernetes_audit_logs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesKubernetesAuditLogs(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_datasources_malware_protection(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesMalwareProtection(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesMalwareProtection(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_datasources_all(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_datasourcesAll(true, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "true"),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "false"),
				),
			},
			{
				Config: testAccDetectorConfig_datasourcesAll(false, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.kubernetes.0.audit_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.malware_protection.0.scan_ec2_instance_with_findings.0.ebs_volumes.0.enable", "false"),
				),
			},
		},
	})
}

func testAccDetector_features_s3_data_events(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_features_s3_data_events(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: testAccDetectorConfig_features_s3_data_events(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
		},
	})
}

func testAccDetector_features_eks_audit_logs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_features_eks_audit_logs(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_features_s3_data_events(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
		},
	})
}

func testAccDetector_features_ebs_malware_protection(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_features_ebs_malware_protection(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_features_ebs_malware_protection(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
		},
	})
}

func testAccDetector_features_rds_login_events(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_features_rds_login_events(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_features_rds_login_events(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
		},
	})
}

func testAccDetector_features_eks_runtime_monitoring(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDetectorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorConfig_features_eks_runtime_monitoring(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDetectorConfig_features_rds_login_events(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDetectorExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "datasources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "datasources.0.s3_logs.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.#", "9"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.name", "CLOUD_TRAIL"),
					resource.TestCheckResourceAttr(resourceName, "feature.0.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.name", "DNS_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.1.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.name", "FLOW_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.2.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.name", "S3_DATA_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.3.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.name", "EKS_AUDIT_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.4.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.name", "EBS_MALWARE_PROTECTION"),
					resource.TestCheckResourceAttr(resourceName, "feature.5.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.name", "RDS_LOGIN_EVENTS"),
					resource.TestCheckResourceAttr(resourceName, "feature.6.enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.name", "EKS_RUNTIME_MONITORING"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.name", "EKS_ADDON_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "feature.7.additional_configuration.0.enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.name", "LAMBDA_NETWORK_LOGS"),
					resource.TestCheckResourceAttr(resourceName, "feature.8.enable", "true"),
				),
			},
		},
	})
}

func testAccCheckDetectorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_guardduty_detector" {
				continue
			}

			_, err := tfguardduty.FindDetectorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GuardDuty Detector %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDetectorExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GuardDutyConn(ctx)

		_, err := tfguardduty.FindDetectorByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

const testAccDetectorConfig_basic = `
resource "aws_guardduty_detector" "test" {}
`

const testAccDetectorConfig_disable = `
resource "aws_guardduty_detector" "test" {
  enable = false
}
`

const testAccDetectorConfig_enable = `
resource "aws_guardduty_detector" "test" {
  enable = true
}
`

const testAccDetectorConfig_findingPublishingFrequency = `
resource "aws_guardduty_detector" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
}
`

func testAccDetectorConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccDetectorConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDetectorConfig_datasourcesS3Logs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    s3_logs {
      enable = %[1]t
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesKubernetesAuditLogs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesMalwareProtection(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = %[1]t
        }
      }
    }
  }
}
`, enable)
}

func testAccDetectorConfig_datasourcesAll(enableK8s, enableS3, enableMalware bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  datasources {
    kubernetes {
      audit_logs {
        enable = %[1]t
      }
    }
    s3_logs {
      enable = %[2]t
    }

    malware_protection {
      scan_ec2_instance_with_findings {
        ebs_volumes {
          enable = %[3]t
        }
      }
    }
  }
}
`, enableK8s, enableS3, enableMalware)
}

func testAccDetectorConfig_features_s3_data_events(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  feature {
    name   = "S3_DATA_EVENTS"
    enable = %[1]t
  }
}
`, enable)
}

func testAccDetectorConfig_features_eks_audit_logs(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  feature {
    name   = "EKS_AUDIT_LOGS"
    enable = %[1]t
  }
}
`, enable)
}

func testAccDetectorConfig_features_ebs_malware_protection(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  feature {
    name   = "EBS_MALWARE_PROTECTION"
    enable = %[1]t
  }
}
`, enable)
}

func testAccDetectorConfig_features_rds_login_events(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  feature {
    name   = "RDS_LOGIN_EVENTS"
    enable = %[1]t
  }
}
`, enable)
}

func testAccDetectorConfig_features_eks_runtime_monitoring(enable bool) string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
  feature {
    name   = "EKS_RUNTIME_MONITORING"
    enable = %[1]t

    additional_configuration {
      name   = "EKS_ADDON_MANAGEMENT"
      enable = %[1]t
    }
  }
}
`, enable)
}
