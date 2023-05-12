// schema is a high-level framework for easily writing new providers
// for Terraform. Usage of schema is recommended over attempting to write
// to the low-level plugin interfaces manually.
//
// schema breaks down provider creation into simple CRUD operations for
// resources. The logic of diffing, destroying before creating, updating
// or creating, etc. is all handled by the framework. The plugin author
// only needs to implement a configuration schema and the CRUD operations and
// everything else is meant to just work.
//
// A good starting point is to view the Provider structure.
package schema

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/hcl2shim"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Schema describes the structure and type information of a value, whether
// sourced from configuration, plan, or state data. Schema is used in Provider
// and Resource types (for managed resources and data resources) and is
// fundamental to the implementations of ResourceData and ResourceDiff.
//
// The Type field must always be set. At least one of Required, Optional,
// Optional and Computed, or Computed must be enabled unless the Schema is
// directly an implementation of an Elem field of another Schema.
type Schema struct {
	// Type is the type of the value and must be one of the ValueType values.
	//
	// This type not only determines what type is expected/valid in configuring
	// this value, but also what type is returned when ResourceData.Get is
	// called. The types returned by Get are:
	//
	//   TypeBool - bool
	//   TypeInt - int
	//   TypeFloat - float64
	//   TypeString - string
	//   TypeList - []interface{}
	//   TypeMap - map[string]interface{}
	//   TypeSet - *schema.Set
	//
	Type ValueType

	// ConfigMode allows for overriding the default behaviors for mapping
	// schema entries onto configuration constructs.
	//
	// By default, the Elem field is used to choose whether a particular
	// schema is represented in configuration as an attribute or as a nested
	// block; if Elem is a *schema.Resource then it's a block and it's an
	// attribute otherwise.
	//
	// If Elem is *schema.Resource then setting ConfigMode to
	// SchemaConfigModeAttr will force it to be represented in configuration
	// as an attribute, which means that the Computed flag can be used to
	// provide default elements when the argument isn't set at all, while still
	// allowing the user to force zero elements by explicitly assigning an
	// empty list.
	//
	// When Computed is set without Optional, the attribute is not settable
	// in configuration at all and so SchemaConfigModeAttr is the automatic
	// behavior, and SchemaConfigModeBlock is not permitted.
	ConfigMode SchemaConfigMode

	// Required indicates whether the practitioner must enter a value in the
	// configuration for this attribute. Required cannot be used with Computed
	// Default, DefaultFunc, DiffSuppressFunc, DiffSuppressOnRefresh,
	// InputDefault, Optional, or StateFunc. At least one of Required,
	// Optional, Optional and Computed, or Computed must be enabled.
	Required bool

	// Optional indicates whether the practitioner can choose to not enter
	// a value in the configuration for this attribute. Optional cannot be used
	// with Required.
	//
	// If also using Default or DefaultFunc, Computed should also be enabled,
	// otherwise Terraform can output warning logs or "inconsistent result
	// after apply" errors.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this attribute or not. Computed cannot be used with Required. If
	// Required and Optional are both false, the attribute will be considered
	// "read only" for the practitioner, with only the provider able to set
	// its value.
	Computed bool

	// ForceNew indicates whether a change in this value requires the
	// replacement (destroy and create) of the managed resource instance,
	// rather than an in-place update. This field is only valid when the
	// encapsulating Resource is a managed resource.
	//
	// If conditional replacement logic is needed, use the Resource type
	// CustomizeDiff field to call the ResourceDiff type ForceNew method.
	ForceNew bool

	// If this is non-nil, the provided function will be used during diff
	// of this field. If this is nil, a default diff for the type of the
	// schema will be used.
	//
	// This allows comparison based on something other than primitive, list
	// or map equality - for example SSH public keys may be considered
	// equivalent regardless of trailing whitespace.
	//
	// If CustomizeDiffFunc makes this field ForceNew=true, the
	// following DiffSuppressFunc will come in with the value of old being
	// empty, as if creating a new resource.
	//
	// By default, DiffSuppressFunc is considered only when deciding whether
	// a configuration value is significantly different than the prior state
	// value during planning. Set DiffSuppressOnRefresh to opt in to checking
	// this also during the refresh step.
	DiffSuppressFunc SchemaDiffSuppressFunc

	// DiffSuppressOnRefresh enables using the DiffSuppressFunc to ignore
	// normalization-classified changes returned by the resource type's
	// "Read" or "ReadContext" function, in addition to the default behavior of
	// doing so during planning.
	//
	// This is a particularly good choice for attributes which take strings
	// containing "microsyntaxes" where various different values are packed
	// together in some serialization where there are many ways to express the
	// same information. For example, attributes which accept JSON data can
	// include different whitespace characters without changing meaning, and
	// case-insensitive identifiers may refer to the same object using different
	// characters.
	//
	// This is valid only for attributes of primitive types, because
	// DiffSuppressFunc itself is only compatible with primitive types.
	//
	// The key benefit of activating this flag is that the result of Read or
	// ReadContext will be cleaned of normalization-only changes in the same
	// way as the planning result would normaly be, which therefore prevents
	// churn for downstream expressions deriving from this attribute and
	// prevents incorrect "Values changed outside of Terraform" messages
	// when the remote API returns values which have the same meaning as the
	// prior state but in a different serialization.
	//
	// This is an opt-in because it was a later addition to the DiffSuppressFunc
	// functionality which would cause some significant changes in behavior
	// for existing providers if activated everywhere all at once.
	DiffSuppressOnRefresh bool

	// Default indicates a value to set if this attribute is not set in the
	// configuration. Default cannot be used with DefaultFunc or Required.
	// Default is only supported if the Type is TypeBool, TypeFloat, TypeInt,
	// or TypeString. Default cannot be used if the Schema is directly an
	// implementation of an Elem field of another Schema, such as trying to
	// set a default value for a TypeList or TypeSet.
	//
	// Changing either Default can be a breaking change, especially if the
	// attribute has ForceNew enabled. If a default needs to change to align
	// with changing assumptions in an upstream API, then it may be necessary
	// to also implement resource state upgrade functionality to change the
	// state to match or update read operation logic to align with the new
	// default.
	Default interface{}

	// DefaultFunc can be specified to compute a dynamic default when this
	// attribute is not set in the configuration. DefaultFunc cannot be used
	// with Default. For legacy reasons, DefaultFunc can be used with Required
	// attributes in a Provider schema, which will prompt practitioners for
	// input if the result of this function is nil.
	//
	// The return value should be stable to avoid generating confusing
	// plan differences. Changing the return value can be a breaking change,
	// especially if ForceNew is enabled. If a default needs to change to align
	// with changing assumptions in an upstream API, then it may be necessary
	// to also implement resource state upgrade functionality to change the
	// state to match or update read operation logic to align with the new
	// default.
	DefaultFunc SchemaDefaultFunc

	// Description is used as the description for docs, the language server and
	// other user facing usage. It can be plain-text or markdown depending on the
	// global DescriptionKind setting.
	Description string

	// InputDefault is the default value to use for when inputs are requested.
	// This differs from Default in that if Default is set, no input is
	// asked for. If Input is asked, this will be the default value offered.
	InputDefault string

	// StateFunc is a function called to change the value of this before
	// storing it in the state (and likewise before comparing for diffs).
	// The use for this is for example with large strings, you may want
	// to simply store the hash of it.
	StateFunc SchemaStateFunc

	// Elem represents the element type for a TypeList, TypeSet, or TypeMap
	// attribute or block. The only valid types are *Schema and *Resource.
	// Only TypeList and TypeSet support *Resource.
	//
	// If the Elem is a *Schema, the surrounding Schema represents a single
	// attribute with a single element type for underlying elements. In
	// practitioner configurations, an equals sign (=) is required to set
	// the value. Refer to the following documentation:
	//
	//   https://www.terraform.io/docs/language/syntax/configuration.html
	//
	// The underlying *Schema is only required to implement Type. ValidateFunc
	// or ValidateDiagFunc can be used to validate each element value.
	//
	// If the Elem is a *Resource, the surrounding Schema represents a
	// configuration block. Blocks can contain underlying attributes or blocks.
	// In practitioner configurations, an equals sign (=) cannot be used to
	// set the value. Blocks are instead repeated as necessary, or require
	// the use of dynamic block expressions. Refer to the following
	// documentation:
	//
	//   https://www.terraform.io/docs/language/syntax/configuration.html
	//   https://www.terraform.io/docs/language/expressions/dynamic-blocks.html
	//
	// The underlying *Resource must only implement the Schema field.
	Elem interface{}

	// MaxItems defines a maximum amount of items that can exist within a
	// TypeSet or TypeList.
	MaxItems int

	// MinItems defines a minimum amount of items that can exist within a
	// TypeSet or TypeList.
	//
	// If the field Optional is set to true then MinItems is ignored and thus
	// effectively zero.
	MinItems int

	// Set defines custom hash algorithm for each TypeSet element. If not
	// defined, the SDK implements a default hash algorithm based on the
	// underlying structure and type information of the Elem field.
	Set SchemaSetFunc

	// ComputedWhen is a set of queries on the configuration. Whenever any
	// of these things is changed, it will require a recompute (this requires
	// that Computed is set to true).
	//
	// Deprecated: This functionality is not implemented and this field
	// declaration should be removed.
	ComputedWhen []string

	// ConflictsWith is a set of attribute paths, including this attribute,
	// whose configurations cannot be set simultaneously. This implements the
	// validation logic declaratively within the schema and can trigger earlier
	// in Terraform operations, rather than using create or update logic which
	// only triggers during apply.
	//
	// Only absolute attribute paths, ones starting with top level attribute
	// names, are supported. Attribute paths cannot be accurately declared
	// for TypeList (if MaxItems is greater than 1), TypeMap, or TypeSet
	// attributes. To reference an attribute under a single configuration block
	// (TypeList with Elem of *Resource and MaxItems of 1), the syntax is
	// "parent_block_name.0.child_attribute_name".
	ConflictsWith []string

	// ExactlyOneOf is a set of attribute paths, including this attribute,
	// where only one attribute out of all specified can be configured. It will
	// return a validation error if none are specified as well. This implements
	// the validation logic declaratively within the schema and can trigger
	// earlier in Terraform operations, rather than using create or update
	// logic which only triggers during apply.
	//
	// Only absolute attribute paths, ones starting with top level attribute
	// names, are supported. Attribute paths cannot be accurately declared
	// for TypeList (if MaxItems is greater than 1), TypeMap, or TypeSet
	// attributes. To reference an attribute under a single configuration block
	// (TypeList with Elem of *Resource and MaxItems of 1), the syntax is
	// "parent_block_name.0.child_attribute_name".
	ExactlyOneOf []string

	// AtLeastOneOf is a set of attribute paths, including this attribute,
	// in which at least one of the attributes must be configured. This
	// implements the validation logic declaratively within the schema and can
	// trigger earlier in Terraform operations, rather than using create or
	// update logic which only triggers during apply.
	//
	// Only absolute attribute paths, ones starting with top level attribute
	// names, are supported. Attribute paths cannot be accurately declared
	// for TypeList (if MaxItems is greater than 1), TypeMap, or TypeSet
	// attributes. To reference an attribute under a single configuration block
	// (TypeList with Elem of *Resource and MaxItems of 1), the syntax is
	// "parent_block_name.0.child_attribute_name".
	AtLeastOneOf []string

	// RequiredWith is a set of attribute paths, including this attribute,
	// that must be set simultaneously. This implements the validation logic
	// declaratively within the schema and can trigger earlier in Terraform
	// operations, rather than using create or update logic which only triggers
	// during apply.
	//
	// Only absolute attribute paths, ones starting with top level attribute
	// names, are supported. Attribute paths cannot be accurately declared
	// for TypeList (if MaxItems is greater than 1), TypeMap, or TypeSet
	// attributes. To reference an attribute under a single configuration block
	// (TypeList with Elem of *Resource and MaxItems of 1), the syntax is
	// "parent_block_name.0.child_attribute_name".
	RequiredWith []string

	// Deprecated defines warning diagnostic details to display when
	// practitioner configurations use this attribute or block. The warning
	// diagnostic summary is automatically set to "Argument is deprecated"
	// along with configuration source file and line information.
	//
	// Set this field to a practitioner actionable message such as:
	//
	//  - "Configure other_attribute instead. This attribute will be removed
	//    in the next major version of the provider."
	//  - "Remove this attribute's configuration as it no longer is used and
	//    the attribute will be removed in the next major version of the
	//    provider."
	//
	// In Terraform 1.2.7 and later, this warning diagnostic is displayed any
	// time a practitioner attempts to configure a known value for this
	// attribute and certain scenarios where this attribute is referenced.
	//
	// In Terraform 1.2.6 and earlier, this warning diagnostic is only
	// displayed when the attribute is Required or Optional, and if the
	// practitioner configuration attempts to set the attribute value to a
	// known value. It cannot detect practitioner configuration values that
	// are unknown ("known after apply").
	//
	// Additional information about deprecation enhancements for read-only
	// attributes can be found in:
	//
	//  - https://github.com/hashicorp/terraform/issues/7569
	Deprecated string

	// ValidateFunc allows individual fields to define arbitrary validation
	// logic. It is yielded the provided config value as an interface{} that is
	// guaranteed to be of the proper Schema type, and it can yield warnings or
	// errors based on inspection of that value.
	//
	// ValidateFunc is honored only when the schema's Type is set to TypeInt,
	// TypeFloat, TypeString, TypeBool, or TypeMap. It is ignored for all other types.
	ValidateFunc SchemaValidateFunc

	// ValidateDiagFunc allows individual fields to define arbitrary validation
	// logic. It is yielded the provided config value as an interface{} that is
	// guaranteed to be of the proper Schema type, and it can yield diagnostics
	// based on inspection of that value.
	//
	// ValidateDiagFunc is honored only when the schema's Type is set to TypeInt,
	// TypeFloat, TypeString, TypeBool, or TypeMap. It is ignored for all other types.
	//
	// ValidateDiagFunc is also yielded the cty.Path the SDK has built up to this
	// attribute. The SDK will automatically set the AttributePath of any returned
	// Diagnostics to this path. Therefore the developer does not need to set
	// the AttributePath for primitive types.
	//
	// In the case of TypeMap to provide the most precise information, please
	// set an AttributePath with the additional cty.IndexStep:
	//
	//  AttributePath: cty.IndexStringPath("key_name")
	//
	// Or alternatively use the passed in path to create the absolute path:
	//
	//  AttributePath: append(path, cty.IndexStep{Key: cty.StringVal("key_name")})
	ValidateDiagFunc SchemaValidateDiagFunc

	// Sensitive ensures that the attribute's value does not get displayed in
	// the Terraform user interface output. It should be used for password or
	// other values which should be hidden.
	//
	// Terraform does not support conditional sensitivity, so if the value may
	// only be sensitive in certain scenarios, a pragmatic choice will be
	// necessary upfront of whether or not to always hide the value. Some
	// providers may opt to split up resources based on sensitivity, to ensure
	// that practitioners without sensitive values do not have values
	// unnecessarily hidden.
	//
	// Terraform does not support passing sensitivity from configurations to
	// providers. For example, if a sensitive value is configured via another
	// attribute, this attribute is not marked Sensitive, and the value is used
	// in this attribute value, the sensitivity is not transitive. The value
	// will be displayed as normal.
	//
	// Sensitive values propagate when referenced in other parts of a
	// configuration unless the nonsensitive() configuration function is used.
	// Certain configuration usage may also expand the sensitivity. For
	// example, including the sensitive value in a set may mark the whole set
	// as sensitive. Any outputs containing a sensitive value must enable the
	// output sensitive argument.
	Sensitive bool
}

// SchemaConfigMode is used to influence how a schema item is mapped into a
// corresponding configuration construct, using the ConfigMode field of
// Schema.
type SchemaConfigMode int

const (
	SchemaConfigModeAuto SchemaConfigMode = iota
	SchemaConfigModeAttr
	SchemaConfigModeBlock
)

// SchemaDiffSuppressFunc is a function which can be used to determine
// whether a detected diff on a schema element is "valid" or not, and
// suppress it from the plan if necessary.
//
// Return true if the diff should be suppressed, false to retain it.
type SchemaDiffSuppressFunc func(k, oldValue, newValue string, d *ResourceData) bool

// SchemaDefaultFunc is a function called to return a default value for
// a field.
type SchemaDefaultFunc func() (interface{}, error)

// EnvDefaultFunc is a helper function that returns the value of the
// given environment variable, if one exists, or the default value
// otherwise.
func EnvDefaultFunc(k string, dv interface{}) SchemaDefaultFunc {
	return func() (interface{}, error) {
		if v := os.Getenv(k); v != "" {
			return v, nil
		}

		return dv, nil
	}
}

// MultiEnvDefaultFunc is a helper function that returns the value of the first
// environment variable in the given list that returns a non-empty value. If
// none of the environment variables return a value, the default value is
// returned.
func MultiEnvDefaultFunc(ks []string, dv interface{}) SchemaDefaultFunc {
	return func() (interface{}, error) {
		for _, k := range ks {
			if v := os.Getenv(k); v != "" {
				return v, nil
			}
		}
		return dv, nil
	}
}

// SchemaSetFunc is a function that must return a unique ID for the given
// element. This unique ID is used to store the element in a hash.
type SchemaSetFunc func(interface{}) int

// SchemaStateFunc is a function used to convert some type to a string
// to be stored in the state.
type SchemaStateFunc func(interface{}) string

// SchemaValidateFunc is a function used to validate a single field in the
// schema.
//
// Deprecated: please use SchemaValidateDiagFunc
type SchemaValidateFunc func(interface{}, string) ([]string, []error)

// SchemaValidateDiagFunc is a function used to validate a single field in the
// schema and has Diagnostic support.
type SchemaValidateDiagFunc func(interface{}, cty.Path) diag.Diagnostics

func (s *Schema) GoString() string {
	return fmt.Sprintf("*%#v", *s)
}

// Returns a default value for this schema by either reading Default or
// evaluating DefaultFunc. If neither of these are defined, returns nil.
func (s *Schema) DefaultValue() (interface{}, error) {
	if s.Default != nil {
		return s.Default, nil
	}

	if s.DefaultFunc != nil {
		defaultValue, err := s.DefaultFunc()
		if err != nil {
			return nil, fmt.Errorf("error loading default: %s", err)
		}
		return defaultValue, nil
	}

	return nil, nil
}

// Returns a zero value for the schema.
func (s *Schema) ZeroValue() interface{} {
	// If it's a set then we'll do a bit of extra work to provide the
	// right hashing function in our empty value.
	if s.Type == TypeSet {
		setFunc := s.Set
		if setFunc == nil {
			// Default set function uses the schema to hash the whole value
			elem := s.Elem
			switch t := elem.(type) {
			case *Schema:
				setFunc = HashSchema(t)
			case *Resource:
				setFunc = HashResource(t)
			default:
				panic("invalid set element type")
			}
		}
		return &Set{F: setFunc}
	} else {
		return s.Type.Zero()
	}
}

func (s *Schema) finalizeDiff(d *terraform.ResourceAttrDiff, customized bool) *terraform.ResourceAttrDiff {
	if d == nil {
		return d
	}

	if s.Type == TypeBool {
		normalizeBoolString := func(s string) string {
			switch s {
			case "0":
				return "false"
			case "1":
				return "true"
			}
			return s
		}
		d.Old = normalizeBoolString(d.Old)
		d.New = normalizeBoolString(d.New)
	}

	if s.Computed && !d.NewRemoved && d.New == "" {
		// Computed attribute without a new value set
		d.NewComputed = true
	}

	if s.ForceNew {
		// ForceNew, mark that this field is requiring new under the
		// following conditions, explained below:
		//
		//   * Old != New - There is a change in value. This field
		//       is therefore causing a new resource.
		//
		//   * NewComputed - This field is being computed, hence a
		//       potential change in value, mark as causing a new resource.
		d.RequiresNew = d.Old != d.New || d.NewComputed
	}

	if d.NewRemoved {
		return d
	}

	if s.Computed {
		// FIXME: This is where the customized bool from getChange finally
		//        comes into play.  It allows the previously incorrect behavior
		//        of an empty string being used as "unset" when the value is
		//        computed. This should be removed once we can properly
		//        represent an unset/nil value from the configuration.
		if !customized {
			if d.Old != "" && d.New == "" {
				// This is a computed value with an old value set already,
				// just let it go.
				log.Println("[DEBUG] A computed value with the empty string as the new value and a non-empty old value was found. Interpreting the empty string as \"unset\" to align with legacy behavior.")
				return nil
			}
		}

		if d.New == "" && !d.NewComputed {
			// Computed attribute without a new value set
			d.NewComputed = true
		}
	}

	if s.Sensitive {
		// Set the Sensitive flag so output is hidden in the UI
		d.Sensitive = true
	}

	return d
}

func (s *Schema) validateFunc(decoded interface{}, k string, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if s.ValidateDiagFunc != nil {
		diags = s.ValidateDiagFunc(decoded, path)
		for i := range diags {
			if !diags[i].AttributePath.HasPrefix(path) {
				diags[i].AttributePath = append(path, diags[i].AttributePath...)
			}
		}
	} else if s.ValidateFunc != nil {
		ws, es := s.ValidateFunc(decoded, k)
		for _, w := range ws {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       w,
				AttributePath: path,
			})
		}
		for _, e := range es {
			diags = append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       e.Error(),
				AttributePath: path,
			})
		}
	}

	return diags
}

// InternalMap is used to aid in the transition to the new schema types and
// protocol. The name is not meant to convey any usefulness, as this is not to
// be used directly by any providers.
type InternalMap = schemaMap

// schemaMap is a wrapper that adds nice functions on top of schemas.
type schemaMap map[string]*Schema

func (m schemaMap) panicOnError() bool {
	return os.Getenv("TF_ACC") != ""
}

// Data returns a ResourceData for the given schema, state, and diff.
//
// The diff is optional.
func (m schemaMap) Data(
	s *terraform.InstanceState,
	d *terraform.InstanceDiff) (*ResourceData, error) {
	return &ResourceData{
		schema:       m,
		state:        s,
		diff:         d,
		panicOnError: m.panicOnError(),
	}, nil
}

// DeepCopy returns a copy of this schemaMap. The copy can be safely modified
// without affecting the original.
func (m *schemaMap) DeepCopy() schemaMap {
	copiedMap, err := copystructure.Config{Lock: true}.Copy(m)
	if err != nil {
		panic(err)
	}
	return *copiedMap.(*schemaMap)
}

// Diff returns the diff for a resource given the schema map,
// state, and configuration.
func (m schemaMap) Diff(
	ctx context.Context,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig,
	customizeDiff CustomizeDiffFunc,
	meta interface{},
	handleRequiresNew bool) (*terraform.InstanceDiff, error) {
	result := new(terraform.InstanceDiff)
	result.Attributes = make(map[string]*terraform.ResourceAttrDiff)

	// Make sure to mark if the resource is tainted
	if s != nil {
		result.DestroyTainted = s.Tainted
		result.RawConfig = s.RawConfig
		result.RawState = s.RawState
		result.RawPlan = s.RawPlan
	}

	d := &ResourceData{
		schema:       m,
		state:        s,
		config:       c,
		panicOnError: m.panicOnError(),
	}

	for k, schema := range m {
		err := m.diff(ctx, k, schema, result, d, false)
		if err != nil {
			return nil, err
		}
	}

	// Remove any nil diffs just to keep things clean
	for k, v := range result.Attributes {
		if v == nil {
			delete(result.Attributes, k)
		}
	}

	// If this is a non-destroy diff, call any custom diff logic that has been
	// defined.
	if !result.DestroyTainted && customizeDiff != nil {
		mc := m.DeepCopy()
		rd := newResourceDiff(mc, c, s, result)

		logging.HelperSchemaTrace(ctx, "Calling downstream")
		err := customizeDiff(ctx, rd, meta)
		logging.HelperSchemaTrace(ctx, "Called downstream")

		if err != nil {
			return nil, err
		}
		for _, k := range rd.UpdatedKeys() {
			err := m.diff(ctx, k, mc[k], result, rd, false)
			if err != nil {
				return nil, err
			}
		}
	}

	if handleRequiresNew {
		// If the diff requires a new resource, then we recompute the diff
		// so we have the complete new resource diff, and preserve the
		// RequiresNew fields where necessary so the user knows exactly what
		// caused that.
		if result.RequiresNew() {
			// Create the new diff
			result2 := new(terraform.InstanceDiff)
			result2.Attributes = make(map[string]*terraform.ResourceAttrDiff)

			// Preserve the DestroyTainted flag
			result2.DestroyTainted = result.DestroyTainted
			result2.RawConfig = result.RawConfig
			result2.RawPlan = result.RawPlan
			result2.RawState = result.RawState

			// Reset the data to not contain state. We have to call init()
			// again in order to reset the FieldReaders.
			d.state = nil
			d.init()

			// Perform the diff again
			for k, schema := range m {
				err := m.diff(ctx, k, schema, result2, d, false)
				if err != nil {
					return nil, err
				}
			}

			// Re-run customization
			if !result2.DestroyTainted && customizeDiff != nil {
				mc := m.DeepCopy()
				rd := newResourceDiff(mc, c, d.state, result2)
				if err := customizeDiff(ctx, rd, meta); err != nil {
					return nil, err
				}
				for _, k := range rd.UpdatedKeys() {
					err := m.diff(ctx, k, mc[k], result2, rd, false)
					if err != nil {
						return nil, err
					}
				}
			}

			// Force all the fields to not force a new since we know what we
			// want to force new.
			for k, attr := range result2.Attributes {
				if attr == nil {
					continue
				}

				if attr.RequiresNew {
					attr.RequiresNew = false
				}

				if s != nil {
					attr.Old = s.Attributes[k]
				}
			}

			// Now copy in all the requires new diffs...
			for k, attr := range result.Attributes {
				if attr == nil {
					continue
				}

				newAttr, ok := result2.Attributes[k]
				if !ok {
					newAttr = attr
				}

				if attr.RequiresNew {
					newAttr.RequiresNew = true
				}

				result2.Attributes[k] = newAttr
			}

			// And set the diff!
			result = result2
		}

	}

	// Go through and detect all of the ComputedWhens now that we've
	// finished the diff.
	// TODO

	if result.Empty() {
		// If we don't have any diff elements, just return nil
		return nil, nil
	}

	return result, nil
}

// Validate validates the configuration against this schema mapping.
func (m schemaMap) Validate(c *terraform.ResourceConfig) diag.Diagnostics {
	return m.validateObject("", m, c, cty.Path{})
}

// InternalValidate validates the format of this schema. This should be called
// from a unit test (and not in user-path code) to verify that a schema
// is properly built.
func (m schemaMap) InternalValidate(topSchemaMap schemaMap) error {
	return m.internalValidate(topSchemaMap, false)
}

func (m schemaMap) internalValidate(topSchemaMap schemaMap, attrsOnly bool) error {
	if topSchemaMap == nil {
		topSchemaMap = m
	}
	for k, v := range m {
		if v.Type == TypeInvalid {
			return fmt.Errorf("%s: Type must be specified", k)
		}

		if v.Optional && v.Required {
			return fmt.Errorf("%s: Optional or Required must be set, not both", k)
		}

		if v.Required && v.Computed {
			return fmt.Errorf("%s: Cannot be both Required and Computed", k)
		}

		if !v.Required && !v.Optional && !v.Computed {
			return fmt.Errorf("%s: One of optional, required, or computed must be set", k)
		}

		computedOnly := v.Computed && !v.Optional

		switch v.ConfigMode {
		case SchemaConfigModeBlock:
			if _, ok := v.Elem.(*Resource); !ok {
				return fmt.Errorf("%s: ConfigMode of block is allowed only when Elem is *schema.Resource", k)
			}
			if attrsOnly {
				return fmt.Errorf("%s: ConfigMode of block cannot be used in child of schema with ConfigMode of attribute", k)
			}
			if computedOnly {
				return fmt.Errorf("%s: ConfigMode of block cannot be used for computed schema", k)
			}
		case SchemaConfigModeAttr:
			// anything goes
		case SchemaConfigModeAuto:
			// Since "Auto" for Elem: *Resource would create a nested block,
			// and that's impossible inside an attribute, we require it to be
			// explicitly overridden as mode "Attr" for clarity.
			if _, ok := v.Elem.(*Resource); ok {
				if attrsOnly {
					return fmt.Errorf("%s: in *schema.Resource with ConfigMode of attribute, so must also have ConfigMode of attribute", k)
				}
			}
		default:
			return fmt.Errorf("%s: invalid ConfigMode value", k)
		}

		if v.Computed && v.Default != nil {
			return fmt.Errorf("%s: Default must be nil if computed", k)
		}

		if v.Required && v.Default != nil {
			return fmt.Errorf("%s: Default cannot be set with Required", k)
		}

		if len(v.ComputedWhen) > 0 && !v.Computed {
			return fmt.Errorf("%s: ComputedWhen can only be set with Computed", k)
		}

		if len(v.ConflictsWith) > 0 && v.Required {
			return fmt.Errorf("%s: ConflictsWith cannot be set with Required", k)
		}

		if len(v.ExactlyOneOf) > 0 && v.Required {
			return fmt.Errorf("%s: ExactlyOneOf cannot be set with Required", k)
		}

		if len(v.AtLeastOneOf) > 0 && v.Required {
			return fmt.Errorf("%s: AtLeastOneOf cannot be set with Required", k)
		}

		if len(v.ConflictsWith) > 0 {
			err := checkKeysAgainstSchemaFlags(k, v.ConflictsWith, topSchemaMap, v, false)
			if err != nil {
				return fmt.Errorf("ConflictsWith: %+v", err)
			}
		}

		if len(v.RequiredWith) > 0 {
			err := checkKeysAgainstSchemaFlags(k, v.RequiredWith, topSchemaMap, v, true)
			if err != nil {
				return fmt.Errorf("RequiredWith: %+v", err)
			}
		}

		if len(v.ExactlyOneOf) > 0 {
			err := checkKeysAgainstSchemaFlags(k, v.ExactlyOneOf, topSchemaMap, v, true)
			if err != nil {
				return fmt.Errorf("ExactlyOneOf: %+v", err)
			}
		}

		if len(v.AtLeastOneOf) > 0 {
			err := checkKeysAgainstSchemaFlags(k, v.AtLeastOneOf, topSchemaMap, v, true)
			if err != nil {
				return fmt.Errorf("AtLeastOneOf: %+v", err)
			}
		}

		if v.DiffSuppressOnRefresh && v.DiffSuppressFunc == nil {
			return fmt.Errorf("%s: cannot set DiffSuppressOnRefresh without DiffSuppressFunc", k)
		}

		if v.Type == TypeList || v.Type == TypeSet {
			if v.Elem == nil {
				return fmt.Errorf("%s: Elem must be set for lists", k)
			}

			if v.Default != nil {
				return fmt.Errorf("%s: Default is not valid for lists or sets", k)
			}

			if v.Type != TypeSet && v.Set != nil {
				return fmt.Errorf("%s: Set can only be set for TypeSet", k)
			}

			switch t := v.Elem.(type) {
			case *Resource:
				attrsOnly := attrsOnly || v.ConfigMode == SchemaConfigModeAttr

				if err := schemaMap(t.Schema).internalValidate(topSchemaMap, attrsOnly); err != nil {
					return err
				}
			case *Schema:
				bad := t.Computed || t.Optional || t.Required
				if bad {
					return fmt.Errorf(
						"%s: Elem must have only Type set", k)
				}
			}
		} else {
			if v.MaxItems > 0 || v.MinItems > 0 {
				return fmt.Errorf("%s: MaxItems and MinItems are only supported on lists or sets", k)
			}
		}

		if v.Type == TypeMap && v.Elem != nil {
			switch v.Elem.(type) {
			case *Resource:
				return fmt.Errorf("%s: TypeMap with Elem *Resource not supported,"+
					"use TypeList/TypeSet with Elem *Resource or TypeMap with Elem *Schema", k)
			}
		}

		if computedOnly {
			if len(v.AtLeastOneOf) > 0 {
				return fmt.Errorf("%s: AtLeastOneOf is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if len(v.ConflictsWith) > 0 {
				return fmt.Errorf("%s: ConflictsWith is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.Default != nil {
				return fmt.Errorf("%s: Default is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.DefaultFunc != nil {
				return fmt.Errorf("%s: DefaultFunc is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.DiffSuppressFunc != nil {
				return fmt.Errorf("%s: DiffSuppressFunc is for suppressing differences"+
					" between config and state representation. "+
					"There is no config for computed-only field, nothing to compare.", k)
			}
			if len(v.ExactlyOneOf) > 0 {
				return fmt.Errorf("%s: ExactlyOneOf is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.InputDefault != "" {
				return fmt.Errorf("%s: InputDefault is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.MaxItems > 0 {
				return fmt.Errorf("%s: MaxItems is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.MinItems > 0 {
				return fmt.Errorf("%s: MinItems is for configurable attributes,"+
					"there's nothing to configure on computed-only field", k)
			}
			if v.StateFunc != nil {
				return fmt.Errorf("%s: StateFunc is extraneous, "+
					"value should just be changed before setting on computed-only field", k)
			}
			if v.ValidateFunc != nil {
				return fmt.Errorf("%s: ValidateFunc is for validating user input, "+
					"there's nothing to validate on computed-only field", k)
			}
			if v.ValidateDiagFunc != nil {
				return fmt.Errorf("%s: ValidateDiagFunc is for validating user input, "+
					"there's nothing to validate on computed-only field", k)
			}
		}

		if v.ValidateFunc != nil || v.ValidateDiagFunc != nil {
			switch v.Type {
			case TypeList, TypeSet:
				return fmt.Errorf("%s: ValidateFunc and ValidateDiagFunc are not yet supported on lists or sets.", k)
			}
		}

		if v.ValidateFunc != nil && v.ValidateDiagFunc != nil {
			return fmt.Errorf("%s: ValidateFunc and ValidateDiagFunc cannot both be set", k)
		}

		if v.Deprecated == "" {
			if !isValidFieldName(k) {
				return fmt.Errorf("%s: Field name may only contain lowercase alphanumeric characters & underscores.", k)
			}
		}
	}

	return nil
}

func checkKeysAgainstSchemaFlags(k string, keys []string, topSchemaMap schemaMap, self *Schema, allowSelfReference bool) error {
	for _, key := range keys {
		parts := strings.Split(key, ".")
		sm := topSchemaMap
		var target *Schema
		for idx, part := range parts {
			// Skip index fields if 0
			partInt, err := strconv.Atoi(part)

			if err == nil {
				if partInt != 0 {
					return fmt.Errorf("%s configuration block reference (%s) can only use the .0. index for TypeList and MaxItems: 1 configuration blocks", k, key)
				}

				continue
			}

			var ok bool
			if target, ok = sm[part]; !ok {
				return fmt.Errorf("%s references unknown attribute (%s) at part (%s)", k, key, part)
			}

			subResource, ok := target.Elem.(*Resource)

			if !ok {
				continue
			}

			// Skip Type/MaxItems check if not the last element
			if (target.Type == TypeSet || target.MaxItems != 1) && idx+1 != len(parts) {
				return fmt.Errorf("%s configuration block reference (%s) can only be used with TypeList and MaxItems: 1 configuration blocks", k, key)
			}

			sm = schemaMap(subResource.Schema)
		}

		if target == nil {
			return fmt.Errorf("%s cannot find target attribute (%s), sm: %#v", k, key, sm)
		}

		if target == self && !allowSelfReference {
			return fmt.Errorf("%s cannot reference self (%s)", k, key)
		}

		if target.Required {
			return fmt.Errorf("%s cannot contain Required attribute (%s)", k, key)
		}

		if len(target.ComputedWhen) > 0 {
			return fmt.Errorf("%s cannot contain Computed(When) attribute (%s)", k, key)
		}
	}

	return nil
}

var validFieldNameRe = regexp.MustCompile("^[a-z0-9_]+$")

func isValidFieldName(name string) bool {
	return validFieldNameRe.MatchString(name)
}

// resourceDiffer is an interface that is used by the private diff functions.
// This helps facilitate diff logic for both ResourceData and ResoureDiff with
// minimal divergence in code.
type resourceDiffer interface {
	diffChange(string) (interface{}, interface{}, bool, bool, bool)
	Get(string) interface{}
	GetChange(string) (interface{}, interface{})
	GetOk(string) (interface{}, bool)
	HasChange(string) bool
	HasChanges(...string) bool
	Id() string
}

func (m schemaMap) diff(
	ctx context.Context,
	k string,
	schema *Schema,
	diff *terraform.InstanceDiff,
	d resourceDiffer,
	all bool) error {

	unsupressedDiff := new(terraform.InstanceDiff)
	unsupressedDiff.Attributes = make(map[string]*terraform.ResourceAttrDiff)

	var err error
	switch schema.Type {
	case TypeBool, TypeInt, TypeFloat, TypeString:
		err = m.diffString(k, schema, unsupressedDiff, d, all)
	case TypeList:
		err = m.diffList(ctx, k, schema, unsupressedDiff, d, all)
	case TypeMap:
		err = m.diffMap(k, schema, unsupressedDiff, d, all)
	case TypeSet:
		err = m.diffSet(ctx, k, schema, unsupressedDiff, d, all)
	default:
		err = fmt.Errorf("%s: unknown type %#v", k, schema.Type)
	}

	for attrK, attrV := range unsupressedDiff.Attributes {
		switch rd := d.(type) {
		case *ResourceData:
			if schema.DiffSuppressFunc != nil && attrV != nil &&
				schema.DiffSuppressFunc(attrK, attrV.Old, attrV.New, rd) {
				// If this attr diff is suppressed, we may still need it in the
				// overall diff if it's contained within a set. Rather than
				// dropping the diff, make it a NOOP.
				if !all {
					continue
				}

				logging.HelperSchemaDebug(ctx, "Ignoring change due to DiffSuppressFunc", map[string]interface{}{logging.KeyAttributePath: attrK})
				attrV = &terraform.ResourceAttrDiff{
					Old: attrV.Old,
					New: attrV.Old,
				}
			}
		}
		diff.Attributes[attrK] = attrV
	}

	return err
}

func (m schemaMap) diffList(
	ctx context.Context,
	k string,
	schema *Schema,
	diff *terraform.InstanceDiff,
	d resourceDiffer,
	all bool) error {
	o, n, _, computedList, customized := d.diffChange(k)
	if computedList {
		n = nil
	}
	nSet := n != nil

	// If we have an old value and no new value is set or will be
	// computed once all variables can be interpolated and we're
	// computed, then nothing has changed.
	if o != nil && n == nil && !computedList && schema.Computed {
		return nil
	}

	if o == nil {
		o = []interface{}{}
	}
	if n == nil {
		n = []interface{}{}
	}
	if s, ok := o.(*Set); ok {
		o = s.List()
	}
	if s, ok := n.(*Set); ok {
		n = s.List()
	}
	os := o.([]interface{})
	vs := n.([]interface{})

	// If the new value was set, and the two are equal, then we're done.
	// We have to do this check here because sets might be NOT
	// reflect.DeepEqual so we need to wait until we get the []interface{}
	if !all && nSet && reflect.DeepEqual(os, vs) {
		return nil
	}

	// Get the counts
	oldLen := len(os)
	newLen := len(vs)
	oldStr := strconv.FormatInt(int64(oldLen), 10)

	// If the whole list is computed, then say that the # is computed
	if computedList {
		diff.Attributes[k+".#"] = &terraform.ResourceAttrDiff{
			Old:         oldStr,
			NewComputed: true,
			RequiresNew: schema.ForceNew,
		}
		return nil
	}

	// If the counts are not the same, then record that diff
	changed := oldLen != newLen
	computed := oldLen == 0 && newLen == 0 && schema.Computed
	if changed || computed || all {
		countSchema := &Schema{
			Type:     TypeInt,
			Computed: schema.Computed,
			ForceNew: schema.ForceNew,
		}

		newStr := ""
		if !computed {
			newStr = strconv.FormatInt(int64(newLen), 10)
		} else {
			oldStr = ""
		}

		finalizedAttr := countSchema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old: oldStr,
				New: newStr,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[k+".#"] = finalizedAttr
		} else {
			delete(diff.Attributes, k+".#")
		}
	}

	// Figure out the maximum
	maxLen := oldLen
	if newLen > maxLen {
		maxLen = newLen
	}

	switch t := schema.Elem.(type) {
	case *Resource:
		// This is a complex resource
		for i := 0; i < maxLen; i++ {
			for k2, schema := range t.Schema {
				subK := fmt.Sprintf("%s.%d.%s", k, i, k2)
				err := m.diff(ctx, subK, schema, diff, d, all)
				if err != nil {
					return err
				}
			}
		}
	case *Schema:
		// Copy the schema so that we can set Computed/ForceNew from
		// the parent schema (the TypeList).
		t2 := *t
		t2.ForceNew = schema.ForceNew

		// This is just a primitive element, so go through each and
		// just diff each.
		for i := 0; i < maxLen; i++ {
			subK := fmt.Sprintf("%s.%d", k, i)
			err := m.diff(ctx, subK, &t2, diff, d, all)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%s: unknown element type (internal)", k)
	}

	return nil
}

func (m schemaMap) diffMap(
	k string,
	schema *Schema,
	diff *terraform.InstanceDiff,
	d resourceDiffer,
	all bool) error {
	prefix := k + "."

	// First get all the values from the state
	var stateMap, configMap map[string]string
	o, n, _, nComputed, customized := d.diffChange(k)
	if err := mapstructure.WeakDecode(o, &stateMap); err != nil {
		return fmt.Errorf("%s: %s", k, err)
	}
	if err := mapstructure.WeakDecode(n, &configMap); err != nil {
		return fmt.Errorf("%s: %s", k, err)
	}

	// Keep track of whether the state _exists_ at all prior to clearing it
	stateExists := o != nil

	// Delete any count values, since we don't use those
	delete(configMap, "%")
	delete(stateMap, "%")

	// Check if the number of elements has changed.
	oldLen, newLen := len(stateMap), len(configMap)
	changed := oldLen != newLen
	if oldLen != 0 && newLen == 0 && schema.Computed {
		changed = false
	}

	// It is computed if we have no old value, no new value, the schema
	// says it is computed, and it didn't exist in the state before. The
	// last point means: if it existed in the state, even empty, then it
	// has already been computed.
	computed := oldLen == 0 && newLen == 0 && schema.Computed && !stateExists

	// If the count has changed or we're computed, then add a diff for the
	// count. "nComputed" means that the new value _contains_ a value that
	// is computed. We don't do granular diffs for this yet, so we mark the
	// whole map as computed.
	if changed || computed || nComputed {
		countSchema := &Schema{
			Type:     TypeInt,
			Computed: schema.Computed || nComputed,
			ForceNew: schema.ForceNew,
		}

		oldStr := strconv.FormatInt(int64(oldLen), 10)
		newStr := ""
		if !computed && !nComputed {
			newStr = strconv.FormatInt(int64(newLen), 10)
		} else {
			oldStr = ""
		}

		finalizedAttr := countSchema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old: oldStr,
				New: newStr,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[k+".%"] = finalizedAttr
		} else {
			delete(diff.Attributes, k+".%")
		}
	}

	// If the new map is nil and we're computed, then ignore it.
	if n == nil && schema.Computed {
		return nil
	}

	// Now we compare, preferring values from the config map
	for k, v := range configMap {
		old, ok := stateMap[k]
		delete(stateMap, k)

		if old == v && ok && !all {
			continue
		}

		finalizedAttr := schema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old: old,
				New: v,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[prefix+k] = finalizedAttr
		}
	}
	for k, v := range stateMap {
		finalizedAttr := schema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old:        v,
				NewRemoved: true,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[prefix+k] = finalizedAttr
		}
	}

	return nil
}

func (m schemaMap) diffSet(
	ctx context.Context,
	k string,
	schema *Schema,
	diff *terraform.InstanceDiff,
	d resourceDiffer,
	all bool) error {

	o, n, _, computedSet, customized := d.diffChange(k)
	if computedSet {
		n = nil
	}
	nSet := n != nil

	// If we have an old value and no new value is set or will be
	// computed once all variables can be interpolated and we're
	// computed, then nothing has changed.
	if o != nil && n == nil && !computedSet && schema.Computed {
		return nil
	}

	if o == nil {
		o = schema.ZeroValue().(*Set)
	}
	if n == nil {
		n = schema.ZeroValue().(*Set)
	}
	os := o.(*Set)
	ns := n.(*Set)

	// If the new value was set, compare the listCode's to determine if
	// the two are equal. Comparing listCode's instead of the actual values
	// is needed because there could be computed values in the set which
	// would result in false positives while comparing.
	if !all && nSet && reflect.DeepEqual(os.listCode(), ns.listCode()) {
		return nil
	}

	// Get the counts
	oldLen := os.Len()
	newLen := ns.Len()
	oldStr := strconv.Itoa(oldLen)
	newStr := strconv.Itoa(newLen)

	// Build a schema for our count
	countSchema := &Schema{
		Type:     TypeInt,
		Computed: schema.Computed,
		ForceNew: schema.ForceNew,
	}

	// If the set computed then say that the # is computed
	if computedSet || schema.Computed && !nSet {
		// If # already exists, equals 0 and no new set is supplied, there
		// is nothing to record in the diff
		count, ok := d.GetOk(k + ".#")
		if ok && count.(int) == 0 && !nSet && !computedSet {
			return nil
		}

		// Set the count but make sure that if # does not exist, we don't
		// use the zeroed value
		countStr := strconv.Itoa(count.(int))
		if !ok {
			countStr = ""
		}

		finalizedAttr := countSchema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old:         countStr,
				NewComputed: true,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[k+".#"] = finalizedAttr
		}
		return nil
	}

	// If the counts are not the same, then record that diff
	changed := oldLen != newLen
	if changed || all {
		finalizedAttr := countSchema.finalizeDiff(
			&terraform.ResourceAttrDiff{
				Old: oldStr,
				New: newStr,
			},
			customized,
		)
		if finalizedAttr != nil {
			diff.Attributes[k+".#"] = finalizedAttr
		}
	}

	// Build the list of codes that will make up our set. This is the
	// removed codes as well as all the codes in the new codes.
	codes := make([][]string, 2)
	codes[0] = os.Difference(ns).listCode()
	codes[1] = ns.listCode()
	for _, list := range codes {
		for _, code := range list {
			switch t := schema.Elem.(type) {
			case *Resource:
				// This is a complex resource
				for k2, schema := range t.Schema {
					subK := fmt.Sprintf("%s.%s.%s", k, code, k2)
					err := m.diff(ctx, subK, schema, diff, d, true)
					if err != nil {
						return err
					}
				}
			case *Schema:
				// Copy the schema so that we can set Computed/ForceNew from
				// the parent schema (the TypeSet).
				t2 := *t
				t2.ForceNew = schema.ForceNew

				// This is just a primitive element, so go through each and
				// just diff each.
				subK := fmt.Sprintf("%s.%s", k, code)
				err := m.diff(ctx, subK, &t2, diff, d, true)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("%s: unknown element type (internal)", k)
			}
		}
	}

	return nil
}

func (m schemaMap) diffString(
	k string,
	schema *Schema,
	diff *terraform.InstanceDiff,
	d resourceDiffer,
	all bool) error {
	var originalN interface{}
	var os, ns string
	o, n, _, computed, customized := d.diffChange(k)
	if schema.StateFunc != nil && n != nil {
		originalN = n
		n = schema.StateFunc(n)
	}
	nraw := n
	if nraw == nil && o != nil {
		nraw = schema.Type.Zero()
	}
	if err := mapstructure.WeakDecode(o, &os); err != nil {
		return fmt.Errorf("%s: %s", k, err)
	}
	if err := mapstructure.WeakDecode(nraw, &ns); err != nil {
		return fmt.Errorf("%s: %s", k, err)
	}

	if os == ns && !all && !computed {
		// They're the same value. If there old value is not blank or we
		// have an ID, then return right away since we're already setup.
		if os != "" || d.Id() != "" {
			return nil
		}

		// Otherwise, only continue if we're computed
		if !schema.Computed {
			return nil
		}
	}

	removed := false
	if o != nil && n == nil && !computed {
		removed = true
	}
	if removed && schema.Computed {
		return nil
	}

	finalizedAttr := schema.finalizeDiff(
		&terraform.ResourceAttrDiff{
			Old:         os,
			New:         ns,
			NewExtra:    originalN,
			NewRemoved:  removed,
			NewComputed: computed,
		},
		customized,
	)
	if finalizedAttr != nil {
		diff.Attributes[k] = finalizedAttr
	}

	return nil
}

// handleDiffSuppressOnRefresh visits each of the attributes set in "new" and,
// if the corresponding schema sets both DiffSuppressFunc and
// DiffSuppressOnRefresh, checks whether the new value is materially different
// than the old and if not it overwrites the new value with the old one,
// in-place.
func (m schemaMap) handleDiffSuppressOnRefresh(ctx context.Context, oldState, newState *terraform.InstanceState) {
	if newState == nil || oldState == nil {
		return // nothing to do, then
	}

	// We'll populate this in the loop below only if we find at least one
	// attribute which needs this analysis.
	var d *ResourceData

	oldAttrs := oldState.Attributes
	newAttrs := newState.Attributes
	for k, newV := range newAttrs {
		oldV, ok := oldAttrs[k]
		if !ok {
			continue // no old value to compare with
		}
		if newV == oldV {
			continue // no change to test
		}

		schemaList := addrToSchema(strings.Split(k, "."), m)
		if len(schemaList) == 0 {
			continue // no schema? weird, but not our responsibility to handle
		}
		schema := schemaList[len(schemaList)-1]
		if !schema.DiffSuppressOnRefresh || schema.DiffSuppressFunc == nil {
			continue // not relevant
		}

		if d == nil {
			// We populate "d" only on demand, to avoid the cost for most
			// existing schemas where DiffSuppressOnRefresh won't be set.
			var err error
			d, err = m.Data(newState, nil)
			if err != nil {
				// Should not happen if we got far enough to be doing this
				// analysis, but if it does then we'll bail out.
				tfsdklog.Warn(ctx, fmt.Sprintf("schemaMap.handleDiffSuppressOnRefresh failed to construct ResourceData: %s", err))
				return
			}
		}

		if schema.DiffSuppressFunc(k, oldV, newV, d) {
			tfsdklog.Debug(ctx, fmt.Sprintf("ignoring change of %q due to DiffSuppressFunc", k))
			newState.Attributes[k] = oldV // keep the old value, then
		}
	}
}

func (m schemaMap) validate(
	k string,
	schema *Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	raw, ok := c.Get(k)
	if !ok && schema.DefaultFunc != nil {
		// We have a dynamic default. Check if we have a value.
		var err error
		raw, err = schema.DefaultFunc()
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Failed to determine default value",
				Detail:        err.Error(),
				AttributePath: path,
			})
		}

		// We're okay as long as we had a value set
		ok = raw != nil
	}

	err := validateExactlyOneAttribute(k, schema, c)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Invalid combination of arguments",
			Detail:        err.Error(),
			AttributePath: path,
		})
	}

	err = validateAtLeastOneAttribute(k, schema, c)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Missing required argument",
			Detail:        err.Error(),
			AttributePath: path,
		})
	}

	if !ok {
		if schema.Required {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Missing required argument",
				Detail:        fmt.Sprintf("The argument %q is required, but no definition was found.", k),
				AttributePath: path,
			})
		}
		return diags
	}

	if !schema.Required && !schema.Optional {
		// This is a computed-only field
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Value for unconfigurable attribute",
			Detail:        fmt.Sprintf("Can't configure a value for %q: its value will be decided automatically based on the result of applying this configuration.", k),
			AttributePath: path,
		})
	}

	err = validateRequiredWithAttribute(k, schema, c)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Missing required argument",
			Detail:        err.Error(),
			AttributePath: path,
		})
	}

	// If the value is unknown then we can't validate it yet.
	// In particular, this avoids spurious type errors where downstream
	// validation code sees UnknownVariableValue as being just a string.
	// The SDK has to allow the unknown value through initially, so that
	// Required fields set via an interpolated value are accepted.
	if !isWhollyKnown(raw) {
		return nil
	}

	err = validateConflictingAttributes(k, schema, c)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Conflicting configuration arguments",
			Detail:        err.Error(),
			AttributePath: path,
		})
	}

	return m.validateType(k, raw, schema, c, path)
}

// isWhollyKnown returns false if the argument contains an UnknownVariableValue
func isWhollyKnown(raw interface{}) bool {
	switch raw := raw.(type) {
	case string:
		if raw == hcl2shim.UnknownVariableValue {
			return false
		}
	case []interface{}:
		for _, v := range raw {
			if !isWhollyKnown(v) {
				return false
			}
		}
	case map[string]interface{}:
		for _, v := range raw {
			if !isWhollyKnown(v) {
				return false
			}
		}
	}
	return true
}
func validateConflictingAttributes(
	k string,
	schema *Schema,
	c *terraform.ResourceConfig) error {

	if len(schema.ConflictsWith) == 0 {
		return nil
	}

	for _, conflictingKey := range schema.ConflictsWith {
		if raw, ok := c.Get(conflictingKey); ok {
			if raw == hcl2shim.UnknownVariableValue {
				// An unknown value might become unset (null) once known, so
				// we must defer validation until it's known.
				continue
			}
			return fmt.Errorf(
				"%q: conflicts with %s", k, conflictingKey)
		}
	}

	return nil
}

func removeDuplicates(elements []string) []string {
	encountered := make(map[string]struct{}, 0)
	result := []string{}

	for v := range elements {
		if _, ok := encountered[elements[v]]; !ok {
			encountered[elements[v]] = struct{}{}
			result = append(result, elements[v])
		}
	}

	return result
}

func validateRequiredWithAttribute(
	k string,
	schema *Schema,
	c *terraform.ResourceConfig) error {

	if len(schema.RequiredWith) == 0 {
		return nil
	}

	allKeys := removeDuplicates(append(schema.RequiredWith, k))
	sort.Strings(allKeys)

	for _, key := range allKeys {
		if _, ok := c.Get(key); !ok {
			return fmt.Errorf("%q: all of `%s` must be specified", k, strings.Join(allKeys, ","))
		}
	}

	return nil
}

func validateExactlyOneAttribute(
	k string,
	schema *Schema,
	c *terraform.ResourceConfig) error {

	if len(schema.ExactlyOneOf) == 0 {
		return nil
	}

	allKeys := removeDuplicates(append(schema.ExactlyOneOf, k))
	sort.Strings(allKeys)
	specified := make([]string, 0)
	unknownVariableValueCount := 0
	for _, exactlyOneOfKey := range allKeys {
		if c.IsComputed(exactlyOneOfKey) {
			unknownVariableValueCount++
			continue
		}

		_, ok := c.Get(exactlyOneOfKey)
		if ok {
			specified = append(specified, exactlyOneOfKey)
		}
	}

	if len(specified) == 0 && unknownVariableValueCount == 0 {
		return fmt.Errorf("%q: one of `%s` must be specified", k, strings.Join(allKeys, ","))
	}

	if len(specified) > 1 {
		return fmt.Errorf("%q: only one of `%s` can be specified, but `%s` were specified.", k, strings.Join(allKeys, ","), strings.Join(specified, ","))
	}

	return nil
}

func validateAtLeastOneAttribute(
	k string,
	schema *Schema,
	c *terraform.ResourceConfig) error {

	if len(schema.AtLeastOneOf) == 0 {
		return nil
	}

	allKeys := removeDuplicates(append(schema.AtLeastOneOf, k))
	sort.Strings(allKeys)

	for _, atLeastOneOfKey := range allKeys {
		if _, ok := c.Get(atLeastOneOfKey); ok {
			// We can ignore hcl2shim.UnknownVariable by assuming it's been set and additional validation elsewhere
			// will uncover this if it is in fact null.
			return nil
		}
	}

	return fmt.Errorf("%q: one of `%s` must be specified", k, strings.Join(allKeys, ","))
}

func (m schemaMap) validateList(
	k string,
	raw interface{},
	schema *Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	// first check if the list is wholly unknown
	if s, ok := raw.(string); ok {
		if s == hcl2shim.UnknownVariableValue {
			return diags
		}
	}

	// schemaMap can't validate nil
	if raw == nil {
		return diags
	}

	// We use reflection to verify the slice because you can't
	// case to []interface{} unless the slice is exactly that type.
	rawV := reflect.ValueOf(raw)

	if rawV.Kind() != reflect.Slice {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Attribute must be a list",
			AttributePath: path,
		})
	}

	// We can't validate list length if this came from a dynamic block.
	// Since there's no way to determine if something was from a dynamic block
	// at this point, we're going to skip validation in the new protocol if
	// there are any unknowns. Validate will eventually be called again once
	// all values are known.
	if !isWhollyKnown(raw) {
		return diags
	}

	// Validate length
	if schema.MaxItems > 0 && rawV.Len() > schema.MaxItems {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Too many list items",
			Detail:        fmt.Sprintf("Attribute %s supports %d item maximum, but config has %d declared.", k, schema.MaxItems, rawV.Len()),
			AttributePath: path,
		})
	}

	if schema.MinItems > 0 && rawV.Len() < schema.MinItems {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Not enough list items",
			Detail:        fmt.Sprintf("Attribute %s requires %d item minimum, but config has only %d declared.", k, schema.MinItems, rawV.Len()),
			AttributePath: path,
		})
	}

	// Now build the []interface{}
	raws := make([]interface{}, rawV.Len())
	for i := range raws {
		raws[i] = rawV.Index(i).Interface()
	}

	for i, raw := range raws {
		key := fmt.Sprintf("%s.%d", k, i)

		// Reify the key value from the ResourceConfig.
		// If the list was computed we have all raw values, but some of these
		// may be known in the config, and aren't individually marked as Computed.
		if r, ok := c.Get(key); ok {
			raw = r
		}

		p := append(path, cty.IndexStep{Key: cty.NumberIntVal(int64(i))})

		switch t := schema.Elem.(type) {
		case *Resource:
			// This is a sub-resource
			diags = append(diags, m.validateObject(key, t.Schema, c, p)...)
		case *Schema:
			diags = append(diags, m.validateType(key, raw, t, c, p)...)
		}

	}

	return diags
}

func (m schemaMap) validateMap(
	k string,
	raw interface{},
	schema *Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	// first check if the list is wholly unknown
	if s, ok := raw.(string); ok {
		if s == hcl2shim.UnknownVariableValue {
			return diags
		}
	}

	// schemaMap can't validate nil
	if raw == nil {
		return diags
	}
	// We use reflection to verify the slice because you can't
	// case to []interface{} unless the slice is exactly that type.
	rawV := reflect.ValueOf(raw)
	switch rawV.Kind() {
	case reflect.String:
		// If raw and reified are equal, this is a string and should
		// be rejected.
		reified, reifiedOk := c.Get(k)
		if reifiedOk && raw == reified && !c.IsComputed(k) {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Attribute must be a map",
				AttributePath: path,
			})
		}
		// Otherwise it's likely raw is an interpolation.
		return diags
	case reflect.Map:
	case reflect.Slice:
	default:
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Attribute must be a map",
			AttributePath: path,
		})
	}

	// If it is not a slice, validate directly
	if rawV.Kind() != reflect.Slice {
		mapIface := rawV.Interface()
		diags = append(diags, validateMapValues(k, mapIface.(map[string]interface{}), schema, path)...)
		if diags.HasError() {
			return diags
		}

		return schema.validateFunc(mapIface, k, path)
	}

	// It is a slice, verify that all the elements are maps
	raws := make([]interface{}, rawV.Len())
	for i := range raws {
		raws[i] = rawV.Index(i).Interface()
	}

	for _, raw := range raws {
		v := reflect.ValueOf(raw)
		if v.Kind() != reflect.Map {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Attribute must be a map",
				AttributePath: path,
			})
		}
		mapIface := v.Interface()
		diags = append(diags, validateMapValues(k, mapIface.(map[string]interface{}), schema, path)...)
		if diags.HasError() {
			return diags
		}
	}

	validatableMap := make(map[string]interface{})
	for _, raw := range raws {
		for k, v := range raw.(map[string]interface{}) {
			validatableMap[k] = v
		}
	}

	return schema.validateFunc(validatableMap, k, path)
}

func validateMapValues(k string, m map[string]interface{}, schema *Schema, path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	for key, raw := range m {
		valueType, err := getValueType(k, schema)
		p := append(path, cty.IndexStep{Key: cty.StringVal(key)})
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				AttributePath: p,
			})
		}

		switch valueType {
		case TypeBool:
			var n bool
			if err := mapstructure.WeakDecode(raw, &n); err != nil {
				return append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       err.Error(),
					AttributePath: p,
				})
			}
		case TypeInt:
			var n int
			if err := mapstructure.WeakDecode(raw, &n); err != nil {
				return append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       err.Error(),
					AttributePath: p,
				})
			}
		case TypeFloat:
			var n float64
			if err := mapstructure.WeakDecode(raw, &n); err != nil {
				return append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       err.Error(),
					AttributePath: p,
				})
			}
		case TypeString:
			var n string
			if err := mapstructure.WeakDecode(raw, &n); err != nil {
				return append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       err.Error(),
					AttributePath: p,
				})
			}
		default:
			panic(fmt.Sprintf("Unknown validation type: %#v", valueType))
		}
	}
	return diags
}

func getValueType(k string, schema *Schema) (ValueType, error) {
	if schema.Elem == nil {
		return TypeString, nil
	}
	if vt, ok := schema.Elem.(ValueType); ok {
		return vt, nil
	}

	// If a Schema is provided to a Map, we use the Type of that schema
	// as the type for each element in the Map.
	if s, ok := schema.Elem.(*Schema); ok {
		return s.Type, nil
	}

	if _, ok := schema.Elem.(*Resource); ok {
		// TODO: We don't actually support this (yet)
		// but silently pass the validation, until we decide
		// how to handle nested structures in maps
		return TypeString, nil
	}
	return 0, fmt.Errorf("%s: unexpected map value type: %#v", k, schema.Elem)
}

func (m schemaMap) validateObject(
	k string,
	schema map[string]*Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	raw, _ := c.Get(k)

	// schemaMap can't validate nil
	if raw == nil {
		return diags
	}

	if _, ok := raw.(map[string]interface{}); !ok && !c.IsComputed(k) {
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Expected Object Type",
			Detail:        fmt.Sprintf("Expected object, got %s", reflect.ValueOf(raw).Kind()),
			AttributePath: path,
		})
	}

	for subK, s := range schema {
		key := subK
		if k != "" {
			key = fmt.Sprintf("%s.%s", k, subK)
		}
		diags = append(diags, m.validate(key, s, c, append(path, cty.GetAttrStep{Name: subK}))...)
	}

	// Detect any extra/unknown keys and report those as errors.
	if m, ok := raw.(map[string]interface{}); ok {
		for subk := range m {
			if _, ok := schema[subk]; !ok {
				if subk == TimeoutsConfigKey {
					continue
				}
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Invalid or unknown key",
					AttributePath: append(path, cty.GetAttrStep{Name: subk}),
				})
			}
		}
	}

	return diags
}

func (m schemaMap) validatePrimitive(
	k string,
	raw interface{},
	schema *Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics

	// a nil value shouldn't happen in the old protocol, and in the new
	// protocol the types have already been validated. Either way, we can't
	// reflect on nil, so don't panic.
	if raw == nil {
		return diags
	}

	// Catch if the user gave a complex type where a primitive was
	// expected, so we can return a friendly error message that
	// doesn't contain Go type system terminology.
	switch reflect.ValueOf(raw).Type().Kind() {
	case reflect.Slice:
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Attribute must be a single value, not a list",
			AttributePath: path,
		})
	case reflect.Map:
		return append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Attribute must be a single value, not a map",
			AttributePath: path,
		})
	default: // ok
	}

	if c.IsComputed(k) {
		// If the key is being computed, then it is not an error as
		// long as it's not a slice or map.
		return diags
	}

	var decoded interface{}
	switch schema.Type {
	case TypeBool:
		// Verify that we can parse this as the correct type
		var n bool
		if err := mapstructure.WeakDecode(raw, &n); err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				AttributePath: path,
			})
		}
		decoded = n
	case TypeInt:
		// We need to verify the type precisely, because WeakDecode will
		// decode a float as an integer.

		// the config shims only use int for integral number values
		// also accept a string, just as the TypeBool and TypeFloat cases do
		if v, ok := raw.(int); ok {
			decoded = v
		} else if _, ok := raw.(string); ok {
			var n int
			if err := mapstructure.WeakDecode(raw, &n); err != nil {
				return append(diags, diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       err.Error(),
					AttributePath: path,
				})
			}
			decoded = n
		} else {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       fmt.Sprintf("Attribute must be a whole number, got %v", raw),
				AttributePath: path,
			})
		}
	case TypeFloat:
		// Verify that we can parse this as a float
		var n float64
		if err := mapstructure.WeakDecode(raw, &n); err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				AttributePath: path,
			})
		}
		decoded = n
	case TypeString:
		// Verify that we can parse this as a string
		var n string
		if err := mapstructure.WeakDecode(raw, &n); err != nil {
			return append(diags, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				AttributePath: path,
			})
		}
		decoded = n
	default:
		panic(fmt.Sprintf("Unknown validation type: %#v", schema.Type))
	}

	return append(diags, schema.validateFunc(decoded, k, path)...)
}

func (m schemaMap) validateType(
	k string,
	raw interface{},
	schema *Schema,
	c *terraform.ResourceConfig,
	path cty.Path) diag.Diagnostics {

	var diags diag.Diagnostics
	switch schema.Type {
	case TypeList:
		diags = m.validateList(k, raw, schema, c, path)
	case TypeSet:
		// indexing into sets is not representable in the current protocol
		// best we can do is associate the path up to this attribute.
		diags = m.validateList(k, raw, schema, c, path)
		if len(diags) > 0 {
			log.Printf("[WARN] Truncating attribute path of %d diagnostics for TypeSet", len(diags))
			for i := range diags {
				diags[i].AttributePath = path
			}
		}
	case TypeMap:
		diags = m.validateMap(k, raw, schema, c, path)
	default:
		diags = m.validatePrimitive(k, raw, schema, c, path)
	}

	if schema.Deprecated != "" {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       "Argument is deprecated",
			Detail:        schema.Deprecated,
			AttributePath: path,
		})
	}

	return diags
}

// Zero returns the zero value for a type.
func (t ValueType) Zero() interface{} {
	switch t {
	case TypeInvalid:
		return nil
	case TypeBool:
		return false
	case TypeInt:
		return 0
	case TypeFloat:
		return 0.0
	case TypeString:
		return ""
	case TypeList:
		return []interface{}{}
	case TypeMap:
		return map[string]interface{}{}
	case TypeSet:
		return new(Set)
	case typeObject:
		return map[string]interface{}{}
	default:
		panic(fmt.Sprintf("unknown type %s", t))
	}
}
