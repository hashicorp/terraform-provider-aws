package iam_test

import (
	"encoding/json"
	"reflect"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"testing"
)

func TestUnMarshallOrderOfPrincipalsShouldNotMatter(t *testing.T) {
	policy1 := `
		  {
			"Action": "sts:AssumeRole",
			"Principal": {
			  "Service": ["lambda.amazonaws.com", "service2.amazonaws.com"]
			},
			"Effect": "Allow",
			"Sid": ""
		  }`
	// Service order is different, but should be the same object for terraform
	policy2 := `
		  {
			"Action": "sts:AssumeRole",
			"Principal": {
			  "Service": ["service2.amazonaws.com", "lambda.amazonaws.com"]
			},
			"Effect": "Allow",
			"Sid": ""
		  }`

	var data1 tfiam.IAMPolicyStatement
	var data2 tfiam.IAMPolicyStatement
	err := json.Unmarshal([]byte(policy1), &data1)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(policy2), &data2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data1, data2) {
		t.Fatalf("should be equal, but was:\n%#v\nVS\n%#v\n", data1, data2)
	}
}
