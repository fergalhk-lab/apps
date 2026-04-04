// billsplit/internal/service/testhelper_test.go
package service_test

import (
	"testing"

	localstore "github.com/fergalhk-lab/apps/billsplit/internal/store"
	"github.com/fergalhk-lab/apps/billsplit/internal/testutil"
)

func newTestStore(t *testing.T) localstore.Store {
	t.Helper()
	return testutil.NewTestStore(t)
}
