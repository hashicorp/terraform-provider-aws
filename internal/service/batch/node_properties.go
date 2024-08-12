// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/batch"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
)

type nodeProperties struct {
	MainNode            *int64
	NodeRangeProperties []*nodeRangeProperty
	NumNodes            *int64
}

type nodeRangeProperty struct {
	Container   *containerProperties
	TargetNodes *string
}

func (np *nodeProperties) Reduce() error {
	// Deal with Environment objects which may be re-ordered in the API
	for _, node := range np.NodeRangeProperties {
		cp := node.Container
		if err := cp.Reduce(); err != nil {
			return err
		}
	}

	return nil
}

// EquivalentNodePropertiesJSON determines equality between two Batch NodeProperties JSON strings
func EquivalentNodePropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var np1, np2 nodeProperties

	if err := json.Unmarshal([]byte(str1), &np1); err != nil {
		return false, err
	}

	if err := np1.Reduce(); err != nil {
		return false, err
	}

	canonicalJson1, err := jsonutil.BuildJSON(np1)

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(str2), &np2); err != nil {
		return false, err
	}

	if err := np2.Reduce(); err != nil {
		return false, err
	}

	canonicalJson2, err := jsonutil.BuildJSON(np2)

	if err != nil {
		return false, err
	}

	equal := bytes.Equal(canonicalJson1, canonicalJson2)

	if !equal {
		log.Printf("[DEBUG] Canonical Batch Node Properties JSON are not equal.\nFirst: %s\nSecond: %s\n", canonicalJson1, canonicalJson2)
	}

	return equal, nil
}

func expandJobNodeProperties(rawProps string) (*batch.NodeProperties, error) {
	var props *batch.NodeProperties

	err := json.Unmarshal([]byte(rawProps), &props)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %s", err)
	}

	return props, nil
}

// Convert batch.NodeProperties object into its JSON representation
func flattenNodeProperties(nodeProperties *batch.NodeProperties) (string, error) {
	b, err := jsonutil.BuildJSON(nodeProperties)

	if err != nil {
		return "", err
	}

	return string(b), nil
}
