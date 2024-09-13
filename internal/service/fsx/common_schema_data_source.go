// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newSnapshotFilterList(s *schema.Set) []awstypes.SnapshotFilter {
	if s == nil {
		return []awstypes.SnapshotFilter{}
	}

	return tfslices.ApplyToAll(s.List(), func(tfList interface{}) awstypes.SnapshotFilter {
		tfMap := tfList.(map[string]interface{})
		return awstypes.SnapshotFilter{
			Name:   awstypes.SnapshotFilterName(tfMap[names.AttrName].(string)),
			Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
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
				names.AttrValues: {
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

func newStorageVirtualMachineFilterList(s *schema.Set) []awstypes.StorageVirtualMachineFilter {
	if s == nil {
		return []awstypes.StorageVirtualMachineFilter{}
	}

	return tfslices.ApplyToAll(s.List(), func(tfList interface{}) awstypes.StorageVirtualMachineFilter {
		tfMap := tfList.(map[string]interface{})
		return awstypes.StorageVirtualMachineFilter{
			Name:   awstypes.StorageVirtualMachineFilterName(tfMap[names.AttrName].(string)),
			Values: flex.ExpandStringValueList(tfMap[names.AttrValues].([]interface{})),
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
				names.AttrValues: {
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
