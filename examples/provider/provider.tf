# Option 1: Environment variables (recommended)
#   export SEDAI_BASE_URL="https://your-sedai-instance.com"
#   export SEDAI_API_TOKEN="your-api-token"

# Option 2: .env file (load before running terraform)
#   Create a .env file:
#     SEDAI_BASE_URL=https://your-sedai-instance.com
#     SEDAI_API_TOKEN=your-api-token
#   Then load it: terraform apply

terraform {
  required_providers {
    sedai = {
      source  = "hashicorp.com/io/sedai"
      version = "~> 1.1"
    }
  }
}

provider "sedai" {}
