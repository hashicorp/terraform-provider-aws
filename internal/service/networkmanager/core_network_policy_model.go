package networkmanager

import (
	"encoding/json"
	"fmt"
	"sort"
)

const (
	policyModelMarshallJSONStartSliceSize = 2
)

type CoreNetworkPolicyDoc struct {
	Version  string                      `json:",omitempty"`
	Id       string                      `json:",omitempty"`
	Segments []*CoreNetworkPolicySegment `json:"Segment"`
}

type CoreNetworkPolicySegment struct {
	Name                        string
	AllowFilter                 interface{} `json:"AllowFilter,omitempty"`
	DenyFilter                  interface{} `json:"DenyFilter,omitempty"`
	EdgeLocations               interface{} `json:"EdgeLocations,omitempty"`
	IsolateAttachments          bool        `json:"IsolateAttachments,omitempty"`
	RequireAttachmentAcceptance bool        `json:"RequireAttachmentAcceptance,omitempty"`
}

type CoreNetworkPolicySegmentPrincipal struct {
	Type        string
	Identifiers interface{}
}

type CoreNetworkPolicySegmentCondition struct {
	Test     string
	Variable string
	Values   interface{}
}

type CoreNetworkPolicySegmentPrincipalSet []CoreNetworkPolicySegmentPrincipal
type CoreNetworkPolicySegmentConditionSet []CoreNetworkPolicySegmentCondition

func (s *CoreNetworkPolicyDoc) Merge(newDoc *CoreNetworkPolicyDoc) {
	// adopt newDoc's Id
	if len(newDoc.Id) > 0 {
		s.Id = newDoc.Id
	}

	// let newDoc upgrade our Version
	if newDoc.Version > s.Version {
		s.Version = newDoc.Version
	}

	// merge in newDoc's statements, overwriting any existing Names
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

func (ps CoreNetworkPolicySegmentPrincipalSet) MarshalJSON() ([]byte, error) {
	raw := map[string]interface{}{}

	// Although CoreNetwork documentation says, that "*" and {"AWS": "*"} are equivalent
	// (https://docs.aws.amazon.com/CoreNetwork/latest/UserGuide/reference_policies_elements_principal.html),
	// in practice they are not for CoreNetwork roles. CoreNetwork will return an error if trust
	// policy have "*" or {"*": "*"} as principal, but will accept {"AWS": "*"}.
	// Only {"*": "*"} should be normalized to "*".
	if len(ps) == 1 {
		p := ps[0]
		if p.Type == "*" {
			if sv, ok := p.Identifiers.(string); ok && sv == "*" {
				return []byte(`"*"`), nil
			}

			if av, ok := p.Identifiers.([]string); ok && len(av) == 1 && av[0] == "*" {
				return []byte(`"*"`), nil
			}
		}
	}

	for _, p := range ps {
		switch i := p.Identifiers.(type) {
		case []string:
			switch v := raw[p.Type].(type) {
			case nil:
				raw[p.Type] = make([]string, 0, len(i))
			case string:
				// Convert to []string to prevent panic
				raw[p.Type] = make([]string, 0, len(i)+1)
				raw[p.Type] = append(raw[p.Type].([]string), v)
			}
			sort.Sort(sort.Reverse(sort.StringSlice(i)))
			raw[p.Type] = append(raw[p.Type].([]string), i...)
		case string:
			switch v := raw[p.Type].(type) {
			case nil:
				raw[p.Type] = i
			case string:
				// Convert to []string to stop drop of principals
				raw[p.Type] = make([]string, 0, policyModelMarshallJSONStartSliceSize)
				raw[p.Type] = append(raw[p.Type].([]string), v)
				raw[p.Type] = append(raw[p.Type].([]string), i)
			case []string:
				raw[p.Type] = append(raw[p.Type].([]string), i)
			}
		default:
			return []byte{}, fmt.Errorf("Unsupported data type %T for CoreNetworkPolicySegmentPrincipalSet", i)
		}
	}

	return json.Marshal(&raw)
}

func (ps *CoreNetworkPolicySegmentPrincipalSet) UnmarshalJSON(b []byte) error {
	var out CoreNetworkPolicySegmentPrincipalSet

	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	switch t := data.(type) {
	case string:
		out = append(out, CoreNetworkPolicySegmentPrincipal{Type: "*", Identifiers: []string{"*"}})
	case map[string]interface{}:
		for key, value := range data.(map[string]interface{}) {
			switch vt := value.(type) {
			case string:
				out = append(out, CoreNetworkPolicySegmentPrincipal{Type: key, Identifiers: value.(string)})
			case []interface{}:
				values := []string{}
				for _, v := range value.([]interface{}) {
					values = append(values, v.(string))
				}
				out = append(out, CoreNetworkPolicySegmentPrincipal{Type: key, Identifiers: values})
			default:
				return fmt.Errorf("Unsupported data type %T for CoreNetworkPolicySegmentPrincipalSet.Identifiers", vt)
			}
		}
	default:
		return fmt.Errorf("Unsupported data type %T for CoreNetworkPolicySegmentPrincipalSet", t)
	}

	*ps = out
	return nil
}

func (cs CoreNetworkPolicySegmentConditionSet) MarshalJSON() ([]byte, error) {
	raw := map[string]map[string]interface{}{}

	for _, c := range cs {
		if _, ok := raw[c.Test]; !ok {
			raw[c.Test] = map[string]interface{}{}
		}
		switch i := c.Values.(type) {
		case []string:
			if _, ok := raw[c.Test][c.Variable]; !ok {
				raw[c.Test][c.Variable] = make([]string, 0, len(i))
			}
			// order matters with values so not sorting here
			raw[c.Test][c.Variable] = append(raw[c.Test][c.Variable].([]string), i...)
		case string:
			raw[c.Test][c.Variable] = i
		default:
			return nil, fmt.Errorf("Unsupported data type for CoreNetworkPolicySegmentConditionSet: %s", i)
		}
	}

	return json.Marshal(&raw)
}

func (cs *CoreNetworkPolicySegmentConditionSet) UnmarshalJSON(b []byte) error {
	var out CoreNetworkPolicySegmentConditionSet

	var data map[string]map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for test_key, test_value := range data {
		for var_key, var_values := range test_value {
			switch var_values := var_values.(type) {
			case string:
				out = append(out, CoreNetworkPolicySegmentCondition{Test: test_key, Variable: var_key, Values: []string{var_values}})
			case []interface{}:
				values := []string{}
				for _, v := range var_values {
					values = append(values, v.(string))
				}
				out = append(out, CoreNetworkPolicySegmentCondition{Test: test_key, Variable: var_key, Values: values})
			}
		}
	}

	*cs = out
	return nil
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
