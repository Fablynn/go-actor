package async

import "runtime/debug"

var (
	catch func(string, ...interface{})
)

func Init(f func(string, ...interface{})) {
	catch = f
}

func SafeRecover(cb func(string, ...interface{}), f func()) {
	defer func() {
		if err := recover(); err != nil {
			if cb != nil {
				cb("%v stack: %v", err, string(debug.Stack()))
			} else if catch != nil {
				catch("%v stack: %v", err, string(debug.Stack()))
			}
		}
	}()
	f()
}

func SafeGo(cb func(string, ...interface{}), f func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if cb != nil {
					cb("%v stack: %v", err, string(debug.Stack()))
				} else if catch != nil {
					catch("%v stack: %v", err, string(debug.Stack()))
				}
			}
		}()
		f()
	}()
}
