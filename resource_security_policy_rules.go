package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sky-uk/gonsx"
	"github.com/sky-uk/gonsx/api/securitypolicy"
	"log"
)

func resourceSecurityPolicyRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecurityPolicyRuleCreate,
		Read:   resourceSecurityPolicyRuleRead,
		Delete: resourceSecurityPolicyRuleDelete,

		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"securitypolicyname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"direction": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"securitygroupids": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"serviceids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSecurityPolicyRuleCreate(d *schema.ResourceData, m interface{}) error {
	nsxclient := m.(*gonsx.NSXClient)
	var name, securitypolicyname, action, direction string
	var securitygroupids, serviceids []string

	// Gather the attributes for the resource.

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}

	if v, ok := d.GetOk("securitypolicyname"); ok {
		securitypolicyname = v.(string)
	} else {
		return fmt.Errorf("securitypolicyname argument is required")
	}

	if v, ok := d.GetOk("action"); ok {
		action = v.(string)
	} else {
		return fmt.Errorf("action argument is required")
	}

	if v, ok := d.GetOk("direction"); ok {
		direction = v.(string)
	} else {
		return fmt.Errorf("direction argument is required")
	}

	if v, ok := d.GetOk("securitygroupids"); ok {
		list := v.([]interface{})

		securitygroupids = make([]string, len(list))
		for i, value := range list {
			groupID, ok := value.(string)
			if !ok {
				return fmt.Errorf("empty element found in securitygroups")
			}
			securitygroupids[i] = groupID
		}
	} else {
		if action == "outbound" {
			return fmt.Errorf("securitygroupids argument is required")
		}
	}

	if v, ok := d.GetOk("serviceids"); ok {
		list := v.([]interface{})

		serviceids = make([]string, len(list))
		for i, value := range list {
			serviceID, ok := value.(string)
			if !ok {
				return fmt.Errorf("empty element found in services")
			}
			serviceids[i] = serviceID
		}
	} else {
		return fmt.Errorf("serviceids argument is required")
	}

	log.Print("Getting policy object to modify")
	policyToModify, err := getSingleSecurityPolicy(securitypolicyname, nsxclient)
	log.Printf("[DEBUG] - policyTOModify :%s", policyToModify)

	if err != nil {
		return err
	}

	existingAction := policyToModify.GetFirewallRuleByName(name)
	if existingAction.Name != "" {
		return fmt.Errorf("Firewall rule with same name already exists in this security policy")
	}

	if direction == "inbound" {
		log.Printf("[DEBUG] policyToModify.AddInboundFirewallAction(%s, %s, %s, %s)", name, action, direction, serviceids)
		modifyErr := policyToModify.AddInboundFirewallAction(name, action, direction, securitygroupids, serviceids)
		if err != nil {
			return fmt.Errorf("Error in adding the rule to policy object: %s", modifyErr)
		}
	} else {
		log.Printf(fmt.Sprintf("[DEBUG] policyToModify.AddOutboundFirewallAction(%s, %s, %s, %s, %s)", name, action, direction, securitygroupids, serviceids))
		modifyErr := policyToModify.AddOutboundFirewallAction(name, action, direction, securitygroupids, serviceids)
		if err != nil {
			return fmt.Errorf("Error in adding the rule to policy object: %s", modifyErr)
		}
	}

	log.Printf("[DEBUG] - policyTOModify :%s", policyToModify)
	policyToModify.Revision += policyToModify.Revision
	updateAPI := securitypolicy.NewUpdate(policyToModify.ObjectID, policyToModify)

	err = nsxclient.Do(updateAPI)

	if err != nil {
		return fmt.Errorf("Error creating security group: %v", err)
	}

	if updateAPI.StatusCode() != 200 {
		return fmt.Errorf("%s", updateAPI.ResponseObject())
	}

	d.SetId(name)
	return resourceSecurityPolicyRuleRead(d, m)
}

func resourceSecurityPolicyRuleRead(d *schema.ResourceData, m interface{}) error {
	nsxclient := m.(*gonsx.NSXClient)
	var name string
	var securitypolicyname string

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}
	if v, ok := d.GetOk("securitypolicyname"); ok {
		securitypolicyname = v.(string)
	} else {
		return fmt.Errorf("securitypolicyname argument is required")
	}

	policyToRead, err := getSingleSecurityPolicy(securitypolicyname, nsxclient)
	log.Printf("[DEBUG] - policyToRead :%s", policyToRead)

	if err != nil {
		return err
	}

	existingAction := policyToRead.GetFirewallRuleByName(name)
	id := existingAction.VsmUUID
	log.Printf("[DEBUG] VsmUUID := %s", id)

	// If the resource has been removed manually, notify Terraform of this fact.
	if id == "" {
		d.SetId("")
	}
	return nil
}

func resourceSecurityPolicyRuleDelete(d *schema.ResourceData, m interface{}) error {
	nsxclient := m.(*gonsx.NSXClient)
	var name string
	var securityPolicyName string

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}

	if v, ok := d.GetOk("securitypolicyname"); ok {
		securityPolicyName = v.(string)
	} else {
		return fmt.Errorf("securitypolicyname argument is required")
	}

	log.Print("Getting policy object to modify")
	policyToModify, err := getSingleSecurityPolicy(securityPolicyName, nsxclient)
	log.Printf("[DEBUG] - policyTOModify :%s", policyToModify)

	if err != nil {
		return err
	}

	log.Printf(fmt.Sprintf("[DEBUG] policyToModify.Remove(%s)", name))
	// FIXME:  RemoveFirewallActionByName probably return a error for consistency
	policyToModify.RemoveFirewallActionByName(name)
	log.Printf("[DEBUG] - policyTOModify :%s", policyToModify)
	updateAPI := securitypolicy.NewUpdate(policyToModify.ObjectID, policyToModify)

	err = nsxclient.Do(updateAPI)

	if err != nil {
		return fmt.Errorf("Error creating security group: %v", err)
	}

	if updateAPI.StatusCode() != 200 {
		return fmt.Errorf("%s", updateAPI.ResponseObject())
	}

	err = waitForRuleDeleted(securityPolicyName, name, 3, nsxclient)

	if err != nil {
		return err
	}

	d.SetId("")
	log.Printf("[DEBUG] firewall rule with name %s from securitypolicy %s deleted.", name, securityPolicyName)
	return nil
}

// Waits for the Rule to be deleted by querying the API and checking it is still there.
//
// Reason:
// This might seem unnecessary but unfortunately if this is not done, resources that are used by this rule (i.e. service)
// will fail to delete for a short amount of time (~1 second) after the deletion of the rule.
// By reading back the security policy and confirming the rule has been removed this does not happen anymore and
// is preferable to a sleep(1 second)
func waitForRuleDeleted(securityPolicyName string, name string, iterations int, nsxclient *gonsx.NSXClient) error {

	if iterations == 0 {
		return nil
	}

	for i := 1; i < iterations+1; i++ {
		log.Printf("[DEBUG] Check if SecurityPolicy %s/%s is deleted: Iteration %d", name, securityPolicyName, i)

		policyToRead, err := getSingleSecurityPolicy(securityPolicyName, nsxclient)
		if err != nil {
			return err
		}

		existingAction := policyToRead.GetFirewallRuleByName(name)
		id := existingAction.VsmUUID

		if id == "" {
			log.Printf(fmt.Sprintf("[DEBUG] Confirmed SecurityPolicy %s/%s deleted in %d iteration", name, securityPolicyName, i))
			return nil
		}
	}

	return fmt.Errorf("firewall rule with name %s from securitypolicy %s not deleted.", name, securityPolicyName)

}
