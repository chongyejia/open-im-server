package discovery

import (
	"testing"

	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
)

func TestDiscoveryTypeForRuntime(t *testing.T) {
	for _, tc := range []struct {
		name, configured, runtime, want string
		wantErr                         bool
	}{
		{name: "legacy k8s", configured: "k8s", runtime: config.KUBERNETES, want: config.KUBERNETES},
		{name: "upstream kubernetes", configured: config.KUBERNETES, runtime: config.KUBERNETES, want: config.KUBERNETES},
		{name: "implicit kubernetes", runtime: config.KUBERNETES, want: config.KUBERNETES},
		{name: "etcd", configured: config.ETCD, runtime: "source", want: config.ETCD},
		{name: "legacy k8s outside kubernetes", configured: "k8s", runtime: "source", wantErr: true},
		{name: "upstream kubernetes outside kubernetes", configured: config.KUBERNETES, runtime: "source", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := discoveryTypeForRuntime(tc.configured, tc.runtime)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("type=%q error=%v, want type=%q error=%v", got, err, tc.want, tc.wantErr)
			}
		})
	}
}

func TestKubernetesServiceName(t *testing.T) {
	for target, want := range map[string]string{"gateway:grpc": "gateway", "gateway:88": "gateway", "gateway": "gateway"} {
		if got := KubernetesServiceName(target); got != want {
			t.Fatalf("%q: %q != %q", target, got, want)
		}
	}
}
