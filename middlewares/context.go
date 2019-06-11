package middlewares

// User context key is stored in spec so both spec & other packages importing spec can access & set user values in context
// Context key types & values must match in ctx.Value, meaning keys have package uniquness.
// Therefore if spec or other package wants ctx key value we must use these helper functions.
// For example, we need user fields in our Logging middleware.

import (
	"context"
	"errors"
)

type contextKey string

var userContextKey = contextKey("user")

// OrgIDFromContext retrieves an organization ID value stored in a context
func OrgIDFromContext(ctx context.Context) (string, error) {
	u := ctx.Value(userContextKey)
	if u == nil {
		return "", errors.New("User is not stored in given context")
	}
	orgID, ok := u.(map[string]interface{})["orgID"]
	if !ok {
		return "", errors.New("OrgID is not stored in given context")
	}
	orgIDString, ok := orgID.(string)
	if !ok || orgIDString == "" {
		return "", errors.New("OrgID not stored as string in given context")
	}
	return orgIDString, nil
}

// UserIDFromContext retrieves an organization ID value stored in a context
func UserIDFromContext(ctx context.Context) (string, error) {
	u := ctx.Value(userContextKey)
	if u == nil {
		return "", errors.New("User is not stored in given context")
	}
	v, ok := u.(map[string]interface{})["userID"]
	if !ok {
		return "", errors.New("UserID is not stored in given context")
	}
	vString, ok := v.(string)
	if !ok || vString == "" {
		return "", errors.New("userID not stored as string in given context")
	}
	return vString, nil
}

// ContextWithUser places a user ID value, org Id value, and admin bool into a context using the same context user key
func ContextWithUser(ctx context.Context, userID string, orgID string, admin bool) context.Context {
	userValues := map[string]interface{}{
		"orgID":  orgID,
		"userID": userID,
		"admin":  admin,
	}
	return context.WithValue(ctx, userContextKey, userValues)
}

// IsAdminFromContext returns a boolean indicating whether the user is an admin or not
func IsAdminFromContext(ctx context.Context) bool {
	u := ctx.Value(userContextKey)
	if u == nil {
		return false
	}

	if v, ok := u.(map[string]interface{})["admin"]; !ok || v == false {
		return false
	}
	return true
}
