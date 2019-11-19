package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestSecurityGroupIntegrated(t *testing.T) {
	scopeID := loadServiceScopeId(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccServiceWithPrefixDontExist(scopeID, "tf_testing"),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "nsx_security_group" "sg" {
					  count   = "1"
					  name    = "tf_testing_sg1"
					  scopeid = "%[1]s"

					  dynamic_membership = [
					    {
					      set_operator   = "OR"
					      rules_operator = "AND"

					      rules = [
					        {
					          key      = "VM.SECURITY_TAG"
					          value    = "something_that_does_not_exist"
					          criteria = "="
					        },
					      ]
					    },
					  ]
					}

					resource "nsx_security_policy" "policy" {
					  count          = "1"
					  name           = "tf_testing_sp1"
					  description    = "TF Testing Security Policy"
					  precedence     = "1337"
					  securitygroups = ["${nsx_security_group.sg.id}"]
					}

					resource "nsx_security_policy_rule" "rule" {
					  name               = "tf_testing_security_policy_rule"
					  securitypolicyname = "${nsx_security_policy.policy.name}"
					  action             = "allow"
					  direction          = "inbound"
					  // securitygroupids   = ["${var.common_security_group_id}"]
					  serviceids         = ["${nsx_service.service0.id}"]
					}

					resource "nsx_service" "service0" {
					  name        = "tf_testing_service0"
					  scopeid     = "%[1]s"
					  description = "Informix"
					  protocol    = "TCP"
					  ports = "1924,1925"
					}`, scopeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nsx_service.service0", "ports", "1924,1925"),
					resource.TestCheckResourceAttrSet("nsx_service.service0", "scopeid"),
					testAccServiceWithNameExists(scopeID, "tf_testing_service0"),
				),
			},
			// Delete Service and rule
			//
			{
				Config: fmt.Sprintf(`
					resource "nsx_security_group" "sg" {
					  count   = "1"
					  name    = "tf_testing_sg1"
					  scopeid = "%[1]s"

					  dynamic_membership = [
					    {
					      set_operator   = "OR"
					      rules_operator = "AND"

					      rules = [
					        {
					          key      = "VM.SECURITY_TAG"
					          value    = "something_that_does_not_exist"
					          criteria = "="
					        },
					      ]
					    },
					  ]
					}

					resource "nsx_security_policy" "policy" {
					  count          = "1"
					  name           = "tf_testing_sp1"
					  description    = "TF Testing Security Policy"
					  precedence     = "1337"
					  securitygroups = ["${nsx_security_group.sg.id}"]
					}`, scopeID),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceWithPrefixDontExist(scopeID, "tf_testing_service0"),
				),
			},
		},
	})
}
