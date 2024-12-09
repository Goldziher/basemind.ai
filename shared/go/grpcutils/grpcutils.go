package grpcutils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/exc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"os"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	loggingmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateInterceptorLogger creates a logger interceptor for use with grpc middleware.
// uses Zerolog and relies on the global zerolog config.
func CreateInterceptorLogger(isDebug bool) (loggingmiddleware.Logger, []loggingmiddleware.Option) {
	loggingEvents := []loggingmiddleware.LoggableEvent{
		loggingmiddleware.StartCall, loggingmiddleware.FinishCall,
	}

	if isDebug {
		loggingEvents = append(
			loggingEvents,
			loggingmiddleware.PayloadSent, loggingmiddleware.PayloadReceived,
		)
	}

	loggingOptions := []loggingmiddleware.Option{
		loggingmiddleware.WithLogOnEvents(loggingEvents...),
	}

	handlerFunc := loggingmiddleware.LoggerFunc(
		func(ctx context.Context, lvl loggingmiddleware.Level, msg string, fields ...any) {
			logger := log.Logger.With().Fields(fields).Logger()

			switch lvl {
			case loggingmiddleware.LevelDebug:
				logger.Debug().Msg(msg)
			case loggingmiddleware.LevelInfo:
				logger.Info().Msg(msg)
			case loggingmiddleware.LevelWarn:
				logger.Warn().Msg(msg)
			case loggingmiddleware.LevelError:
				logger.Error().Msg(msg)
			default:
				panic(fmt.Sprintf("unknown level %v", lvl))
			}
		},
	)

	return handlerFunc, loggingOptions
}

// ServiceRegistrar is a function that registers a service with a grpc server.
type ServiceRegistrar func(s grpc.ServiceRegistrar)

// Options is a struct that contains options for creating a grpc server.
type Options struct {
	// AuthHandler is the auth handler function for the service.
	AuthHandler auth.AuthFunc
	// Environment is the current environment. Can be 'testing', 'development' or 'production'
	Environment string
	// ServiceName is the name of the service. Used for logging.
	ServiceName string
	// ServiceRegistrars is a list of functions that register services with the grpc server.
	ServiceRegistrars []ServiceRegistrar
}

// RecoveryHandler is a handler for the grpc recovery interceptor.
func RecoveryHandler(p any) (err error) {
	log.Error().Bytes("stack", debug.Stack()).Msgf("panic triggered: %v", p)
	return status.Error(codes.Internal, "an internal error occurred")
}

// CreateGRPCServer creates a grpc server with the given options.
func CreateGRPCServer(opts Options, serverOpts ...grpc.ServerOption) *grpc.Server {
	if opts.Environment != "test" {
		interceptorLogger, loggingOptions := CreateInterceptorLogger(
			opts.Environment != "production",
		)

		serverOpts = append(
			serverOpts,
			grpc.ChainUnaryInterceptor(
				loggingmiddleware.UnaryServerInterceptor(interceptorLogger, loggingOptions...),
				auth.UnaryServerInterceptor(opts.AuthHandler),
				recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(RecoveryHandler)),
			),
			grpc.ChainStreamInterceptor(
				loggingmiddleware.StreamServerInterceptor(interceptorLogger, loggingOptions...),
				auth.StreamServerInterceptor(opts.AuthHandler),
				recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(RecoveryHandler)),
			),
		)
	}

	server := grpc.NewServer(serverOpts...) // skipcq: GO-S0902

	for _, registrar := range opts.ServiceRegistrars {
		registrar(server)
	}

	// enable the health check protocol
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	return server
}

// NewConnection creates a new grpc connection.
// if the GRPC_USE_TLS environment variable is set, a TLS connection is used.
// otherwise, an insecure connection is used.
func NewConnection(host string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithAuthority(host))

	if os.Getenv("GRPC_USE_TLS") != "" {
		log.Info().Msg("using TLS for gRPC connection")
		systemRoots := exc.MustResult(x509.SystemCertPool())
		cred := credentials.NewTLS(&tls.Config{ //nolint:gosec
			RootCAs: systemRoots,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	} else {
		log.Info().Msg("using insecure connection for gRPC")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.Dial(host, opts...)
}
