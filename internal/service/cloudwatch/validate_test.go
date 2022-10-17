package cloudwatch

import (
	"strings"
	"testing"
)

func TestValidDashboardName(t *testing.T) {
	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello-world-012345",
	}
	for _, v := range validNames {
		_, errors := validDashboardName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CloudWatch dashboard name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"slash/in-the-middle",
		"dot.in-the-middle",
		strings.Repeat("W", 256), // > 255
	}
	for _, v := range invalidNames {
		_, errors := validDashboardName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CloudWatch dashboard name", v)
		}
	}
}

func TestValidEC2AutomateARN(t *testing.T) {
	validNames := []string{
		"arn:aws:automate:us-east-1:ec2:reboot",    //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:recover",   //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:stop",      //lintignore:AWSAT003,AWSAT005
		"arn:aws:automate:us-east-1:ec2:terminate", //lintignore:AWSAT003,AWSAT005
	}
	for _, v := range validNames {
		_, errors := validEC2AutomateARN(v, "test_property")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ARN: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		"arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment", // lintignore:AWSAT003,AWSAT005 // Beanstalk
		"arn:aws:iam::123456789012:user/David",                                             // lintignore:AWSAT005          // IAM User
		"arn:aws:rds:eu-west-1:123456789012:db:mysql-db",                                   // lintignore:AWSAT003,AWSAT005 // RDS
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                               // lintignore:AWSAT005          // S3 object
		"arn:aws:events:us-east-1:319201112229:rule/rule_name",                             // lintignore:AWSAT003,AWSAT005 // CloudWatch Rule
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",                  // lintignore:AWSAT003,AWSAT005 // Lambda function
		"arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction:Qualifier",        // lintignore:AWSAT003,AWSAT005 // Lambda func qualifier
		"arn:aws-us-gov:s3:::corp_bucket/object.png",                                       // lintignore:AWSAT005          // GovCloud ARN
		"arn:aws-us-gov:kms:us-gov-west-1:123456789012:key/some-uuid-abc123",               // lintignore:AWSAT003,AWSAT005 // GovCloud KMS ARN
	}
	for _, v := range invalidNames {
		_, errors := validEC2AutomateARN(v, "test_property")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ARN", v)
		}
	}
}
