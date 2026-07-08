// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package smithymodel parses a Smithy 2.0 JSON model file into a queryable
// in-memory representation. It covers the subset of Smithy shapes needed by
// the drift-detect AWS extractor: service, resource, operation, structure,
// enum, map, and the scalar types.
//
// Only the traits relevant to field extraction are decoded:
//   - smithy.api#required     — marks a member as required
//   - smithy.api#httpLabel    — marks a member as a URL path segment (suppress)
//   - smithy.api#httpQuery    — marks a member as a URL query param (suppress)
//   - smithy.api#idempotencyToken — marks a member as an idempotency token (suppress)
//   - smithy.api#input / output  — marks a structure as op input/output
package smithymodel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Raw JSON deserialization types
// ---------------------------------------------------------------------------

// rawModel is the top-level shape of a Smithy 2.0 JSON file.
type rawModel struct {
	Smithy string              `json:"smithy"`
	Shapes map[string]rawShape `json:"shapes"`
}

// rawShape captures any shape. The type discriminator is in "type".
// Members, key, value, and traits all need careful handling.
type rawShape struct {
	Type        string                 `json:"type"`
	Input       *rawTarget             `json:"input"`      // operation input
	Output      *rawTarget             `json:"output"`     // operation output
	Create      *rawTarget             `json:"create"`     // resource create operation
	Put         *rawTarget             `json:"put"`        // resource put operation
	Read        *rawTarget             `json:"read"`       // resource read operation
	Update      *rawTarget             `json:"update"`     // resource update operation
	Delete      *rawTarget             `json:"delete"`     // resource delete operation
	List        *rawTarget             `json:"list"`       // resource list operation
	Operations  []rawTarget            `json:"operations"` // resource operations
	Key         *rawTarget             `json:"key"`        // map key type
	Value       *rawTarget             `json:"value"`      // map value type
	Members     map[string]rawMember   `json:"members"`    // structure/enum members
	Identifiers map[string]rawTarget   `json:"identifiers"`
	Target      string                 `json:"target"` // used for simple typedef aliases
	Traits      map[string]rawAnything `json:"traits"`
}

// rawTarget is a shape reference — {"target": "com.amazonaws.foo#Bar"}
type rawTarget struct {
	Target string `json:"target"`
}

// rawMember is a structure member or enum member.
type rawMember struct {
	Target string                 `json:"target"`
	Traits map[string]rawAnything `json:"traits"`
}

// rawAnything accepts any JSON value without deserializing it.
type rawAnything = json.RawMessage

// ---------------------------------------------------------------------------
// Public model types
// ---------------------------------------------------------------------------

// Model is the parsed representation of a Smithy model file.
type Model struct {
	shapes map[string]*Shape // absolute shape ID → Shape
}

// ShapeKind classifies a Smithy shape.
type ShapeKind string

const (
	KindService    ShapeKind = "service"
	KindResource   ShapeKind = "resource"
	KindOperation  ShapeKind = "operation"
	KindStructure  ShapeKind = "structure"
	KindEnum       ShapeKind = "enum"
	KindMap        ShapeKind = "map"
	KindString     ShapeKind = "string"
	KindBoolean    ShapeKind = "boolean"
	KindInteger    ShapeKind = "integer"
	KindLong       ShapeKind = "long"
	KindFloat      ShapeKind = "float"
	KindDouble     ShapeKind = "double"
	KindTimestamp  ShapeKind = "timestamp"
	KindBlob       ShapeKind = "blob"
	KindByte       ShapeKind = "byte"
	KindShort      ShapeKind = "short"
	KindBigInteger ShapeKind = "bigInteger"
	KindBigDecimal ShapeKind = "bigDecimal"
	KindList       ShapeKind = "list"
	KindSet        ShapeKind = "set"
	KindUnion      ShapeKind = "union"
	KindDocument   ShapeKind = "document"
	KindOther      ShapeKind = "other"
)

// Shape is the parsed representation of a single Smithy shape.
type Shape struct {
	// ID is the absolute shape ID, e.g. "com.amazonaws.sqs#CreateQueue".
	ID   string
	Kind ShapeKind

	// Operation fields
	InputTarget  string // shape ID of input structure (operations only)
	OutputTarget string // shape ID of output structure (operations only)

	// Resource fields
	CreateTarget string            // resource create operation target
	PutTarget    string            // resource put operation target
	ReadTarget   string            // resource read operation target
	UpdateTarget string            // resource update operation target
	DeleteTarget string            // resource delete operation target
	ListTarget   string            // resource list operation target
	Operations   []string          // other resource operations
	Identifiers  map[string]string // identifier member name -> shape ID

	// Structure / enum members. For structures: Member.Target is the shape ID
	// of the member type; Traits carries the member-level traits.
	// For enums: Member.EnumValue is the string value of the enum member.
	Members map[string]*Member

	// Map fields
	KeyTarget   string // shape ID of map key type
	ValueTarget string // shape ID of map value type

	// Traits on the shape itself (not on members).
	Traits ShapeTraits
}

// Member is a structure or enum member.
type Member struct {
	// Target is the absolute shape ID this member points to.
	Target string

	// EnumValue is set for enum members (the string value of the constant).
	EnumValue string

	// Traits parsed from this member's traits map.
	Traits MemberTraits
}

// ShapeTraits holds the subset of shape-level traits we care about.
type ShapeTraits struct {
	IsInput  bool // smithy.api#input
	IsOutput bool // smithy.api#output
}

// MemberTraits holds the subset of member-level traits used for extraction.
type MemberTraits struct {
	Required         bool // smithy.api#required
	HTTPLabel        bool // smithy.api#httpLabel
	HTTPQuery        bool // smithy.api#httpQuery (value = query param name)
	IdempotencyToken bool // smithy.api#idempotencyToken
}

// IsSuppressible returns true when the member should be excluded from the
// resource IR (it is a URL routing concern, not a resource body field).
func (t MemberTraits) IsSuppressible() bool {
	return t.HTTPLabel || t.IdempotencyToken
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// LoadFile reads and parses the Smithy 2.0 JSON model at path.
func LoadFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading smithy model: %w", err)
	}

	return loadBytes(data)
}

// LoadURL fetches and parses the Smithy 2.0 JSON model at url.
func LoadURL(url string) (*Model, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching smithy model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching smithy model: unexpected status %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading smithy model response: %w", err)
	}

	return loadBytes(data)
}

func loadBytes(data []byte) (*Model, error) {

	var raw rawModel
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing smithy model: %w", err)
	}

	m := &Model{shapes: make(map[string]*Shape, len(raw.Shapes))}
	for id, rs := range raw.Shapes {
		m.shapes[id] = parseShape(id, rs)
	}
	return m, nil
}

// Shape returns the shape with the given absolute ID, or nil if not found.
func (m *Model) Shape(id string) *Shape {
	return m.shapes[id]
}

// ResolveToKind walks typedef chains until it reaches a shape whose kind is
// not an alias of another shape, returning the terminal shape.
// This handles patterns like:
//
//	com.amazonaws.amp#WorkspaceAlias → { type: "string" }  (direct)
//	com.amazonaws.amp#WorkspaceId    → { type: "string" }  (direct)
func (m *Model) ResolveToKind(id string) ShapeKind {
	// Built-in smithy prelude types resolve directly.
	if k, ok := builtinKind(id); ok {
		return k
	}

	s := m.shapes[id]
	if s == nil {
		return KindOther
	}

	// A string/boolean/integer/etc. shape in the model is a terminal type.
	if isPrimitiveKind(s.Kind) {
		return s.Kind
	}

	// Non-primitive: return as-is.
	return s.Kind
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func parseShape(id string, rs rawShape) *Shape {
	s := &Shape{
		ID:   id,
		Kind: parseKind(rs.Type),
	}

	// Operation: extract input/output targets.
	if rs.Input != nil {
		s.InputTarget = rs.Input.Target
	}
	if rs.Output != nil {
		s.OutputTarget = rs.Output.Target
	}

	// Resource: extract operation targets and identifiers.
	if rs.Create != nil {
		s.CreateTarget = rs.Create.Target
	}
	if rs.Put != nil {
		s.PutTarget = rs.Put.Target
	}
	if rs.Read != nil {
		s.ReadTarget = rs.Read.Target
	}
	if rs.Update != nil {
		s.UpdateTarget = rs.Update.Target
	}
	if rs.Delete != nil {
		s.DeleteTarget = rs.Delete.Target
	}
	if rs.List != nil {
		s.ListTarget = rs.List.Target
	}
	if len(rs.Operations) > 0 {
		s.Operations = make([]string, 0, len(rs.Operations))
		for _, op := range rs.Operations {
			if op.Target == "" {
				continue
			}
			s.Operations = append(s.Operations, op.Target)
		}
	}
	if len(rs.Identifiers) > 0 {
		s.Identifiers = make(map[string]string, len(rs.Identifiers))
		for name, target := range rs.Identifiers {
			s.Identifiers[name] = target.Target
		}
	}

	// Map: extract key/value targets.
	if rs.Key != nil {
		s.KeyTarget = rs.Key.Target
	}
	if rs.Value != nil {
		s.ValueTarget = rs.Value.Target
	}

	// Members (structure members and enum members).
	if len(rs.Members) > 0 {
		s.Members = make(map[string]*Member, len(rs.Members))
		for name, rm := range rs.Members {
			m := &Member{
				Target: rm.Target,
			}
			// Enum members carry smithy.api#enumValue in their traits.
			if ev, ok := rm.Traits["smithy.api#enumValue"]; ok {
				var sv string
				if err := json.Unmarshal(ev, &sv); err == nil {
					m.EnumValue = sv
				}
			}
			m.Traits = parseMemberTraits(rm.Traits)
			s.Members[name] = m
		}
	}

	// Shape-level traits.
	_, s.Traits.IsInput = rs.Traits["smithy.api#input"]
	_, s.Traits.IsOutput = rs.Traits["smithy.api#output"]

	return s
}

func parseMemberTraits(traits map[string]rawAnything) MemberTraits {
	var t MemberTraits
	_, t.Required = traits["smithy.api#required"]
	_, t.HTTPLabel = traits["smithy.api#httpLabel"]
	_, t.HTTPQuery = traits["smithy.api#httpQuery"]
	_, t.IdempotencyToken = traits["smithy.api#idempotencyToken"]
	return t
}

func parseKind(s string) ShapeKind {
	switch strings.ToLower(s) {
	case "service":
		return KindService
	case "resource":
		return KindResource
	case "operation":
		return KindOperation
	case "structure":
		return KindStructure
	case "enum":
		return KindEnum
	case "map":
		return KindMap
	case "string":
		return KindString
	case "boolean":
		return KindBoolean
	case "integer":
		return KindInteger
	case "long":
		return KindLong
	case "float":
		return KindFloat
	case "double":
		return KindDouble
	case "timestamp":
		return KindTimestamp
	case "blob":
		return KindBlob
	case "byte":
		return KindByte
	case "short":
		return KindShort
	case "biginteger":
		return KindBigInteger
	case "bigdecimal":
		return KindBigDecimal
	case "list":
		return KindList
	case "set":
		return KindSet
	case "union":
		return KindUnion
	case "document":
		return KindDocument
	default:
		return KindOther
	}
}

// builtinKind maps Smithy prelude shape IDs to their primitive kind.
// These are well-known shapes that do not appear as explicit entries in model
// JSON files but are referenced by target IDs.
func builtinKind(id string) (ShapeKind, bool) {
	switch id {
	case "smithy.api#String":
		return KindString, true
	case "smithy.api#Boolean":
		return KindBoolean, true
	case "smithy.api#Integer":
		return KindInteger, true
	case "smithy.api#Long":
		return KindLong, true
	case "smithy.api#Float":
		return KindFloat, true
	case "smithy.api#Double":
		return KindDouble, true
	case "smithy.api#Timestamp":
		return KindTimestamp, true
	case "smithy.api#Blob":
		return KindBlob, true
	case "smithy.api#Byte":
		return KindByte, true
	case "smithy.api#Short":
		return KindShort, true
	case "smithy.api#BigInteger":
		return KindBigInteger, true
	case "smithy.api#BigDecimal":
		return KindBigDecimal, true
	case "smithy.api#Document":
		return KindDocument, true
	case "smithy.api#PrimitiveByte",
		"smithy.api#PrimitiveShort",
		"smithy.api#PrimitiveInteger":
		return KindInteger, true
	case "smithy.api#PrimitiveLong":
		return KindLong, true
	case "smithy.api#PrimitiveFloat":
		return KindFloat, true
	case "smithy.api#PrimitiveDouble":
		return KindDouble, true
	case "smithy.api#PrimitiveBoolean":
		return KindBoolean, true
	case "smithy.api#Unit":
		return KindOther, true // enum unit members
	}
	return KindOther, false
}

// isPrimitiveKind returns true for scalar Smithy kinds that map directly to
// a Go FieldType without further resolution.
func isPrimitiveKind(k ShapeKind) bool {
	switch k {
	case KindString, KindBoolean, KindInteger, KindLong,
		KindFloat, KindDouble, KindTimestamp, KindBlob,
		KindByte, KindShort, KindBigInteger, KindBigDecimal:
		return true
	}
	return false
}
