package js

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

// ErrTimeout is returned by RunVM if the scripts run longer than its timeout.
var ErrTimeout = errors.New("js: script execution timed out")

// unmarshalTimeout limits the execution time of the UnmarshalObject scripts.
var unmarshalTimeout = 3 * time.Second

// RunVM calls fn with a new VM, interrupting the scripts it runs once the
// timeout expires. Since scripts usually come from untrusted pages, any panic
// in fn, e.g., of the VM, is recovered and returned as an error, which is
// ErrTimeout for the timeout.
func RunVM(timeout time.Duration, fn func(vm *otto.Otto) error) (err error) {
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	timer := time.AfterFunc(timeout, func() {
		vm.Interrupt <- func() { panic(ErrTimeout) }
	})
	defer timer.Stop()

	defer func() {
		if r := recover(); r != nil {
			if r == ErrTimeout {
				err = ErrTimeout
			} else {
				err = fmt.Errorf("js: script panicked: %v", r)
			}
		}
	}()

	return fn(vm)
}

func UnmarshalObject[T ~string | ~[]byte](jsCode T, objName string, i any) error {
	if len(jsCode) == 0 {
		return errors.New("empty JS code snippet")
	}

	return RunVM(unmarshalTimeout, func(vm *otto.Otto) error {
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
	})
}
