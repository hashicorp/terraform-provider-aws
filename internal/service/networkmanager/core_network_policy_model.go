package networkmanager

import (
	"sort"
)

const (
	policyModelMarshallJSONStartSliceSize = 2
)

type CoreNetworkPolicyDoc struct {
	Version            string                         `json:",omitempty"`
	Id                 string                         `json:",omitempty"`
	AttachmentPolicies []*CoreNetworkAttachmentPolicy `json:"AttachmentPolicies"`
	Segments           []*CoreNetworkPolicySegment    `json:"Segments"`
	// SegmentActions           []*CoreNetworkPolicySegment                `json:"Segments"`
	CoreNetworkConfiguration *CoreNetworkPolicyCoreNetworkConfiguration `json:"CoreNetworkConfiguration"`
}

type CoreNetworkAttachmentPolicy struct {
	RuleNumber     int
	Action         *CoreNetworkAttachmentPolicyAction
	Conditions     []*CoreNetworkAttachmentPolicyCondition
	Description    string `json:"Description,omitempty"`
	ConditionLogic string `json:"ConditionLogic,omitempty"`
}

type CoreNetworkAttachmentPolicyAction struct {
	AssociationMethod string
	Segment           string `json:"Segment,omitempty"`
	TagValueOfKey     string `json:"TagValueOfKey,omitempty"`
	RequireAcceptance bool   `json:"RequireAcceptance,omitempty"`
}

type CoreNetworkAttachmentPolicyCondition struct {
	Type     string
	Operator string `json:"Operator,omitempty"`
	Key      string `json:"Key,omitempty"`
	Value    string `json:"Value,omitempty"`
}

type CoreNetworkPolicySegment struct {
	Name                        string
	AllowFilter                 interface{} `json:"AllowFilter,omitempty"`
	DenyFilter                  interface{} `json:"DenyFilter,omitempty"`
	EdgeLocations               interface{} `json:"EdgeLocations,omitempty"`
	IsolateAttachments          bool        `json:"IsolateAttachments,omitempty"`
	RequireAttachmentAcceptance bool        `json:"RequireAttachmentAcceptance,omitempty"`
}

type CoreNetworkPolicyCoreNetworkConfiguration struct {
	AsnRanges        interface{}
	VpnEcmpSupport   bool                       `json:"VpnEcmpSupport,omitempty"`
	EdgeLocations    []*CoreNetworkEdgeLocation `json:"EdgeLocations,omitempty"`
	InsideCidrBlocks interface{}                `json:"InsideCidrBlocks,omitempty"`
}

type CoreNetworkEdgeLocation struct {
	Location         string
	Asn              int         `json:"Asn,omitempty"`
	InsideCidrBlocks interface{} `json:"InsideCidrBlocks,omitempty"`
}

func (s *CoreNetworkPolicyDoc) Merge(newDoc *CoreNetworkPolicyDoc) {
	// adopt newDoc's Id
	if len(newDoc.Id) > 0 {
		s.Id = newDoc.Id
	}

	// let newDoc upgrade our Version
	if newDoc.Version > s.Version {
		s.Version = newDoc.Version
	}

	// merge in newDoc's segments, overwriting any existing Names
	var seen bool
	for _, newSegment := range newDoc.Segments {
		if len(newSegment.Name) == 0 {
			s.Segments = append(s.Segments, newSegment)
			continue
		}
		seen = false
		for i, existingSegment := range s.Segments {
			if existingSegment.Name == newSegment.Name {
				s.Segments[i] = newSegment
				seen = true
				break
			}
		}
		if !seen {
			s.Segments = append(s.Segments, newSegment)
		}
	}
}

func CoreNetworkPolicyDecodeConfigStringList(lI []interface{}) interface{} {
	if len(lI) == 1 {
		return lI[0].(string)
	}
	ret := make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}
