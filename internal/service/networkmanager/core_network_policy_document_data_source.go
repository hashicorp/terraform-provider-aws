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

			"core_network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"asn_ranges": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// validate like <asn>-<asn> ?
						},
						"vpn_ecmp_support": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"edge_locations": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 17,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"location": {
										Type:     schema.TypeString,
										Required: true,
										// a-z, 0-9
										// ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
									},
									"asn": {
										Type:     schema.TypeInt,
										Default:  false,
										Optional: true,
										// validate asn-like
									},
									"inside_cidr_blocks": {
										Type:     schema.TypeList,
										Optional: true,
										// validate either ipv4 or 6?
										Elem: &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"inside_cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							// validate either ipv4 or 6?
							Elem: &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				// DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
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

	// TODO: segments is required
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
