package grpcutils_test

import (
	"bytes"
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/grpcutils"
	"github.com/basemind-ai/monorepo/internal/utils/testutils"
	loggingmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := testutils.CreateNamespaceTestDBModule("grpc-utils")
	defer cleanup()
	m.Run()
}

func TestGrpcUtils(t *testing.T) {
	t.Run("CreateGRPCServer creates a server instance", func(t *testing.T) {
		server := grpcutils.CreateGRPCServer(grpcutils.Options{
			AuthHandler: func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			},
			Environment: "development",
			ServiceName: "test-service",
			ServiceRegistrars: []grpcutils.ServiceRegistrar{
				func(s grpc.ServiceRegistrar) {},
			},
		})
		assert.NotNil(t, server)
	})

	t.Run("CreateInterceptorLogger", func(t *testing.T) {
		t.Run("returns the expected values", func(t *testing.T) {
			loggerFunc, opts := grpcutils.CreateInterceptorLogger(true)
			assert.NotNil(t, loggerFunc)
			assert.NotNil(t, opts)
		})

		t.Run(
			"handler functions logs at all levels correctly",
			func(t *testing.T) {
				loggerFunc, _ := grpcutils.CreateInterceptorLogger(true)

				var buf bytes.Buffer
				log.Logger = log.Output(&buf)

				loggerFunc.Log(context.TODO(), loggingmiddleware.LevelDebug, "debug message")
				loggerFunc.Log(context.TODO(), loggingmiddleware.LevelInfo, "info message")
				loggerFunc.Log(context.TODO(), loggingmiddleware.LevelWarn, "warn message")
				loggerFunc.Log(context.TODO(), loggingmiddleware.LevelError, "error message")

				assert.Contains(t, buf.String(), "debug message")
				assert.Contains(t, buf.String(), "info message")
				assert.Contains(t, buf.String(), "warn message")
				assert.Contains(t, buf.String(), "error message")
			},
		)

		t.Run("panics if an unknown level is passed", func(t *testing.T) {
			loggerFunc, _ := grpcutils.CreateInterceptorLogger(true)
			assert.Panics(t, func() {
				loggerFunc.Log(context.TODO(), 100, "unknown level")
			})
		})
	})

	t.Run(
		"RecoveryHandler logs panic message and formats the error as expected",
		func(t *testing.T) {
			var buf bytes.Buffer
			log.Logger = log.Output(&buf)

			err := grpcutils.RecoveryHandler("test panic")

			assert.Error(t, err)
			assert.Contains(t, buf.String(), "panic triggered: test panic")
		},
	)

	t.Run("NewConnection", func(t *testing.T) {
		host := "localhost:50051"
		t.Run("creates insecure connection when GRPC_USE_TLS is not set", func(t *testing.T) {
			_ = os.Unsetenv("GRPC_USE_TLS")
			conn, err := grpcutils.NewConnection(host)
			assert.NoError(t, err)
			assert.NotNil(t, conn)
		})
		t.Run("creates secure connection when GRPC_USE_TLS is set", func(t *testing.T) {
			t.Setenv("GRPC_USE_TLS", "true")
			conn, err := grpcutils.NewConnection(host)
			assert.NoError(t, err)
			assert.NotNil(t, conn)
		})
	})
}
