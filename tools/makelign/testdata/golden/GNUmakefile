SHELL := /bin/bash

# Minimal GNUmakefile fixture used by makelign tests.
#
# It includes one example of each interesting case:
#   - documented target in .PHONY (good)
#   - documented CI target referenced in CI doc (good)
#   - documented target NOT in .PHONY (phony-missing)
#   - .PHONY entry that points at no target (phony-ghost)
#   - documented target NOT in cheat sheet (cheatsheet-missing via deps)

GO_VER ?= go

default: build

build: ## Build provider
	@echo build

ci: build ## [CI] Run all CI checks
	@echo ci

docs: ## [CI] Documentation checks
	@echo docs

deps: ## Install dependencies
	@echo deps

missing-from-phony: ## Documented but missing from .PHONY
	@echo missing

.PHONY: \
	build \
	ci \
	docs \
	ghost-target \
	deps
