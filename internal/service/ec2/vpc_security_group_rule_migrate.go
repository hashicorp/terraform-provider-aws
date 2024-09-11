// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func securityGroupRuleMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Security Group State v0; migrating to v1")
		return migrateSGRuleStateV0toV1(is)
	case 1:
		log.Println("[INFO] Found AWS Security Group State v1; migrating to v2")
		// migrating to version 2 of the schema is the same as 0->1, since the
		// method signature has changed now and will use the security group id in
		// the hash
		return migrateSGRuleStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateSGRuleStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	perm, err := migrateExpandIPPerm(is.Attributes)

	if err != nil {
		return nil, fmt.Errorf("making new IP Permission in Security Group migration")
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	newID := securityGroupRuleCreateID(is.Attributes["security_group_id"], is.Attributes[names.AttrType], perm)
	is.Attributes[names.AttrID] = newID
	is.ID = newID
	log.Printf("[DEBUG] Attributes after migration: %#v, new id: %s", is.Attributes, newID)
	return is, nil
}

func migrateExpandIPPerm(attrs map[string]string) (*awstypes.IpPermission, error) {
	var perm awstypes.IpPermission
	tp, err := strconv.Atoi(attrs["to_port"])
	if err != nil {
		return nil, fmt.Errorf("converting to_port in Security Group migration")
	}

	fp, err := strconv.Atoi(attrs["from_port"])
	if err != nil {
		return nil, fmt.Errorf("converting from_port in Security Group migration")
	}

	perm.ToPort = aws.Int32(int32(tp))
	perm.FromPort = aws.Int32(int32(fp))
	perm.IpProtocol = aws.String(attrs[names.AttrProtocol])

	groups := make(map[string]bool)
	if attrs["self"] == "true" {
		groups[attrs["security_group_id"]] = true
	}

	if attrs["source_security_group_id"] != "" {
		groups[attrs["source_security_group_id"]] = true
	}

	if len(groups) > 0 {
		perm.UserIdGroupPairs = make([]awstypes.UserIdGroupPair, len(groups))
		// build string list of group name/ids
		var gl []string
		for k := range groups {
			gl = append(gl, k)
		}

		for i, name := range gl {
			perm.UserIdGroupPairs[i] = awstypes.UserIdGroupPair{
				GroupId: aws.String(name),
			}
		}
	}

	var cb []string
	for k, v := range attrs {
		if k != "cidr_blocks.#" && strings.HasPrefix(k, "cidr_blocks") {
			cb = append(cb, v)
		}
	}
	if len(cb) > 0 {
		perm.IpRanges = make([]awstypes.IpRange, len(cb))
		for i, v := range cb {
			perm.IpRanges[i] = awstypes.IpRange{CidrIp: aws.String(v)}
		}
	}

	return &perm, nil
}
