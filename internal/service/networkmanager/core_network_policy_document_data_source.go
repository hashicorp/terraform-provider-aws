package networkmanager

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func DataSourceCoreNetworkPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		Read: dataSourceCoreNetworkPolicyDocumentRead,

		Schema: map[string]*schema.Schema{

			// "core_network_configuration": {
			// 	Type:     schema.TypeList,
			// 	Optional: true,
			// 	MaxItems: 1,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"asn_ranges": {
			// 				Type:          schema.TypeList,
			// 				Optional:      true,
			// 				Elem:          &schema.Schema{Type: schema.TypeString},
			// 			},
			// 		},
			// 	},
			// 	DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			// },
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segments": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_filter": setOfString,
						"deny_filter":  setOfString,
						"name": {
							Type:     schema.TypeString,
							Required: true,
							// a-z, 0-9
							// ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
						},
						"edge_locations": setOfString,
						"isolate_attachments": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"require_attachment_acceptance": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2021.12",
				ValidateFunc: validation.StringInSlice([]string{
					"2021.12",
				}, false),
			},
		},
	}
}

func dataSourceCoreNetworkPolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	mergedDoc := &CoreNetworkPolicyDoc{}

	doc := &CoreNetworkPolicyDoc{
		Version: d.Get("version").(string),
	}

	if cfgSgmts, hasCfgSgmts := d.GetOk("segments"); hasCfgSgmts {
		var cfgSgmtIntf = cfgSgmts.([]interface{})
		Sgmts := make([]*CoreNetworkPolicySegment, len(cfgSgmtIntf))
		nameMap := make(map[string]struct{})

		for i, sgmtI := range cfgSgmtIntf {
			cfgSgmt := sgmtI.(map[string]interface{})
			sgmt := &CoreNetworkPolicySegment{}

			if name, ok := cfgSgmt["name"]; ok {
				if _, ok := nameMap[name.(string)]; ok {
					return fmt.Errorf("duplicate Name (%s). Remove the Name or ensure the Name is unique.", name.(string))
				}
				sgmt.Name = name.(string)
				if len(sgmt.Name) > 0 {
					nameMap[sgmt.Name] = struct{}{}
				}
			}
			if actions := cfgSgmt["allow_filter"].(*schema.Set).List(); len(actions) > 0 {
				sgmt.AllowFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
			}
			if actions := cfgSgmt["deny_filter"].(*schema.Set).List(); len(actions) > 0 {
				sgmt.DenyFilter = CoreNetworkPolicyDecodeConfigStringList(actions)
			}
			if b, ok := cfgSgmt["require_attachment_acceptance"]; ok {
				sgmt.RequireAttachmentAcceptance = b.(bool)
			}
			if b, ok := cfgSgmt["isolate_attachments"]; ok {
				sgmt.IsolateAttachments = b.(bool)
			}

			// if resources := cfgSgmt["resources"].(*schema.Set).List(); len(resources) > 0 {
			// 	var err error
			// 	sgmt.Resources, err = dataSourcePolicyDocumentReplaceVarsInList(
			// 		iamPolicyDecodeConfigStringList(resources), doc.Version,
			// 	)
			// 	if err != nil {
			// 		return fmt.Errorf("error reading resources: %w", err)
			// 	}
			// }
			// if notResources := cfgSgmt["not_resources"].(*schema.Set).List(); len(notResources) > 0 {
			// 	var err error
			// 	sgmt.NotResources, err = dataSourcePolicyDocumentReplaceVarsInList(
			// 		iamPolicyDecodeConfigStringList(notResources), doc.Version,
			// 	)
			// 	if err != nil {
			// 		return fmt.Errorf("error reading not_resources: %w", err)
			// 	}
			// }

			// if principals := cfgSgmt["principals"].(*schema.Set).List(); len(principals) > 0 {
			// 	var err error
			// 	sgmt.Principals, err = dataSourcePolicyDocumentMakePrincipals(principals, doc.Version)
			// 	if err != nil {
			// 		return fmt.Errorf("error reading principals: %w", err)
			// 	}
			// }

			// if notPrincipals := cfgSgmt["not_principals"].(*schema.Set).List(); len(notPrincipals) > 0 {
			// 	var err error
			// 	sgmt.NotPrincipals, err = dataSourcePolicyDocumentMakePrincipals(notPrincipals, doc.Version)
			// 	if err != nil {
			// 		return fmt.Errorf("error reading not_principals: %w", err)
			// 	}
			// }

			// if conditions := cfgSgmt["condition"].(*schema.Set).List(); len(conditions) > 0 {
			// 	var err error
			// 	sgmt.Conditions, err = dataSourcePolicyDocumentMakeConditions(conditions, doc.Version)
			// 	if err != nil {
			// 		return fmt.Errorf("error reading condition: %w", err)
			// 	}
			// }

			Sgmts[i] = sgmt
		}

		doc.Segments = Sgmts

	}
	mergedDoc.Merge(doc)
	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(create.StringHashcode(jsonString)))

	return nil
}
