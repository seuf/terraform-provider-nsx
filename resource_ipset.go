package main

import (
	"fmt"
	"github.com/gregsteel/gonsx/api/ipset"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sky-uk/gonsx"
	"log"
	"strings"
)

func getSingleIPSet(scopeid, name string, nsxclient *gonsx.NSXClient) (*ipset.IPSet, error) {
	getAllAPI := ipset.NewGetAll(scopeid)
	err := nsxclient.Do(getAllAPI)

	if err != nil {
		return nil, err
	}

	if getAllAPI.StatusCode() != 200 {
		return nil, fmt.Errorf("Status code: %d, Response: %s", getAllAPI.StatusCode(), getAllAPI.ResponseObject())
	}

	ipSet := getAllAPI.GetResponse().FilterByName(name)

	if ipSet.ObjectID == "" {
		return nil, fmt.Errorf("Not found %s", name)
	}

	return ipSet, nil
}

func resourceIPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPSetCreate,
		Read:   resourceIPSetRead,
		Delete: resourceIPSetDelete,
		Update: resourceIPSetUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceIPSetImport,
		},

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

			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceIPSetCreate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var name, scopeid, description, value string

	// Gather the attributes for the resource.
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}

	if v, ok := d.GetOk("scopeid"); ok {
		scopeid = v.(string)
	} else {
		return fmt.Errorf("scopeid argument is required")
	}

	if v, ok := d.GetOk("description"); ok {
		description = v.(string)
	} else {
		return fmt.Errorf("description argument is required")
	}

	if v, ok := d.GetOk("value"); ok {
		value = v.(string)
	} else {
		return fmt.Errorf("value argument is required")
	}

	// Create the API, use it and check for errors.
	log.Printf(fmt.Sprintf("[DEBUG] ipset.NewCreate(%s, %s, %s, %s, %s)", scopeid, name, description, value))

	ipSet := ipset.IPSet{Value: value, Name: name, Description: description}
	createAPI := ipset.NewCreate(scopeid, &ipSet)
	err := nsxclient.Do(createAPI)

	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	if createAPI.StatusCode() != 201 {
		return fmt.Errorf("%s", createAPI.ResponseObject())
	}

	// If we get here, everything is OK.  Set the ID for the Terraform state
	// and return the response from the READ method.
	d.SetId(createAPI.GetResponse())
	return resourceIPSetRead(d, meta)
}

func resourceIPSetRead(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var scopeid, name string

	// Gather the attributes for the resource.
	if v, ok := d.GetOk("scopeid"); ok {
		scopeid = v.(string)
	} else {
		return fmt.Errorf("scopeid argument is required")
	}

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}

	// Gather all the resources that are associated with the specified
	// scopeid.
	log.Printf(fmt.Sprintf("[DEBUG] ipset.NewGetAll(%s)", scopeid))
	api := ipset.NewGetAll(scopeid)
	err := nsxclient.Do(api)

	if err != nil {
		return err
	}

	// See if we can find our specifically named resource within the list of
	// resources associated with the scopeid.
	log.Printf(fmt.Sprintf("[DEBUG] api.GetResponse().FilterByName(\"%s\").ObjectID", name))
	ipsetObject, err := getSingleIPSet(scopeid, name, nsxclient)
	id := ipsetObject.ObjectID
	d.SetId(id)
	log.Printf(fmt.Sprintf("[DEBUG] id := %s", id))

	// If the resource has been removed manually, notify Terraform of this fact.
	if id == "" {
		d.SetId("")
	}

	return nil
}

func resourceIPSetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	ipset_id := strings.Split(d.Id(), "_")
	d.Set("scopeid", ipset_id[0])
	d.Set("name", ipset_id[1])
	err := resourceIPSetRead(d, meta)
	if err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceIPSetDelete(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var name, scopeid string

	// Gather the attributes for the resource.
	if v, ok := d.GetOk("scopeid"); ok {
		scopeid = v.(string)
	} else {
		return fmt.Errorf("scopeid argument is required")
	}

	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		return fmt.Errorf("name argument is required")
	}

	// Gather all the resources that are associated with the specified
	// scopeid.
	log.Printf(fmt.Sprintf("[DEBUG] ipset.NewGetAll(%s)", scopeid))
	api := ipset.NewGetAll(scopeid)
	err := nsxclient.Do(api)

	if err != nil {
		return err
	}

	// See if we can find our specifically named resource within the list of
	// resources associated with the scopeid.
	log.Printf(fmt.Sprintf("[DEBUG] api.GetResponse().FilterByName(\"%s\").ObjectID", name))
	ipsetObject, err := getSingleIPSet(scopeid, name, nsxclient)
	id := ipsetObject.ObjectID
	log.Printf(fmt.Sprintf("[DEBUG] id := %s", id))

	// If the resource has been removed manually, notify Terraform of this fact.
	if id == "" {
		d.SetId("")
		return nil
	}

	// If we got here, the resource exists, so we attempt to delete it.
	deleteAPI := ipset.NewDelete(id)
	err = nsxclient.Do(deleteAPI)

	if err != nil {
		return err
	}

	// If we got here, the resource had existed, we deleted it and there was
	// no error.  Notify Terraform of this fact and return successful
	// completion.
	d.SetId("")
	log.Printf(fmt.Sprintf("[DEBUG] id %s deleted.", id))

	return nil
}

func resourceIPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	nsxclient := meta.(*gonsx.NSXClient)
	var scopeid string
	hasChanges := false

	// Gather the attributes for the resource.
	if v, ok := d.GetOk("scopeid"); ok {
		scopeid = v.(string)
	} else {
		return fmt.Errorf("scopeid argument is required")
	}

	// Do a GetAll on ipset resources that are associated with the specified scopeid.
	log.Printf(fmt.Sprintf("[DEBUG] ipset.NewGetAll(%s)", scopeid))
	api := ipset.NewGetAll(scopeid)
	err := nsxclient.Do(api)
	if err != nil {
		log.Printf(fmt.Sprintf("[DEBUG] Error during getting all ipset resources: %s", err))
		return err
	}

	// Find the resource with current name within all the scopeid resources.
	oldName, newName := d.GetChange("name")
	log.Printf(fmt.Sprintf("[DEBUG] api.GetResponse().FilterByName(\"%s\").ObjectID", oldName.(string)))
	ipsetObject, err := getSingleIPSet(scopeid, oldName.(string), nsxclient)
	id := ipsetObject.ObjectID
	log.Printf(fmt.Sprintf("[DEBUG] id := %s", id))

	// If the resource has been removed manually, notify Terraform of this fact.
	if id == "" {
		d.SetId("")
		log.Printf(fmt.Sprintf("[DEBUG] Could not find the ipset resource with %s name, state will be cleared", oldName))
		return nil
	}

	if d.HasChange("name") {
		hasChanges = true
		ipsetObject.Name = newName.(string)
		log.Printf(fmt.Sprintf("[DEBUG] Changing name of ipset from %s to %s", oldName.(string), newName.(string)))
	}

	if d.HasChange("description") {
		hasChanges = true
		oldDesc, newDesc := d.GetChange("description")
		ipsetObject.Description = newDesc.(string)
		log.Printf(fmt.Sprintf("[DEBUG] Changing description of ipset from %s to %s", oldDesc.(string), newDesc.(string)))
	}

	if d.HasChange("value") {
		hasChanges = true
		oldValue, newValue := d.GetChange("value")
		ipsetObject.Value = newValue.(string)
		log.Printf(fmt.Sprintf("[DEBUG] Changing value of ipset from %s to %s", oldValue.(string), newValue.(string)))
	}

	if hasChanges {
		updateAPI := ipset.NewUpdate(id, ipsetObject)
		err = nsxclient.Do(updateAPI)

		if err != nil {
			log.Printf(fmt.Sprintf("[DEBUG] Error updating ipset resource: %s", err))
			return err
		}
	}
	return resourceIPSetRead(d, meta)
}
