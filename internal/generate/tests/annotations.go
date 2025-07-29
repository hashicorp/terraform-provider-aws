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
)

type CommonArgs struct {
	CheckDestroyNoop  bool
	DestroyTakesT     bool
	GoImports         []GoImport
	InitCodeBlocks    []CodeBlock
	AdditionalTfVars_ map[string]TFVar
}

func (d CommonArgs) AdditionalTfVars() map[string]TFVar {
	return tfmaps.ApplyToAllKeys(d.AdditionalTfVars_, func(k string) string {
		return acctestgen.ConstOrQuote(k)
	})
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
