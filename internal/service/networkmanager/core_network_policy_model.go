package networkmanager

import (
	"encoding/json"
	"sort"
)

type CoreNetworkPolicyDoc struct {
	Version                  string                                     `json:"version,omitempty"`
	CoreNetworkConfiguration *CoreNetworkPolicyCoreNetworkConfiguration `json:"core-network-configuration"`
	Segments                 []*CoreNetworkPolicySegment                `json:"segments"`
	AttachmentPolicies       []*CoreNetworkAttachmentPolicy             `json:"attachment-policies,omitempty"`
	SegmentActions           []*CoreNetworkPolicySegmentAction          `json:"segment-actions,omitempty"`
}

type CoreNetworkPolicySegmentAction struct {
	Action                string      `json:"action"`
	Destinations          interface{} `json:"destinations,omitempty"`
	DestinationCidrBlocks interface{} `json:"destination-cidr-blocks,omitempty"`
	Mode                  string      `json:"mode,omitempty"`
	Segment               string      `json:"segment,omitempty"`
	ShareWith             interface{} `json:"share-with,omitempty"`
	ShareWithExcept       interface{} `json:",omitempty"`
}

type CoreNetworkAttachmentPolicy struct {
	RuleNumber     int                                     `json:"rule-number,omitempty"`
	Action         *CoreNetworkAttachmentPolicyAction      `json:"action"`
	Conditions     []*CoreNetworkAttachmentPolicyCondition `json:"conditions"`
	Description    string                                  `json:"description,omitempty"`
	ConditionLogic string                                  `json:"condition-logic,omitempty"`
}

type CoreNetworkAttachmentPolicyAction struct {
	AssociationMethod string `json:"association-method,omitempty"`
	Segment           string `json:"segment,omitempty"`
	TagValueOfKey     string `json:"tag-value-of-key,omitempty"`
	RequireAcceptance bool   `json:"require-acceptance,omitempty"`
}

type CoreNetworkAttachmentPolicyCondition struct {
	Type     string `json:"type,omitempty"`
	Operator string `json:"operator,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
}

type CoreNetworkPolicySegment struct {
	Name                        string      `json:"name"`
	Description                 string      `json:"description,omitempty"`
	AllowFilter                 interface{} `json:"allow-filter,omitempty"`
	DenyFilter                  interface{} `json:"deny-filter,omitempty"`
	EdgeLocations               interface{} `json:"edge-locations,omitempty"`
	IsolateAttachments          bool        `json:"isolate-attachments,omitempty"`
	RequireAttachmentAcceptance bool        `json:"require-attachment-acceptance,omitempty"`
}

type CoreNetworkPolicyCoreNetworkConfiguration struct {
	AsnRanges        interface{}                `json:"asn-ranges"`
	VpnEcmpSupport   bool                       `json:"vpn-ecmp-support"`
	EdgeLocations    []*CoreNetworkEdgeLocation `json:"edge-locations,omitempty"`
	InsideCidrBlocks interface{}                `json:"inside-cidr-blocks,omitempty"`
}

type CoreNetworkEdgeLocation struct {
	Location         string      `json:"location"`
	Asn              int         `json:"asn,omitempty"`
	InsideCidrBlocks interface{} `json:"inside-cidr-blocks,omitempty"`
}

func (c *CoreNetworkPolicySegmentAction) MarshalJSON() ([]byte, error) {
	type Alias CoreNetworkPolicySegmentAction

	var share interface{}

	if c.ShareWith != nil {
		sWIntf := c.ShareWith.([]string)

		if sWIntf[0] == "*" {
			share = sWIntf[0]
		} else {
			share = sWIntf
		}
	}

	if c.ShareWithExcept != nil {
		share = c.ShareWithExcept.([]string)
	}

	return json.Marshal(&Alias{
		Action:                c.Action,
		Mode:                  c.Mode,
		Destinations:          c.Destinations,
		DestinationCidrBlocks: c.DestinationCidrBlocks,
		Segment:               c.Segment,
		ShareWith:             share,
	})
}

func CoreNetworkPolicyDecodeConfigStringList(lI []interface{}) interface{} {
	ret := make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}
