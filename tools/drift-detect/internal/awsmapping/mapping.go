// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package awsmapping loads the YAML resource mapping file that bridges
// Terraform resource names to their corresponding AWS Smithy model operations.
package awsmapping

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// File is the top-level structure of aws_resources.yaml.
type File struct {
	// Services maps Terraform service names to AWS API model directory names
	// when the two differ (e.g. "prometheus" → "amp"). Checked before
	// auto-discovery; if an entry exists the renamed value is used directly.
	Services map[string]string `yaml:"services"`

	Resources map[string]*ResourceMapping `yaml:"resources"`
}

// ResourceMapping holds the configuration for one Terraform resource.
type ResourceMapping struct {
	// SmithyModel is the path to the Smithy JSON file, relative to the
	// api-models-aws root directory.
	SmithyModel string `yaml:"smithy_model"`

	// SmithyNamespace is the AWS namespace prefix, e.g. "com.amazonaws.sqs".
	SmithyNamespace string `yaml:"smithy_namespace"`

	// SmithyResource is the Smithy resource shape name (without namespace),
	// e.g. "WorkspaceResourcePolicy". When set, extraction can infer lifecycle
	// operations from the resource shape's put/create/read/update/delete/list
	// targets and other operations.
	SmithyResource string `yaml:"smithy_resource"`

	// Lifecycle names the Smithy operations (without namespace) for each
	// CRUD verb. Only Create and Read are required for Phase 1.
	Lifecycle Lifecycle `yaml:"lifecycle"`

	// SuppressFields lists member names that must be dropped from the
	// extracted IR because they are AWS-internal (URL path identifiers,
	// idempotency tokens, pagination cursors, etc.).
	SuppressFields []string `yaml:"suppress_fields"`

	// FieldRenames maps AWS member names to Terraform attribute names.
	// When a name is not listed here the extractor applies the default
	// CamelCase → snake_case conversion.
	FieldRenames map[string]string `yaml:"field_renames"`

	// --- Attribute-map extraction fields ---

	// AttributeMapEnum is the Smithy shape name (without namespace) of the
	// enum whose member values encode the real attribute names.
	AttributeMapEnum string `yaml:"attribute_map_enum"`

	// ExplicitFields is used when field names are not machine-readable in the
	// model. Each entry is a hand-authored
	// field descriptor.
	ExplicitFields []ExplicitField `yaml:"explicit_fields"`
}

// Lifecycle names the operation (without namespace) for each resource verb.
type Lifecycle struct {
	Create     string   `yaml:"create"`
	Put        string   `yaml:"put"`
	Read       string   `yaml:"read"`
	Update     string   `yaml:"update"`
	Delete     string   `yaml:"delete"`
	List       string   `yaml:"list"`
	Operations []string `yaml:"operations"`
}

// ExplicitField is a hand-authored field descriptor for services where
// field names are only present in documentation strings.
type ExplicitField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // "string", "bool", "int64", "float64"
	Required bool   `yaml:"required"`
}

// LoadFile reads and parses the YAML mapping file at path.
func LoadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading mapping file: %w", err)
	}

	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing mapping file: %w", err)
	}
	if f.Services == nil {
		f.Services = make(map[string]string)
	}
	if f.Resources == nil {
		f.Resources = make(map[string]*ResourceMapping)
	}

	return &f, nil
}

// IsSuppressed reports whether fieldName appears in the suppress list.
func (m *ResourceMapping) IsSuppressed(fieldName string) bool {
	for _, s := range m.SuppressFields {
		if s == fieldName {
			return true
		}
	}
	return false
}

// TFName returns the Terraform attribute name for an AWS member name:
// it checks FieldRenames first, then falls back to CamelToSnake.
func (m *ResourceMapping) TFName(awsName string) string {
	if m.FieldRenames != nil {
		if override, ok := m.FieldRenames[awsName]; ok {
			return override
		}
	}
	return CamelToSnake(awsName)
}
