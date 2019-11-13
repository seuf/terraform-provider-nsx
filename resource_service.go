package main

import (
	"encoding/xml"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sky-uk/gonsx"
	"github.com/sky-uk/gonsx/api/service"
	"log"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreate,
		Read:   resourceServiceRead,
		Delete: resourceServiceDelete,
		Update: resourceServiceUpdate,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"scopeid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Required: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ports": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func getService(nsxclient *gonsx.NSXClient, applicationID string) (*service.ApplicationService, error) {
	api := service.NewGet(applicationID)
	err := nsxclient.Do(api)

	// API Error
	if err != nil {
		return nil, fmt.Errorf("Could not fetch ApplicationService: %s: %s", applicationID, err)
	}

	// Found
	if api.StatusCode() == 200 {
		service := api.GetResponse()
		return service, nil
	}

	// Does not exist
	if api.StatusCode() == 404 {
		return nil, nil
	}

	// Unknown Status Codes => Error
	return nil, fmt.Errorf("Could not fetch ApplicationService: %s: Status code: %d, Response: %s", applicationID, api.StatusCode(), api.ResponseObject())
}

func printService(rule *service.ApplicationService) {
	rule_xml, err := xml.MarshalIndent(rule, "", "  ")
	if err != nil {
		log.Printf("Error: %v", err)
	}
	log.Printf(string(rule_xml))
}

func resourceServiceCreate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var name, scopeid, description, protocol, ports string

	// Gather the attributes for the resource.
	name = d.Get("name").(string)
	scopeid = d.Get("scopeid").(string)
	description = d.Get("description").(string)
	protocol = d.Get("protocol").(string)
	ports = d.Get("ports").(string)

	// Create the API, use it and check for errors.
	createAPI := service.NewCreate(scopeid, name, description, protocol, ports)
	err := nsxclient.Do(createAPI)

	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	if createAPI.StatusCode() != 201 {
		return fmt.Errorf("%s", createAPI.ResponseObject())
	}

	id := createAPI.GetResponse()
	d.SetId(id)

	return resourceServiceRead(d, meta)
}

func resourceServiceRead(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var id = d.Id()

	log.Printf(fmt.Sprintf("[DEBUG] ServiceID %s", id))

	service, err := getService(nsxclient, id)

	if err != nil {
		return err
	}

	if service == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", service.Name)
	d.Set("description", service.Description)

	if len(service.Element) != 0 {
		d.Set("protocol", service.Element[0].ApplicationProtocol)
		d.Set("ports", service.Element[0].Value)
	} else {
		d.Set("protocol", nil)
		d.Set("ports", nil)
	}

	return nil
}

func resourceServiceDelete(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var id = d.Id()

	deleteAPI := service.NewDelete(id)
	err := nsxclient.Do(deleteAPI)

	log.Printf(fmt.Sprintf("Delete: %s: Status code: %d, Response: %s", id, deleteAPI.StatusCode(), deleteAPI.ResponseObject()))

	if err != nil {
		return fmt.Errorf("Error: %v", err)
	} else if deleteAPI.StatusCode() != 200 {
		return fmt.Errorf("Failed to delete %s", id)
	}

	d.SetId("")
	return nil
}

func resourceServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var id = d.Id()
	hasChanges := false

	log.Printf(fmt.Sprintf("[DEBUG] ServiceID %s", id))

	serviceObject, err := getService(nsxclient, id)

	if err != nil {
		return err
	}

	if serviceObject == nil {
		d.SetId("")
		return nil
	}

	if d.HasChange("name") {
		hasChanges = true
		oldName, newName := d.GetChange("name")
		serviceObject.Name = newName.(string)
		log.Printf(fmt.Sprintf("[DEBUG] Changing name of service from %s to %s", oldName.(string), newName.(string)))
	}

	if d.HasChange("description") {
		hasChanges = true
		_, newDesc := d.GetChange("description")
		serviceObject.Description = newDesc.(string)
	}

	if d.HasChange("protocol") || d.HasChange("ports") {
		hasChanges = true
		oldProtocol, newProtocol := d.GetChange("protocol")
		oldPorts, newPorts := d.GetChange("ports")
		newElement := service.Element{ApplicationProtocol: newProtocol.(string), Value: newPorts.(string)}
		serviceObject.Element = []service.Element{newElement}
		log.Printf(fmt.Sprintf("[DEBUG] Changing protocol and/or ports of service from %s:%s to %s:%s",
			oldProtocol.(string), oldPorts.(string), newProtocol.(string), newPorts.(string)))
	}

	if hasChanges {
		serviceObject.Revision = serviceObject.Revision + 1
		updateAPI := service.NewUpdate(id, serviceObject)
		log.Printf(updateAPI.Endpoint())
		err = nsxclient.Do(updateAPI)

		if err != nil {
			log.Printf(fmt.Sprintf("[DEBUG] Error updating service resource: %s", err))
			return err
		} else if updateAPI.StatusCode() != 200 {
			return fmt.Errorf("Failed to update ServiceApplication: %d, %s", updateAPI.StatusCode(), updateAPI.GetResponse())
		}
	}
	return resourceServiceRead(d, meta)
}
