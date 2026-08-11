// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
)

func Validate[T Valueser[T]]() schema.SchemaValidateDiagFunc {
	return validate[T](false)
}

func ValidateIgnoreCase[T Valueser[T]]() schema.SchemaValidateDiagFunc {
	return validate[T](true)
}

func validate[T Valueser[T]](ignoreCase bool) schema.SchemaValidateDiagFunc {
	id := validateIdentity{
		typ:        reflect.TypeFor[T](),
		ignoreCase: ignoreCase,
	}

	s, ok := validateCache.Load(id)
	if ok {
		return s
	}

	// Separates the slow path so that the fast path can be inlined
	return validateSlow[T](id)
}

type validateIdentity struct {
	typ        reflect.Type
	ignoreCase bool
}

var validateCache tfsync.Map[validateIdentity, schema.SchemaValidateDiagFunc]

func validateSlow[T Valueser[T]](id validateIdentity) schema.SchemaValidateDiagFunc {
	s, _ := validateCache.LoadOrStore(
		id,
		validation.ToDiagFunc(validation.StringInSlice(Values[T](), id.ignoreCase)),
	)
	return s
}

func FrameworkValidateIgnoreCase[T Valueser[T]]() validator.String {
	return stringvalidator.OneOfCaseInsensitive(Values[T]()...)
}

// TODO Move to internal/framework/validators or replace with custom types.
func FrameworkValidate[T Valueser[T]]() validator.String {
	return stringvalidator.OneOf(Values[T]()...)
}
