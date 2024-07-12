// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

type coreNetworkPolicyDocument struct {
	Version                  string                                     `json:"version,omitempty"`
	CoreNetworkConfiguration *coreNetworkPolicyCoreNetworkConfiguration `json:"core-network-configuration"`
	Segments                 []*coreNetworkPolicySegment                `json:"segments"`
	NetworkFunctionGroups    []*coreNetworkPolicyNetworkFunctionGroup   `json:"network-function-groups"`
	SegmentActions           []*coreNetworkPolicySegmentAction          `json:"segment-actions,omitempty"`
	AttachmentPolicies       []*coreNetworkPolicyAttachmentPolicy       `json:"attachment-policies,omitempty"`
}

type coreNetworkPolicyCoreNetworkConfiguration struct {
	AsnRanges        interface{}                                 `json:"asn-ranges"`
	InsideCidrBlocks interface{}                                 `json:"inside-cidr-blocks,omitempty"`
	VpnEcmpSupport   bool                                        `json:"vpn-ecmp-support"`
	EdgeLocations    []*coreNetworkPolicyCoreNetworkEdgeLocation `json:"edge-locations,omitempty"`
}

type coreNetworkPolicyCoreNetworkEdgeLocation struct {
	Location         string      `json:"location"`
	Asn              int64       `json:"asn,omitempty"`
	InsideCidrBlocks interface{} `json:"inside-cidr-blocks,omitempty"`
}

type coreNetworkPolicySegment struct {
	Name                        string      `json:"name"`
	Description                 string      `json:"description,omitempty"`
	EdgeLocations               interface{} `json:"edge-locations,omitempty"`
	IsolateAttachments          bool        `json:"isolate-attachments"`
	RequireAttachmentAcceptance bool        `json:"require-attachment-acceptance"`
	DenyFilter                  interface{} `json:"deny-filter,omitempty"`
	AllowFilter                 interface{} `json:"allow-filter,omitempty"`
}

type coreNetworkPolicyNetworkFunctionGroup struct {
	Name                        string `json:"name"`
	Description                 string `json:"description,omitempty"`
	RequireAttachmentAcceptance bool   `json:"require-attachment-acceptance"`
}

type coreNetworkPolicySegmentAction struct {
	Action                string                                    `json:"action"`
	Segment               string                                    `json:"segment,omitempty"`
	Mode                  string                                    `json:"mode,omitempty"`
	ShareWith             interface{}                               `json:"share-with,omitempty"`
	ShareWithExcept       interface{}                               `json:",omitempty"`
	DestinationCidrBlocks interface{}                               `json:"destination-cidr-blocks,omitempty"`
	Destinations          interface{}                               `json:"destinations,omitempty"`
	Description           string                                    `json:"description,omitempty"`
	WhenSentTo            *coreNetworkPolicySegmentActionWhenSentTo `json:"when-sent-to,omitempty"`
	Via                   *coreNetworkPolicySegmentActionVia        `json:"via,omitempty"`
}

type coreNetworkPolicySegmentActionWhenSentTo struct {
	Segments interface{} `json:"segments,omitempty"`
}

type coreNetworkPolicySegmentActionVia struct {
	NetworkFunctionGroups interface{}                                      `json:"network-function-groups,omitempty"`
	WithEdgeOverrides     []*coreNetworkPolicySegmentActionViaEdgeOverride `json:"with-edge-overrides,omitempty"`
}
type coreNetworkPolicySegmentActionViaEdgeOverride struct {
	EdgeSets interface{} `json:"edge-sets,omitempty"`
	UseEdge  string      `json:"use-edge,omitempty"`
}

type coreNetworkPolicyAttachmentPolicy struct {
	RuleNumber     int                                           `json:"rule-number,omitempty"`
	Description    string                                        `json:"description,omitempty"`
	ConditionLogic string                                        `json:"condition-logic,omitempty"`
	Conditions     []*coreNetworkPolicyAttachmentPolicyCondition `json:"conditions"`
	Action         *coreNetworkPolicyAttachmentPolicyAction      `json:"action"`
}

type coreNetworkPolicyAttachmentPolicyCondition struct {
	Type     string `json:"type,omitempty"`
	Operator string `json:"operator,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
}

type coreNetworkPolicyAttachmentPolicyAction struct {
	AssociationMethod         string `json:"association-method,omitempty"`
	Segment                   string `json:"segment,omitempty"`
	TagValueOfKey             string `json:"tag-value-of-key,omitempty"`
	RequireAcceptance         bool   `json:"require-acceptance,omitempty"`
	AddToNetworkFunctionGroup string `json:"add-to-network-function-group,omitempty"`
}

func (c coreNetworkPolicySegmentAction) MarshalJSON() ([]byte, error) {
	type Alias coreNetworkPolicySegmentAction
	var share interface{}

	if v := c.ShareWith; v != nil {
		v := v.([]string)
		if v[0] == "*" {
			share = v[0]
		} else {
			share = v
		}
	} else if v := c.ShareWithExcept; v != nil {
		share = map[string]interface{}{
			"except": v.([]string),
		}
	}

	return json.Marshal(&Alias{
		Action:                c.Action,
		Mode:                  c.Mode,
		Destinations:          c.Destinations,
		DestinationCidrBlocks: c.DestinationCidrBlocks,
		Segment:               c.Segment,
		ShareWith:             share,
		Via:                   c.Via,
		WhenSentTo:            c.WhenSentTo,
	})
}

func coreNetworkPolicyExpandStringList(configured []interface{}) interface{} {
	vs := flex.ExpandStringValueList(configured)
	sort.Sort(sort.Reverse(sort.StringSlice(vs)))

	return vs
}
