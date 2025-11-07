# 1. Build & install

make install

# 2. Create examples/main.tf

```terraform
terraform {
  required_providers {
    kineticpanel = {
      source = "sidler1/kineticpanel"
      version = "0.1.0"
    }
  }
}
```

# 3. terraform init && terraform plan