package js

import (
	"encoding/json"
	"fmt"

	"github.com/robertkrimen/otto"
)

func UnmarshalObject(jsCode any, objName string, i any) error {
	vm := otto.New()
	_, _ = vm.Run(jsCode)

	v, err := vm.Get(objName)
	if err != nil {
		return err
	}
	if !v.IsObject() {
		err = fmt.Errorf("object not found for `%s`", objName)
	}
	b, err := v.Object().MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i)
}
