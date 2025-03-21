// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"encoding/json"
	"slices"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

type coreNetworkPolicyDocument struct {
	Version                  string                                     `json:"version,omitempty"`
	CoreNetworkConfiguration *coreNetworkPolicyCoreNetworkConfiguration `json:"core-network-configuration"`
	Segments                 []*coreNetworkPolicySegment                `json:"segments"`
	NetworkFunctionGroups    []*coreNetworkPolicyNetworkFunctionGroup   `json:"network-function-groups,omitempty"`
	SegmentActions           []*coreNetworkPolicySegmentAction          `json:"segment-actions,omitempty"`
	AttachmentPolicies       []*coreNetworkPolicyAttachmentPolicy       `json:"attachment-policies,omitempty"`
}

type coreNetworkPolicyCoreNetworkConfiguration struct {
	AsnRanges        any                                         `json:"asn-ranges"`
	InsideCidrBlocks any                                         `json:"inside-cidr-blocks,omitempty"`
	VpnEcmpSupport   bool                                        `json:"vpn-ecmp-support"`
	EdgeLocations    []*coreNetworkPolicyCoreNetworkEdgeLocation `json:"edge-locations,omitempty"`
}

type coreNetworkPolicyCoreNetworkEdgeLocation struct {
	Location         string `json:"location"`
	Asn              int64  `json:"asn,omitempty"`
	InsideCidrBlocks any    `json:"inside-cidr-blocks,omitempty"`
}

type coreNetworkPolicySegment struct {
	Name                        string `json:"name"`
	Description                 string `json:"description,omitempty"`
	EdgeLocations               any    `json:"edge-locations,omitempty"`
	IsolateAttachments          bool   `json:"isolate-attachments"`
	RequireAttachmentAcceptance bool   `json:"require-attachment-acceptance"`
	DenyFilter                  any    `json:"deny-filter,omitempty"`
	AllowFilter                 any    `json:"allow-filter,omitempty"`
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
	ShareWith             any                                       `json:"share-with,omitempty"`
	ShareWithExcept       any                                       `json:",omitempty"`
	DestinationCidrBlocks any                                       `json:"destination-cidr-blocks,omitempty"`
	Destinations          any                                       `json:"destinations,omitempty"`
	Description           string                                    `json:"description,omitempty"`
	WhenSentTo            *coreNetworkPolicySegmentActionWhenSentTo `json:"when-sent-to,omitempty"`
	Via                   *coreNetworkPolicySegmentActionVia        `json:"via,omitempty"`
}

type coreNetworkPolicySegmentActionWhenSentTo struct {
	Segments any `json:"segments,omitempty"`
}

type coreNetworkPolicySegmentActionVia struct {
	NetworkFunctionGroups any                                              `json:"network-function-groups,omitempty"`
	WithEdgeOverrides     []*coreNetworkPolicySegmentActionViaEdgeOverride `json:"with-edge-overrides,omitempty"`
}
type coreNetworkPolicySegmentActionViaEdgeOverride struct {
	EdgeSets        [][]string `json:"edge-sets,omitempty"`
	UseEdgeLocation string     `json:"use-edge-location,omitempty"`
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
	var share any
	var whenSentTo *coreNetworkPolicySegmentActionWhenSentTo

	if v := c.ShareWith; v != nil {
		v := v.([]string)
		if v[0] == "*" {
			share = v[0]
		} else {
			share = v
		}
	} else if v := c.ShareWithExcept; v != nil {
		share = map[string]any{
			"except": v.([]string),
		}
	}

	if v := c.WhenSentTo; v != nil {
		if s := v.Segments; s != nil {
			s := s.([]string)
			if s[0] == "*" {
				whenSentTo = &coreNetworkPolicySegmentActionWhenSentTo{Segments: s[0]}
			} else {
				whenSentTo = c.WhenSentTo
			}
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
		WhenSentTo:            whenSentTo,
	})
}

func coreNetworkPolicyExpandStringList(configured []any) any {
	vs := flex.ExpandStringValueList(configured)
	slices.Sort(vs)
	slices.Reverse(vs)

	return vs
}
