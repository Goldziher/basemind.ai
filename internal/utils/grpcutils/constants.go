package grpcutils

type contextKeyType int

const (
	// ApplicationIDContextKey is the key used to store the application id in the context.
	ApplicationIDContextKey contextKeyType = iota
	ProjectIDContextKey     contextKeyType = iota
)
