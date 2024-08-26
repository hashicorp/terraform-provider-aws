// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestProtocolStateFunc(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    interface{}
		expected string
	}{
		{
			input:    "tcp",
			expected: "tcp",
		},
		{
			input:    6,
			expected: "",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "all",
			expected: "-1",
		},
		{
			input:    "-1",
			expected: "-1",
		},
		{
			input:    -1,
			expected: "",
		},
		{
			input:    acctest.Ct1,
			expected: "icmp",
		},
		{
			input:    "icmp",
			expected: "icmp",
		},
		{
			input:    1,
			expected: "",
		},
		{
			input:    "icmpv6",
			expected: "icmpv6",
		},
		{
			input:    "58",
			expected: "icmpv6",
		},
		{
			input:    58,
			expected: "",
		},
	}
	for _, c := range cases {
		result := tfec2.ProtocolStateFunc(c.input)
		if result != c.expected {
			t.Errorf("Error matching protocol, expected (%s), got (%s)", c.expected, result)
		}
	}
}

func TestProtocolForValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "tcp",
			expected: "tcp",
		},
		{
			input:    "6",
			expected: "tcp",
		},
		{
			input:    "udp",
			expected: "udp",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "all",
			expected: "-1",
		},
		{
			input:    "-1",
			expected: "-1",
		},
		{
			input:    "tCp",
			expected: "tcp",
		},
		{
			input:    "6",
			expected: "tcp",
		},
		{
			input:    "UDp",
			expected: "udp",
		},
		{
			input:    "17",
			expected: "udp",
		},
		{
			input:    "ALL",
			expected: "-1",
		},
		{
			input:    "icMp",
			expected: "icmp",
		},
		{
			input:    acctest.Ct1,
			expected: "icmp",
		},
		{
			input:    "icMpv6",
			expected: "icmpv6",
		},
		{
			input:    "58",
			expected: "icmpv6",
		},
	}

	for _, c := range cases {
		result := tfec2.ProtocolForValue(c.input)
		if result != c.expected {
			t.Errorf("Error matching protocol, expected (%s), got (%s)", c.expected, result)
		}
	}
}

func calcSecurityGroupChecksum(rules []interface{}) int {
	sum := 0
	for _, rule := range rules {
		sum += tfec2.SecurityGroupRuleHash(rule)
	}
	return sum
}

func TestSecurityGroupExpandCollapseRules(t *testing.T) {
	t.Parallel()

	expected_compact_list := []interface{}{
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with description",
			"self":                true,
			"cidr_blocks": []interface{}{
				"10.0.0.1/32",
				"10.0.0.2/32",
				"10.0.0.3/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with another description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"192.168.0.1/32",
				"192.168.0.2/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::1/128",
				"fd00::2/128",
			},
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-11111",
				"sg-22222",
				"sg-33333",
			}),
		},
		map[string]interface{}{
			names.AttrProtocol:    "udp",
			"from_port":           int(10000),
			"to_port":             int(10000),
			names.AttrDescription: "",
			"self":                false,
			"prefix_list_ids": []interface{}{
				"pl-111111",
				"pl-222222",
			},
		},
	}

	expected_expanded_list := []interface{}{
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with description",
			"self":                true,
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"10.0.0.1/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"10.0.0.2/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"10.0.0.3/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with another description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"192.168.0.1/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "tcp",
			"from_port":           int(443),
			"to_port":             int(443),
			names.AttrDescription: "block with another description",
			"self":                false,
			"cidr_blocks": []interface{}{
				"192.168.0.2/32",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::1/128",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			"ipv6_cidr_blocks": []interface{}{
				"fd00::2/128",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-11111",
			}),
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
		},
		map[string]interface{}{
			names.AttrProtocol:    "-1",
			"from_port":           int(8000),
			"to_port":             int(8080),
			names.AttrDescription: "",
			"self":                false,
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-33333",
			}),
		},
		map[string]interface{}{
			names.AttrProtocol:    "udp",
			"from_port":           int(10000),
			"to_port":             int(10000),
			names.AttrDescription: "",
			"self":                false,
			"prefix_list_ids": []interface{}{
				"pl-111111",
			},
		},
		map[string]interface{}{
			names.AttrProtocol:    "udp",
			"from_port":           int(10000),
			"to_port":             int(10000),
			names.AttrDescription: "",
			"self":                false,
			"prefix_list_ids": []interface{}{
				"pl-222222",
			},
		},
	}

	expected_compact_set := schema.NewSet(tfec2.SecurityGroupRuleHash, expected_compact_list)
	actual_expanded_list := tfec2.SecurityGroupExpandRules(expected_compact_set).List()

	if calcSecurityGroupChecksum(expected_expanded_list) != calcSecurityGroupChecksum(actual_expanded_list) {
		t.Fatalf("error matching expanded set for tfec2.SecurityGroupExpandRules()")
	}

	actual_collapsed_list := tfec2.SecurityGroupCollapseRules("ingress", expected_expanded_list)

	if calcSecurityGroupChecksum(expected_compact_list) != calcSecurityGroupChecksum(actual_collapsed_list) {
		t.Fatalf("error matching collapsed set for tfec2.SecurityGroupCollapseRules()")
	}
}

func TestSecurityGroupIPPermGather(t *testing.T) {
	t.Parallel()

	raw := []awstypes.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(1),
			ToPort:     aws.Int32(int32(-1)),
			IpRanges:   []awstypes.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
				{
					GroupId:     aws.String("sg-11111"),
					Description: aws.String("desc"),
				},
			},
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(80),
			ToPort:     aws.Int32(80),
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
				// VPC
				{
					GroupId: aws.String("sg-22222"),
				},
			},
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(443),
			ToPort:     aws.Int32(443),
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
				{
					UserId:    aws.String("amazon-elb"),
					GroupId:   aws.String("sg-d2c979d3"),
					GroupName: aws.String("amazon-elb-sg"),
				},
			},
		},
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int32(0),
			ToPort:     aws.Int32(0),
			PrefixListIds: []awstypes.PrefixListId{
				{
					PrefixListId: aws.String("pl-12345678"),
					Description:  aws.String("desc"),
				},
			},
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
				// VPC
				{
					GroupId: aws.String("sg-22222"),
				},
			},
		},
	}

	local := []map[string]interface{}{
		{
			names.AttrProtocol:    "tcp",
			"from_port":           int64(1),
			"to_port":             int64(-1),
			"cidr_blocks":         []string{"0.0.0.0/0"},
			"self":                true,
			names.AttrDescription: "desc",
		},
		{
			names.AttrProtocol: "tcp",
			"from_port":        int64(80),
			"to_port":          int64(80),
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
		},
		{
			names.AttrProtocol: "-1",
			"from_port":        int64(0),
			"to_port":          int64(0),
			"prefix_list_ids":  []string{"pl-12345678"},
			names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{
				"sg-22222",
			}),
			names.AttrDescription: "desc",
		},
	}

	out := tfec2.SecurityGroupIPPermGather("sg-11111", raw, aws.String("12345"))
	for _, i := range out {
		// loop and match rules, because the ordering is not guarneteed
		for _, l := range local {
			if i["from_port"] == l["from_port"] {
				if i["to_port"] != l["to_port"] {
					t.Fatalf("to_port does not match")
				}

				if _, ok := i["cidr_blocks"]; ok {
					if !reflect.DeepEqual(i["cidr_blocks"], l["cidr_blocks"]) {
						t.Fatalf("error matching cidr_blocks")
					}
				}

				if _, ok := i[names.AttrSecurityGroups]; ok {
					outSet := i[names.AttrSecurityGroups].(*schema.Set)
					localSet := l[names.AttrSecurityGroups].(*schema.Set)

					if !outSet.Equal(localSet) {
						t.Fatalf("Security Group sets are not equal")
					}
				}
			}
		}
	}
}

func TestExpandIPPerms(t *testing.T) {
	t.Parallel()

	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			names.AttrProtocol: "icmp",
			"from_port":        1,
			"to_port":          -1,
			"cidr_blocks":      []interface{}{"0.0.0.0/0"},
			names.AttrSecurityGroups: schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
			names.AttrDescription: "desc",
		},
		map[string]interface{}{
			names.AttrProtocol: "icmp",
			"from_port":        1,
			"to_port":          -1,
			"self":             true,
		},
	}
	group := &awstypes.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}
	perms, err := tfec2.ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []awstypes.IpPermission{
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int32(1),
			ToPort:     aws.Int32(int32(-1)),
			IpRanges: []awstypes.IpRange{
				{
					CidrIp:      aws.String("0.0.0.0/0"),
					Description: aws.String("desc"),
				},
			},
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
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
			FromPort:   aws.Int32(1),
			ToPort:     aws.Int32(int32(-1)),
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
				{
					GroupId: aws.String("foo"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.ToInt32(exp.FromPort) != aws.ToInt32(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToInt32(perm.FromPort),
			aws.ToInt32(exp.FromPort))
	}

	if aws.ToString(exp.IpRanges[0].CidrIp) != aws.ToString(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.IpRanges[0].CidrIp),
			aws.ToString(exp.IpRanges[0].CidrIp))
	}

	if aws.ToString(exp.UserIdGroupPairs[0].UserId) != aws.ToString(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[0].UserId),
			aws.ToString(exp.UserIdGroupPairs[0].UserId))
	}

	if aws.ToString(exp.UserIdGroupPairs[0].GroupId) != aws.ToString(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[0].GroupId),
			aws.ToString(exp.UserIdGroupPairs[0].GroupId))
	}

	if aws.ToString(exp.UserIdGroupPairs[1].GroupId) != aws.ToString(perm.UserIdGroupPairs[1].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[1].GroupId),
			aws.ToString(exp.UserIdGroupPairs[1].GroupId))
	}

	exp = expected[1]
	perm = perms[1]

	if aws.ToString(exp.UserIdGroupPairs[0].GroupId) != aws.ToString(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[0].GroupId),
			aws.ToString(exp.UserIdGroupPairs[0].GroupId))
	}
}

func TestExpandIPPerms_NegOneProtocol(t *testing.T) {
	t.Parallel()

	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			names.AttrProtocol: "-1",
			"from_port":        0,
			"to_port":          0,
			"cidr_blocks":      []interface{}{"0.0.0.0/0"},
			names.AttrSecurityGroups: schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	group := &awstypes.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	perms, err := tfec2.ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []awstypes.IpPermission{
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int32(0),
			ToPort:     aws.Int32(0),
			IpRanges:   []awstypes.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
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

	if aws.ToInt32(exp.FromPort) != aws.ToInt32(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToInt32(perm.FromPort),
			aws.ToInt32(exp.FromPort))
	}

	if aws.ToString(exp.IpRanges[0].CidrIp) != aws.ToString(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.IpRanges[0].CidrIp),
			aws.ToString(exp.IpRanges[0].CidrIp))
	}

	if aws.ToString(exp.UserIdGroupPairs[0].UserId) != aws.ToString(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[0].UserId),
			aws.ToString(exp.UserIdGroupPairs[0].UserId))
	}

	// Now test the error case. This *should* error when either from_port
	// or to_port is not zero, but protocol is "-1".
	errorCase := []interface{}{
		map[string]interface{}{
			names.AttrProtocol: "-1",
			"from_port":        0,
			"to_port":          65535,
			"cidr_blocks":      []interface{}{"0.0.0.0/0"},
			names.AttrSecurityGroups: schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	securityGroups := &awstypes.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	_, expandErr := tfec2.ExpandIPPerms(securityGroups, errorCase)
	if expandErr == nil {
		t.Fatal("ExpandIPPerms should have errored!")
	}
}

func TestExpandIPPerms_AllProtocol(t *testing.T) {
	t.Parallel()

	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			names.AttrProtocol: "all",
			"from_port":        0,
			"to_port":          0,
			"cidr_blocks":      []interface{}{"0.0.0.0/0"},
			names.AttrSecurityGroups: schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	group := &awstypes.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	perms, err := tfec2.ExpandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []awstypes.IpPermission{
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int32(0),
			ToPort:     aws.Int32(0),
			IpRanges:   []awstypes.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []awstypes.UserIdGroupPair{
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

	if aws.ToInt32(exp.FromPort) != aws.ToInt32(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToInt32(perm.FromPort),
			aws.ToInt32(exp.FromPort))
	}

	if aws.ToString(exp.IpRanges[0].CidrIp) != aws.ToString(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.IpRanges[0].CidrIp),
			aws.ToString(exp.IpRanges[0].CidrIp))
	}

	if aws.ToString(exp.UserIdGroupPairs[0].UserId) != aws.ToString(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.ToString(perm.UserIdGroupPairs[0].UserId),
			aws.ToString(exp.UserIdGroupPairs[0].UserId))
	}

	// Now test the error case. This *should* error when either from_port
	// or to_port is not zero, but protocol is "all".
	errorCase := []interface{}{
		map[string]interface{}{
			names.AttrProtocol: "all",
			"from_port":        0,
			"to_port":          65535,
			"cidr_blocks":      []interface{}{"0.0.0.0/0"},
			names.AttrSecurityGroups: schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	securityGroups := &awstypes.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	_, expandErr := tfec2.ExpandIPPerms(securityGroups, errorCase)
	if expandErr == nil {
		t.Fatal("ExpandIPPerms should have errored!")
	}
}

func TestFlattenSecurityGroups(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ownerId  *string
		pairs    []awstypes.UserIdGroupPair
		expected []*tfec2.GroupIdentifier
	}{
		// simple, no user id included (we ignore it mostly)
		{
			ownerId: aws.String("user1234"),
			pairs: []awstypes.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
			expected: []*tfec2.GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},
		{
			ownerId: aws.String("user1234"),
			pairs: []awstypes.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user1234"),
				},
			},
			expected: []*tfec2.GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},
		{
			ownerId: aws.String("user1234"),
			pairs: []awstypes.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user4321"),
				},
			},
			expected: []*tfec2.GroupIdentifier{
				{
					GroupId: aws.String("user4321/sg-12345"),
				},
			},
		},

		// include description
		{
			ownerId: aws.String("user1234"),
			pairs: []awstypes.UserIdGroupPair{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
			expected: []*tfec2.GroupIdentifier{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
		},
	}

	for _, c := range cases {
		out := tfec2.FlattenSecurityGroups(c.pairs, c.ownerId)
		if !reflect.DeepEqual(out, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", out, c.expected)
		}
	}
}

func TestAccVPCSecurityGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`security-group/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "revoke_rules_on_delete", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSecurityGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroup_noVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_noVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "data.aws_vpc.default", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config:   testAccVPCSecurityGroupConfig_defaultVPC(rName),
				PlanOnly: true,
			},
			{
				Config:   testAccVPCSecurityGroupConfig_noVPC(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCSecurityGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
func TestAccVPCSecurityGroup_nameTerraformPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix("terraform-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_namePrefix(rName, "tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17017
func TestAccVPCSecurityGroup_namePrefixTerraform(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_namePrefix(rName, "terraform-test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "terraform-test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_allowAll(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_allowAll(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_sourceSecurityGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_sourceSecurityGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ipRangeAndSecurityGroupWithSameRules(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ipRangeAndSecurityGroupWithSameRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ipRangesWithSameRules(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ipRangesWithSameRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_egressMode(t *testing.T) {
	ctx := acctest.Context(t)
	var securityGroup1, securityGroup2, securityGroup3 awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_egressModeBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup1),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccVPCSecurityGroupConfig_egressModeNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup2),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCSecurityGroupConfig_egressModeZeroed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup3),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_ingressMode(t *testing.T) {
	ctx := acctest.Context(t)
	var securityGroup1, securityGroup2, securityGroup3 awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ingressModeBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup1),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccVPCSecurityGroupConfig_ingressModeNoBlocks(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup2),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
				),
			},
			{
				Config: testAccVPCSecurityGroupConfig_ingressModeZeroed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &securityGroup3),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_ruleGathering(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ruleGathering(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct0,
						names.AttrDescription: "egress for all ipv6",
						"from_port":           acctest.Ct0,
						"ipv6_cidr_blocks.#":  acctest.Ct1,
						"ipv6_cidr_blocks.0":  "::/0",
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "-1",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "0.0.0.0/0",
						names.AttrDescription: "egress for all ipv4",
						"from_port":           acctest.Ct0,
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "-1",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "192.168.0.0/16",
						names.AttrDescription: "ingress from 192.168.0.0/16",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct0,
						names.AttrDescription: "ingress from all ipv6",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct1,
						"ipv6_cidr_blocks.0":  "::/0",
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct2,
						"cidr_blocks.0":       "10.0.2.0/24",
						"cidr_blocks.1":       "10.0.3.0/24",
						names.AttrDescription: "ingress from 10.0.0.0/16",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "80",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct2,
						"cidr_blocks.0":       "10.0.0.0/24",
						"cidr_blocks.1":       "10.0.1.0/24",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtTrue,
						"to_port":             "80",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// This test should fail to destroy the Security Groups and VPC, due to a
// dependency cycle added outside of terraform's management. There is a sweeper
// 'aws_vpc' and 'aws_security_group' that cleans these up, however, the test is
// written to allow Terraform to clean it up because we do go and revoke the
// cyclic rules that were added.
func TestAccVPCSecurityGroup_forceRevokeRulesTrue(t *testing.T) {
	ctx := acctest.Context(t)
	var primary awstypes.SecurityGroup
	var secondary awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.primary"
	resourceName2 := "aws_security_group.secondary"

	// Add rules to create a cycle between primary and secondary. This prevents
	// Terraform/AWS from being able to destroy the groups
	testAddCycle := testAddRuleCycle(ctx, &primary, &secondary)
	// Remove the rules that created the cycle; Terraform/AWS can now destroy them
	testRemoveCycle := testRemoveRuleCycle(ctx, &primary, &secondary)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted
			{
				Config: testAccVPCSecurityGroupConfig_revokeBase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &primary),
					testAccCheckSecurityGroupExists(ctx, resourceName2, &secondary),
					testAddCycle,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Because of the cyclic dependency created in testAddCycle, we add data outside of terraform to this resource.
				// During an import this cannot be accounted for and should be ignored.
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "egress"},
			},
			// Verify the DependencyViolation error by using a configuration with the
			// groups removed. Terraform tries to destroy them but cannot. Expect a
			// DependencyViolation error
			{
				Config:      testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName),
				ExpectError: regexache.MustCompile("DependencyViolation"),
			},
			// Restore the config (a no-op plan) but also remove the dependencies
			// between the groups with testRemoveCycle
			{
				Config: testAccVPCSecurityGroupConfig_revokeBase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &primary),
					testAccCheckSecurityGroupExists(ctx, resourceName2, &secondary),
					testRemoveCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work
			{
				Config: testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName),
			},
			////
			// now test with revoke_rules_on_delete
			////
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted. In this
			// configuration, each Security Group has `revoke_rules_on_delete`
			// specified, and should delete with no issue
			{
				Config: testAccVPCSecurityGroupConfig_revokeTrue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &primary),
					testAccCheckSecurityGroupExists(ctx, resourceName2, &secondary),
					testAddCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work,
			// because we've told the SGs to forcefully revoke their rules first
			{
				Config: testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName),
			},
		},
	})
}

func TestAccVPCSecurityGroup_forceRevokeRulesFalse(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var primary awstypes.SecurityGroup
	var secondary awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.primary"
	resourceName2 := "aws_security_group.secondary"

	// Add rules to create a cycle between primary and secondary. This prevents
	// Terraform/AWS from being able to destroy the groups
	testAddCycle := testAddRuleCycle(ctx, &primary, &secondary)
	// Remove the rules that created the cycle; Terraform/AWS can now destroy them
	testRemoveCycle := testRemoveRuleCycle(ctx, &primary, &secondary)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create the configuration with 2 security groups, then create a
			// dependency cycle such that they cannot be deleted. These Security
			// Groups are configured to explicitly not revoke rules on delete,
			// `revoke_rules_on_delete = false`
			{
				Config: testAccVPCSecurityGroupConfig_revokeFalse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &primary),
					testAccCheckSecurityGroupExists(ctx, resourceName2, &secondary),
					testAddCycle,
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Because of the cyclic dependency created in testAddCycle, we add data outside of terraform to this resource.
				// During an import this cannot be accounted for and should be ignored.
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "egress"},
			},
			// Verify the DependencyViolation error by using a configuration with the
			// groups removed, and the Groups not configured to revoke their ruls.
			// Terraform tries to destroy them but cannot. Expect a
			// DependencyViolation error
			{
				Config:      testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName),
				ExpectError: regexache.MustCompile("DependencyViolation"),
			},
			// Restore the config (a no-op plan) but also remove the dependencies
			// between the groups with testRemoveCycle
			{
				Config: testAccVPCSecurityGroupConfig_revokeFalse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &primary),
					testAccCheckSecurityGroupExists(ctx, resourceName2, &secondary),
					testRemoveCycle,
				),
			},
			// Again try to apply the config with the sgs removed; it should work
			{
				Config: testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName),
			},
		},
	})
}

func TestAccVPCSecurityGroup_change(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			{
				Config: testAccVPCSecurityGroupConfig_changed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "9000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct2,
						"cidr_blocks.0":       "0.0.0.0/0",
						"cidr_blocks.1":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct0,
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct1,
						"ipv6_cidr_blocks.0":  "::/0",
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct0,
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct1,
						"ipv6_cidr_blocks.0":  "::/0",
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_self(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	checkSelf := func(s *terraform.State) (err error) {
		if len(group.IpPermissions) > 0 &&
			len(group.IpPermissions[0].UserIdGroupPairs) > 0 &&
			aws.ToString(group.IpPermissions[0].UserIdGroupPairs[0].GroupId) == aws.ToString(group.GroupId) {
			return nil
		}

		return fmt.Errorf("Security Group does not contain \"self\" rule: %#v", group)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_self(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "80",
						"to_port":          "8000",
						"self":             acctest.CtTrue,
					}),
					checkSelf,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "80",
						"to_port":          "8000",
						"cidr_blocks.#":    acctest.Ct1,
						"cidr_blocks.0":    "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						names.AttrProtocol: "tcp",
						"from_port":        "80",
						"to_port":          "8000",
						"cidr_blocks.#":    acctest.Ct1,
						"cidr_blocks.0":    "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_vpcNegOneIngress(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_vpcNegativeOneIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "-1",
						"from_port":        acctest.Ct0,
						"to_port":          acctest.Ct0,
						"cidr_blocks.#":    acctest.Ct1,
						"cidr_blocks.0":    "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_vpcProtoNumIngress(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_vpcProtocolNumberIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						names.AttrProtocol: "50",
						"from_port":        acctest.Ct0,
						"to_port":          acctest.Ct0,
						"cidr_blocks.#":    acctest.Ct1,
						"cidr_blocks.0":    "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_multiIngress(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_multiIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_vpcAllEgress(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_vpcAllEgress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						names.AttrProtocol: "-1",
						"from_port":        acctest.Ct0,
						"to_port":          acctest.Ct0,
						"cidr_blocks.#":    acctest.Ct1,
						"cidr_blocks.0":    "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ruleDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ruleDescription(rName, "Egress description", "Ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "Egress description",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "Ingress description",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
			// Change just the rule descriptions.
			{
				Config: testAccVPCSecurityGroupConfig_ruleDescription(rName, "New egress description", "New ingress description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "New egress description",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "New ingress description",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
			// Remove just the rule descriptions.
			{
				Config: testAccVPCSecurityGroupConfig_emptyRuleDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_defaultEgressVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_defaultEgress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

// Testing drift detection with groups containing the same port and types
func TestAccVPCSecurityGroup_driftComplex(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_driftComplex(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "206.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "206.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// In rules with cidr_block drift, import only creates a single ingress
				// rule with the cidr_blocks de-normalized. During subsequent apply, its
				// normalized to create the 2 ingress rules seen in checks above.
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "ingress", "egress"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_invalidCIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupConfig_invalidIngressCIDR,
				ExpectError: regexache.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccVPCSecurityGroupConfig_invalidEgressCIDR,
				ExpectError: regexache.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccVPCSecurityGroupConfig_invalidIPv6IngressCIDR,
				ExpectError: regexache.MustCompile("invalid CIDR address: ::/244"),
			},
			{
				Config:      testAccVPCSecurityGroupConfig_invalidIPv6EgressCIDR,
				ExpectError: regexache.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func TestAccVPCSecurityGroup_cidrAndGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_combinedCIDRAndGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ingressWithCIDRAndSGsVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ingressWithCIDRAndSGs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "10.0.0.0/8",
						names.AttrDescription: "",
						"from_port":           "80",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "8000",
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "192.168.0.1/32",
						names.AttrDescription: "",
						"from_port":           "22",
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						names.AttrProtocol:    "tcp",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             "22",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_egressWithPrefixList(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_prefixListEgress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ingressWithPrefixList(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_prefixListIngress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_ipv4AndIPv6Egress(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_ipv4andIPv6Egress(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct1,
						"cidr_blocks.0":       "0.0.0.0/0",
						names.AttrDescription: "",
						"from_port":           acctest.Ct0,
						"ipv6_cidr_blocks.#":  acctest.Ct0,
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "-1",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"cidr_blocks.#":       acctest.Ct0,
						names.AttrDescription: "",
						"from_port":           acctest.Ct0,
						"ipv6_cidr_blocks.#":  acctest.Ct1,
						"ipv6_cidr_blocks.0":  "::/0",
						"prefix_list_ids.#":   acctest.Ct0,
						names.AttrProtocol:    "-1",
						"security_groups.#":   acctest.Ct0,
						"self":                acctest.CtFalse,
						"to_port":             acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete", "egress"},
			},
		},
	})
}

func TestAccVPCSecurityGroup_failWithDiffMismatch(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_failWithDiffMismatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct2),
				),
			},
		},
	})
}

var ruleLimit int

// testAccSecurityGroup_ruleLimit sets the global "ruleLimit" and is only called once
// but does not run in parallel slowing down tests. It cannot run in parallel since
// it is called by another test and double paralleling is a panic.
func testAccSecurityGroup_ruleLimit(t *testing.T) {
	ctx := acctest.Context(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// get limit
			{
				Config: testAccVPCSecurityGroupConfig_getLimit(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleLimit("data.aws_servicequotas_service_quota.test", &ruleLimit),
				),
				PreventPostDestroyRefresh: true, // saves a few seconds
			},
		},
	})
}

func TestAccVPCSecurityGroup_RuleLimit_exceededAppend(t *testing.T) {
	ctx := acctest.Context(t)
	if ruleLimit == 0 {
		testAccSecurityGroup_ruleLimit(t)
	}

	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
					resource.TestCheckResourceAttr(resourceName, "egress.#", strconv.Itoa(ruleLimit)),
				),
			},
			// append a rule to step over the limit
			{
				Config:      testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit+1),
				ExpectError: regexache.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original rules still
					err := testSecurityGroupRuleCount(ctx, aws.ToString(group.GroupId), 0, ruleLimit)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
					resource.TestCheckResourceAttr(resourceName, "egress.#", strconv.Itoa(ruleLimit)),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_RuleLimit_cidrBlockExceededAppend(t *testing.T) {
	ctx := acctest.Context(t)
	if ruleLimit == 0 {
		testAccSecurityGroup_ruleLimit(t)
	}

	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccVPCSecurityGroupConfig_cidrBlockRuleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, 1),
				),
			},
			// append a rule to step over the limit
			{
				Config:      testAccVPCSecurityGroupConfig_cidrBlockRuleLimit(rName, 0, ruleLimit+1),
				ExpectError: regexache.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original cidr blocks still in 1 rule
					err := testSecurityGroupRuleCount(ctx, aws.ToString(group.GroupId), 0, 1)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}

					id := aws.ToString(group.GroupId)

					conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

					match, err := tfec2.FindSecurityGroupByID(ctx, conn, id)
					if tfresource.NotFound(err) {
						t.Fatalf("PreConfig check failed: Security Group (%s) not found: %s", id, err)
					}
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}

					if cidrCount := len(match.IpPermissionsEgress[0].IpRanges); cidrCount != ruleLimit {
						t.Fatalf("PreConfig check failed: rule does not have previous IP ranges, has %d", cidrCount)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccVPCSecurityGroupConfig_cidrBlockRuleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, 1),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_RuleLimit_exceededPrepend(t *testing.T) {
	ctx := acctest.Context(t)
	if ruleLimit == 0 {
		testAccSecurityGroup_ruleLimit(t)
	}

	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
				),
			},
			// prepend a rule to step over the limit
			{
				Config:      testAccVPCSecurityGroupConfig_ruleLimit(rName, 1, ruleLimit+1),
				ExpectError: regexache.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				PreConfig: func() {
					// should have the original rules still (limit - 1 because of the shift)
					err := testSecurityGroupRuleCount(ctx, aws.ToString(group.GroupId), 0, ruleLimit-1)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_RuleLimit_exceededAllNew(t *testing.T) {
	ctx := acctest.Context(t)
	if ruleLimit == 0 {
		testAccSecurityGroup_ruleLimit(t)
	}

	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// create a valid SG just under the limit
			{
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
				),
			},
			// add a rule to step over the limit with entirely new rules
			{
				Config:      testAccVPCSecurityGroupConfig_ruleLimit(rName, 100, ruleLimit+1),
				ExpectError: regexache.MustCompile("RulesPerSecurityGroupLimitExceeded"),
			},
			{
				// all the rules should have been revoked and the add failed
				PreConfig: func() {
					err := testSecurityGroupRuleCount(ctx, aws.ToString(group.GroupId), 0, 0)
					if err != nil {
						t.Fatalf("PreConfig check failed: %s", err)
					}
				},
				// running the original config again now should restore the rules
				Config: testAccVPCSecurityGroupConfig_ruleLimit(rName, 0, ruleLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, 0, ruleLimit),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroup_rulesDropOnError(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			// Create a valid security group with some rules and make sure it exists
			{
				Config: testAccVPCSecurityGroupConfig_rulesDropOnErrorInit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
				),
			},
			// Add a bad rule to trigger API error
			{
				Config:      testAccVPCSecurityGroupConfig_rulesDropOnErrorAddBadRule(rName),
				ExpectError: regexache.MustCompile("InvalidGroupId.Malformed"),
			},
			// All originally added rules must survive. This will return non-empty plan if anything changed.
			{
				Config:   testAccVPCSecurityGroupConfig_rulesDropOnErrorInit(rName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccVPCSecurityGroup_emrDependencyViolation is very complex but captures
// a problem seen in EMR and other services. The main gist is that a security
// group can have 0 rules and still have dependencies. Services, like EMR,
// create rules in security groups. If a 0-rule SG is listed as the source of
// a rule in another SG, it could not previously be deleted.
func TestAccVPCSecurityGroup_emrDependencyViolation(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var group awstypes.SecurityGroup
	resourceName := "aws_security_group.allow_access"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupConfig_emrLinkedRulesDestroy(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, resourceName, &group),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`security-group/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "egress.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "revoke_rules_on_delete", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, "aws_vpc.test", names.AttrID),
				),
				ExpectError: regexache.MustCompile("unexpected state"),
			},
		},
	})
}

// cycleIPPermForGroup returns an IpPermission struct with a configured
// UserIdGroupPair for the groupid given. Used in
// TestAccAWSSecurityGroup_forceRevokeRules_should_fail to create a cyclic rule
// between 2 security groups
func cycleIPPermForGroup(groupId string) awstypes.IpPermission {
	var perm awstypes.IpPermission
	perm.FromPort = aws.Int32(0)
	perm.ToPort = aws.Int32(0)
	perm.IpProtocol = aws.String("icmp")
	perm.UserIdGroupPairs = make([]awstypes.UserIdGroupPair, 1)
	perm.UserIdGroupPairs[0] = awstypes.UserIdGroupPair{
		GroupId: aws.String(groupId),
	}
	return perm
}

// testAddRuleCycle returns a TestCheckFunc to use at the end of a test, such
// that a Security Group Rule cyclic dependency will be created between the two
// Security Groups. A companion function, testRemoveRuleCycle, will undo this.
func testAddRuleCycle(ctx context.Context, primary, secondary *awstypes.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if primary.GroupId == nil {
			return fmt.Errorf("Primary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}
		if secondary.GroupId == nil {
			return fmt.Errorf("Secondary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		// cycle from primary to secondary
		perm1 := cycleIPPermForGroup(aws.ToString(secondary.GroupId))
		// cycle from secondary to primary
		perm2 := cycleIPPermForGroup(aws.ToString(primary.GroupId))

		req1 := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       primary.GroupId,
			IpPermissions: []awstypes.IpPermission{perm1},
		}
		req2 := &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       secondary.GroupId,
			IpPermissions: []awstypes.IpPermission{perm2},
		}

		var err error
		_, err = conn.AuthorizeSecurityGroupEgress(ctx, req1)
		if err != nil {
			return fmt.Errorf("Error authorizing primary security group %s rules: %w", aws.ToString(primary.GroupId), err)
		}
		_, err = conn.AuthorizeSecurityGroupEgress(ctx, req2)
		if err != nil {
			return fmt.Errorf("Error authorizing secondary security group %s rules: %w", aws.ToString(secondary.GroupId), err)
		}
		return nil
	}
}

// testRemoveRuleCycle removes the cyclic dependency between two security groups
// that was added in testAddRuleCycle
func testRemoveRuleCycle(ctx context.Context, primary, secondary *awstypes.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if primary.GroupId == nil {
			return fmt.Errorf("Primary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}
		if secondary.GroupId == nil {
			return fmt.Errorf("Secondary SG not set for TestAccAWSSecurityGroup_forceRevokeRules_should_fail")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)
		for _, sg := range []*awstypes.SecurityGroup{primary, secondary} {
			var err error
			if sg.IpPermissions != nil && len(sg.IpPermissions) > 0 {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}

				if _, err = conn.RevokeSecurityGroupIngress(ctx, req); err != nil {
					return fmt.Errorf("Error revoking default ingress rule for Security Group in testRemoveCycle (%s): %w", aws.ToString(primary.GroupId), err)
				}
			}

			if sg.IpPermissionsEgress != nil && len(sg.IpPermissionsEgress) > 0 {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}

				if _, err = conn.RevokeSecurityGroupEgress(ctx, req); err != nil {
					return fmt.Errorf("Error revoking default egress rule for Security Group in testRemoveCycle (%s): %w", aws.ToString(sg.GroupId), err)
				}
			}
		}
		return nil
	}
}

func testAccCheckSecurityGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_security_group" {
				continue
			}

			_, err := tfec2.FindSecurityGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.EC2, create.ErrActionCheckingDestroyed, "Security Group", rs.Primary.ID, err)
			}

			return fmt.Errorf("VPC Security Group (%s) still exists.", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecurityGroupExists(ctx context.Context, n string, v *awstypes.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Security Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSecurityGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.EC2, create.ErrActionCheckingExistence, "Security Group", rs.Primary.ID, err)
		}

		*v = *output

		return nil
	}
}

func testAccCheckSecurityGroupRuleLimit(n string, v *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Quotas ID is set")
		}

		limit, err := strconv.Atoi(rs.Primary.Attributes[names.AttrValue])
		if err != nil {
			return fmt.Errorf("converting value to int: %s", err)
		}

		*v = limit

		return nil
	}
}

func testAccCheckSecurityGroupRuleCount(ctx context.Context, group *awstypes.SecurityGroup, expectedIngressCount, expectedEgressCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := aws.ToString(group.GroupId)
		return testSecurityGroupRuleCount(ctx, id, expectedIngressCount, expectedEgressCount)
	}
}

func testSecurityGroupRuleCount(ctx context.Context, id string, expectedIngressCount, expectedEgressCount int) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	group, err := tfec2.FindSecurityGroupByID(ctx, conn, id)
	if tfresource.NotFound(err) {
		return fmt.Errorf("Security Group (%s) not found: %w", id, err)
	}
	if err != nil {
		return create.Error(names.EC2, create.ErrActionChecking, "Security Group", id, err)
	}

	if actual := len(group.IpPermissions); actual != expectedIngressCount {
		return fmt.Errorf("Security group ingress rule count %d does not match %d", actual, expectedIngressCount)
	}

	if actual := len(group.IpPermissionsEgress); actual != expectedEgressCount {
		return fmt.Errorf("Security group egress rule count %d does not match %d", actual, expectedEgressCount)
	}

	return nil
}

func testAccVPCSecurityGroupConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_nameGenerated(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_namePrefix(rName, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name_prefix = %[2]q
  vpc_id      = aws_vpc.test.id
}
`, rName, namePrefix)
}

func testAccVPCSecurityGroupConfig_noVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

data "aws_vpc" "default" {
  default = true
}
`, rName)
}

func testAccVPCSecurityGroupConfig_defaultVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = data.aws_vpc.default.id
}

data "aws_vpc" "default" {
  default = true
}
`, rName)
}

func testAccVPCSecurityGroupConfig_getLimit() string {
	return `
data "aws_servicequotas_service_quota" "test" {
  quota_name   = "Inbound or outbound rules per security group"
  service_code = "vpc"
}
`
}

func testAccVPCSecurityGroupConfig_ruleLimit(rName string, egressStartIndex, egressRulesCount int) string {
	var egressRules strings.Builder
	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		fmt.Fprintf(&egressRules, `
  egress {
    protocol    = "tcp"
    from_port   = "${80 + %[1]d}"
    to_port     = "${80 + %[1]d}"
    cidr_blocks = ["${cidrhost("10.1.0.0/16", %[1]d)}/32"]
  }
`, i)
	}
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  # egress rules to exhaust the limit
  %[2]s
}
`, rName, egressRules.String())
}

func testAccVPCSecurityGroupConfig_cidrBlockRuleLimit(rName string, egressStartIndex, egressRulesCount int) string {
	var cidrBlocks strings.Builder
	for i := egressStartIndex; i < egressRulesCount+egressStartIndex; i++ {
		fmt.Fprintf(&cidrBlocks, `
		"${cidrhost("10.1.0.0/16", %[1]d)}/32",
`, i)
	}

	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  egress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    # cidr_blocks to exhaust the limit
    cidr_blocks = [
		%[2]s
    ]
  }
}
`, rName, cidrBlocks.String())
}

func testAccVPCSecurityGroupConfig_emptyRuleDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = ""
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ipv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol         = "6"
    from_port        = 80
    to_port          = 8000
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    protocol         = "tcp"
    from_port        = 80
    to_port          = 8000
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_revokeBaseRemoved(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_revokeBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "primary" {
  name   = "%[1]s-primary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  timeouts {
    delete = "2m"
  }
}

resource "aws_security_group" "secondary" {
  name   = "%[1]s-secondary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  timeouts {
    delete = "2m"
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_revokeFalse(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "primary" {
  name   = "%[1]s-primary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  revoke_rules_on_delete = false
}

resource "aws_security_group" "secondary" {
  name   = "%[1]s-secondary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  revoke_rules_on_delete = false
}
`, rName)
}

func testAccVPCSecurityGroupConfig_revokeTrue(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "primary" {
  name   = "%[1]s-primary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  revoke_rules_on_delete = true
}

resource "aws_security_group" "secondary" {
  name   = "%[1]s-secondary"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  revoke_rules_on_delete = true
}
`, rName)
}

func testAccVPCSecurityGroupConfig_changed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 9000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ruleDescription(rName, egressDescription, ingressDescription string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = %[2]q
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
    description = %[3]q
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, ingressDescription, egressDescription)
}

func testAccVPCSecurityGroupConfig_self(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = "tcp"
    from_port = 80
    to_port   = 8000
    self      = true
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_vpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_vpcNegativeOneIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_vpcProtocolNumberIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "50"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_multiIngress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 800
    to_port     = 800
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 8000
    security_groups = [aws_security_group.test1.id]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_vpcAllEgress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    protocol    = "all"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_defaultEgress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_driftComplex(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["206.0.0.0/8"]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    security_groups = [aws_security_group.test2.id]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["206.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol        = "tcp"
    from_port       = 22
    to_port         = 22
    security_groups = [aws_security_group.test2.id]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

const testAccVPCSecurityGroupConfig_invalidIngressCIDR = `
resource "aws_security_group" "test" {
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}
`

const testAccVPCSecurityGroupConfig_invalidEgressCIDR = `
resource "aws_security_group" "test" {
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["1.2.3.4/33"]
  }
}
`

const testAccVPCSecurityGroupConfig_invalidIPv6IngressCIDR = `
resource "aws_security_group" "test" {
  ingress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}
`

const testAccVPCSecurityGroupConfig_invalidIPv6EgressCIDR = `
resource "aws_security_group" "test" {
  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/244"]
  }
}
`

func testAccVPCSecurityGroupConfig_combinedCIDRAndGroups(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test3" {
  name   = "%[1]s-3"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test4" {
  name   = "%[1]s-4"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16", "10.1.0.0/16", "10.7.0.0/16"]

    security_groups = [
      aws_security_group.test2.id,
      aws_security_group.test3.id,
      aws_security_group.test4.id,
    ]
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ingressWithCIDRAndSGs(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol  = "tcp"
    from_port = "22"
    to_port   = "22"

    cidr_blocks = [
      "192.168.0.1/32",
    ]
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 8000
    cidr_blocks     = ["10.0.0.0/8"]
    security_groups = [aws_security_group.test2.id]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`, rName)
}

// fails to apply in one pass with the error "diffs didn't match during apply"
// GH-2027
func testAccVPCSecurityGroupConfig_failWithDiffMismatch(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test3" {
  vpc_id = aws_vpc.main.id
  name   = "%[1]s-3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  vpc_id = aws_vpc.main.id
  name   = "%[1]s-2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test1" {
  vpc_id = aws_vpc.main.id
  name   = "%[1]s-1"

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.test2.id]
  }

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.test3.id]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_allowAll(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "allow_all-1" {
  type        = "ingress"
  from_port   = 0
  to_port     = 65535
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]

  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_all-2" {
  type      = "ingress"
  from_port = 65534
  to_port   = 65535
  protocol  = "tcp"

  self              = true
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_sourceSecurityGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test3" {
  name   = "%[1]s-3"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "allow_test2" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test.id
  security_group_id        = aws_security_group.test2.id
}

resource "aws_security_group_rule" "allow_test3" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test.id
  security_group_id        = aws_security_group.test3.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ipRangeAndSecurityGroupWithSameRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "allow_security_group" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test2.id
  security_group_id        = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks       = ["10.0.0.0/32"]
  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks  = ["::/0"]
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ipRangesWithSameRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "allow_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  cidr_blocks       = ["10.0.0.0/32"]
  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "allow_ipv6_cidr_block" {
  type      = "ingress"
  from_port = 0
  to_port   = 0
  protocol  = "tcp"

  ipv6_cidr_blocks  = ["::/0"]
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ipv4andIPv6Egress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_prefixListEgress(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]

  tags = {
    Name = %[1]q
  }

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  egress {
    protocol        = "-1"
    from_port       = 0
    to_port         = 0
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_prefixListIngress(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]

  tags = {
    Name = %[1]q
  }

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol        = "-1"
    from_port       = 0
    to_port         = 0
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ruleGathering(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "source1" {
  name   = "%[1]s-source1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "source2" {
  name   = "%[1]s-source2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.0.0.0/24", "10.0.1.0/24"]
    self        = true
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24"]
    description = "ingress from 10.0.0.0/16"
  }

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["192.168.0.0/16"]
    description = "ingress from 192.168.0.0/16"
  }

  ingress {
    protocol         = "tcp"
    from_port        = 80
    to_port          = 80
    ipv6_cidr_blocks = ["::/0"]
    description      = "ingress from all ipv6"
  }

  ingress {
    protocol        = "tcp"
    from_port       = 80
    to_port         = 80
    security_groups = [aws_security_group.source1.id, aws_security_group.source2.id]
    description     = "ingress from other security groups"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "egress for all ipv4"
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    ipv6_cidr_blocks = ["::/0"]
    description      = "egress for all ipv6"
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    prefix_list_ids = [aws_vpc_endpoint.test.prefix_list_id]
    description     = "egress for vpc endpoints"
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_rulesDropOnErrorInit(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test_ref0" {
  name   = "%[1]s-ref0"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test_ref1" {
  name   = "%[1]s-ref1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    security_groups = [
      aws_security_group.test_ref0.id,
      aws_security_group.test_ref1.id,
    ]
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_rulesDropOnErrorAddBadRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test_ref0" {
  name   = "%[1]s-ref0"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test_ref1" {
  name   = "%[1]s-ref1"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  ingress {
    protocol  = "tcp"
    from_port = "80"
    to_port   = "80"
    security_groups = [
      aws_security_group.test_ref0.id,
      aws_security_group.test_ref1.id,
      "sg-malformed", # non-existent rule to trigger API error
    ]
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_egressModeBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "tcp"
    to_port     = 0
  }

  egress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "udp"
    to_port     = 0
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_egressModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_egressModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  egress = []

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ingressModeBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "tcp"
    to_port     = 0
  }

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = "udp"
    to_port     = 0
  }
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ingressModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccVPCSecurityGroupConfig_ingressModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }

  ingress = []

  vpc_id = aws_vpc.test.id
}
`, rName)
}

// testAccVPCSecurityGroupConfig_emrLinkedRulesDestroy is very involved but captures
// a problem seen in EMR and other contexts.
func testAccVPCSecurityGroupConfig_emrLinkedRulesDestroy(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
# VPC
resource "aws_vpc" "main" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = %[1]q
  }
}

# subnets
resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.1.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "allow_ssh" {
  name        = "%[1]s-ssh"
  description = "ssh"
  vpc_id      = aws_vpc.main.id

  tags = {
    Name = "%[1]s-ssh"
  }
}

# internet gateway
resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

# elastic ip for NAT gateway
resource "aws_eip" "nat" {
  domain = "vpc"
  tags = {
    Name = %[1]q
  }
}

# NAT gateway
resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.private.id

  tags = {
    Name = %[1]q
  }

  # To ensure proper ordering, it is recommended to add an explicit dependency
  # on the Internet Gateway for the VPC.
  depends_on = [aws_internet_gateway.gw]
}

# route tables
# add internet gateway
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.gw.id
  }

  tags = {
    Name = %[1]q
  }
}

# route table for nat
resource "aws_route_table" "nat" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat.id
  }

  tags = {
    Name = %[1]q
  }
}

# associate nat route table with subnet
resource "aws_route_table_association" "nat" {
  subnet_id      = aws_subnet.private.id
  route_table_id = aws_route_table.nat.id
}

resource "aws_security_group" "allow_access" {
  name                   = "%[1]s-allow-access"
  description            = "Allow inbound traffic"
  vpc_id                 = aws_vpc.main.id
  revoke_rules_on_delete = true

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }

  tags = {
    name = "%[1]s-allow-access"
  }
}

resource "aws_security_group" "service_access" {
  name                   = "%[1]s-service-access"
  description            = "Allow inbound traffic"
  vpc_id                 = aws_vpc.main.id
  revoke_rules_on_delete = true

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  ingress {
    from_port       = 8443
    to_port         = 8443
    protocol        = "tcp"
    cidr_blocks     = [aws_vpc.main.cidr_block]
    security_groups = [aws_security_group.allow_access.id]
  }

  ingress {
    from_port       = 9443
    to_port         = 9443
    protocol        = "tcp"
    cidr_blocks     = [aws_vpc.main.cidr_block]
    security_groups = [aws_security_group.allow_access.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }

  tags = {
    name = "%[1]s-service-access"
  }
}

# IAM role for EMR Service
resource "aws_iam_role" "iam_emr_service_role" {
  name = "%[1]s-service-role"

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "iam_emr_service_policy" {
  name = "%[1]s-service-policy"
  role = aws_iam_role.iam_emr_service_role.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOF
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "iam_emr_profile_role" {
  name = "%[1]s-profile-role"

  assume_role_policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "emr_profile" {
  name = "%[1]s-profile"
  role = aws_iam_role.iam_emr_profile_role.name
}

resource "aws_iam_role_policy" "iam_emr_profile_policy" {
  name = "%[1]s-profile-policy"
  role = aws_iam_role.iam_emr_profile_role.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOF
}

resource "aws_emr_cluster" "cluster" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  applications  = ["Spark"]

  additional_info = <<EOF
{
  "instanceAwsClientConfiguration": {
    "proxyPort": 8099,
    "proxyHost": "myproxy.example.com"
  }
}
EOF

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  ec2_attributes {
    subnet_id                         = aws_subnet.private.id
    instance_profile                  = aws_iam_instance_profile.emr_profile.arn
    emr_managed_master_security_group = aws_security_group.allow_access.id
    emr_managed_slave_security_group  = aws_security_group.allow_access.id
    additional_master_security_groups = aws_security_group.allow_ssh.id
    additional_slave_security_groups  = aws_security_group.allow_ssh.id
    service_access_security_group     = aws_security_group.service_access.id
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_type  = "c4.large"
    instance_count = 1

    ebs_config {
      size                 = "40"
      type                 = "gp2"
      volumes_per_instance = 1
    }
  }

  ebs_root_volume_size = 100

  tags = {
    role = "rolename"
    env  = "env"
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations_json = <<EOF
  [
    {
      "Classification": "hadoop-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    },
    {
      "Classification": "spark-env",
      "Configurations": [
        {
          "Classification": "export",
          "Properties": {
            "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
          }
        }
      ],
      "Properties": {}
    }
  ]
EOF

  service_role = aws_iam_role.iam_emr_service_role.arn

  depends_on = [
    aws_route_table_association.nat,
    aws_iam_role_policy.iam_emr_service_policy,
    aws_iam_role_policy.iam_emr_profile_policy
  ]
}
`, rName))
}
