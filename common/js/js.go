package js

import (
	"encoding/json"
	"errors"

	"github.com/robertkrimen/otto"
)

func UnmarshalObject[T ~string | ~[]byte](jsCode T, objName string, i any) error {
	if len(jsCode) == 0 {
		return errors.New("empty JS code snippet")
	}

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
