// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

type TriBoolean uint

const (
	TriBooleanUnset TriBoolean = iota
	TriBooleanTrue
	TriBooleanFalse
)

func TriBool(b bool) TriBoolean {
	if b {
		return TriBooleanTrue
	} else {
		return TriBooleanFalse
	}
}

type Implementation string

const (
	ImplementationFramework Implementation = "framework"
	ImplementationSDK       Implementation = "sdk"
)

type ResourceIdentity struct {
	isARNIdentity                  bool
	isCustomInherentRegionIdentity bool
	isSingletonIdentity            bool
	identityAttributeName          string
	IdentityDuplicateAttrNames     []string
	IdentityAttributes             []IdentityAttribute
	MutableIdentity                bool
	IdentityVersion                int64
	SDKv2IdentityUpgraders         []string
	CustomInherentRegionParser     string
	HasV6_0NullValuesError         bool
	HasV6_0RefreshError            bool
	ImportIDHandler                string
	SetIDAttribute                 bool
}

func (r ResourceIdentity) HasResourceIdentity() bool {
	return r.IsParameterizedIdentity() || r.isARNIdentity || r.isSingletonIdentity || r.isCustomInherentRegionIdentity
}

func (r ResourceIdentity) HasInherentRegionIdentity() bool {
	return r.isARNIdentity || r.isCustomInherentRegionIdentity
}

func (r ResourceIdentity) IsARNIdentity() bool {
	return r.isARNIdentity
}

func (r ResourceIdentity) IsCustomInherentRegionIdentity() bool {
	return r.isCustomInherentRegionIdentity
}

func (r ResourceIdentity) IsParameterizedIdentity() bool {
	return len(r.IdentityAttributes) > 0
}

func (r ResourceIdentity) IsSingleParameterizedIdentity() bool {
	return len(r.IdentityAttributes) == 1
}

func (r ResourceIdentity) IsMultipleParameterizedIdentity() bool {
	return len(r.IdentityAttributes) > 1
}

func (r ResourceIdentity) IsSingletonIdentity() bool {
	return r.isSingletonIdentity
}

func (r ResourceIdentity) IdentityAttribute() string {
	return namesgen.ConstOrQuote(r.IdentityAttributeName())
}

func (r ResourceIdentity) IdentityAttributeName() string {
	if r.identityAttributeName != "" {
		return r.identityAttributeName
	}
	if len(r.IdentityAttributes) == 1 {
		return r.IdentityAttributes[0].Name_
	}
	return ""
}

func (r ResourceIdentity) HasIdentityDuplicateAttrs() bool {
	return len(r.IdentityDuplicateAttrNames) > 0
}

func (r ResourceIdentity) IdentityDuplicateAttrs() []string {
	return tfslices.ApplyToAll(r.IdentityDuplicateAttrNames, func(s string) string {
		return namesgen.ConstOrQuote(s)
	})
}

func (r ResourceIdentity) Validate() error {
	if r.IsMultipleParameterizedIdentity() {
		if r.ImportIDHandler == "" {
			return errors.New("ImportIDHandler required for multiple parameterized identity")
		}
	}
	return nil
}

type IdentityAttribute struct {
	Name_                  string
	Optional               bool
	ResourceAttributeName_ string
	TestNotNull            bool
	ValueType              string
}

func (a IdentityAttribute) Name() string {
	return namesgen.ConstOrQuote(a.Name_)
}

func (a IdentityAttribute) ResourceAttributeName() string {
	return namesgen.ConstOrQuote(a.ResourceAttributeName_)
}

func ParseResourceIdentity(annotationName string, args Args, implementation Implementation, d *ResourceIdentity, goImports *[]GoImport) (errs error) {
	switch annotationName {
	case "ArnIdentity":
		d.isARNIdentity = true
		if len(args.Positional) == 0 {
			d.identityAttributeName = "arn"
		} else {
			d.identityAttributeName = args.Positional[0]
		}

		parseIdentityDuplicateAttrNames(args, implementation, d)

		for k := range args.Keyword {
			errs = errors.Join(errs, fmt.Errorf("annotation \"@ArnIdentity\": unexpected keyword parameter %q", k))
		}

	case "CustomInherentRegionIdentity":
		d.isCustomInherentRegionIdentity = true

		if len(args.Positional) < 2 {
			errs = errors.Join(errs, errors.New("annotation \"@CustomInherentRegionIdentity\": missing required positional parameters"))
		}

		d.identityAttributeName = args.Positional[0]

		parseIdentityDuplicateAttrNames(args, implementation, d)

		for k := range args.Keyword {
			errs = errors.Join(errs, fmt.Errorf("annotation \"@CustomInherentRegionIdentity\": unexpected keyword parameter %q", k))
		}

		attr := args.Positional[1]
		if funcName, importSpec, err := ParseIdentifierSpec(attr); err != nil {
			errs = errors.Join(errs, fmt.Errorf("%q: %w", attr, err))
		} else {
			d.CustomInherentRegionParser = funcName
			if importSpec != nil {
				*goImports = append(*goImports, *importSpec)
			}
		}

	case "IdentityAttribute":
		if len(args.Positional) == 0 {
			errs = errors.Join(errs, errors.New("no Identity attribute name"))
		}

		identityAttribute := IdentityAttribute{
			Name_: args.Positional[0],
		}

		for k := range args.Keyword {
			switch k {
			// Needs to be handled differently than in `parseIdentityDuplicateAttrNames`
			case "identityDuplicateAttributes":
				attr := args.Keyword[k]
				attrs := strings.Split(attr, ";")
				// Sort `id` to first position, the rest alphabetically
				slices.SortFunc(attrs, func(a, b string) int {
					if a == "id" {
						return -1
					} else if b == "id" {
						return 1
					} else {
						return strings.Compare(a, b)
					}
				})
				d.IdentityDuplicateAttrNames = slices.Compact(attrs)

			case "optional":
				attr := args.Keyword[k]
				if b, err := ParseBoolAttr("optional", attr); err != nil {
					errs = errors.Join(errs, err)
				} else {
					identityAttribute.Optional = b
				}

			case "resourceAttributeName":
				identityAttribute.ResourceAttributeName_ = args.Keyword[k]

			case "testNotNull":
				attr := args.Keyword[k]
				if b, err := ParseBoolAttr("testNotNull", attr); err != nil {
					errs = errors.Join(errs, err)
				} else {
					identityAttribute.TestNotNull = b
				}

			case "valueType":
				identityAttribute.ValueType = args.Keyword[k]

			default:
				errs = errors.Join(errs, fmt.Errorf("annotation \"@IdentityAttribute\": unexpected keyword parameter %q", k))
			}
		}

		d.IdentityAttributes = append(d.IdentityAttributes, identityAttribute)

	case "IdentityVersion":
		attr := args.Positional[0]
		if i, err := strconv.ParseInt(attr, 10, 64); err != nil {
			return fmt.Errorf("invalid IdentityVersion value: %q. Should be integer value.", attr)
		} else {
			d.IdentityVersion = i
		}

		for k := range args.Keyword {
			switch k {
			case "sdkV2IdentityUpgraders":
				attr := args.Keyword[k]
				attrs := strings.Split(attr, ";")
				d.SDKv2IdentityUpgraders = attrs

			default:
				errs = errors.Join(errs, fmt.Errorf("annotation \"@IdentityVersion\": unexpected keyword parameter %q", k))
			}
		}

	case "ImportIDHandler":
		attr := args.Positional[0]
		if typeName, importSpec, err := ParseIdentifierSpec(attr); err != nil {
			errs = errors.Join(errs, err)
		} else {
			d.ImportIDHandler = typeName
			if importSpec != nil {
				*goImports = append(*goImports, *importSpec)
			}
		}

		for k := range args.Keyword {
			switch k {
			case "setIDAttribute":
				attr := args.Keyword[k]
				if b, err := strconv.ParseBool(attr); err != nil {
					errs = errors.Join(errs, err)
				} else {
					d.SetIDAttribute = b
				}

			default:
				errs = errors.Join(errs, fmt.Errorf("annotation \"@ImportIDHandler\": unexpected keyword parameter %q", k))
			}
		}

	case "MutableIdentity":
		d.MutableIdentity = true

	case "SingletonIdentity":
		d.isSingletonIdentity = true

		// FIXME: Not actually for Global, but the value is never used
		d.identityAttributeName = "region"

		parseIdentityDuplicateAttrNames(args, implementation, d)

		for k := range args.Keyword {
			switch k {
			default:
				errs = errors.Join(errs, fmt.Errorf("annotation \"@SingletonIdentity\": unexpected keyword parameter %q", k))
			}
		}

	// TODO: allow underscore?
	case "V60SDKv2Fix":
		d.HasV6_0NullValuesError = true

		for k := range args.Keyword {
			switch k {
			case "v60RefreshError":
				attr := args.Keyword[k]
				if b, err := strconv.ParseBool(attr); err != nil {
					errs = errors.Join(errs, err)
				} else {
					d.HasV6_0RefreshError = b
				}

			default:
				errs = errors.Join(errs, fmt.Errorf("annotation \"@V60SDKv2Fix\": unexpected keyword parameter %q", k))
			}
		}
	}

	return errs
}

type GoImport struct {
	Path  string
	Alias string
}

func ParseIdentifierSpec(s string) (string, *GoImport, error) {
	parts := strings.Split(s, ";")
	switch len(parts) {
	case 1:
		return parts[0], nil, nil

	case 2:
		return parts[1], &GoImport{
			Path: parts[0],
		}, nil

	case 3:
		return parts[2], &GoImport{
			Path:  parts[0],
			Alias: parts[1],
		}, nil

	default:
		return "", nil, fmt.Errorf("invalid generator value: %q", s)
	}
}

func ParseBoolAttr(name, value string) (bool, error) {
	if b, err := strconv.ParseBool(value); err != nil {
		return b, fmt.Errorf("invalid %s value %q: Should be boolean value.", name, value)
	} else {
		return b, nil
	}
}

func parseIdentityDuplicateAttrNames(args Args, implementation Implementation, d *ResourceIdentity) {
	var attrs []string
	if attr, ok := args.Keyword["identityDuplicateAttributes"]; ok {
		attrs = strings.Split(attr, ";")
	}
	if implementation == ImplementationSDK {
		attrs = append(attrs, "id")
	}

	// Sort `id` to first position, the rest alphabetically
	slices.SortFunc(attrs, func(a, b string) int {
		if a == "id" {
			return -1
		} else if b == "id" {
			return 1
		} else {
			return strings.Compare(a, b)
		}
	})
	d.IdentityDuplicateAttrNames = slices.Compact(attrs)

	delete(args.Keyword, "identityDuplicateAttributes")
}
