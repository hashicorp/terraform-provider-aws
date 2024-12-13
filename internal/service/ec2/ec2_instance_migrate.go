// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func instanceMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Instance State v0; migrating to v1")
		return migrateInstanceStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateInstanceStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// Delete old count
	delete(is.Attributes, "block_device.#")

	oldBds, err := readV0BlockDevices(is)
	if err != nil {
		return is, err
	}
	// seed count fields for new types
	is.Attributes["ebs_block_device.#"] = "0"
	is.Attributes["ephemeral_block_device.#"] = "0"
	// depending on if state was v0.3.7 or an earlier version, it might have
	// root_block_device defined already
	if _, ok := is.Attributes["root_block_device.#"]; !ok {
		is.Attributes["root_block_device.#"] = "0"
	}
	for _, oldBd := range oldBds {
		writeV1BlockDevice(is, oldBd)
	}
	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}

func readV0BlockDevices(is *terraform.InstanceState) (map[string]map[string]string, error) {
	oldBds := make(map[string]map[string]string)
	for k, v := range is.Attributes {
		if !strings.HasPrefix(k, "block_device.") {
			continue
		}
		path := strings.Split(k, ".")
		if len(path) != 3 {
			return oldBds, fmt.Errorf("Found unexpected block_device field: %#v", k)
		}
		hashcode, attribute := path[1], path[2]
		oldBd, ok := oldBds[hashcode]
		if !ok {
			oldBd = make(map[string]string)
			oldBds[hashcode] = oldBd
		}
		oldBd[attribute] = v
		delete(is.Attributes, k)
	}
	return oldBds, nil
}

func writeV1BlockDevice(is *terraform.InstanceState, oldBd map[string]string) {
	code := create.StringHashcode(oldBd[names.AttrDeviceName])
	bdType := "ebs_block_device"
	if vn, ok := oldBd[names.AttrVirtualName]; ok && strings.HasPrefix(vn, "ephemeral") {
		bdType = "ephemeral_block_device"
	} else if dn, ok := oldBd[names.AttrDeviceName]; ok && dn == "/dev/sda1" {
		bdType = "root_block_device"
	}

	switch bdType {
	case "ebs_block_device":
		delete(oldBd, names.AttrVirtualName)
	case "root_block_device":
		delete(oldBd, names.AttrVirtualName)
		delete(oldBd, names.AttrEncrypted)
		delete(oldBd, names.AttrSnapshotID)
	case "ephemeral_block_device":
		delete(oldBd, names.AttrDeleteOnTermination)
		delete(oldBd, names.AttrEncrypted)
		delete(oldBd, names.AttrIOPS)
		delete(oldBd, names.AttrVolumeSize)
		delete(oldBd, names.AttrVolumeType)
	}
	for attr, val := range oldBd {
		attrKey := fmt.Sprintf("%s.%d.%s", bdType, code, attr)
		is.Attributes[attrKey] = val
	}

	countAttr := fmt.Sprintf("%s.#", bdType)
	count, _ := strconv.Atoi(is.Attributes[countAttr])
	is.Attributes[countAttr] = strconv.Itoa(count + 1)
}
