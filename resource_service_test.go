package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sky-uk/gonsx"
	"github.com/sky-uk/gonsx/api/service"
	"strings"
	"testing"
)

func TestAccService(t *testing.T) {
	scopeID := loadServiceScopeId(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: func(s *terraform.State) error { return testAccServiceWithPrefixDontExist(scopeID, "tf_testing") },
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`   resource "nsx_service" "http" {
										    name = "tf_testing_service_80"
										    scopeid = "%s"
										    description = "testing"
										    protocol = "TCP"
										    ports = "80"
										}
										resource "nsx_service" "http2" {
										    name = "tf_testing_service_8080"
										    scopeid = "%s"
										    description = "testing second"
										    protocol = "TCP"
										    ports = "8080"
										}`, scopeID, scopeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nsx_service.http", "ports", "80"),
					resource.TestCheckResourceAttrSet("nsx_service.http", "scopeid"),
					resource.TestCheckResourceAttr("nsx_service.http2", "ports", "8080"),
					resource.TestCheckResourceAttrSet("nsx_service.http2", "scopeid"),
					testAccServiceWithNameExists(scopeID, "tf_testing_service_80"),
					testAccServiceWithNameExists(scopeID, "tf_testing_service_8080"),
				),
			},
			{
				Config: fmt.Sprintf(`   resource "nsx_service" "http" {
										    name = "tf_testing_service_81"
										    scopeid = "%s"
										    description = "testing"
										    protocol = "TCP"
										    ports = "81"
										}`, scopeID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nsx_service.http", "ports", "81"),
					resource.TestCheckResourceAttrSet("nsx_service.http", "scopeid"),
					testAccServiceWithNameExists(scopeID, "tf_testing_service_81"),
				),
			},
		},
	})
}

func testAccServiceWithPrefixDontExist(scopeid string, prefix string) error {
	nsxClient := testAccProvider.Meta().(*gonsx.NSXClient)

	api := service.NewGetAll(scopeid)
	err := nsxClient.Do(api)
	if err != nil {
		return err
	}

	serviceList := api.GetResponse()

	for _, service := range serviceList.Applications {
		if strings.HasPrefix(service.Name, prefix) {
			return fmt.Errorf("There are still Services leftover: %s", service.Name)
		}
	}

	return nil
}

func testAccServiceWithNameExists(scopeid string, name string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(*gonsx.NSXClient)

		api := service.NewGetAll(scopeid)
		err := nsxClient.Do(api)
		if err != nil {
			return err
		}

		serviceList := api.GetResponse()

		for _, service := range serviceList.Applications {
			if service.Name == name {
				return nil
			}
		}

		return fmt.Errorf("Service with name %s wasn't found", name)
	}
}
