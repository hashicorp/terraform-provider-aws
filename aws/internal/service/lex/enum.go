package lex

import "time"

const (
	LexBotCreateTimeout      = 5 * time.Minute
	LexBotDeleteTimeout      = 5 * time.Minute
	LexBotUpdateTimeout      = 5 * time.Minute
	LexIntentDeleteTimeout   = 5 * time.Minute
	LexSlotTypeDeleteTimeout = 5 * time.Minute
)

const (
	LexBotVersionLatest = "$LATEST"
)
