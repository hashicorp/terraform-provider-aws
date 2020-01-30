package terraformtype

import (
	"go/ast"
	"go/types"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	SchemaFieldAtLeastOneOf     = `AtLeastOneOf`
	SchemaFieldComputed         = `Computed`
	SchemaFieldComputedWhen     = `ComputedWhen`
	SchemaFieldConfigMode       = `ConfigMode`
	SchemaFieldConflictsWith    = `ConflictsWith`
	SchemaFieldDefault          = `Default`
	SchemaFieldDefaultFunc      = `DefaultFunc`
	SchemaFieldDeprecated       = `Deprecated`
	SchemaFieldDescription      = `Description`
	SchemaFieldDiffSuppressFunc = `DiffSuppressFunc`
	SchemaFieldElem             = `Elem`
	SchemaFieldExactlyOneOf     = `ExactlyOneOf`
	SchemaFieldForceNew         = `ForceNew`
	SchemaFieldInputDefault     = `InputDefault`
	SchemaFieldMaxItems         = `MaxItems`
	SchemaFieldMinItems         = `MinItems`
	SchemaFieldOptional         = `Optional`
	SchemaFieldPromoteSingle    = `PromoteSingle`
	SchemaFieldRemoved          = `Removed`
	SchemaFieldRequired         = `Required`
	SchemaFieldSensitive        = `Sensitive`
	SchemaFieldSet              = `Set`
	SchemaFieldStateFunc        = `StateFunc`
	SchemaFieldType             = `Type`
	SchemaFieldValidateFunc     = `ValidateFunc`

	SchemaValueTypeBool   = `TypeBool`
	SchemaValueTypeFloat  = `TypeFloat`
	SchemaValueTypeInt    = `TypeInt`
	SchemaValueTypeList   = `TypeList`
	SchemaValueTypeMap    = `TypeMap`
	SchemaValueTypeSet    = `TypeSet`
	SchemaValueTypeString = `TypeString`

	TypeNameSchema    = `Schema`
	TypeNameSet       = `Set`
	TypeNameValueType = `ValueType`
)

// HelperSchemaSchemaInfo represents all gathered Schema data for easier access
type HelperSchemaSchemaInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	Schema          *schema.Schema
	SchemaValueType string
	TypesInfo       *types.Info
}

// NewHelperSchemaSchemaInfo instantiates a HelperSchemaSchemaInfo
func NewHelperSchemaSchemaInfo(cl *ast.CompositeLit, info *types.Info) *HelperSchemaSchemaInfo {
	result := &HelperSchemaSchemaInfo{
		AstCompositeLit: cl,
		Fields:          astCompositeLitFields(cl),
		Schema:          &schema.Schema{},
		SchemaValueType: helperSchemaTypeSchemaType(cl, info),
		TypesInfo:       info,
	}

	if kvExpr := result.Fields[SchemaFieldComputed]; kvExpr != nil && astBoolValue(kvExpr.Value) != nil {
		result.Schema.Computed = *astBoolValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldConflictsWith]; kvExpr != nil && astExprValue(kvExpr.Value) != nil {
		result.Schema.ConflictsWith = []string{}
	}

	if kvExpr := result.Fields[SchemaFieldDefault]; kvExpr != nil && astExprValue(kvExpr.Value) != nil {
		result.Schema.Default = func() (interface{}, error) { return nil, nil }
	}

	if kvExpr := result.Fields[SchemaFieldDiffSuppressFunc]; kvExpr != nil && astExprValue(kvExpr.Value) != nil {
		result.Schema.DiffSuppressFunc = func(k, old, new string, d *schema.ResourceData) bool { return false }
	}

	if kvExpr := result.Fields[SchemaFieldForceNew]; kvExpr != nil && astBoolValue(kvExpr.Value) != nil {
		result.Schema.ForceNew = *astBoolValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldMaxItems]; kvExpr != nil && astIntValue(kvExpr.Value) != nil {
		result.Schema.MaxItems = *astIntValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldMinItems]; kvExpr != nil && astIntValue(kvExpr.Value) != nil {
		result.Schema.MinItems = *astIntValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldOptional]; kvExpr != nil && astBoolValue(kvExpr.Value) != nil {
		result.Schema.Optional = *astBoolValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldRequired]; kvExpr != nil && astBoolValue(kvExpr.Value) != nil {
		result.Schema.Required = *astBoolValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldSensitive]; kvExpr != nil && astBoolValue(kvExpr.Value) != nil {
		result.Schema.Sensitive = *astBoolValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[SchemaFieldValidateFunc]; kvExpr != nil && astExprValue(kvExpr.Value) != nil {
		result.Schema.ValidateFunc = func(interface{}, string) ([]string, []error) { return nil, nil }
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *HelperSchemaSchemaInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// DeclaresBoolFieldWithZeroValue returns true if the field name is present and is false
func (info *HelperSchemaSchemaInfo) DeclaresBoolFieldWithZeroValue(fieldName string) bool {
	kvExpr := info.Fields[fieldName]

	// Field not declared
	if kvExpr == nil {
		return false
	}

	valuePtr := astBoolValue(kvExpr.Value)

	// Value not readable
	if valuePtr == nil {
		return false
	}

	return !*valuePtr
}

// IsType returns true if the given input is equal to the Type
func (info *HelperSchemaSchemaInfo) IsType(valueType string) bool {
	return info.SchemaValueType == valueType
}

// IsOneOfTypes returns true if one of the given input is equal to the Type
func (info *HelperSchemaSchemaInfo) IsOneOfTypes(valueTypes ...string) bool {
	for _, valueType := range valueTypes {
		if info.SchemaValueType == valueType {
			return true
		}
	}

	return false
}

// IsHelperSchemaTypeSchema returns if the type is Schema from the helper/schema package
func IsHelperSchemaTypeSchema(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperSchemaNamedType(t, TypeNameSchema)
	case *types.Pointer:
		return IsHelperSchemaTypeSchema(t.Elem())
	default:
		return false
	}
}

// IsHelperSchemaValueType returns if the Schema field Type matches
func IsHelperSchemaValueType(e ast.Expr, info *types.Info) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		switch t := info.TypeOf(e).(type) {
		case *types.Named:
			return IsHelperSchemaNamedType(t, TypeNameValueType)
		default:
			return false
		}
	default:
		return false
	}
}

// IsHelperSchemaTypeSet returns if the type is Set from the helper/schema package
// Use IsHelperSchemaTypeSchemaFieldType for verifying Type: schema.TypeSet ValueType
func IsHelperSchemaTypeSet(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsHelperSchemaNamedType(t, TypeNameSet)
	case *types.Pointer:
		return IsHelperSchemaTypeSet(t.Elem())
	default:
		return false
	}
}

// helperSchemaTypeSchemaType extracts the string representation of a Schema Type value
func helperSchemaTypeSchemaType(schema *ast.CompositeLit, info *types.Info) string {
	kvExpr := astCompositeLitField(schema, SchemaFieldType)

	if kvExpr == nil {
		return ""
	}

	if !IsHelperSchemaValueType(kvExpr.Value, info) {
		return ""
	}

	return helperSchemaValueTypeString(kvExpr.Value)
}

// helperSchemaValueTypeString extracts the string representation of a Schema ValueType
func helperSchemaValueTypeString(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}
