MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

CF_DIR := $(MAKEFILE_DIR)

AWS_PROFILE := rvpay-testing
AWS_REGION ?= us-east-1

generate:
	@echo "No generation implemented for infra"

lint:
	@echo "No linting implemented for infra"

test:
	@echo "Validating Cloudformation templates in: $(CF_DIR)"
	for f in $$(find "$(CF_DIR)" -name \*.yaml ); do \
		AWS_PROFILE=$(AWS_PROFILE) AWS_DEFAULT_REGION=$(AWS_REGION) AWS_PAGER="" \
		aws cloudformation validate-template --template-body file://$$f; \
	done

aws-cloudformation-templates/core-components:
	@echo "Initializing Submodule with AWS Cloudformation templates"
	git submodule update --init ./aws-cloudformation-templates

publish: aws-cloudformation-templates/core-components
	AWS_PROFILE=$(AWS_PROFILE) AWS_DEFAULT_REGION=$(AWS_REGION) \
	aws s3 sync $(CF_DIR) s3://$(bucket_name)/cloudformation/$(artifact_hash)

.PHONY: test publish