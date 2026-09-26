package offlinepush

import (
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/emas"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"testing"
)

func TestEmasFactoryFailsClosedAndSelectsOnlyEmas(t *testing.T) {
	t.Setenv("EMAS_OFFLINE_PUSH_URL", "http://mp-api/internal/im/offline-push")
	t.Setenv("EMAS_OFFLINE_PUSH_TOKEN", "")
	if _, err := NewOfflinePusher(&config.Push{Enable: "emas"}, nil, ""); err == nil {
		t.Fatal("missing token must fail startup")
	}
	t.Setenv("EMAS_OFFLINE_PUSH_TOKEN", "test-token")
	p, err := NewOfflinePusher(&config.Push{Enable: "EMAS"}, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.(*emas.Client); !ok {
		t.Fatalf("selected wrong provider: %T", p)
	}
}
