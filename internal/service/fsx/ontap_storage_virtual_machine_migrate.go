// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"log"
	"strings"

	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceONTAPStorageVirtualMachineV0() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active_directory_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"netbios_name": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							ValidateFunc: validation.StringLenBetween(1, 15),
						},
						"self_managed_active_directory_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_ips": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										MaxItems: 3,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsIPAddress,
										},
									},
									names.AttrDomainName: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"file_system_administrators_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"organizational_unit_distinguidshed_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ValidateFunc:  validation.StringLenBetween(1, 2000),
										Deprecated:    "use 'organizational_unit_distinguished_name' instead",
										ConflictsWith: []string{"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name"},
									},
									"organizational_unit_distinguished_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ValidateFunc:  validation.StringLenBetween(1, 2000),
										ConflictsWith: []string{"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name"},
									},
									names.AttrPassword: {
										Type:         schema.TypeString,
										Sensitive:    true,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									names.AttrUsername: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
			},
			names.AttrEndpoints: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iscsi": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"management": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"nfs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"smb": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDNSName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIPAddresses: {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			names.AttrFileSystemID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(11, 21),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"root_volume_security_style": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.StorageVirtualMachineRootVolumeSecurityStyle](),
			},
			"subtype": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"svm_admin_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 50),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceONTAPStorageVirtualMachineStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	log.Printf("[DEBUG] Attributes before migration: %#v", rawState)

	rawState["active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name"] = rawState["active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name"]
	delete(rawState, "active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name")

	log.Printf("[DEBUG] Attributes after migration: %#v", rawState)
	return rawState, nil
}
