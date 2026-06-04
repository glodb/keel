package keel

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // Import pprof for profiling
	"runtime/debug"
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/panicrecovery"
	"github.com/glodb/keel/settings/searchsettings/meilisearch"
	"github.com/glodb/keel/settings/searchsettings/search"
	"github.com/glodb/keel/settings/servicehandler"
	"github.com/glodb/keel/settings/topics"
	"github.com/glodb/keel/settings/tracing"
)

// Boot initialises and runs the named service. env is the deployment environment
// (e.g. "DEV", "PROD") and service is the registered service name (e.g. "SSOSERVICE").
// It returns an error on any fatal startup failure; the caller decides how to handle it.
func Boot(env, service string) (err error) {
	configmanager.Configure(env, service)

	// Recover from panics in the startup/run path and convert to a returned error.
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			if logger.Log() != nil {
				logger.Log().Error("FATAL PANIC IN BOOT - SERVICE CRASHED",
					logger.AnyField("panic", r),
					logger.StringField("stack_trace", stack),
					logger.StringField("service", configmanager.GetInstance().ClassName),
					logger.StringField("timestamp", time.Now().Format(time.RFC3339)))
			}
			time.Sleep(500 * time.Millisecond)
			err = fmt.Errorf("fatal panic: %v", r)
		}
	}()

	logger.Log().Info("Starting service with panic recovery enabled",
		logger.StringField("service", configmanager.GetInstance().ClassName))

	configmanager.GetInstance()

	logger.LogInit(configmanager.GetInstance())

	// Initialize Meilisearch if configured.
	if configmanager.GetInstance().MeilisearchConfig.Host != "" {
		meiliClient := meilisearch.GetInstance()
		if meiliClient != nil {
			search.SetSearch(meiliClient)
			logger.Log().Info("Meilisearch initialized successfully")
		}
	}

	// Initialize tracing with panic recovery.
	panicrecovery.SafeGo(func() {
		tracing.GetInstance()
	}, "tracing initialization")

	if configmanager.GetInstance().UsePprof {
		panicrecovery.SafeGo(func() {
			logger.Log().Info("Starting pprof on :6060")
			if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
				logger.Log().Error("pprof server error", logger.ErrorField("error", err))
			}
		}, "pprof server")
	}

	topics.GetInstance().Register()

	var base interface{}
	initErr := panicrecovery.InitializerWithRecovery(func() error {
		var e error
		base, e = servicehandler.GetInstance().InitializeService(configmanager.GetInstance().ClassName)
		return e
	}, "service initialization")

	if initErr != nil {
		logger.Log().Error("Error in initializing Service", logger.ErrorField("error", initErr))
		return fmt.Errorf("service initialization failed: %w", initErr)
	}

	if base == nil {
		logger.Log().Error("Error in running Service - base is nil")
		return fmt.Errorf("service initialization returned nil for %q", configmanager.GetInstance().ClassName)
	}

	// Run service with panic recovery.
	func() {
		defer panicrecovery.RecoverFromPanic("service run")
		if runnable, ok := base.(interface{ Run() error }); ok {
			if runErr := runnable.Run(); runErr != nil {
				logger.Log().Error("Error in running Service", logger.ErrorField("error", runErr))
			}
		}
	}()

	// Stop service with panic recovery.
	if stoppable, ok := base.(interface{ Stop() }); ok {
		defer panicrecovery.RecoverFromPanic("service stop")
		stoppable.Stop()
	}

	return nil
}
