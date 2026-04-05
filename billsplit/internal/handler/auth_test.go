// internal/handler/auth_test.go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fergalhk-lab/apps/billsplit/internal/handler"
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

// newTestRouter registers alice and returns a router with secureCookie=false
// (httptest doesn't use HTTPS).
func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	st := newTestStore(t)
	auth := service.NewAuthService(st, "test-secret")
	invites := service.NewInviteService(st)
	code, err := invites.GenerateInvite(context.Background(), false)
	require.NoError(t, err)
	require.NoError(t, auth.Register(context.Background(), "alice", "password123", code))

	svc := handler.Services{
		Auth:        auth,
		Groups:      service.NewGroupService(st),
		Expenses:    service.NewExpenseService(st),
		Settlements: service.NewSettlementService(st),
		Invites:     invites,
	}
	return handler.NewRouter(svc, false)
}

func sessionCookie(rr *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session" {
			return c
		}
	}
	return nil
}

func TestLoginHandler_SetsCookieAndReturnsIdentity(t *testing.T) {
	router := newTestRouter(t)

	body := `{"username":"alice","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	cookie := sessionCookie(rr)
	require.NotNil(t, cookie, "expected a session cookie in response")
	assert.True(t, cookie.HttpOnly, "session cookie must be HttpOnly")
	assert.NotEmpty(t, cookie.Value, "session cookie value must not be empty")
	assert.Equal(t, 86400, cookie.MaxAge)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "alice", resp["username"])
	assert.Equal(t, false, resp["isAdmin"])
}

func TestLoginHandler_InvalidCredentials_Returns401(t *testing.T) {
	router := newTestRouter(t)

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Nil(t, sessionCookie(rr), "no cookie should be set on failed login")
}
