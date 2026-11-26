package js

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
)

// RunVM calls fn with a new VM, interrupting the scripts it runs once ctx
// is done.
func RunVM(ctx context.Context, fn func(vm *otto.Otto) error) (err error) {
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)

	stop := context.AfterFunc(ctx, func() {
		vm.Interrupt <- func() { panic(ctx.Err()) }
	})
	defer stop()

	defer func() {
		if r := recover(); r != nil {
			if r == ctx.Err() {
				err = ctx.Err()
			} else {
				err = fmt.Errorf("js: script panicked: %v", r)
			}
		}
	}()

	return fn(vm)
}

const unmarshalTimeout = 3 * time.Second

func UnmarshalObject[T ~string | ~[]byte](jsCode T, objName string, i any) error {
	if len(jsCode) == 0 {
		return errors.New("empty JS code snippet")
	}

	ctx, cancel := context.WithTimeout(context.Background(), unmarshalTimeout)
	defer cancel()

	return RunVM(ctx, func(vm *otto.Otto) error {
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
