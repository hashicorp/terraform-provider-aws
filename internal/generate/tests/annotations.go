// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	"fmt"
	"strconv"

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
	return nil
}
