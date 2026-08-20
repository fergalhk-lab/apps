// billsplit/internal/service/testhelper_test.go
package service_test

import (
	"testing"
	"time"

	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
)

// testAuthExpiry is the token lifetime used by tests that don't exercise
// expiry behaviour themselves.
const testAuthExpiry = 24 * time.Hour

func newTestStore(t *testing.T) localstore.Store {
	t.Helper()
	return testutil.NewTestStore(t)
}
