// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSADAssessment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ad_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckADAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccADAssessmentConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckADAssessmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "assessment_id", regexache.MustCompile(`^da-.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDNSName, domainName),
					resource.TestCheckResourceAttr(resourceName, "customer_dns_ips.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "customer_dns_ips.0", "10.0.0.10"),
					resource.TestCheckResourceAttr(resourceName, "customer_dns_ips.1", "10.0.1.10"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_instance_ids.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "self_managed_instance_ids.0", regexache.MustCompile(`^i-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "self_managed_instance_ids.1", regexache.MustCompile(`^i-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "security_group_ids.0", regexache.MustCompile(`^sg-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "subnet_ids.0", regexache.MustCompile(`^subnet-.+$`)),
					resource.TestMatchResourceAttr(resourceName, "subnet_ids.1", regexache.MustCompile(`^subnet-.+$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrVPCID, regexache.MustCompile(`^vpc-.+$`)),
				),
			},
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccDSADAssessment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName(t)
	resourceName := "aws_directory_service_ad_assessment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckADAssessmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccADAssessmentConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckADAssessmentExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfds.ResourceADAssessment, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckADAssessmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_ad_assessment" {
				continue
			}
			_, err := tfds.FindADAssessmentByID(ctx, conn, rs.Primary.Attributes["assessment_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameADAssessment, rs.Primary.Attributes["assessment_id"], err)
			}

			return create.Error(names.DS, create.ErrActionCheckingDestroyed, tfds.ResNameADAssessment, rs.Primary.Attributes["assessment_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckADAssessmentExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameADAssessment, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["assessment_id"] == "" {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameADAssessment, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		_, err := tfds.FindADAssessmentByID(ctx, conn, rs.Primary.Attributes["assessment_id"])
		if err != nil {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameADAssessment, rs.Primary.Attributes["assessment_id"], err)
		}
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

	input := &directoryservice.ListADAssessmentsInput{}

	_, err := conn.ListADAssessments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccADAssessmentConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		testAccADAssessmentConfig_base(rName, domain),
		fmt.Sprintf(`
resource "aws_directory_service_ad_assessment" "test" {
  customer_dns_ips = ["10.0.0.10", "10.0.1.10"]
  dns_name         = %[2]q
  self_managed_instance_ids = [
    aws_cloudformation_stack.test.outputs.DomainController1InstanceId,
    aws_cloudformation_stack.test.outputs.DomainController2InstanceId,
  ]
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = aws_subnet.test[*].id
  vpc_id             = aws_vpc.test.id
}

`, rName, domain))
}

// lintignore:AWSAT005
func testAccADAssessmentConfig_base(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccLatestWindowsServer2022CoreAMIConfig(),
		fmt.Sprintf(`
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 2)
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s-igw"
  }
}

resource "aws_eip" "test" {
  domain = "vpc"
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    Name = "%[1]s-nat"
  }
  depends_on = [aws_internet_gateway.test]
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.test.id
  }

  tags = {
    Name = "%[1]s-priv"
  }
}

resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.test[count.index].id
  route_table_id = aws_route_table.private.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "%[1]s-pub"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id
  cidr_ipv4         = aws_vpc.test.cidr_block
  ip_protocol       = "-1"
}

resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    Name = %[1]q
  }
}
resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name
}

resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
    AMI                = data.aws_ami.win2022core-ami.id
    SecurityGroupId    = aws_security_group.test.id
    PrivateSubnet1     = aws_subnet.test[0].id
    PrivateSubnet2     = aws_subnet.test[1].id
    EC2InstanceProfile = aws_iam_instance_profile.test.name
    Domain             = %[2]q
  }
  template_body = <<EOF
Parameters:
  AMI:
    Type: AWS::EC2::Image::Id
  SecurityGroupId:
    Type: AWS::EC2::SecurityGroup::Id
  PrivateSubnet1: 
    Type: AWS::EC2::Subnet::Id
  PrivateSubnet2: 
    Type: AWS::EC2::Subnet::Id
  EC2InstanceProfile:
    Type: String
  Domain:
    Type: String
Resources:
  DomainController1:
    Type: AWS::EC2::Instance
    CreationPolicy:
      ResourceSignal:
        Timeout: PT30M
        Count: 1
    Properties:
      ImageId: !Ref AMI
      InstanceType: t3.medium
      SubnetId: !Ref PrivateSubnet1
      SecurityGroupIds:
        - !Ref SecurityGroupId
      IamInstanceProfile: !Ref EC2InstanceProfile
      PrivateIpAddress: 10.0.0.10
      MetadataOptions:
        HttpEndpoint: enabled
        HttpTokens: required
      BlockDeviceMappings:
        - DeviceName: /dev/sda1
          Ebs:
            Encrypted: true
            VolumeType: gp3
            DeleteOnTermination: true
      UserData:
        Fn::Base64: !Sub |
          <powershell>
          # Create a scheduled task to run after restart
          $action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument '-Command "cfn-signal.exe -e 0 --stack $${AWS::StackName} --resource DomainController1 --region $${AWS::Region}"'
          $trigger = New-ScheduledTaskTrigger -AtStartup
          $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount
          Register-ScheduledTask -TaskName "CFNSignal" -Action $action -Trigger $trigger -Principal $principal

          $NewPassword = ConvertTo-SecureString "SuperSecretPassw0rd" -AsPlainText -Force
          
          # Reset Administrator password
          Set-LocalUser -Name "Administrator" -Password $NewPassword
          
          # Enable Administrator account (in case it's disabled)
          Enable-LocalUser -Name "Administrator"

          # Install AD DS Role
          Install-WindowsFeature -Name AD-Domain-Services -IncludeManagementTools

          Start-Sleep -Seconds 5  # Wait for network to stabilize   
      
          # Install new forest

          $maxAttempts = 3
          $delaySeconds = 60
          $attempt = 0
          $success = $false

          while (-not $success -and $attempt -lt $maxAttempts) {
            $attempt++
            try {
                Install-ADDSForest -DomainName "$${Domain}" -SafeModeAdministratorPassword $NewPassword -Force -ErrorAction Stop
                $success = $true
            }
            catch {
                if ($attempt -lt $maxAttempts) {
                    Start-Sleep -Seconds $delaySeconds
                }
                else {
                    cfn-signal.exe -e 1 --stack $${AWS::StackName} --resource DomainController1 --region $${AWS::Region}
                    throw "Install new forest failed after $maxAttempts attempts."
                }
            }
          }
          </powershell>
      Tags:
        - Key: Name
          Value: %[1]s-dc1
        - Key: StackName
          Value: !Ref AWS::StackName
  DomainController2:
    Type: AWS::EC2::Instance
    DependsOn: DomainController1
    CreationPolicy:
      ResourceSignal:
        Timeout: PT30M
        Count: 1
    Properties:
      ImageId: !Ref AMI
      InstanceType: t3.medium
      SubnetId: !Ref PrivateSubnet2
      SecurityGroupIds:
        - !Ref SecurityGroupId
      IamInstanceProfile: !Ref EC2InstanceProfile
      PrivateIpAddress: 10.0.1.10
      MetadataOptions:
        HttpEndpoint: enabled
        HttpTokens: required
      BlockDeviceMappings:
        - DeviceName: /dev/sda1
          Ebs:
            Encrypted: true
            VolumeType: gp3
            DeleteOnTermination: true
      UserData:
        Fn::Base64: !Sub |
          <powershell>
          # Create a scheduled task to run after restart
          $action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument '-Command "cfn-signal.exe -e 0 --stack $${AWS::StackName} --resource DomainController2 --region $${AWS::Region}"'
          $trigger = New-ScheduledTaskTrigger -AtStartup
          $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount
          Register-ScheduledTask -TaskName "CFNSignal" -Action $action -Trigger $trigger -Principal $principal

          # Set DNS to first DC
          $adapter = Get-NetAdapter | Where-Object {$_.Status -eq "Up"}
          Set-DnsClientServerAddress -InterfaceIndex $adapter.InterfaceIndex -ServerAddresses "10.0.0.10"

          Start-Sleep -Seconds 3  # Wait for network to stabilize

          $domain = "$${Domain}"
          $Username = "Administrator"
          $FinalUserName = $domain + '\' + $Username

          $NewPassword = ConvertTo-SecureString "SuperSecretPassw0rd" -AsPlainText -Force
          $Credential = New-Object System.Management.Automation.PsCredential($FinalUserName, $NewPassword)  
          
          # Reset Administrator password
          Set-LocalUser -Name "Administrator" -Password $NewPassword    

          # Enable Administrator account (in case it's disabled)
          Enable-LocalUser -Name "Administrator"  

          # Install AD DS Role
          Install-WindowsFeature -Name AD-Domain-Services -IncludeManagementTools

          $maxAttempts = 3
          $delaySeconds = 60
          $attempt = 0
          $success = $false

          while (-not $success -and $attempt -lt $maxAttempts) {
            $attempt++
            try {
                Install-ADDSDomainController -DomainName $domain -InstallDns:$true -Credential $Credential -SafeModeAdministratorPassword $NewPassword -confirm:$false -ErrorAction Stop
                $success = $true
            }
            catch {
                if ($attempt -lt $maxAttempts) {
                    Start-Sleep -Seconds $delaySeconds
                }
                else {
                    cfn-signal.exe -e 1 --stack $${AWS::StackName} --resource DomainController2 --region $${AWS::Region}
                    throw "Domain controller promotion failed after $maxAttempts attempts."
                }
            }
          }
          </powershell>
      Tags:
        - Key: Name
          Value: %[1]s-dc2
        - Key: StackName
          Value: !Ref AWS::StackName
Outputs:
  DomainController1InstanceId:
    Description: Primary Domain Controller Instance Id
    Value: !GetAtt DomainController1.InstanceId

  DomainController2InstanceId:
    Description: Secondary Domain Controller Instance Id
    Value: !GetAtt DomainController2.InstanceId

EOF
  timeouts {
    create = "30m"
  }
  depends_on = [
    aws_route_table_association.private
  ]
}

`, rName, domain))
}

// testAccLatestWindowsServer2022CoreAMIConfig returns the configuration for a data source that
// describes the latest Microsoft Windows Server 2022 Core AMI.
// The data source is named 'win2022core-ami'.
func testAccLatestWindowsServer2022CoreAMIConfig() string {
	return `
data "aws_ami" "win2022core-ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["Windows_Server-2022-English-Core-Base-*"]
  }
}
`
}
