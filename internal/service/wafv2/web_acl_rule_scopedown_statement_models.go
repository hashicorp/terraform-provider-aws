// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
)

// Level-based scope_down statement models

// webACLRuleScopeDownStatementLevel0Model contains only leaf statements; no logical operators.
type webACLRuleScopeDownStatementLevel0Model struct {
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownAndStatementLevel0Model wraps level-0 statements for an AND operation.
type webACLRuleScopeDownAndStatementLevel0Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel0Model] `tfsdk:"statement"`
}

// webACLRuleScopeDownNotStatementLevel0Model wraps a single level-0 statement for a NOT operation.
type webACLRuleScopeDownNotStatementLevel0Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel0Model] `tfsdk:"statement"`
}

// webACLRuleScopeDownOrStatementLevel0Model wraps level-0 statements for an OR operation.
type webACLRuleScopeDownOrStatementLevel0Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel0Model] `tfsdk:"statement"`
}

// webACLRuleScopeDownStatementLevel1Model contains leaf statements plus and/not/or at level 0.
type webACLRuleScopeDownStatementLevel1Model struct {
	AndStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementLevel0Model]       `tfsdk:"and_statement"`
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel0Model]       `tfsdk:"not_statement"`
	OrStatement                       fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementLevel0Model]        `tfsdk:"or_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownAndStatementLevel1ChildModel contains leaf statements plus not/or at level 0.
// AWS WAF doesn't accept an AND statement directly inside an AND statement.
type webACLRuleScopeDownAndStatementLevel1ChildModel struct {
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel0Model]       `tfsdk:"not_statement"`
	OrStatement                       fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementLevel0Model]        `tfsdk:"or_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownOrStatementLevel1ChildModel contains leaf statements plus and/not at level 0.
// AWS WAF doesn't accept an OR statement directly inside an OR statement.
type webACLRuleScopeDownOrStatementLevel1ChildModel struct {
	AndStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementLevel0Model]       `tfsdk:"and_statement"`
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel0Model]       `tfsdk:"not_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownAndStatementLevel1Model wraps level-1 statements for an AND operation.
type webACLRuleScopeDownAndStatementLevel1Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementLevel1ChildModel] `tfsdk:"statement"`
}

// webACLRuleScopeDownNotStatementLevel1Model wraps a single level-1 statement for a NOT operation.
type webACLRuleScopeDownNotStatementLevel1Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel1Model] `tfsdk:"statement"`
}

// webACLRuleScopeDownOrStatementLevel1Model wraps level-1 statements for an OR operation.
type webACLRuleScopeDownOrStatementLevel1Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementLevel1ChildModel] `tfsdk:"statement"`
}

// webACLRuleScopeDownStatementLevel2Model contains leaf statements plus and/not/or at level 1.
type webACLRuleScopeDownStatementLevel2Model struct {
	AndStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementLevel1Model]       `tfsdk:"and_statement"`
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel1Model]       `tfsdk:"not_statement"`
	OrStatement                       fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementLevel1Model]        `tfsdk:"or_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownAndStatementLevel2ChildModel contains leaf statements plus not/or at level 1.
// AWS WAF doesn't accept an AND statement directly inside an AND statement.
type webACLRuleScopeDownAndStatementLevel2ChildModel struct {
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel1Model]       `tfsdk:"not_statement"`
	OrStatement                       fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementLevel1Model]        `tfsdk:"or_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// webACLRuleScopeDownOrStatementLevel2ChildModel contains leaf statements plus and/not at level 1.
// AWS WAF doesn't accept an OR statement directly inside an OR statement.
type webACLRuleScopeDownOrStatementLevel2ChildModel struct {
	AndStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementLevel1Model]       `tfsdk:"and_statement"`
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementLevel1Model]       `tfsdk:"not_statement"`
	IPSetReferenceStatement           fwtypes.ListNestedObjectValueOf[webACLRuleIPSetReferenceStatementModel]           `tfsdk:"ip_set_reference_statement"`
	GeoMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleGeoMatchStatementModel]                 `tfsdk:"geo_match_statement"`
	ByteMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleByteMatchStatementModel]                `tfsdk:"byte_match_statement"`
	SqliMatchStatement                fwtypes.ListNestedObjectValueOf[webACLRuleSqliMatchStatementModel]                `tfsdk:"sqli_match_statement"`
	XssMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleXssMatchStatementModel]                 `tfsdk:"xss_match_statement"`
	SizeConstraintStatement           fwtypes.ListNestedObjectValueOf[webACLRuleSizeConstraintStatementModel]           `tfsdk:"size_constraint_statement"`
	RegexMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleRegexMatchStatementModel]               `tfsdk:"regex_match_statement"`
	RegexPatternSetReferenceStatement fwtypes.ListNestedObjectValueOf[webACLRuleRegexPatternSetReferenceStatementModel] `tfsdk:"regex_pattern_set_reference_statement"`
	LabelMatchStatement               fwtypes.ListNestedObjectValueOf[webACLRuleLabelMatchStatementModel]               `tfsdk:"label_match_statement"`
	AsnMatchStatement                 fwtypes.ListNestedObjectValueOf[webACLRuleAsnMatchStatementModel]                 `tfsdk:"asn_match_statement"`
}

// Type aliases for the top-level statement model (used by the resource).
type webACLRuleScopeDownStatementModel = webACLRuleScopeDownStatementLevel2Model
type webACLRuleScopeDownAndStatementModel = webACLRuleScopeDownAndStatementLevel1Model
type webACLRuleScopeDownNotStatementModel = webACLRuleScopeDownNotStatementLevel1Model
type webACLRuleScopeDownOrStatementModel = webACLRuleScopeDownOrStatementLevel1Model

// Schema block functions for each level.

// scopeDownLeafStatementBlocks returns only leaf statements (no managed_rule_group, rate_based,
// or rule_group_reference) to avoid initialization cycles when used inside scope_down_statement
// which is itself referenced by managed_rule_group_statement and rate_based_statement.
func scopeDownLeafStatementBlocks(ctx context.Context) map[string]schema.Block {
	return map[string]schema.Block{
		"ip_set_reference_statement":            ipSetReferenceStatementBlock(ctx),
		"geo_match_statement":                   geoMatchStatementBlock(ctx),
		"byte_match_statement":                  byteMatchStatementBlock(ctx),
		"sqli_match_statement":                  sqliMatchStatementBlock(ctx),
		"xss_match_statement":                   xssMatchStatementBlock(ctx),
		"size_constraint_statement":             sizeConstraintStatementBlock(ctx),
		"regex_match_statement":                 regexMatchStatementBlock(ctx),
		"regex_pattern_set_reference_statement": regexPatternSetReferenceStatementBlock(ctx),
		"label_match_statement":                 labelMatchStatementBlock(ctx),
		"asn_match_statement":                   asnMatchStatementBlock(ctx),
	}
}

// Level-based scope_down blocks — mirrors web_acl_rule_statement_models_gen.go.
// Level 0 blocks: leaf statements only (no and/not/or).
// Level 1 blocks: leaf + and/not/or at level 0.
//   AND can't directly contain AND and OR can't directly contain OR, so AND and OR each get their own level-1 statement block.
// scopeDownStatementBlock (used in scope_down_statement) references and/not/or at level 1.

// scopeDownStatementLevel0MultipleBlock: at least two level-0 statements, used by AND/OR at level 0.
var scopeDownStatementLevel0MultipleBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel0Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested leaf statements for logical operations within scope_down_statement.",
	}
})

func scopeDownAndStatementLevel0ChildBlock(ctx context.Context) schema.Block {
	return scopeDownStatementLevel0MultipleBlock(ctx)
}

func scopeDownOrStatementLevel0ChildBlock(ctx context.Context) schema.Block {
	return scopeDownStatementLevel0MultipleBlock(ctx)
}

// scopeDownNotStatementLevel0ChildBlock: exactly one level-0 statement, used by NOT at level 0.
var scopeDownNotStatementLevel0ChildBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel0Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested leaf statement for NOT operation within scope_down_statement.",
	}
})

// scopeDownAndStatementLevel0Block: AND at level 0, wrapping multiple level-0 statements.
var scopeDownAndStatementLevel0Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownAndStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownAndStatementLevel0ChildBlock(ctx),
			},
		},
		Description: "Logical AND statement.",
	}
})

// scopeDownNotStatementLevel0Block: NOT at level 0, wrapping a single level-0 statement.
var scopeDownNotStatementLevel0Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownNotStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownNotStatementLevel0ChildBlock(ctx),
			},
		},
		Description: "Logical NOT statement.",
	}
})

// scopeDownOrStatementLevel0Block: OR at level 0, wrapping multiple level-0 statements.
var scopeDownOrStatementLevel0Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownOrStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownOrStatementLevel0ChildBlock(ctx),
			},
		},
		Description: "Logical OR statement.",
	}
})

// scopeDownAndStatementLevel1ChildBlock: at least two level-1 statements (leaf + not/or at level 0), used by AND at level 1.
var scopeDownAndStatementLevel1ChildBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["not_statement"] = scopeDownNotStatementLevel0Block(ctx)
	blocks["or_statement"] = scopeDownOrStatementLevel0Block(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownAndStatementLevel1ChildModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statements for AND operation within scope_down_statement.",
	}
})

// scopeDownOrStatementLevel1ChildBlock: at least two level-1 statements (leaf + and/not at level 0), used by OR at level 1.
var scopeDownOrStatementLevel1ChildBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementLevel0Block(ctx)
	blocks["not_statement"] = scopeDownNotStatementLevel0Block(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownOrStatementLevel1ChildModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statements for OR operation within scope_down_statement.",
	}
})

// scopeDownNotStatementLevel1ChildBlock: exactly one level-1 statement, used by NOT at level 1.
var scopeDownNotStatementLevel1ChildBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementLevel0Block(ctx)
	blocks["not_statement"] = scopeDownNotStatementLevel0Block(ctx)
	blocks["or_statement"] = scopeDownOrStatementLevel0Block(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel1Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statement for NOT operation within scope_down_statement.",
	}
})

// scopeDownAndStatementLevel1Block: AND at level 1, wrapping multiple level-1 statements.
var scopeDownAndStatementLevel1Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownAndStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownAndStatementLevel1ChildBlock(ctx),
			},
		},
		Description: "Logical AND statement.",
	}
})

// scopeDownNotStatementLevel1Block: NOT at level 1, wrapping a single level-1 statement.
var scopeDownNotStatementLevel1Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownNotStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownNotStatementLevel1ChildBlock(ctx),
			},
		},
		Description: "Logical NOT statement.",
	}
})

// scopeDownOrStatementLevel1Block: OR at level 1, wrapping multiple level-1 statements.
var scopeDownOrStatementLevel1Block = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownOrStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownOrStatementLevel1ChildBlock(ctx),
			},
		},
		Description: "Logical OR statement.",
	}
})

// scopeDownStatementBlock returns the top-level scopedown statement block schema.
var scopeDownStatementBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementLevel1Block(ctx)
	blocks["not_statement"] = scopeDownNotStatementLevel1Block(ctx)
	blocks["or_statement"] = scopeDownOrStatementLevel1Block(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel2Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Scope down statement for managed rule groups.",
	}
})
