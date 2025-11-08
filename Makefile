build:
	go build -o ~/.terraform.d/plugins/registry.terraform.io/sidler2/kineticpanel/0.1.2/$(shell go env GOOS)_$(shell go env GOARCH)/terraform-provider-kineticpanel_v0.1.2
build_release:
	GPG_FINGERPRINT=${GPG} GITHUB_TOKEN=${GITHUB_TOKEN} goreleaser release --clean
install: build
	@echo "Provider installed locally"