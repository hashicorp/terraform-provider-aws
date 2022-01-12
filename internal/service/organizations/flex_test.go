package organizations

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
)

func TestFlattenOrganizationalUnits(t *testing.T) {
	input := []*organizations.OrganizationalUnit{
		{
			Arn:  aws.String("arn:aws:organizations::123456789012:ou/o-abcde12345/ou-ab12-abcd1234"), //lintignore:AWSAT005
			Id:   aws.String("ou-ab12-abcd1234"),
			Name: aws.String("Engineering"),
		},
	}

	expected_output := []map[string]interface{}{
		{
			"arn":  "arn:aws:organizations::123456789012:ou/o-abcde12345/ou-ab12-abcd1234", //lintignore:AWSAT005
			"id":   "ou-ab12-abcd1234",
			"name": "Engineering",
		},
	}

	output := FlattenOrganizationalUnits(input)
	if !reflect.DeepEqual(expected_output, output) {
		t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, expected_output)
	}
}
