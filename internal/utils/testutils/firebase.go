package testutils

import (
	"context"
	"github.com/basemind-ai/monorepo/shared/go/firebaseutils"
	"testing"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/mock"
)

type FirebaseAuthMock struct {
	mock.Mock
}

func (m *FirebaseAuthMock) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	args := m.Called(ctx, idToken)

	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *FirebaseAuthMock) GetUser(
	ctx context.Context,
	firebaseID string,
) (*auth.UserRecord, error) {
	args := m.Called(ctx, firebaseID)

	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

func (m *FirebaseAuthMock) DeleteUser(
	ctx context.Context,
	firebaseID string,
) error {
	args := m.Called(ctx, firebaseID)

	return args.Error(0)
}

func MockFirebaseAuth(t *testing.T) *FirebaseAuthMock {
	t.Helper()

	// we have to exhaust the sync.Once
	firebaseutils.GetFirebaseAuth(context.TODO())

	authMock := &FirebaseAuthMock{}
	firebaseutils.SetFirebaseAuth(authMock)

	return authMock
}
