package testutils

import (
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"runtime"
	"time"
)

// CreatePubsubTestContainer - creates a test container for the gcloud pubsub emulator.
// The pubsub SDK checks for the PUBSUB_EMULATOR_HOST environment variable and uses it if it exists.
// So all that is required to execute pubsub in tests is to run this function.
func CreatePubsubTestContainer() func() {
	environment := map[string]string{}

	if runtime.GOOS == "darwin" {
		environment["TESTCONTAINERS_HOST_OVERRIDE"] = "host.docker.internal"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	container, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				// see: https://hub.docker.com/r/thekevjames/gcloud-pubsub-emulator
				Name:         "gcloud-pubsub-emulator",
				Image:        "thekevjames/gcloud-pubsub-emulator:406.0.0",
				ExposedPorts: []string{"8681/tcp"},
				WaitingFor:   wait.ForListeningPort("8681/tcp"),
				Env:          environment,
			},
			Reuse:   true,
			Started: true,
		},
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pubsub test container")
	}

	exc.LogIfErr(
		os.Setenv("PUBSUB_EMULATOR_HOST", exc.MustResult(container.Endpoint(context.TODO(), ""))),
	)

	return func() {
		cancel()
		exc.LogIfErr(os.Unsetenv("PUBSUB_EMULATOR_HOST"))
		exc.LogIfErr(container.Terminate(context.TODO()))
	}
}
