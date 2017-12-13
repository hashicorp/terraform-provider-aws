package hcldec

// Verify that all of our spec types implement the necessary interfaces
var objectSpecAsSpec Spec = ObjectSpec(nil)
var tupleSpecAsSpec Spec = TupleSpec(nil)
var attrSpecAsSpec Spec = (*AttrSpec)(nil)
var literalSpecAsSpec Spec = (*LiteralSpec)(nil)
var exprSpecAsSpec Spec = (*ExprSpec)(nil)
var blockSpecAsSpec Spec = (*BlockSpec)(nil)
var blockListSpecAsSpec Spec = (*BlockListSpec)(nil)
var blockSetSpecAsSpec Spec = (*BlockSetSpec)(nil)
var blockMapSpecAsSpec Spec = (*BlockMapSpec)(nil)
var blockLabelSpecAsSpec Spec = (*BlockLabelSpec)(nil)
var defaultSpecAsSpec Spec = (*DefaultSpec)(nil)
