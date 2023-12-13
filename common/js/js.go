package js

import (
	"encoding/json"

	"github.com/robertkrimen/otto"
)

func UnmarshalObject(jsCode any, objName string, i any) error {
	vm := otto.New()
	v, _ := vm.Run(jsCode)

	var err error
	if objName != "" {
		v, err = vm.Get(objName)
		if err != nil {
			return err
		}
	}
	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}
