// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
)

type CommonArgs struct {
	CheckDestroyNoop bool
	DestroyTakesT    bool
	GoImports        []GoImport
}

type GoImport struct {
	Path  string
	Alias string
}

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
