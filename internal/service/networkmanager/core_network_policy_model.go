// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type coreNetworkPolicyDocument struct {
	Version                      string                                          `json:"version,omitempty"`
	CoreNetworkConfiguration     *coreNetworkPolicyCoreNetworkConfiguration      `json:"core-network-configuration"`
	Segments                     []*coreNetworkPolicySegment                     `json:"segments"`
	RoutingPolicies              []*coreNetworkPolicyRoutingPolicy               `json:"routing-policies,omitempty"`
	NetworkFunctionGroups        []*coreNetworkPolicyNetworkFunctionGroup        `json:"network-function-groups,omitempty"`
	SegmentActions               []*coreNetworkPolicySegmentAction               `json:"segment-actions,omitempty"`
	AttachmentPolicies           []*coreNetworkPolicyAttachmentPolicy            `json:"attachment-policies,omitempty"`
	AttachmentRoutingPolicyRules []*coreNetworkPolicyAttachmentRoutingPolicyRule `json:"attachment-routing-policy-rules,omitempty"`
}

type coreNetworkPolicyCoreNetworkConfiguration struct {
	AsnRanges                       any                                         `json:"asn-ranges"`
	InsideCidrBlocks                any                                         `json:"inside-cidr-blocks,omitempty"`
	VpnEcmpSupport                  bool                                        `json:"vpn-ecmp-support"`
	EdgeLocations                   []*coreNetworkPolicyCoreNetworkEdgeLocation `json:"edge-locations,omitempty"`
	DnsSupport                      bool                                        `json:"dns-support"`
	SecurityGroupReferencingSupport bool                                        `json:"security-group-referencing-support"`
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

type coreNetworkPolicyRoutingPolicy struct {
	RoutingPolicyName        string                                `json:"routing-policy-name,omitempty"`
	RoutingPolicyDescription string                                `json:"routing-policy-description,omitempty"`
	RoutingPolicyDirection   string                                `json:"routing-policy-direction,omitempty"`
	RoutingPolicyNumber      int                                   `json:"routing-policy-number,omitempty"`
	RoutingPolicyRules       []*coreNetworkPolicyRoutingPolicyRule `json:"routing-policy-rules,omitempty"`
}

type coreNetworkPolicyRoutingPolicyRule struct {
	RuleNumber     int                                           `json:"rule-number,omitempty"`
	RuleDefinition *coreNetworkPolicyRoutingPolicyRuleDefinition `json:"rule-definition,omitempty"`
}

type coreNetworkPolicyRoutingPolicyRuleDefinition struct {
	ConditionLogic  string                                              `json:"condition-logic,omitempty"`
	MatchConditions []*coreNetworkPolicyRoutingPolicyRuleMatchCondition `json:"match-conditions,omitempty"`
	Action          *coreNetworkPolicyRoutingPolicyRuleAction           `json:"action,omitempty"`
}

type coreNetworkPolicyRoutingPolicyRuleMatchCondition struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type coreNetworkPolicyRoutingPolicyRuleAction struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for routing policy rule actions.
// Some action types require arrays of integers: prepend-asn-list, remove-asn-list, replace-asn-list
func (c coreNetworkPolicyRoutingPolicyRuleAction) MarshalJSON() ([]byte, error) {
	// Types that require array of integers (ASN lists)
	asnListTypes := map[string]bool{
		"prepend-asn-list": true,
		"remove-asn-list":  true,
		"replace-asn-list": true,
	}

	if asnListTypes[c.Type] && c.Value != "" {
		// Parse comma-separated ASN values into an array of integers
		parts := strings.Split(c.Value, ",")
		asnList := make([]int64, 0, len(parts))
		for _, part := range parts {
			if asn, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
				asnList = append(asnList, asn)
			}
		}
		return json.Marshal(map[string]any{
			names.AttrType:  c.Type,
			names.AttrValue: asnList,
		})
	}

	// Default: marshal as strings
	type Alias coreNetworkPolicyRoutingPolicyRuleAction
	return json.Marshal((*Alias)(&c))
}

type coreNetworkPolicyNetworkFunctionGroup struct {
	Name                        string `json:"name"`
	Description                 string `json:"description,omitempty"`
	RequireAttachmentAcceptance bool   `json:"require-attachment-acceptance"`
}

type coreNetworkPolicySegmentAction struct {
	Action                  string                                                 `json:"action"`
	Segment                 string                                                 `json:"segment,omitempty"`
	Mode                    string                                                 `json:"mode,omitempty"`
	ShareWith               any                                                    `json:"share-with,omitempty"`
	ShareWithExcept         any                                                    `json:",omitempty"`
	DestinationCidrBlocks   any                                                    `json:"destination-cidr-blocks,omitempty"`
	Destinations            any                                                    `json:"destinations,omitempty"`
	Description             string                                                 `json:"description,omitempty"`
	WhenSentTo              *coreNetworkPolicySegmentActionWhenSentTo              `json:"when-sent-to,omitempty"`
	Via                     *coreNetworkPolicySegmentActionVia                     `json:"via,omitempty"`
	EdgeLocationAssociation *coreNetworkPolicySegmentActionEdgeLocationAssociation `json:"edge-location-association,omitempty"`
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

type coreNetworkPolicySegmentActionEdgeLocationAssociation struct {
	EdgeLocation       string `json:"edge-location,omitempty"`
	PeerEdgeLocation   string `json:"peer-edge-location,omitempty"`
	RoutingPolicyNames any    `json:"routing-policy-names,omitempty"`
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

type coreNetworkPolicyAttachmentRoutingPolicyRule struct {
	RuleNumber    int                                                      `json:"rule-number,omitempty"`
	Description   string                                                   `json:"description,omitempty"`
	EdgeLocations any                                                      `json:"edge-locations,omitempty"`
	Conditions    []*coreNetworkPolicyAttachmentRoutingPolicyRuleCondition `json:"conditions"`
	Action        *coreNetworkPolicyAttachmentRoutingPolicyRuleAction      `json:"action"`
}

type coreNetworkPolicyAttachmentRoutingPolicyRuleCondition struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type coreNetworkPolicyAttachmentRoutingPolicyRuleAction struct {
	AssociateRoutingPolicies any `json:"associate-routing-policies,omitempty"`
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
		Action:                  c.Action,
		Mode:                    c.Mode,
		Destinations:            c.Destinations,
		DestinationCidrBlocks:   c.DestinationCidrBlocks,
		Segment:                 c.Segment,
		ShareWith:               share,
		Via:                     c.Via,
		WhenSentTo:              whenSentTo,
		Description:             c.Description,
		EdgeLocationAssociation: c.EdgeLocationAssociation,
	})
}

// MarshalJSON implements custom JSON marshaling for match conditions.
// Some condition types (asn-in-as-path) require the value to be a number, not a string.
func (c coreNetworkPolicyRoutingPolicyRuleMatchCondition) MarshalJSON() ([]byte, error) {
	// Types that require numeric values
	numericTypes := map[string]bool{
		"asn-in-as-path": true,
	}

	if numericTypes[c.Type] && c.Value != "" {
		// Try to parse the value as an integer
		if numVal, err := strconv.ParseInt(c.Value, 10, 64); err == nil {
			return json.Marshal(map[string]any{
				names.AttrType:  c.Type,
				names.AttrValue: numVal,
			})
		}
	}

	// Default: marshal as strings
	type Alias coreNetworkPolicyRoutingPolicyRuleMatchCondition
	return json.Marshal((*Alias)(&c))
}

func coreNetworkPolicyExpandStringList(configured []any) any {
	vs := flex.ExpandStringValueList(configured)
	slices.Sort(vs)
	slices.Reverse(vs)

	return vs
}
