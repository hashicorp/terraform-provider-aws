---
subcategory: "Launch Wizard"
layout: "aws"
page_title: "AWS: aws_launchwizard_deployment"
description: |-
  Terraform resource for managing an AWS Launch Wizard Deployment. Check the user and API guide for workload-specific details. Currently only SAP workloads are enabled for the API.  
---

# Resource: aws_launchwizard_deployment

Terraform resource for managing an AWS Launch Wizard Deployment. Check the user and API guide for workload-specific details. Currently only SAP workloads are enabled for the API.
SAP Installation Media needs to be provided in an S3 bucket prior to deployment. See [documentation](https://docs.aws.amazon.com/launchwizard/latest/userguide/launch-wizard-sap-structure.html) for more details. Refer to [this helper script by AWS](https://github.com/awslabs/aws-sap-automation/tree/main/software_download) for automating the download.

The API only supports initial deployments. Certain Operations (like adding additional instances in HA deployments) still require console usage. Also, it is not possible to clone deployments that were created using the API.

~> **Note:** Some deployments can run multiple hours. Please consider the maximum session duration of your access token if the deployment fails due to an `ExpiredTokenException`

## Example Usage

### Simple Example: SAP HANA Database - Single Instance

Possible parameters for a SAP HANA Database - Single Instance deployment can be found [here](https://docs.aws.amazon.com/launchwizard/latest/APIReference/launch-wizard-specifications-sap-hana-single.html).

```terraform
resource "aws_launchwizard_deployment" "example_SapHanaSingle" {
  name               = "exampleHanaSingle"
  deployment_pattern = "SapHanaSingle"
  workload_name      = "SAP"
  specifications = {
    "KeyPairName" : "${aws_key_pair.key.key_name}",
    "AvailabilityZone1PrivateSubnet1Id" : "${aws_subnet.private_1.id}",
    "VpcId" : "${aws_vpc.test.id}",
    "Timezone" : "UTC",
    "EnableEbsVolumeEncryption" : "No",
    "CreateSecurityGroup" : "No",
    "DatabaseSecurityGroupId" : "${aws_security_group.test.id}",
    "SapSysGroupId" : "5002",
    "DatabaseSystemId" : "HDB",
    "DatabaseInstanceNumber" : "00",
    "DatabasePassword" : "Password123@1",
    "InstallDatabaseSoftware" : "Yes",
    "DatabaseInstallationMediaS3Uri" : "s3://sap-installation-files/HANA_DB_Software",
    "DatabaseOperatingSystem" : "SuSE-Linux-15-SP4-For-SAP-HVM",
    "DatabaseAmiId" : "${data.aws_ami.suse.id}",
    "DatabasePrimaryHostname" : "hana-primary",
    "DatabaseInstanceType" : "r5.xlarge",
    "InstallAwsBackintAgent" : "No",
    "DisableDeploymentRollback" : "Yes",
    "SaveDeploymentArtifacts" : "No",
    "DatabaseAutomaticRecovery" : "No",
    "DatabaseLogVolumeType" : "gp3",
    "DatabaseDataVolumeType" : "gp3",
  }
}

data "aws_ami" "suse" {
  most_recent = true
  owners      = ["aws-marketplace"]
  filter {
    name   = "name"
    values = ["suse-sles-sap-15-sp4*hvm*"]
  }
}

```

### Complex Example: SAP Netweaver on SAP HANA - High Availability

Possible parameters for a SAP Netweaver on SAP HANA - High Availability deployment can be found [here](https://docs.aws.amazon.com/launchwizard/latest/APIReference/launch-wizard-specifications-sap-netweaver-ha.html).

Some deployments modify resources like route tables (for the HA setup). Use the lifecycle block `ignore_changes` on the route tables and other such route tables depending on the deployment pattern.

```terraform
resource "aws_launchwizard_deployment" "example_SapNWOnHanaHA" {
  name                       = "exampleNWOnHanaHA"
  deployment_pattern         = "SapNWOnHanaHA"
  workload_name              = "SAP"
  skip_destroy_after_failure = true
  specifications = {
    "KeyPairName" : "${aws_key_pair.key.key_name}",
    "VpcId" : "${aws_vpc.test.id}",
    "AvailabilityZone1PrivateSubnet1Id" : "${aws_subnet.private_1.id}",
    "AvailabilityZone2PrivateSubnet1Id" : "${aws_subnet.private_2.id}",
    "Timezone" : "UTC",
    "EnableEbsVolumeEncryption" : "No",
    "CreateSecurityGroup" : "No",
    "DatabaseSecurityGroupId" : "${aws_security_group.test.id}",
    "ApplicationSecurityGroupId" : "${aws_security_group.test.id}",
    "SidAdmUserId" : "7002",
    "SapSysGroupId" : "5001",
    "DatabaseSystemId" : "HYD",
    "SapSid" : "S4K",
    "DatabaseInstanceNumber" : "30",
    "InstallAwsBackintAgent" : "No",
    "AscsOperatingSystem" : "SuSE-Linux-15-SP4-For-SAP-HVM",
    "AscsAmiId" : "${data.aws_ami.suse.id}",
    "AscsInstanceType" : "r5.2xlarge",
    "AscsHostname" : "hallp123",
    "AscsInstanceNumber" : "10",
    "ErsOperatingSystem" : "SuSE-Linux-15-SP4-For-SAP-HVM",
    "ErsAmiId" : "${data.aws_ami.suse.id}",
    "ErsInstanceType" : "r5.2xlarge",
    "ErsHostname" : "hallq123",
    "ErsInstanceNumber" : "11",
    "PasOperatingSystem" : "SuSE-Linux-15-SP4-For-SAP-HVM",
    "PasAmiId" : "${data.aws_ami.suse.id}",
    "PasInstanceType" : "r5.2xlarge",
    "PasHostname" : "hallv123",
    "InstallAas" : "Yes",
    "AasHostCount" : "2",
    "AasInstanceType" : "r5.2xlarge",
    "AasHostnames" : "halgaa123,halhaa124",
    "AasAutomaticRecovery" : "Yes",
    "DatabaseOperatingSystem" : "SuSE-Linux-15-SP4-For-SAP-HVM",
    "DatabaseAmiId" : "${data.aws_ami.suse.id}",
    "DatabaseInstanceType" : "r5.8xlarge",
    "DatabasePrimaryHostname" : "haljdbpri",
    "DatabaseSecondaryHostname" : "hakldbsec",
    "SapPassword" : "Password123@",
    "InstallSap" : "Yes",
    "DatabaseVirtualIpAddress" : "192.168.100.100",
    "DatabasePrimarySiteName" : "dbsitfp",
    "DatabaseSecondarySiteName" : "dbsitshc",
    "DatabasePacemakerTag" : "hallpag",
    "SetupTransportDomainController" : "No",
    "SapInstallationSpecifications" : jsonencode(
      {
        "parameters" : {
          "PRODUCT_ID" : "saps4hana-2020",
          "HDB_SCHEMA_NAME" : "SAPABAP1",
          "CI_INSTANCE_NR" : "50",
          "ASCS_VIRTUAL_HOSTNAME" : "hallvirascs",
          "ERS_VIRTUAL_HOSTNAME" : "hallvirers",
          "ASCS_OVERLAY_IP" : "192.168.100.110",
          "ERS_OVERLAY_IP" : "192.168.100.120",
          "DB_VIRTUAL_HOSTNAME" : "halldbvr1",
          "SAP_PACEMAKER_TAG" : "hllpacetag",
          "SAPINST_CD_SAPCAR" : "s3://sap-installation-files/SAPCAR",
          "SAPINST_CD_SWPM" : "s3://sap-installation-files/SWPM",
          "SAPINST_CD_KERNEL" : "s3://sap-installation-files/Kernel",
          "SAPINST_CD_LOAD" : "s3://sap-installation-files/Exports",
          "SAPINST_CD_RDBMS" : "s3://installation-files/HANA_DB_Software",
          "SAPINST_CD_RDBMS_CLIENT" : "s3://sap-installation-files/HANA_Client_Software"
        },
        "onFailureBehaviour" : "CONTINUE"
      }
    ),
    "SaveDeploymentArtifacts" : "No",
    "DisableDeploymentRollback" : "Yes",
    "ApplicationDataVolumeType" : "gp3",
    "DatabaseLogVolumeType" : "gp3",
    "DatabaseDataVolumeType" : "gp3",
    "PasAutomaticRecovery" : "Yes"
  }
}

data "aws_ami" "suse" {
  most_recent = true
  owners      = ["aws-marketplace"]
  filter {
    name   = "name"
    values = ["suse-sles-sap-15-sp4*hvm*"]
  }
}

resource "aws_route_table" "private_1" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.private[0].id
  }

  lifecycle {
    ignore_changes = [
      route,
    ]
  }
}

resource "aws_route_table" "private_2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.private[1].id
  }

  lifecycle {
    ignore_changes = [
      route, tags
    ]
  }
}

```

## Argument Reference

The following arguments are required:

* `name` - (Required) Deployment Name

* `deployment_pattern` - (Required) Valid deployment pattern. See [API reference](https://docs.aws.amazon.com/launchwizard/latest/APIReference/launch-wizard-specifications-sap.html) for valid options.

* `workload_name` - (Required) Type of workload. Only supported option = `SAP`.

* `specifications` - (Required) A map of deployment parameters. See [API reference](https://docs.aws.amazon.com/launchwizard/latest/APIReference/launch-wizard-specifications-sap.html) for valid options.

The following arguments are optional:

* `skip_destroy_after_failure` - (Optional) Keeps deployment available after a failure to allow troubleshooting. Deployment will then be re-created on next apply. Defaults to `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Deployment ID.

* `status` - Current deployment status.

* `resource_group` - Name of resource group created by the deployment

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `180m`)
* `update` - (Default `180m`)
* `delete` - (Default `180m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Launch Wizard Deployment using the `name`. For example:

```terraform
import {
  to = aws_launchwizard_deployment.example
  id = "exampleHanaSingle"
}
```

Using `terraform import`, import Launch Wizard Deployment using the `name`. For example:

```console
% terraform import aws_launchwizard_deployment.example exampleHanaSingle
```
