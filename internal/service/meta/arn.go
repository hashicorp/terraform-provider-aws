// TODO: Move this to a shared 'types' package.
package meta

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type arnType uint8

const (
	ARNType arnType = iota
)

var (
	_ xattr.TypeWithValidate = ARNType
)

func (t arnType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t arnType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return ARN{Unknown: true}, nil
	}
	if in.IsNull() {
		return ARN{Null: true}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	a, err := arn.Parse(s)
	if err != nil {
		return nil, err
	}
	return ARN{Value: a}, nil
}

// Equal returns true if `o` is also an ARNType.
func (t arnType) Equal(o attr.Type) bool {
	_, ok := o.(arnType)
	return ok
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t arnType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

// String returns a human-friendly description of the ARNType.
func (t arnType) String() string {
	return "types.ARNType"
}

// Validate implements type validation.
func (t arnType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected String value, received %T with value: %v", in, in),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to arn.ARN: %s", err),
		)
		return diags
	}

	if !arn.IsARN(value) {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as an ARN.", value),
		)
		return diags
	}

	return diags
}

func (t arnType) Description() string {
	return `An Amazon Resource Name.`
}

type ARN struct {
	Unknown bool
	Null    bool
	Value   arn.ARN
}

func (a ARN) Type(_ context.Context) attr.Type {
	return ARNType
}

func (a ARN) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := ARNType.TerraformType(ctx)
	if a.Null {
		return tftypes.NewValue(t, nil), nil
	}
	if a.Unknown {
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	}
	return tftypes.NewValue(t, a.Value.String()), nil
}

// Equal returns true if `other` is a *ARN and has the same value as `a`.
func (a ARN) Equal(other attr.Value) bool {
	o, ok := other.(ARN)
	if !ok {
		return false
	}
	if a.Unknown != o.Unknown {
		return false
	}
	if a.Null != o.Null {
		return false
	}
	return a.Value == o.Value
}

// IsNull returns true if the Value is not set, or is explicitly set to null.
func (a ARN) IsNull() bool {
	return a.Null
}

// IsUnknown returns true if the Value is not yet known.
func (a ARN) IsUnknown() bool {
	return a.Unknown
}

// String returns a summary representation of either the underlying Value,
// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
// or NullValueString (`<null>`) when IsNull() return true.
//
// This is an intentionally lossy representation, that are best suited for
// logging and error reporting, as they are not protected by
// compatibility guarantees within the framework.
func (a ARN) String() string {
	if a.IsUnknown() {
		return attr.UnknownValueString
	}

	if a.IsNull() {
		return attr.NullValueString
	}

	return a.Value.String()
}
