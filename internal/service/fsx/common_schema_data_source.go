// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newSnapshotFilterList(s *schema.Set) []*fsx.SnapshotFilter {
	if s == nil {
		return []*fsx.SnapshotFilter{}
	}

	return tfslices.ApplyToAll(s.List(), func(tfList interface{}) *fsx.SnapshotFilter {
		tfMap := tfList.(map[string]interface{})
		return &fsx.SnapshotFilter{
			Name:   aws.String(tfMap[names.AttrName].(string)),
			Values: flex.ExpandStringList(tfMap["values"].([]interface{})),
		}
	})
}

func snapshotFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"values": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func newStorageVirtualMachineFilterList(s *schema.Set) []*fsx.StorageVirtualMachineFilter {
	if s == nil {
		return []*fsx.StorageVirtualMachineFilter{}
	}

	return tfslices.ApplyToAll(s.List(), func(tfList interface{}) *fsx.StorageVirtualMachineFilter {
		tfMap := tfList.(map[string]interface{})
		return &fsx.StorageVirtualMachineFilter{
			Name:   aws.String(tfMap[names.AttrName].(string)),
			Values: flex.ExpandStringList(tfMap["values"].([]interface{})),
		}
	})
}

func storageVirtualMachineFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				"values": {
					Type:     schema.TypeList,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}
