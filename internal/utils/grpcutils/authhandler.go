package grpcutils

import (
	"context"
	"github.com/basemind-ai/monorepo/internal/utils/exc"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/basemind-ai/monorepo/shared/go/jwtutils"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthHandler is an Auth handler function fulfilling the type specified by
// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/auth/auth.go#L24
type AuthHandler struct {
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler instance without requiring its underlying attributes be exposed.
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{
		jwtSecret: jwtSecret,
	}
}

// HandleAuth handles authentication for a request.
// It expects the request to have a bearer token in the metadata.
func (handler *AuthHandler) HandleAuth(ctx context.Context) (context.Context, error) {
	token, metadataErr := auth.AuthFromMD(ctx, "bearer")
	if metadataErr != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to get metadata: %v", metadataErr)
	}

	claims, tokenErr := jwtutils.ParseJWT(token, []byte(handler.jwtSecret))
	if tokenErr != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", tokenErr)
	}

	sub, subErr := claims.GetSubject()
	if subErr != nil || sub == "" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", subErr)
	}

	apiKeyID := exc.MustResult(db.StringToUUID(sub))

	ids, retrieveErr := db.GetQueries().RetrieveApplicationDataForAPIKey(ctx, *apiKeyID)
	if retrieveErr != nil {
		return nil, status.Errorf(
			codes.Unauthenticated,
			"failed to retrieve application id for token: %v",
			retrieveErr,
		)
	}
	applicationIDContext := context.WithValue(ctx, ApplicationIDContextKey, ids.ApplicationID)
	return context.WithValue(applicationIDContext, ProjectIDContextKey, ids.ProjectID), nil
}
