resource "hatchet_management_token" "ci" {
  name     = "GitHub Actions CI"
  duration = "30D"
  tags     = ["ci", "github-actions"]
}

output "ci_token" {
  value     = hatchet_management_token.ci.token
  sensitive = true
}
