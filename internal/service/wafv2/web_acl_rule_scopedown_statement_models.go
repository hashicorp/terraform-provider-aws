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

// webACLRuleScopeDownStatementLevel1InAndModel contains leaf statements plus not/or at level 0.
// AWS WAF doesn't accept an AND statement directly inside an AND statement.
type webACLRuleScopeDownStatementLevel1InAndModel struct {
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

// webACLRuleScopeDownStatementLevel1InOrModel contains leaf statements plus and/not at level 0.
// AWS WAF doesn't accept an OR statement directly inside an OR statement.
type webACLRuleScopeDownStatementLevel1InOrModel struct {
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
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel1InAndModel] `tfsdk:"statement"`
}

// webACLRuleScopeDownNotStatementLevel1Model wraps a single level-1 statement for a NOT operation.
type webACLRuleScopeDownNotStatementLevel1Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel1Model] `tfsdk:"statement"`
}

// webACLRuleScopeDownOrStatementLevel1Model wraps level-1 statements for an OR operation.
type webACLRuleScopeDownOrStatementLevel1Model struct {
	Statements fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownStatementLevel1InOrModel] `tfsdk:"statement"`
}

// Type aliases so the rest of the code (flattenScopeDownStatement, Expand, etc.) keeps working unchanged.
type webACLRuleScopeDownAndStatementModel = webACLRuleScopeDownAndStatementLevel1Model
type webACLRuleScopeDownNotStatementModel = webACLRuleScopeDownNotStatementLevel1Model
type webACLRuleScopeDownOrStatementModel = webACLRuleScopeDownOrStatementLevel1Model

type webACLRuleScopeDownStatementModel struct {
	AndStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownAndStatementModel]             `tfsdk:"and_statement"`
	NotStatement                      fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownNotStatementModel]             `tfsdk:"not_statement"`
	OrStatement                       fwtypes.ListNestedObjectValueOf[webACLRuleScopeDownOrStatementModel]              `tfsdk:"or_statement"`
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
// Level 1 blocks: leaf + and/not/or at level 0. As in AWS WAF, AND can't directly contain AND and
// OR can't directly contain OR, so AND and OR each get their own level-1 statement block.
// scopeDownStatementBlock (used in scope_down_statement) references and/not/or at level 1.
// No var/var cycles since each level only refers to the level below.

// scopeDownStatementBlockLevel0Multiple: at least two level-0 statements, used by AND/OR at level 0.
var scopeDownStatementBlockLevel0Multiple = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel0Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: scopeDownLeafStatementBlocks(ctx),
		},
		Description: "Nested leaf statements for logical operations within scope_down_statement.",
	}
})

// scopeDownStatementBlockLevel0Single: exactly one level-0 statement, used by NOT at level 0.
var scopeDownStatementBlockLevel0Single = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel0Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: scopeDownLeafStatementBlocks(ctx),
		},
		Description: "Nested leaf statement for NOT operation within scope_down_statement.",
	}
})

var scopeDownAndStatementBlockLevel0 = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownAndStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel0Multiple(ctx),
			},
		},
		Description: "Logical AND statement.",
	}
})

var scopeDownNotStatementBlockLevel0 = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownNotStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel0Single(ctx),
			},
		},
		Description: "Logical NOT statement.",
	}
})

var scopeDownOrStatementBlockLevel0 = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownOrStatementLevel0Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel0Multiple(ctx),
			},
		},
		Description: "Logical OR statement.",
	}
})

// scopeDownStatementBlockLevel1InAnd: at least two level-1 statements (leaf + not/or at level 0), used by AND at level 1.
var scopeDownStatementBlockLevel1InAnd = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["not_statement"] = scopeDownNotStatementBlockLevel0(ctx)
	blocks["or_statement"] = scopeDownOrStatementBlockLevel0(ctx)
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel1InAndModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statements for AND operation within scope_down_statement.",
	}
})

// scopeDownStatementBlockLevel1InOr: at least two level-1 statements (leaf + and/not at level 0), used by OR at level 1.
var scopeDownStatementBlockLevel1InOr = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementBlockLevel0(ctx)
	blocks["not_statement"] = scopeDownNotStatementBlockLevel0(ctx)
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel1InOrModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(2),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statements for OR operation within scope_down_statement.",
	}
})

// scopeDownStatementBlockLevel1Single: exactly one level-1 statement, used by NOT at level 1.
var scopeDownStatementBlockLevel1Single = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementBlockLevel0(ctx)
	blocks["not_statement"] = scopeDownNotStatementBlockLevel0(ctx)
	blocks["or_statement"] = scopeDownOrStatementBlockLevel0(ctx)
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementLevel1Model](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Nested statement for NOT operation within scope_down_statement.",
	}
})

// scopeDownAndStatementBlock: AND at level 1, wrapping level-1 statements other than AND.
// Type alias: webACLRuleScopeDownAndStatementModel = webACLRuleScopeDownAndStatementLevel1Model.
var scopeDownAndStatementBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownAndStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel1InAnd(ctx),
			},
		},
		Description: "Logical AND statement.",
	}
})

// scopeDownNotStatementBlock: NOT at level 1, wrapping a single level-1 statement.
// Type alias: webACLRuleScopeDownNotStatementModel = webACLRuleScopeDownNotStatementLevel1Model.
var scopeDownNotStatementBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownNotStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel1Single(ctx),
			},
		},
		Description: "Logical NOT statement.",
	}
})

// scopeDownOrStatementBlock: OR at level 1, wrapping level-1 statements other than OR.
// Type alias: webACLRuleScopeDownOrStatementModel = webACLRuleScopeDownOrStatementLevel1Model.
var scopeDownOrStatementBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownOrStatementLevel1Model](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"statement": scopeDownStatementBlockLevel1InOr(ctx),
			},
		},
		Description: "Logical OR statement.",
	}
})

// scopeDownStatementBlock adds logical statements (and/or/not) to the leaf statement blocks.
// Uses scope-down-specific models to avoid type cycles with managed_rule_group and rate_based.
var scopeDownStatementBlock = tfsync.OnceValueCtx(func(ctx context.Context) schema.Block {
	blocks := scopeDownLeafStatementBlocks(ctx)
	blocks["and_statement"] = scopeDownAndStatementBlock(ctx)
	blocks["not_statement"] = scopeDownNotStatementBlock(ctx)
	blocks["or_statement"] = scopeDownOrStatementBlock(ctx)

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[webACLRuleScopeDownStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: blocks,
		},
		Description: "Scope down statement for managed rule groups.",
	}
})
