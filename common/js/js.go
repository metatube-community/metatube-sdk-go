package js

import (
	"encoding/json"

	"github.com/robertkrimen/otto"
)

func UnmarshalObject(jsCode any, objName string, i any) error {
	vm := otto.New()
	_, _ = vm.Run(jsCode)

	v, err := vm.Get(objName)
	if err != nil {
		return err
	}
	b, err := v.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}
