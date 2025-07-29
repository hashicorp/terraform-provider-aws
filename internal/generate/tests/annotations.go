// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	"fmt"
	"strconv"
	"strings"

	acctestgen "github.com/hashicorp/terraform-provider-aws/internal/acctest/generate"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
)

type CommonArgs struct {
	// CheckDestroy
	CheckDestroyNoop bool
	DestroyTakesT    bool

	// CheckExists
	HasExistsFunc  bool
	ExistsTypeName string
	ExistsTakesT   bool

	// Import
	NoImport               bool
	ImportStateID          string
	importStateIDAttribute string
	ImportStateIDFunc      string
	ImportIgnore           []string
	plannableImportAction  importAction

	GoImports         []GoImport
	InitCodeBlocks    []CodeBlock
	AdditionalTfVars_ map[string]TFVar
}

func InitCommonArgs() CommonArgs {
	return CommonArgs{
		AdditionalTfVars_: make(map[string]TFVar),
		HasExistsFunc:     true,
	}
}

func (c CommonArgs) HasImportStateIDAttribute() bool {
	return c.importStateIDAttribute != ""
}

func (c CommonArgs) ImportStateIDAttribute() string {
	return namesgen.ConstOrQuote(c.importStateIDAttribute)
}

func (c CommonArgs) HasImportIgnore() bool {
	return len(c.ImportIgnore) > 0
}

func (c CommonArgs) PlannableResourceAction() string {
	return c.plannableImportAction.String()
}

func (c CommonArgs) AdditionalTfVars() map[string]TFVar {
	return tfmaps.ApplyToAllKeys(c.AdditionalTfVars_, func(k string) string {
		return acctestgen.ConstOrQuote(k)
	})
}

type importAction int

const (
	importActionNoop importAction = iota
	importActionUpdate
	importActionReplace
)

func (i importAction) String() string {
	switch i {
	case importActionNoop:
		return "NoOp"

	case importActionUpdate:
		return "Update"

	case importActionReplace:
		return "Replace"

	default:
		return ""
	}
}

type GoImport struct {
	Path  string
	Alias string
}

type CodeBlock struct {
	Code string
}

type TFVar struct {
	GoVarName string
	Type      TFVarType
}

type TFVarType string

const (
	TFVarTypeString TFVarType = "string"
	TFVarTypeInt    TFVarType = "int"
)

func ParseTestingAnnotations(args common.Args, stuff *CommonArgs) error {
	// DestroyCheck
	if attr, ok := args.Keyword["checkDestroyNoop"]; ok {
		if b, err := strconv.ParseBool(attr); err != nil {
			return fmt.Errorf("invalid checkDestroyNoop value: %q: Should be boolean value.", attr)
		} else {
			stuff.CheckDestroyNoop = b
			stuff.GoImports = append(stuff.GoImports,
				GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
	}

	if attr, ok := args.Keyword["destroyTakesT"]; ok {
		if b, err := strconv.ParseBool(attr); err != nil {
			return fmt.Errorf("invalid destroyTakesT value %q: Should be boolean value.", attr)
		} else {
			stuff.DestroyTakesT = b
		}
	}

	// ExistsCheck
	if attr, ok := args.Keyword["hasExistsFunction"]; ok {
		if b, err := strconv.ParseBool(attr); err != nil {
			return fmt.Errorf("invalid existsFunction value %q: Should be boolean value.", attr)
		} else {
			stuff.HasExistsFunc = b
		}
	}

	if attr, ok := args.Keyword["existsType"]; ok {
		if typeName, importSpec, err := ParseIdentifierSpec(attr); err != nil {
			return fmt.Errorf("%s: %w", attr, err)
		} else {
			stuff.ExistsTypeName = typeName
			if importSpec != nil {
				stuff.GoImports = append(stuff.GoImports, *importSpec)
			}
		}
	}

	if attr, ok := args.Keyword["existsTakesT"]; ok {
		if b, err := strconv.ParseBool(attr); err != nil {
			return fmt.Errorf("invalid existsTakesT value %q: Should be boolean value.", attr)
		} else {
			stuff.ExistsTakesT = b
		}
	}

	// Import
	if attr, ok := args.Keyword["importIgnore"]; ok {
		stuff.ImportIgnore = strings.Split(attr, ";")
		for i, val := range stuff.ImportIgnore {
			stuff.ImportIgnore[i] = namesgen.ConstOrQuote(val)
		}
		stuff.plannableImportAction = importActionUpdate
	}
	if attr, ok := args.Keyword["importStateId"]; ok {
		stuff.ImportStateID = attr
	}
	if attr, ok := args.Keyword["importStateIdAttribute"]; ok {
		stuff.importStateIDAttribute = attr
	}
	if attr, ok := args.Keyword["importStateIdFunc"]; ok {
		stuff.ImportStateIDFunc = attr
	}
	if attr, ok := args.Keyword["noImport"]; ok {
		if b, err := strconv.ParseBool(attr); err != nil {
			return fmt.Errorf("invalid noImport value %q: Should be boolean value.", attr)
		} else {
			stuff.NoImport = b
		}
	}
	if attr, ok := args.Keyword["plannableImportAction"]; ok {
		switch attr {
		case importActionNoop.String():
			stuff.plannableImportAction = importActionNoop

		case importActionUpdate.String():
			stuff.plannableImportAction = importActionUpdate

		case importActionReplace.String():
			stuff.plannableImportAction = importActionReplace

		default:
			return fmt.Errorf("invalid plannableImportAction value %q: Must be one of %s.", attr, []string{importActionNoop.String(), importActionUpdate.String(), importActionReplace.String()})
		}
	}

	// TF Variables
	if attr, ok := args.Keyword["emailAddress"]; ok {
		varName := "address"
		if len(attr) > 0 {
			varName = attr
		}
		stuff.GoImports = append(stuff.GoImports,
			GoImport{
				Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf(
				`domain := acctest.RandomDomainName()
%s := acctest.RandomEmailAddress(domain)`, varName),
		})
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: varName,
			Type:      TFVarTypeString,
		}
	}

	if attr, ok := args.Keyword["domainTfVar"]; ok {
		varName := "domain"
		if len(attr) > 0 {
			varName = attr
		}
		stuff.GoImports = append(stuff.GoImports,
			GoImport{
				Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf(`%s := acctest.RandomDomainName()`, varName),
		})
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: varName,
			Type:      TFVarTypeString,
		}
	}

	if attr, ok := args.Keyword["subdomainTfVar"]; ok {
		parentName := "domain"
		varName := "subdomain"
		parts := strings.Split(attr, ";")
		if len(parts) > 1 {
			if len(parts[0]) > 0 {
				parentName = parts[0]
			}
			if len(parts[1]) > 0 {
				varName = parts[1]
			}
		}
		stuff.GoImports = append(stuff.GoImports,
			GoImport{
				Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf(`%s := acctest.RandomDomain()`, parentName),
		})
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf(`%s := %s.RandomSubdomain()`, varName, parentName),
		})
		stuff.AdditionalTfVars_[parentName] = TFVar{
			GoVarName: fmt.Sprintf("%s.String()", parentName),
			Type:      TFVarTypeString,
		}
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: fmt.Sprintf("%s.String()", varName),
			Type:      TFVarTypeString,
		}
	}

	if attr, ok := args.Keyword["randomBgpAsn"]; ok {
		parts := strings.Split(attr, ";")
		varName := "rBgpAsn"
		stuff.GoImports = append(stuff.GoImports,
			GoImport{
				Path:  "github.com/hashicorp/terraform-plugin-testing/helper/acctest",
				Alias: "sdkacctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf("%s := sdkacctest.RandIntRange(%s,%s)", varName, parts[0], parts[1]),
		})
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: varName,
			Type:      TFVarTypeInt,
		}
	}

	if attr, ok := args.Keyword["randomIPv4Address"]; ok {
		varName := "rIPv4Address"
		stuff.GoImports = append(stuff.GoImports,
			GoImport{
				Path:  "github.com/hashicorp/terraform-plugin-testing/helper/acctest",
				Alias: "sdkacctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf(`%s, err := sdkacctest.RandIpAddress("%s")
if err != nil {
	t.Fatal(err)
}
`, varName, attr),
		})
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: varName,
			Type:      TFVarTypeString,
		}
	}

	return nil
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
