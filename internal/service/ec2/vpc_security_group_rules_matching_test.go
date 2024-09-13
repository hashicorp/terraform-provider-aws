// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testing rulesForGroupPermissions
func TestRulesMixedMatching(t *testing.T) {
	t.Parallel()

	cases := []struct {
		local  []interface{}
		remote []map[string]interface{}
		saves  []map[string]interface{}
	}{
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []interface{}{"172.8.0.0/16", "10.0.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "10.0.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "10.0.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "10.0.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16"},
				},
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"192.168.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16"},
				},
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"192.168.0.0/16"},
				},
			},
		},
		{
			local: []interface{}{},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.0.0.0/16"},
				},
			},
		},
		// test lower/ uppercase handling
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "TCP",
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
				},
			},
		},
		// local and remote differ
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"10.0.0.0/16"},
				},
			},
			// Because this is the remote rule being saved, we need to check for int32
			// encoding. We could convert this code, but ultimately Terraform doesn't
			// care it's for the reflect.DeepEqual in this test
			saves: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"10.0.0.0/16"},
				},
			},
		},
		// local with more rules and the remote (the remote should then be saved)
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16", "10.8.0.0/16", "192.168.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16"},
				},
			},
		},
		// 3 local rules
		// this should trigger a diff (not shown)
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16"},
				},
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"10.8.0.0/16"},
				},
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"192.168.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16"},
				},
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"192.168.0.0/16"},
				},
			},
		},
		// a local rule with 2 cidrs, remote has 4 cidrs, should be saved to match
		// the local but also an extra rule found
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16", "10.8.0.0/16"},
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16", "10.8.0.0/16", "206.8.0.0/16"},
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "10.8.0.0/16"},
				},
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"192.168.0.0/16", "206.8.0.0/16"},
				},
			},
		},
		// testing some SGS
		{
			local: []interface{}{},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(22),
					"to_port":                int32(22),
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876"}),
				},
			},
			saves: []map[string]interface{}{
				{
					// we're saving the remote, so it will be int32 encoded
					"from_port":              int32(22),
					"to_port":                int32(22),
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876"}),
				},
			},
		},
		// two local blocks that match a single remote group, but are saved as two
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              22,
					"to_port":                22,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876"}),
				},
				map[string]interface{}{
					"from_port":              22,
					"to_port":                22,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					names.AttrSecurityGroups: schema.NewSet(
						schema.HashString,
						[]interface{}{
							"sg-9876",
							"sg-4444",
						},
					),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        22,
					"to_port":          22,
					names.AttrProtocol: "tcp",
					names.AttrSecurityGroups: schema.NewSet(
						schema.HashString,
						[]interface{}{
							"sg-9876",
						},
					),
				},
				{
					"from_port":        22,
					"to_port":          22,
					names.AttrProtocol: "tcp",
					names.AttrSecurityGroups: schema.NewSet(
						schema.HashString,
						[]interface{}{
							"sg-4444",
						},
					),
				},
			},
		},
		// test self with other rules
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              22,
					"to_port":                22,
					names.AttrProtocol:       "tcp",
					"self":                   true,
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
		},
		// test self
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        22,
					"to_port":          22,
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":              22,
					"to_port":                22,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(22),
					"to_port":          int32(22),
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
		},
		// mix of sgs and cidrs
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16", "10.8.0.0/16", "192.168.0.0/16"},
				},
				map[string]interface{}{
					"from_port":              80,
					"to_port":                8000,
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "192.168.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        int32(80),
					"to_port":          int32(8000),
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []string{"172.8.0.0/16", "192.168.0.0/16"},
				},
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
		},
		{
			local: []interface{}{
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"cidr_blocks":      []interface{}{"172.8.0.0/16", "10.8.0.0/16", "192.168.0.0/16"},
				},
				map[string]interface{}{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"self":             true,
				},
			},
			remote: []map[string]interface{}{
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "192.168.0.0/16"},
					"self":                   true,
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
			saves: []map[string]interface{}{
				{
					"from_port":        80,
					"to_port":          8000,
					names.AttrProtocol: "tcp",
					"self":             true,
				},
				{
					"from_port":              int32(80),
					"to_port":                int32(8000),
					names.AttrProtocol:       "tcp",
					"cidr_blocks":            []string{"172.8.0.0/16", "192.168.0.0/16"},
					names.AttrSecurityGroups: schema.NewSet(schema.HashString, []interface{}{"sg-9876", "sg-4444"}),
				},
			},
		},
	}
	for i, c := range cases {
		saves := tfec2.MatchRules("ingress", c.local, c.remote)
		log.Printf("\n======\n\nSaves:\n%#v\n\nCS Saves:\n%#v\n\n======\n", saves, c.saves)
		log.Printf("\n\tTest %d:\n", i)

		if len(saves) != len(c.saves) {
			t.Fatalf("Expected %d saves, got %d", len(c.saves), len(saves))
		}

		shouldFind := len(c.saves)
		var found int
		for _, s := range saves {
			for _, cs := range c.saves {
				// deep equal cannot compare schema.Set's directly
				// make sure we're not failing the reflect b/c of ports/type
				for _, attr := range []string{"to_port", "from_port", names.AttrType} {
					if s[attr] != cs[attr] {
						continue
					}
				}

				var numExpectedCidrs, numExpectedSGs, numRemoteCidrs, numRemoteSGs int
				// var matchingCidrs []string
				// var matchingSGs []string

				var cidrsMatch, sGsMatch bool

				if _, ok := s["cidr_blocks"]; ok {
					switch s["cidr_blocks"].(type) {
					case []string:
						numExpectedCidrs = len(s["cidr_blocks"].([]string))
					default:
						numExpectedCidrs = len(s["cidr_blocks"].([]interface{}))
					}
				}
				if _, ok := s[names.AttrSecurityGroups]; ok {
					numExpectedSGs = len(s[names.AttrSecurityGroups].(*schema.Set).List())
				}

				if _, ok := cs["cidr_blocks"]; ok {
					numRemoteCidrs = len(cs["cidr_blocks"].([]string))
				}

				if _, ok := cs[names.AttrSecurityGroups]; ok {
					numRemoteSGs = len(cs[names.AttrSecurityGroups].(*schema.Set).List())
				}

				// skip early
				if numExpectedSGs != numRemoteSGs {
					log.Printf("\n\ncontinuing on numRemoteSGs \n\n")
					continue
				}
				if numExpectedCidrs != numRemoteCidrs {
					log.Printf("\n\ncontinuing numRemoteCidrs\n\n")
					continue
				}

				if numExpectedCidrs == 0 {
					cidrsMatch = true
				}
				if numExpectedSGs == 0 {
					sGsMatch = true
				}

				// convert save cidrs to set
				var lcs []interface{}
				if _, ok := s["cidr_blocks"]; ok {
					switch s["cidr_blocks"].(type) {
					case []string:
						for _, c := range s["cidr_blocks"].([]string) {
							lcs = append(lcs, c)
						}
					default:
						lcs = append(lcs, s["cidr_blocks"].([]interface{})...)
					}
				}
				savesCidrs := schema.NewSet(schema.HashString, lcs)

				// convert cs cidrs to set
				var cslcs []interface{}
				if _, ok := cs["cidr_blocks"]; ok {
					for _, c := range cs["cidr_blocks"].([]string) {
						cslcs = append(cslcs, c)
					}
				}
				csCidrs := schema.NewSet(schema.HashString, cslcs)

				if csCidrs.Equal(savesCidrs) {
					log.Printf("\nmatched cidrs")
					cidrsMatch = true
				}

				if rawS, ok := s[names.AttrSecurityGroups]; ok {
					outSet := rawS.(*schema.Set)
					if rawL, ok := cs[names.AttrSecurityGroups]; ok {
						localSet := rawL.(*schema.Set)
						if outSet.Equal(localSet) {
							log.Printf("\nmatched sgs")
							sGsMatch = true
						}
					}
				}

				var lSelf bool
				var rSelf bool
				if _, ok := s["self"]; ok {
					lSelf = s["self"].(bool)
				}
				if _, ok := cs["self"]; ok {
					rSelf = cs["self"].(bool)
				}

				if (sGsMatch && cidrsMatch) && (lSelf == rSelf) {
					found++
				}
			}
		}

		if found != shouldFind {
			t.Fatalf("Bad sg rule matches (%d / %d)", found, shouldFind)
		}
	}
}
