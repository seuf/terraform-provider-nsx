package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/sky-uk/gonsx/api"
)

func getListOfStructs(v interface{}) []map[string]interface{} {
	if vvSet, ok := v.(*schema.Set); ok {
		v = vvSet.List()
	}
	vvv := []map[string]interface{}{}
	for _, vv := range v.([]interface{}) {
		vvv = append(vvv, vv.(map[string]interface{}))
	}
	return vvv
}

func checkerr(api api.NSXApi) error {
	if api.StatusCode() >= 200 && api.StatusCode() <= 399 {
		return nil
	}
	return fmt.Errorf(string(api.RawResponse()))
}
