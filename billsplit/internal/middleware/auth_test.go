// internal/middleware/auth_test.go
package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) localstore.Store {
	t.Helper()
	minio := testutil.StartMinIO(t)
	ctx := context.Background()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(
			func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{AccessKeyID: minio.AccessKey, SecretAccessKey: minio.SecretKey}, nil
			},
		)),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(svc, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: minio.Endpoint, HostnameImmutable: true}, nil
			},
		)),
	)
	require.NoError(t, err)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })
	bucket := "test-bucket"
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	require.NoError(t, err)
	return localstore.NewS3Store(client, bucket)
}

func setupAuth(t *testing.T) (*service.AuthService, string) {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "test-secret")
	invites := service.NewInviteService(st)
	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(context.Background(), "alice", "password123", code))
	token, _, err := auth.Login(context.Background(), "alice", "password123")
	require.NoError(t, err)
	return auth, token
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequireAuth_ValidCookie_Passes(t *testing.T) {
	auth, token := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireAuth_ValidCookie_SetsUsernameInContext(t *testing.T) {
	auth, token := setupAuth(t)
	var gotUsername string
	handler := middleware.RequireAuth(auth, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUsername = middleware.UsernameFromCtx(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "alice", gotUsername)
}

func TestRequireAuth_NoCookie_Returns401(t *testing.T) {
	auth, _ := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_InvalidTokenInCookie_Returns401(t *testing.T) {
	auth, _ := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "not-a-valid-jwt"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRequireAuth_BearerHeader_Returns401(t *testing.T) {
	// Bearer tokens must no longer be accepted — cookie only.
	auth, token := setupAuth(t)
	handler := middleware.RequireAuth(auth, okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
