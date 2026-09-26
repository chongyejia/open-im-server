package discovery

import "testing"

func TestKubernetesServiceName(t *testing.T) {
	for target, want := range map[string]string{"gateway:grpc": "gateway", "gateway:88": "gateway", "gateway": "gateway"} {
		if got := KubernetesServiceName(target); got != want {
			t.Fatalf("%q: %q != %q", target, got, want)
		}
	}
}
