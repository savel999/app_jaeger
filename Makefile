SERVICE_NAME ?= service-multimedia
GQLGEN ?= github.com/99designs/gqlgen
REFLEX ?= github.com/cespare/reflex
DATALOADEN ?= github.com/vektah/dataloaden
SRC_GEN = git.rn/gm/service-multimedia/internal/graphql
SRC_GEN_DATALOADERS ?= ./internal/graphql/dataloaders
DOT_ENV ?= github.com/joho/godotenv
GOLANGCI_LINT ?= $(GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.38.0
CGO_CFLAGS_ALLOW=-Xpreprocessor
SEP			 ?= "========================================================"


.PHONY: gen-it-client-services
gen-ft-client-services:
	cd ./tools && go run ./... --config .gqlgenc-service-multimedia.yml

define _info
	$(call _echoColor,$1,6)
endef

define _hint
	$(call _echoColor,$1,8)
endef

define _succ
	$(call _echoColor,$1,2)
endef

define _warn
	$(call _echoColor,$1,3)
endef

define _mega
	$(call _echoColor,$1,13)
endef

define _error
	$(call _echoColor,$1,1)
endef

define _echoColor
	@tput setaf $2
	@echo $1
	@tput sgr0
endef

################################################################################################################
