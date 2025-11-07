build:
	go build -o ~/.terraform.d/plugins/registry.terraform.io/yourname/kineticpanel/0.1.0/$(shell go env GOOS)_$(shell go env GOARCH)/terraform-provider-kineticpanel_v0.1.0

install: build
	@echo "Provider installed locally"