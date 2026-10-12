.PHONY: help build test infra-plan infra-apply configure verify destroy

help:
	@printf '%s\n' 'Planned targets: build test infra-plan infra-apply configure verify destroy' 'Components are not implemented yet.'

build test infra-plan infra-apply configure verify destroy:
	@printf '%s\n' "Target '$@' is not implemented yet." >&2
	@exit 1
