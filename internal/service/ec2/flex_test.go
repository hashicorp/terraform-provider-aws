package ec2

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestFlattenGroupIdentifiers(t *testing.T) {
	expanded := []*ec2.GroupIdentifier{
		{GroupId: aws.String("sg-001")},
		{GroupId: aws.String("sg-002")},
	}

	result := FlattenGroupIdentifiers(expanded)

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if result[0] != "sg-001" {
		t.Fatalf("expected id to be sg-001, but was %s", result[0])
	}

	if result[1] != "sg-002" {
		t.Fatalf("expected id to be sg-002, but was %s", result[1])
	}
}

func TestFlattenSecurityGroups(t *testing.T) {
	cases := []struct {
		ownerId  *string
		pairs    []*ec2.UserIdGroupPair
		expected []*GroupIdentifier
	}{
		// simple, no user id included (we ignore it mostly)
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},
		// include the owner id, but keep it consitent with the same account. Tests
		// EC2 classic situation
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user1234"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},

		// include the owner id, but from a different account. This is reflects
		// EC2 Classic when referring to groups by name
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId:   aws.String("sg-12345"),
					GroupName: aws.String("somegroup"), // GroupName is only included in Classic
					UserId:    aws.String("user4321"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId:   aws.String("sg-12345"),
					GroupName: aws.String("user4321/somegroup"),
				},
			},
		},

		// include the owner id, but from a different account. This reflects in
		// EC2 VPC when referring to groups by id
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user4321"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("user4321/sg-12345"),
				},
			},
		},

		// include description
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
		},
	}

	for _, c := range cases {
		out := FlattenSecurityGroups(c.pairs, c.ownerId)
		if !reflect.DeepEqual(out, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", out, c.expected)
		}
	}
}
