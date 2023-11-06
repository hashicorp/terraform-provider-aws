package connectcases

const (
	fieldTypeText         = "Text"
	fieldTypeNumber       = "Number"
	fieldTypeBoolean      = "Boolean"
	fieldTypeDateTime     = "DateTime"
	fieldTypeSingleSelect = "SingleSelect"
	fieldTypeUrl          = "Url"
)

func fieldType_Values() []string {
	return []string{
		fieldTypeText,
		fieldTypeNumber,
		fieldTypeBoolean,
		fieldTypeDateTime,
		fieldTypeSingleSelect,
		fieldTypeUrl,
	}
}
