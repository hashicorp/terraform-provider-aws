package lambda

import (
	"strings"
	"testing"
)

func TestValidFunctionName(t *testing.T) {
	validNames := []string{
		"arn:aws:lambda:us-west-2:123456789012:function:ThumbNail",            //lintignore:AWSAT003,AWSAT005
		"arn:aws-us-gov:lambda:us-west-2:123456789012:function:ThumbNail",     //lintignore:AWSAT003,AWSAT005
		"arn:aws-us-gov:lambda:us-gov-west-1:123456789012:function:ThumbNail", //lintignore:AWSAT003,AWSAT005
		"FunctionName",
		"function-name",
	}
	for _, v := range validNames {
		_, errors := validFunctionName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda function name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"/FunctionNameWithSlash",
		"function.name.with.dots",
		// length > 140
		"arn:aws:lambda:us-west-2:123456789012:function:TooLoooooo" + //lintignore:AWSAT003,AWSAT005
			"ooooooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooongFunctionName",
	}
	for _, v := range invalidNames {
		_, errors := validFunctionName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda function name", v)
		}
	}
}

func TestValidPermissionAction(t *testing.T) {
	validNames := []string{
		"lambda:*",
		"lambda:InvokeFunction",
		"*",
	}
	for _, v := range validNames {
		_, errors := validPermissionAction(v, "action")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda permission action: %q", v, errors)
		}
	}

	invalidNames := []string{
		"yada",
		"lambda:123",
		"*:*",
		"lambda:Invoke*",
	}
	for _, v := range invalidNames {
		_, errors := validPermissionAction(v, "action")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda permission action", v)
		}
	}
}

func TestValidPermissionEventSourceToken(t *testing.T) {
	validTokens := []string{
		"amzn1.ask.skill.80c92c86-e6dd-4c4b-8d0d-000000000000",
		"test-event-source-token",
		strings.Repeat(".", 256),
	}
	for _, v := range validTokens {
		_, errors := validPermissionEventSourceToken(v, "event_source_token")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda permission event source token", v)
		}
	}

	invalidTokens := []string{
		"!",
		"test event source token",
		strings.Repeat(".", 257),
	}
	for _, v := range invalidTokens {
		_, errors := validPermissionEventSourceToken(v, "event_source_token")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda permission event source token", v)
		}
	}
}

func TestValidQualifier(t *testing.T) {
	validNames := []string{
		"123",
		"prod",
		"PROD",
		"MyTestEnv",
		"contains-dashes",
		"contains_underscores",
		"$LATEST",
	}
	for _, v := range validNames {
		_, errors := validQualifier(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Lambda function qualifier: %q", v, errors)
		}
	}

	invalidNames := []string{
		// No ARNs allowed
		"arn:aws:lambda:us-west-2:123456789012:function:prod", //lintignore:AWSAT003,AWSAT005
		// length > 128
		"TooLooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooooooooooooooooooooooooooooooooooooo" +
			"oooooooooooongQualifier",
	}
	for _, v := range invalidNames {
		_, errors := validQualifier(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Lambda function qualifier", v)
		}
	}
}

func TestValidPolicyStatementID(t *testing.T) {
	validNames := []string{
		"YadaHereAndThere",
		"Valid-5tatement_Id",
		"1234",
	}
	for _, v := range validNames {
		_, errors := validPolicyStatementID(v, "statement_id")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Statement ID: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Invalid/StatementId/with/slashes",
		"InvalidStatementId.with.dots",
		// length > 100
		"TooooLoooooooooooooooooooooooooooooooooooooooooooo" +
			"ooooooooooooooooooooooooooooooooooooooooStatementId",
	}
	for _, v := range invalidNames {
		_, errors := validPolicyStatementID(v, "statement_id")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Statement ID", v)
		}
	}
}
