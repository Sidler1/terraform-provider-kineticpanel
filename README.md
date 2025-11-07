# 1. Build & install

make install

# 2. Create examples/main.tf

provider "kineticpanel" {
host = "https://kineticpanel.net"
api_key = "your-app-key-here"
use_application = true
}

# 3. terraform init && terraform plan