package panicrecovery

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/glodb/keel/settings/logger"
)

// RecoverFromPanic recovers from panics and logs them
// Use this in defer statements at the start of critical functions
func RecoverFromPanic(context string) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		logger.Log().Error("PANIC RECOVERED",
			logger.StringField("context", context),
			logger.AnyField("panic", r),
			logger.StringField("stack_trace", stack),
			logger.StringField("timestamp", time.Now().Format(time.RFC3339)))
	}
}

// RecoverFromPanicWithCallback recovers from panics and executes a callback
func RecoverFromPanicWithCallback(context string, callback func(panicValue interface{}, stackTrace string)) {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		logger.Log().Error("PANIC RECOVERED WITH CALLBACK",
			logger.StringField("context", context),
			logger.AnyField("panic", r),
			logger.StringField("stack_trace", stack))

		if callback != nil {
			callback(r, stack)
		}
	}
}

// SafeGo runs a goroutine with panic recovery
// Use this instead of 'go func()' for better error handling
func SafeGo(fn func(), context string) {
	go func() {
		defer RecoverFromPanic(fmt.Sprintf("goroutine: %s", context))
		fn()
	}()
}

// SafeGoWithReturn runs a goroutine with panic recovery and error channel
func SafeGoWithReturn(fn func() error, context string, errChan chan<- error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Log().Error("PANIC IN GOROUTINE WITH RETURN",
					logger.StringField("context", context),
					logger.AnyField("panic", r),
					logger.StringField("stack_trace", stack))

				if errChan != nil {
					errChan <- fmt.Errorf("panic in %s: %v", context, r)
				}
			}
		}()

		if err := fn(); err != nil && errChan != nil {
			errChan <- err
		}
	}()
}

// WrapWithRecovery wraps any function with panic recovery
func WrapWithRecovery(fn func(), context string) func() {
	return func() {
		defer RecoverFromPanic(context)
		fn()
	}
}

// WrapWithRecoveryAndError wraps any function that returns error with panic recovery
func WrapWithRecoveryAndError(fn func() error, context string) func() error {
	return func() error {
		defer RecoverFromPanic(context)
		return fn()
	}
}

// HTTPRecoveryMiddleware returns a recovery middleware for HTTP handlers
// This should be used in Gin middleware
func HTTPRecoveryMiddleware() func(interface{}) {
	return func(c interface{}) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Log().Error("PANIC IN HTTP HANDLER",
					logger.AnyField("panic", r),
					logger.StringField("stack_trace", stack),
					logger.StringField("timestamp", time.Now().Format(time.RFC3339)))
			}
		}()
	}
}

// InitializerWithRecovery wraps initialization code with panic recovery
// Returns true if initialization succeeded, false if panicked
func InitializerWithRecovery(fn func() error, context string) error {
	var err error
	var panicked bool
	var panicValue interface{}

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				panicValue = r
				stack := string(debug.Stack())
				logger.Log().Error("PANIC DURING INITIALIZATION",
					logger.StringField("context", context),
					logger.AnyField("panic", r),
					logger.StringField("stack_trace", stack))
			}
		}()

		err = fn()
	}()

	if panicked {
		return fmt.Errorf("panic during initialization of %s: %v", context, panicValue)
	}

	return err
}

// LogPanicAndExit logs panic and exits the program
// Use this only in main() or critical startup code
func LogPanicAndExit() {
	if r := recover(); r != nil {
		stack := string(debug.Stack())
		logger.Log().Error("FATAL PANIC - SERVICE SHUTTING DOWN",
			logger.AnyField("panic", r),
			logger.StringField("stack_trace", stack),
			logger.StringField("timestamp", time.Now().Format(time.RFC3339)))

		// Give logger time to flush
		time.Sleep(500 * time.Millisecond)
		panic(r) // Re-panic to exit with error
	}
}
