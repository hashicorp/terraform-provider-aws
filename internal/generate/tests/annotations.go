// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tests

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	acctestgen "github.com/hashicorp/terraform-provider-aws/internal/acctest/generate"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	namesgen "github.com/hashicorp/terraform-provider-aws/names/generate"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CommonArgs struct {
	Name           string // Resource Type Name
	TypeName       string // Terraform Type Name
	Implementation common.Implementation

	// CheckDestroy
	CheckDestroyNoop bool
	destroyTakesNoT  bool

	// CheckExists
	HasExistsFunc  bool
	ExistsTypeName string
	existsTakesNoT bool

	// Import
	NoImport                   bool
	ImportStateID              string
	importStateIDAttribute     string
	importStateIDAttributes    []string
	importStateIDAttributesSep string
	ImportStateIDFunc          string
	ImportIgnore               []string
	plannableImportAction      importAction

	// Serialization
	Serialize              bool
	SerializeDelay         bool
	SerializeParallelTests bool

	// PreChecks
	PreChecks           []CodeBlock
	PreCheckRegions     []string
	PreChecksWithRegion []CodeBlock

	UseAlternateAccount     bool
	AlternateRegionProvider bool
	AlternateRegionTfVars   bool

	Generator     string
	generatorSeen bool

	RequiredEnvVars      []string
	RequiredEnvVarValues []string

	GoImports         []common.GoImport
	InitCodeBlocks    []CodeBlock
	AdditionalTfVars_ map[string]TFVar

	// Resource Identity Versions
	HasNoPreExistingResource bool
	PreIdentityVersion       *version.Version
	IdentityVersions         map[int64]*version.Version
}

func InitCommonArgs() CommonArgs {
	return CommonArgs{
		AdditionalTfVars_: make(map[string]TFVar, 0),
		HasExistsFunc:     true,
		IdentityVersions:  make(map[int64]*version.Version, 0),
	}
}

func (c CommonArgs) HasImportStateIDAttribute() bool {
	return c.importStateIDAttribute != ""
}

func (c CommonArgs) ImportStateIDAttribute() string {
	return namesgen.ConstOrQuote(c.importStateIDAttribute)
}

func (c *CommonArgs) SetImportStateIDAttribute(attrName string) {
	c.importStateIDAttribute = attrName
}

func (c CommonArgs) HasImportStateIDAttributes() bool {
	return len(c.importStateIDAttributes) > 0
}

func (c CommonArgs) ImportStateIDAttributes() string {
	quoted := make([]string, len(c.importStateIDAttributes))
	for i, attr := range c.importStateIDAttributes {
		quoted[i] = namesgen.ConstOrQuote(attr)
	}
	return strings.Join(quoted, ", ")
}

func (c CommonArgs) ImportStateIDAttributesFirst() string {
	if len(c.importStateIDAttributes) > 0 {
		return namesgen.ConstOrQuote(c.importStateIDAttributes[0])
	}
	return ""
}

func (c CommonArgs) ImportStateIDAttributesSep() string {
	if c.importStateIDAttributesSep == "" {
		return namesgen.ConstOrQuote(",")
	}
	sep := c.importStateIDAttributesSep
	// Check if it looks like a package-qualified constant (e.g., "intflex.ResourceIdSeparator")
	if strings.Contains(sep, ".") {
		return sep // Return as-is (constant reference)
	}
	// Check if it starts with uppercase (unqualified constant)
	if len(sep) > 0 && sep[0] >= 'A' && sep[0] <= 'Z' {
		return sep // Return as-is (constant reference)
	}
	return namesgen.ConstOrQuote(sep) // Quote it (string literal)
}

func (c *CommonArgs) SetImportStateIDAttributes(attrNames []string, sep string) {
	c.importStateIDAttributes = attrNames
	c.importStateIDAttributesSep = sep
}

func (c CommonArgs) HasImportIgnore() bool {
	return len(c.ImportIgnore) > 0
}

func (c CommonArgs) PlannableResourceAction() string {
	if c.plannableImportAction == importActionUnset {
		return importActionNoop.String()
	}
	return c.plannableImportAction.String()
}

func (c CommonArgs) AdditionalTfVars() map[string]TFVar {
	return tfmaps.ApplyToAllKeys(c.AdditionalTfVars_, func(k string) string {
		return acctestgen.ConstOrQuote(k)
	})
}

func (c CommonArgs) DestroyTakesT() bool {
	return !c.destroyTakesNoT
}

func (c CommonArgs) ExistsTakesT() bool {
	return !c.existsTakesNoT
}

type importAction int

const (
	importActionUnset importAction = iota
	importActionNoop
	importActionUpdate
	importActionReplace
)

func (i importAction) String() string {
	switch i {
	case importActionNoop:
		return "NoOp"

	case importActionUpdate:
		return "Update"

	case importActionReplace:
		return "Replace"

	default:
		return ""
	}
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
	if attr, ok := args.Keyword["name"]; ok {
		stuff.Name = strings.ReplaceAll(attr, " ", "")
	}

	// DestroyCheck
	if attr, ok := args.Keyword["checkDestroyNoop"]; ok {
		if b, err := common.ParseBoolAttr("checkDestroyNoop", attr); err != nil {
			return err
		} else {
			stuff.CheckDestroyNoop = b
			stuff.GoImports = append(stuff.GoImports,
				common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
	}

	if attr, ok := args.Keyword["destroyTakesT"]; ok {
		if b, err := common.ParseBoolAttr("destroyTakesT", attr); err != nil {
			return err
		} else {
			stuff.destroyTakesNoT = !b
		}
	}

	// ExistsCheck
	if attr, ok := args.Keyword["hasExistsFunction"]; ok {
		if b, err := common.ParseBoolAttr("hasExistsFunction", attr); err != nil {
			return err
		} else {
			stuff.HasExistsFunc = b
		}
	}

	if attr, ok := args.Keyword["existsType"]; ok {
		if typeName, importSpec, err := common.ParseIdentifierSpec(attr); err != nil {
			return fmt.Errorf("%s: %w", attr, err)
		} else {
			stuff.ExistsTypeName = typeName
			if importSpec != nil {
				stuff.GoImports = append(stuff.GoImports, *importSpec)
			}
		}
	}

	if attr, ok := args.Keyword["existsTakesT"]; ok {
		if b, err := common.ParseBoolAttr("existsTakesT", attr); err != nil {
			return err
		} else {
			stuff.existsTakesNoT = !b
		}
	}

	// Import
	if attr, ok := args.Keyword["importIgnore"]; ok {
		stuff.ImportIgnore = strings.Split(attr, ";")
		for i, val := range stuff.ImportIgnore {
			stuff.ImportIgnore[i] = namesgen.ConstOrQuote(val)
		}
		if stuff.plannableImportAction == importActionUnset {
			stuff.plannableImportAction = importActionUpdate
		}
	}

	if attr, ok := args.Keyword["importStateId"]; ok {
		stuff.ImportStateID = attr
	}

	if attr, ok := args.Keyword["importStateIdAttribute"]; ok {
		stuff.importStateIDAttribute = attr
	}

	if attr, ok := args.Keyword["importStateIdAttributes"]; ok {
		attrs := strings.Split(attr, ";")
		for i, a := range attrs {
			attrs[i] = strings.TrimSpace(a)
		}
		sep := ","
		if s, ok := args.Keyword["importStateIdAttributesSep"]; ok {
			sep = s
			// If separator references intflex, add the import
			switch {
			case strings.Contains(sep, "intflex."):
				stuff.GoImports = append(stuff.GoImports, common.GoImport{
					Path:  "github.com/hashicorp/terraform-provider-aws/internal/flex",
					Alias: "intflex",
				})
			case strings.Contains(sep, "flex."):
				stuff.GoImports = append(stuff.GoImports, common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/flex",
				})
			}
		}
		stuff.SetImportStateIDAttributes(attrs, sep)
	}

	if attr, ok := args.Keyword["importStateIdFunc"]; ok {
		stuff.ImportStateIDFunc = attr
	}

	if attr, ok := args.Keyword["plannableImportAction"]; ok {
		switch attr {
		case importActionNoop.String():
			stuff.plannableImportAction = importActionNoop

		case importActionUpdate.String():
			stuff.plannableImportAction = importActionUpdate

		case importActionReplace.String():
			stuff.plannableImportAction = importActionReplace

		default:
			return fmt.Errorf("invalid plannableImportAction value %q: Must be one of %s.", attr, []string{importActionNoop.String(), importActionUpdate.String(), importActionReplace.String()})
		}
	}

	// Serialization
	if attr, ok := args.Keyword["serialize"]; ok {
		if b, err := common.ParseBoolAttr("serialize", attr); err != nil {
			return err
		} else {
			stuff.Serialize = b
		}
	}

	if attr, ok := args.Keyword["serializeParallelTests"]; ok {
		if b, err := common.ParseBoolAttr("serializeParallelTests", attr); err != nil {
			return err
		} else {
			stuff.SerializeParallelTests = b
		}
	}

	if attr, ok := args.Keyword["serializeDelay"]; ok {
		if b, err := common.ParseBoolAttr("serializeDelay", attr); err != nil {
			return err
		} else {
			stuff.SerializeDelay = b
		}
	}

	// PreChecks
	if attr, ok := args.Keyword["preCheck"]; ok {
		if code, importSpec, err := common.ParseIdentifierSpec(attr); err != nil {
			return fmt.Errorf("%s: %w", attr, err)
		} else {
			stuff.PreChecks = append(stuff.PreChecks, CodeBlock{
				Code: fmt.Sprintf("%s(ctx, t)", code),
			})
			if importSpec != nil {
				stuff.GoImports = append(stuff.GoImports, *importSpec)
			}
		}
	}

	if attr, ok := args.Keyword["preCheckRegion"]; ok {
		regions := strings.Split(attr, ";")
		stuff.PreCheckRegions = tfslices.ApplyToAll(regions, func(s string) string {
			return endpointsConstOrQuote(s)
		})
		stuff.GoImports = append(stuff.GoImports,
			common.GoImport{
				Path: "github.com/hashicorp/aws-sdk-go-base/v2/endpoints",
			},
		)
	}

	if attr, ok := args.Keyword["preCheckWithRegion"]; ok {
		if code, importSpec, err := common.ParseIdentifierSpec(attr); err != nil {
			return fmt.Errorf("%s: %w", attr, err)
		} else {
			stuff.PreChecksWithRegion = append(stuff.PreChecksWithRegion, CodeBlock{
				Code: code,
			})
			if importSpec != nil {
				stuff.GoImports = append(stuff.GoImports, *importSpec)
			}
		}
	}

	if attr, ok := args.Keyword["requireEnvVar"]; ok {
		stuff.RequiredEnvVars = append(stuff.RequiredEnvVars, attr)
	}

	if attr, ok := args.Keyword["requireEnvVarValue"]; ok {
		stuff.RequiredEnvVarValues = append(stuff.RequiredEnvVarValues, attr)
	}

	if attr, ok := args.Keyword["useAlternateAccount"]; ok {
		if b, err := common.ParseBoolAttr("useAlternateAccount", attr); err != nil {
			return err
		} else if b {
			stuff.UseAlternateAccount = true
			stuff.PreChecks = append(stuff.PreChecks, CodeBlock{
				Code: "acctest.PreCheckAlternateAccount(t)",
			})
			stuff.GoImports = append(stuff.GoImports,
				// Required to initialize `schema.Provider` map.
				common.GoImport{
					Path: "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema",
				},
				common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
	}

	if attr, ok := args.Keyword["altRegionProvider"]; ok {
		if b, err := common.ParseBoolAttr("altRegionProvider", attr); err != nil {
			return err
		} else {
			stuff.AlternateRegionProvider = b
			stuff.PreChecks = append(stuff.PreChecks, CodeBlock{
				Code: "acctest.PreCheckMultipleRegion(t, 2)",
			})
			stuff.GoImports = append(stuff.GoImports,
				common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
	}

	if attr, ok := args.Keyword["altRegionTfVars"]; ok {
		if b, err := common.ParseBoolAttr("altRegionTfVars", attr); err != nil {
			return err
		} else {
			stuff.AlternateRegionTfVars = b
			stuff.PreChecks = append(stuff.PreChecks, CodeBlock{
				Code: "acctest.PreCheckMultipleRegion(t, 2)",
			})
			stuff.GoImports = append(stuff.GoImports,
				common.GoImport{
					Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
				},
			)
		}
	}

	// TF Variables
	if attr, ok := args.Keyword["generator"]; ok {
		stuff.generatorSeen = true
		if attr != "false" {
			if funcName, importSpec, err := common.ParseIdentifierSpec(attr); err != nil {
				return fmt.Errorf("%s: %w", attr, err)
			} else {
				stuff.Generator = funcName
				if importSpec != nil {
					stuff.GoImports = append(stuff.GoImports, *importSpec)
				}
			}
		}
	}

	if attr, ok := args.Keyword["emailAddress"]; ok {
		varName := "address"
		if len(attr) > 0 {
			varName = attr
		}
		stuff.GoImports = append(stuff.GoImports,
			common.GoImport{
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
			common.GoImport{
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
			common.GoImport{
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
			common.GoImport{
				Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
			},
		)
		stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
			Code: fmt.Sprintf("%s := acctest.RandIntRange(t, %s,%s)", varName, parts[0], parts[1]),
		})
		stuff.AdditionalTfVars_[varName] = TFVar{
			GoVarName: varName,
			Type:      TFVarTypeInt,
		}
	}

	if attr, ok := args.Keyword["randomIPv4Address"]; ok {
		varName := "rIPv4Address"
		stuff.GoImports = append(stuff.GoImports,
			common.GoImport{
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

	if attr, ok := args.Keyword["tlsEcdsaPublicKeyPem"]; ok {
		if _, err := common.ParseBoolAttr("tlsEcdsaPublicKeyPem", attr); err != nil {
			return err
		} else {
			varName := "rTlsEcdsaPublicKeyPem"
			stuff.InitCodeBlocks = append(stuff.InitCodeBlocks, CodeBlock{
				Code: fmt.Sprintf(`privateKey := acctest.TLSECDSAPrivateKeyPEM(t, "P-384")
%s, _ := acctest.TLSECDSAPublicKeyPEM(t, privateKey)`, varName),
			})
			stuff.AdditionalTfVars_[varName] = TFVar{
				GoVarName: varName,
				Type:      TFVarTypeString,
			}
		}
	}

	// Resource Identity Versions
	if attr, ok := args.Keyword["preIdentityVersion"]; ok {
		version, err := version.NewVersion(attr)
		if err != nil {
			return fmt.Errorf("invalid preIdentityVersion value: %q. Should be version value.", attr)
		}
		stuff.PreIdentityVersion = version
	}

	if attr, ok := args.Keyword["hasNoPreExistingResource"]; ok {
		if b, err := common.ParseBoolAttr("hasNoPreExistingResource", attr); err != nil {
			return err
		} else {
			stuff.HasNoPreExistingResource = b
		}
	}

	if attr, ok := args.Keyword["identityVersion"]; ok {
		parts := strings.Split(attr, ";")
		if len(parts) != 2 {
			return fmt.Errorf("invalid identityVersion value: %q. Should be in format <identity version>;<provider version>.", attr)
		}
		var identityVersion int64
		if i, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
			return fmt.Errorf("invalid identity version value: %q. Should be integer value.", parts[0])
		} else {
			identityVersion = i
		}
		providerVersion, err := version.NewVersion(parts[1])
		if err != nil {
			return fmt.Errorf("invalid provider version value: %q. Should be version value.", parts[1])
		}
		stuff.IdentityVersions[identityVersion] = providerVersion
	}

	return nil
}

func endpointsConstOrQuote(region string) string {
	var buf strings.Builder
	buf.WriteString("endpoints.")

	caser := cases.Title(language.Und, cases.NoLower)
	for part := range strings.SplitSeq(region, "-") {
		buf.WriteString(caser.String(part))
	}
	buf.WriteString("RegionID")

	return buf.String()
}

func Configure(d *CommonArgs) error {
	if d.Name == "" {
		return errors.New("no name parameter set")
	}

	if !d.generatorSeen {
		d.Generator = "acctest.RandomWithPrefix(t, acctest.ResourcePrefix)"
		d.GoImports = append(d.GoImports,
			common.GoImport{
				Path: "github.com/hashicorp/terraform-provider-aws/internal/acctest",
			},
		)
	}

	return nil
}
