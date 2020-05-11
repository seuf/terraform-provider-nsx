module github.com/sky-uk/terraform-provider-nsx

go 1.12

require (
	github.com/gregsteel/gonsx v0.0.0-20181119232610-c7cb4960ab68
	// github.com/orange-cloudfoundry/gonsx v0.4.1
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/hcl v0.0.0-20170509225359-392dba7d905e // indirect
	github.com/hashicorp/terraform v0.12.7
	github.com/sky-uk/gonsx v0.0.0-20180122153724-c3caef9aee9b
)

//replace github.com/sky-uk/gonsx => github.com/sgdigital-devops/gonsx v0.3.17-4
//replace github.com/sky-uk/gonsx => github.com/orange-cloudfoundry/gonsx v0.4.1
replace github.com/sky-uk/gonsx => github.com/seuf/gonsx v0.3.18-2

