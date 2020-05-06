module github.com/sky-uk/terraform-provider-nsx

go 1.12

require (
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/hcl v0.0.0-20170509225359-392dba7d905e // indirect
	github.com/hashicorp/terraform v0.12.7
	github.com/sky-uk/gonsx v0.0.0-20180122153724-c3caef9aee9b
)

replace github.com/sky-uk/gonsx => github.com/sgdigital-devops/gonsx v0.3.17-4
