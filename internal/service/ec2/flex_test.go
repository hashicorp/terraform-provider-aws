package ec2

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func TestExpandIPPerms(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "icmp",
			"from_port":   1,
			"to_port":     -1,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
			"description": "desc",
		},
		map[string]interface{}{
			"protocol":  "icmp",
			"from_port": 1,
			"to_port":   -1,
			"self":      true,
		},
	}
	group := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}
	perms, err := ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(1),
			ToPort:     aws.Int64(int64(-1)),
			IpRanges: []*ec2.IpRange{
				{
					CidrIp:      aws.String("0.0.0.0/0"),
					Description: aws.String("desc"),
				},
			},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					UserId:      aws.String("foo"),
					GroupId:     aws.String("sg-22222"),
					Description: aws.String("desc"),
				},
				{
					GroupId:     aws.String("sg-11111"),
					Description: aws.String("desc"),
				},
			},
		},
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(1),
			ToPort:     aws.Int64(int64(-1)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("foo"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].UserId) != aws.StringValue(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].UserId),
			aws.StringValue(exp.UserIdGroupPairs[0].UserId))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupId) != aws.StringValue(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupId))
	}

	if aws.StringValue(exp.UserIdGroupPairs[1].GroupId) != aws.StringValue(perm.UserIdGroupPairs[1].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[1].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[1].GroupId))
	}

	exp = expected[1]
	perm = perms[1]

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupId) != aws.StringValue(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupId))
	}
}

func TestExpandIPPerms_NegOneProtocol(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   0,
			"to_port":     0,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	group := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	perms, err := ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int64(0),
			ToPort:     aws.Int64(0),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					UserId:  aws.String("foo"),
					GroupId: aws.String("sg-22222"),
				},
				{
					GroupId: aws.String("sg-11111"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].UserId) != aws.StringValue(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].UserId),
			aws.StringValue(exp.UserIdGroupPairs[0].UserId))
	}

	// Now test the error case. This *should* error when either from_port
	// or to_port is not zero, but protocol is "-1".
	errorCase := []interface{}{
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   0,
			"to_port":     65535,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	securityGroups := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	_, expandErr := ExpandIPPerms(securityGroups, errorCase)
	if expandErr == nil {
		t.Fatal("ExpandIPPerms should have errored!")
	}
}

func TestExpandIPPerms_nonVPC(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "icmp",
			"from_port":   1,
			"to_port":     -1,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
		map[string]interface{}{
			"protocol":  "icmp",
			"from_port": 1,
			"to_port":   -1,
			"self":      true,
		},
	}
	group := &ec2.SecurityGroup{
		GroupName: aws.String("foo"),
	}
	perms, err := ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(1),
			ToPort:     aws.Int64(int64(-1)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupName: aws.String("sg-22222"),
				},
				{
					GroupName: aws.String("sg-11111"),
				},
			},
		},
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(1),
			ToPort:     aws.Int64(int64(-1)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupName: aws.String("foo"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupName) != aws.StringValue(perm.UserIdGroupPairs[0].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupName))
	}

	if aws.StringValue(exp.UserIdGroupPairs[1].GroupName) != aws.StringValue(perm.UserIdGroupPairs[1].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[1].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[1].GroupName))
	}

	exp = expected[1]
	perm = perms[1]

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupName) != aws.StringValue(perm.UserIdGroupPairs[0].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupName))
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

func TestExpandUserBucket(t *testing.T) {
	expanded := map[string]interface{}{
		"s3_bucket": "bucket",
		"s3_key":    "key",
	}

	bucket := ExpandUserBucket(expanded)

	expected := &ec2.UserBucket{
		S3Bucket: aws.String("bucket"),
		S3Key:    aws.String("key"),
	}

	if aws.StringValue(expected.S3Bucket) != aws.StringValue(bucket.S3Bucket) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(bucket.S3Bucket),
			aws.StringValue(expected.S3Bucket))
	}

	if aws.StringValue(expected.S3Key) != aws.StringValue(bucket.S3Key) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(bucket.S3Key),
			aws.StringValue(expected.S3Key))
	}
}

func TestExpandClientData(t *testing.T) {
	expanded := map[string]interface{}{
		"comment":      "bucket",
		"upload_end":   "2018-05-13T07:44:12Z",
		"upload_start": "2018-05-13T07:44:12Z",
		"upload_size":  123.123,
	}

	data := ExpandClientData(expanded)

	ti, _ := time.Parse(time.RFC3339, "2018-05-13T07:44:12Z")

	expected := &ec2.ClientData{
		Comment:     aws.String("bucket"),
		UploadEnd:   aws.Time(ti),
		UploadSize:  aws.Float64(123.123),
		UploadStart: aws.Time(ti),
	}

	if aws.StringValue(expected.Comment) != aws.StringValue(data.Comment) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(data.Comment),
			aws.StringValue(expected.Comment))
	}

	if !aws.TimeValue(expected.UploadEnd).Equal(aws.TimeValue(data.UploadEnd)) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.TimeValue(data.UploadEnd),
			aws.TimeValue(expected.UploadEnd))
	}

	if aws.Float64Value(expected.UploadSize) != aws.Float64Value(data.UploadSize) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Float64Value(data.UploadSize),
			aws.Float64Value(expected.UploadSize))
	}

	if !aws.TimeValue(expected.UploadStart).Equal(aws.TimeValue(data.UploadStart)) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.TimeValue(data.UploadStart),
			aws.TimeValue(expected.UploadStart))
	}

}
