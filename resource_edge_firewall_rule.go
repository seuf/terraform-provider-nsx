package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/sky-uk/gonsx"
	"github.com/sky-uk/gonsx/api/edgefirewall"
	"log"
	"strings"
)

func resourceEdgeFirewallRule() *schema.Resource {

	return &schema.Resource{

		Create: resourceEdgeFirewallRuleCreate,
		Read:   resourceEdgeFirewallRuleRead,
		Update: resourceEdgeFirewallRuleUpdate,
		Delete: resourceEdgeFirewallRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceEdgeFirewallRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"edgeid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rule_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "user",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"logging_enabled": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"action": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "accept",
				ValidateFunc: validation.StringInSlice([]string{
					"accept",
					"deny",
					"reject",
				}, true),
			},
			"source": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"ip_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"grouping_object_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"destination": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"ip_address": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"grouping_object_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"application": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceEdgeFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	edgeId := d.Get("edgeid").(string)
	name := d.Get("name").(string)
	r, err := getEdgeFirewallRuleByName(edgeId, name, meta)
	if err != nil {
		return err
	}
	if r.Name != "" {
		return fmt.Errorf("Rule %s already exists on edge %s", name, edgeId)
	}

	rules, err := getRuleFromSchema(d)

	fRuleCreate := edgefirewall.NewCreateRule(edgeId, rules)
	err = nsxclient.Do(fRuleCreate)
	if err != nil {
		return err
	}

	err = checkerr(fRuleCreate)
	if err != nil {
		return err
	}

	// Rule is created Fetch it's Id
	r, err = getEdgeFirewallRuleByName(edgeId, name, meta)
	if err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("[DEBUG] RULE CREATED %+v", r))
	d.SetId(fmt.Sprintf("%s_%s", edgeId, name))

	return nil
}

func resourceEdgeFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	edgeId := d.Get("edgeid").(string)
	name := d.Get("name").(string)
	rule, err := getEdgeFirewallRuleByName(edgeId, name, meta)
	if err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("[DEBUG] resourceEdgeFirewallRuleRead RULE READ %+v", rule))

	d.Set("name", rule.Name)
	d.Set("rule_type", rule.RuleType)
	d.Set("enabled", rule.Enabled)
	d.Set("description", rule.Description)
	d.Set("action", rule.Action)
	d.Set("source", map[string]interface{}{
		"exclude":            rule.Source.Exclude,
		"ip_address":         rule.Source.IpAddress,
		"grouping_object_id": strings.Join(rule.Source.GroupingObjectId, ","),
	})
	d.Set("destination", map[string]interface{}{
		"exclude":            rule.Destination.Exclude,
		"ip_address":         rule.Destination.IpAddress,
		"grouping_object_id": strings.Join(rule.Destination.GroupingObjectId, ","),
	})
	d.Set("application", map[string]interface{}{
		"application_id": strings.Join(rule.Application.ApplicationId, ","),
	})

	return nil
}

func resourceEdgeFirewallRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	id := strings.Split(d.Id(), "_")
	d.Set("edgeid", id[0])
	d.Set("name", id[1])

	err := resourceEdgeFirewallRuleRead(d, meta)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceEdgeFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	edgeId := d.Get("edgeid").(string)
	name := d.Get("name").(string)
	rule, err := getEdgeFirewallRuleByName(edgeId, name, meta)
	if err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("[DEBUG] resourceEdgeFirewallRuleUpdate RULE READ %+v", rule))

	rules, err := getRuleFromSchema(d)
	if err != nil {
		return err
	}

	fRuleUpdate := edgefirewall.NewUpdateRule(edgeId, rule.RuleId, rules)
	err = nsxclient.Do(fRuleUpdate)
	if err != nil {
		return err
	}

	return nil
}

func resourceEdgeFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	edgeId := d.Get("edgeid").(string)
	name := d.Get("name").(string)
	rule, err := getEdgeFirewallRuleByName(edgeId, name, meta)
	if err != nil {
		return err
	}
	log.Printf(fmt.Sprintf("[DEBUG] RULE READ %+v", rule))

	fRuleDelete := edgefirewall.NewDeleteRule(edgeId, rule.RuleId)
	err = nsxclient.Do(fRuleDelete)
	if err != nil {
		return err
	}
	return nil
}

func getRuleFromSchema(d *schema.ResourceData) (*edgefirewall.FirewallRules, error) {

	rule := edgefirewall.FirewallRule{
		Name:           d.Get("name").(string),
		RuleType:       d.Get("rule_type").(string),
		Enabled:        d.Get("enabled").(bool),
		LoggingEnabled: d.Get("logging_enabled").(bool),
		Description:    d.Get("description").(string),
		Action:         d.Get("action").(string),
	}

	var src map[string]interface{}
	var dst map[string]interface{}
	var app map[string]interface{}
	var src_grouping_object_id []string
	var dst_grouping_object_id []string
	var application_id []string
	if v, ok := d.GetOk("source"); ok {
		for _, vv := range v.([]interface{}) {
			src = vv.(map[string]interface{})
			src_grouping_object_id = strings.Split(src["grouping_object_id"].(string), ",")
			rule.Source = edgefirewall.Source{
				Exclude:          false,
				IpAddress:        src["ip_address"].(string),
				GroupingObjectId: src_grouping_object_id,
			}
		}
	}
	if v, ok := d.GetOk("destination"); ok {
		for _, vv := range v.([]interface{}) {
			dst = vv.(map[string]interface{})
			dst_grouping_object_id = strings.Split(dst["grouping_object_id"].(string), ",")
			rule.Destination = edgefirewall.Destination{
				Exclude:          false,
				IpAddress:        dst["ip_address"].(string),
				GroupingObjectId: dst_grouping_object_id,
			}
		}
	}
	if v, ok := d.GetOk("application"); ok {
		for _, vv := range v.([]interface{}) {
			app = vv.(map[string]interface{})
			application_id = strings.Split(app["application_id"].(string), ",")
			rule.Application = edgefirewall.Application{
				ApplicationId: application_id,
			}
		}
	}

	log.Printf(fmt.Sprintf("[DEBUG] ============ Rule : %+v", rule))
	rules := edgefirewall.FirewallRules{
		FirewallRule: []edgefirewall.FirewallRule{rule},
	}

	return &rules, nil
}

func getEdgeFirewallRuleByName(edgeId string, name string, meta interface{}) (edgefirewall.FirewallRule, error) {
	nsxclient := meta.(*gonsx.NSXClient)

	fConfig := edgefirewall.NewGetEdgeFirewallConfig(edgeId)
	err := nsxclient.Do(fConfig)
	if err != nil {
		return edgefirewall.FirewallRule{}, err
	}
	edgeFirewallConfig := fConfig.GetResponse()
	for _, r := range edgeFirewallConfig.FirewallRules.FirewallRule {
		if r.Name == name {
			return r, nil
		}
	}
	return edgefirewall.FirewallRule{}, nil
}
