// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff

import (
	"context"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func HasChanges(ctx context.Context, plan, state any) ([]fwflex.AutoFlexOptionsFunc, bool) {
	p, s := reflect.ValueOf(plan), reflect.ValueOf(state)
	typeOfP, typesOfS := p.Type(), s.Type()
	var ignoredFields []fwflex.AutoFlexOptionsFunc

	if typeOfP != typesOfS {
		tflog.Debug(ctx, "Type mismatch between plan and state", map[string]any{
			"plan_type":  typeOfP,
			"state_type": typesOfS,
		})
		return nil, false
	}

	var result bool
	for i := 0; i < p.NumField(); i++ {
		name := typeOfP.Field(i).Name
		log.Printf("name: %s", name)
		if name == "Tags" || name == "TagsAll" || name == "Timeouts" {
			//ignoredFields = append(ignoredFields, fwflex.WithIgnoredFieldNamesAppend(name))
			continue
		}

		_, sHasField := typesOfS.FieldByName(name)
		if sHasField {
			typeForP, typeForS := reflect.TypeFor[attr.Value](), reflect.TypeFor[attr.Value]()
			if !p.FieldByName(name).Type().Implements(typeForS) || !s.FieldByName(name).Type().Implements(typeForP) {
				continue
			}

			pValue := p.FieldByName(name).Interface().(attr.Value)
			sValue := s.FieldByName(name).Interface().(attr.Value)

			// check that the types are the same so that they can be compared
			if pValue.Type(ctx) != sValue.Type(ctx) {
				continue
			}

			if ok := !pValue.Equal(sValue); ok {
				result = ok
			} else {
				ignoredFields = append(ignoredFields, fwflex.WithIgnoredFieldNamesAppend(name))
			}
		}
	}

	return ignoredFields, result
}
