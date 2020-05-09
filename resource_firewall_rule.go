package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/sky-uk/gonsx"
	"github.com/sky-uk/gonsx/api/firewall"
	"strconv"
)

func resourceFirewallRule() *schema.Resource {

	return &schema.Resource{

		Create: resourceFirewallRuleCreate,
		Read:   resourceFirewallRuleRead,
		Update: resourceFirewallRuleUpdate,
		Delete: resourceFirewallRuleDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"logged": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"action": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "allow",
				ValidateFunc: validation.StringInSlice([]string{
					"allow",
					"deny",
					"reject",
				}, true),
			},
			"direction": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "inout",
				ValidateFunc: validation.StringInSlice([]string{
					"in",
					"out",
					"inout",
				}, true),
			},
			"sectionid": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"packet_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "any",
			},
			"applied_to": {
				Type:       schema.TypeSet,
				ConfigMode: schema.SchemaConfigModeAttr,
				Optional:   true,
				Computed:   true,
				Set:        setRuleElement,
				Elem:       schemaRuleElement(),
			},
			"source": {
				Type:          schema.TypeSet,
				ConfigMode:    schema.SchemaConfigModeAttr,
				Optional:      true,
				Set:           setRuleElement,
				Elem:          schemaRuleElement(),
				ConflictsWith: []string{"source_excluded"},
			},
			"source_excluded": {
				Type:          schema.TypeSet,
				ConfigMode:    schema.SchemaConfigModeAttr,
				Optional:      true,
				Set:           setRuleElement,
				Elem:          schemaRuleElement(),
				ConflictsWith: []string{"source"},
			},
			"destination": {
				Type:          schema.TypeSet,
				ConfigMode:    schema.SchemaConfigModeAttr,
				Optional:      true,
				Set:           setRuleElement,
				Elem:          schemaRuleElement(),
				ConflictsWith: []string{"destination_excluded"},
			},
			"destination_excluded": {
				Type:          schema.TypeSet,
				ConfigMode:    schema.SchemaConfigModeAttr,
				Optional:      true,
				Set:           setRuleElement,
				Elem:          schemaRuleElement(),
				ConflictsWith: []string{"destination"},
			},
			"service": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      setRuleElement,
				Elem:     schemaRuleElement(),
			},
		},
	}
}

func setRuleElement(v interface{}) int {
	elem := v.(map[string]interface{})
	return hashcode.String(fmt.Sprintf(
		"%s-%d",
		elem["value"],
		elem["type"],
	))
}

func schemaRuleElement() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"ALL_EDGES",
					"Application",
					"ApplicationGroup",
					"Datacenter",
					"DISTRIBUTED_FIREWALL",
					"DistributedVirtualPortgroup",
					"Edge",
					"GlobalRoot",
					"IPSet",
					"Ipv4Address",
					"Ipv6Address",
					"VirtualWire",
					"MACSet",
					"Network",
					"ALL_PROFILE_BINDINGS",
					"ResourcePool",
					"SecurityGroup",
					"Vnic",
				}, false),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_valid": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func schemaRuleElementToElement(d interface{}) []firewall.Element {
	schemaToListMap := getListOfStructs(d)
	elems := make([]firewall.Element, len(schemaToListMap))

	for i, elem := range schemaToListMap {
		elems[i] = firewall.Element{
			Name:    elem["name"].(string),
			Value:   elem["value"].(string),
			Type:    firewall.ElemType(elem["type"].(string)),
			IsValid: elem["is_valid"].(bool),
		}
	}

	return elems
}

func elementToSchemaRuleElement(elems []firewall.Element) []map[string]interface{} {
	elemsMap := make([]map[string]interface{}, len(elems))

	for i, elem := range elems {
		elemsMap[i] = map[string]interface{}{
			"name":     elem.Name,
			"type":     elem.Type,
			"value":    elem.Value,
			"is_valid": elem.IsValid,
		}
	}

	return elemsMap
}

func tfRuleToFirewallRule(d *schema.ResourceData) firewall.Rule {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		id = 0
	}

	var sources *firewall.Sources

	if len(d.Get("source").(*schema.Set).List()) > 0 {
		sources = &firewall.Sources{
			Elements: schemaRuleElementToElement(d.Get("source")),
		}
	} else if len(d.Get("source_excluded").(*schema.Set).List()) > 0 {
		sources = &firewall.Sources{
			Excluded: true,
			Elements: schemaRuleElementToElement(d.Get("source_excluded")),
		}
	}

	var destinations *firewall.Destinations

	if len(d.Get("destination").(*schema.Set).List()) > 0 {
		destinations = &firewall.Destinations{
			Elements: schemaRuleElementToElement(d.Get("destination")),
		}
	} else if len(d.Get("destination_excluded").(*schema.Set).List()) > 0 {
		destinations = &firewall.Destinations{
			Excluded: true,
			Elements: schemaRuleElementToElement(d.Get("destination_excluded")),
		}
	}

	var services *firewall.Services

	if len(d.Get("service").(*schema.Set).List()) > 0 {
		services = &firewall.Services{
			Elements: schemaRuleElementToElement(d.Get("service")),
		}
	}

	return firewall.Rule{
		Name:       d.Get("name").(string),
		ID:         id,
		SectionId:  d.Get("sectionid").(int),
		Direction:  firewall.Direction(d.Get("direction").(string)),
		Action:     firewall.Action(d.Get("action").(string)),
		PacketType: d.Get("packet_type").(string),
		Disabled:   d.Get("disabled").(bool),
		Logged:     d.Get("logged").(bool),
		Notes:      d.Get("description").(string),
		AppliedToList: &firewall.AppliedToList{
			Elements: schemaRuleElementToElement(d.Get("applied_to")),
		},
		Sources:      sources,
		Destinations: destinations,
		Services:     services,
	}
}

func firewallRuleToTfRule(d *schema.ResourceData, rule *firewall.Rule) {
	d.SetId(fmt.Sprintf("%d", rule.ID))
	d.Set("name", rule.Name)
	d.Set("sectionid", rule.SectionId)
	d.Set("direction", string(rule.Direction))
	d.Set("action", string(rule.Action))
	d.Set("packet_type", rule.PacketType)
	d.Set("disabled", rule.Disabled)
	d.Set("logged", rule.Logged)
	d.Set("description", rule.Notes)
	d.Set("applied_to", elementToSchemaRuleElement(rule.AppliedToList.Elements))

	if rule.Sources != nil && rule.Sources.Elements != nil {
		if !rule.Sources.Excluded {
			d.Set("source", elementToSchemaRuleElement(rule.Sources.Elements))
		} else {
			d.Set("source_excluded", elementToSchemaRuleElement(rule.Sources.Elements))
		}
	}

	if rule.Destinations != nil && rule.Destinations.Elements != nil {
		if !rule.Destinations.Excluded {
			d.Set("destination", elementToSchemaRuleElement(rule.Destinations.Elements))
		} else {
			d.Set("destination_excluded", elementToSchemaRuleElement(rule.Destinations.Elements))
		}
	}

	if rule.Services != nil && rule.Services.Elements != nil {
		d.Set("service", elementToSchemaRuleElement(rule.Services.Elements))
	}
}

func resourceFirewallRuleCreate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	fConfig := firewall.NewGetFirewallConfig()
	err := nsxclient.Do(fConfig)
	if err != nil {
		return err
	}
	d.Set("etag", fConfig.ResponseHeaders().Get("Etag"))

	rule := tfRuleToFirewallRule(d)

	fRuleCreate := firewall.NewCreateRule(rule.SectionId, d.Get("etag").(string), &rule)
	err = nsxclient.Do(fRuleCreate)
	if err != nil {
		return err
	}

	err = checkerr(fRuleCreate)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%d", fRuleCreate.GetResponse().ID))
	return nil
}

func resourceFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	fConfig := firewall.NewGetFirewallConfig()
	err := nsxclient.Do(fConfig)
	if err != nil {
		return err
	}
	d.Set("etag", fConfig.ResponseHeaders().Get("Etag"))
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	fRuleRead := firewall.NewGetRule(d.Get("sectionid").(int), d.Get("etag").(string), id)

	err = nsxclient.Do(fRuleRead)
	if err != nil {
		return err
	}
	firewallRuleToTfRule(d, fRuleRead.GetResponse())
	return nil
}

func resourceFirewallRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	rule := tfRuleToFirewallRule(d)
	fRuleUpdate := firewall.NewUpdateRule(rule.SectionId, d.Get("etag").(string), id, rule)
	err = nsxclient.Do(fRuleUpdate)
	if err != nil {
		return err
	}
	return nil
}

func resourceFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)

	fConfig := firewall.NewGetFirewallConfig()
	err := nsxclient.Do(fConfig)
	if err != nil {
		return err
	}
	etag := fConfig.ResponseHeaders().Get("Etag")

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	fRuleDelete := firewall.NewDeleteRule(d.Get("sectionid").(int), etag, id)
	err = nsxclient.Do(fRuleDelete)
	if err != nil {
		return err
	}
	return nil
}
