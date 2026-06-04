package keel

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // Import pprof for profiling
	"os"
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

func Boot() {
	// CRITICAL: Global panic recovery to prevent silent crashes
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())

			// Try to log if logger is initialized
			if logger.Log() != nil {
				logger.Log().Error("FATAL PANIC IN MAIN - SERVICE CRASHED",
					logger.AnyField("panic", r),
					logger.StringField("stack_trace", stack),
					logger.StringField("service", configmanager.GetInstance().ClassName),
					logger.StringField("timestamp", time.Now().Format(time.RFC3339)))
			} else {
				// Fallback to stderr if logger not available
				_, _ = os.Stderr.WriteString(fmt.Sprintf("\n\n!!! FATAL PANIC IN MAIN !!!\nService: %s\nPanic: %v\nStack:\n%s\n",
					configmanager.GetInstance().ClassName, r, stack))
			}

			// Give logger time to flush
			time.Sleep(500 * time.Millisecond)
			os.Exit(1)
		}
	}()

	logger.Log().Info("Starting service with panic recovery enabled",
		logger.StringField("service", configmanager.GetInstance().ClassName))

	configmanager.GetInstance()

	logger.LogInit(configmanager.GetInstance())

	// Initialize Meilisearch if configured
	if configmanager.GetInstance().MeilisearchConfig.Host != "" {
		meiliClient := meilisearch.GetInstance()
		if meiliClient != nil {
			search.SetSearch(meiliClient)
			logger.Log().Info("Meilisearch initialized successfully")
		}
	}

	// Initialize tracing explicitly with panic recovery
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

	// Wrap service initialization with panic recovery
	var base interface{}
	err := panicrecovery.InitializerWithRecovery(func() error {
		var initErr error
		base, initErr = servicehandler.GetInstance().InitializeService(configmanager.GetInstance().ClassName)
		return initErr
	}, "service initialization")

	if err != nil {
		logger.Log().Error("Error in initializing Service", logger.ErrorField("error", err))
		os.Exit(1)
	}

	if base != nil {
		// Run service with panic recovery
		func() {
			defer panicrecovery.RecoverFromPanic("service run")
			if runnable, ok := base.(interface{ Run() error }); ok {
				if runErr := runnable.Run(); runErr != nil {
					logger.Log().Error("Error in running Service", logger.ErrorField("error", runErr))
				}
			}
		}()
	} else {
		logger.Log().Error("Error in running Service - base is nil")
		os.Exit(1)
	}

	if err != nil {
		logger.Log().Error("Error in running Service", logger.ErrorField("error", err))
	}

	// Stop service with panic recovery
	if base != nil {
		if stoppable, ok := base.(interface{ Stop() }); ok {
			defer panicrecovery.RecoverFromPanic("service stop")
			stoppable.Stop()
		}
	}
}
