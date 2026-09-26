package push

import (
	"context"
	"errors"
	"github.com/openimsdk/protocol/msggateway"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/discovery"
	"google.golang.org/grpc"
	"testing"
)

type fanoutDiscovery struct {
	discovery.Conn
	conns []grpc.ClientConnInterface
	err   error
}

func (d fanoutDiscovery) GetConn(context.Context, string, ...grpc.DialOption) (grpc.ClientConnInterface, error) {
	panic("online fanout must not use a load-balanced connection")
}
func (d fanoutDiscovery) GetConns(context.Context, string, ...grpc.DialOption) ([]grpc.ClientConnInterface, error) {
	return d.conns, d.err
}

type fanoutGateway struct {
	grpc.ClientConnInterface
	online bool
	calls  int
}

func (g *fanoutGateway) Invoke(_ context.Context, _ string, _ interface{}, reply interface{}, _ ...grpc.CallOption) error {
	g.calls++
	reply.(*msggateway.OnlineBatchPushOneMsgResp).SinglePushResult = []*msggateway.SingleMsgToUserResults{{UserID: "receiver", OnlinePush: g.online}}
	return nil
}
func TestOnlinePushVisitsEveryGateway(t *testing.T) {
	for _, count := range []int{1, 2} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			gateways := make([]*fanoutGateway, count)
			conns := make([]grpc.ClientConnInterface, count)
			for i := range gateways {
				gateways[i] = &fanoutGateway{online: i == count-1}
				conns[i] = gateways[i]
			}
			p := NewDefaultAllNode(fanoutDiscovery{conns: conns}, &Config{})
			users := []string{"receiver"}
			msg := &sdkws.MsgData{SendID: "sender"}
			results, err := p.GetConnsAndOnlinePush(context.Background(), msg, users)
			if err != nil {
				t.Fatal(err)
			}
			for i, g := range gateways {
				if g.calls != 1 {
					t.Fatalf("gateway %d calls=%d", i, g.calls)
				}
			}
			if offline := p.GetOnlinePushFailedUserIDs(context.Background(), msg, results, &users); len(offline) != 0 {
				t.Fatalf("online receiver incorrectly offline: %v", offline)
			}
		})
	}
}
func TestOnlinePushDiscoveryError(t *testing.T) {
	want := errors.New("discovery unavailable")
	p := NewDefaultAllNode(fanoutDiscovery{err: want}, &Config{})
	if _, err := p.GetConnsAndOnlinePush(context.Background(), &sdkws.MsgData{}, nil); !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

type namedFanoutDiscovery struct {
	fanoutDiscovery
	name string
}

func (d *namedFanoutDiscovery) GetConns(_ context.Context, name string, _ ...grpc.DialOption) ([]grpc.ClientConnInterface, error) {
	d.name = name
	return d.conns, nil
}
func TestKubernetesGatewayFanout(t *testing.T) {
	d := &namedFanoutDiscovery{fanoutDiscovery: fanoutDiscovery{conns: []grpc.ClientConnInterface{&fanoutGateway{}, &fanoutGateway{}}}}
	cfg := &Config{}
	cfg.Discovery.RpcService.MessageGateway = "gateway:grpc"
	p := NewDefaultAllNode(d, cfg)
	conns, err := p.gatewayConns(context.Background(), "kubernetes")
	if err != nil || len(conns) != 2 || d.name != "gateway" {
		t.Fatalf("connections=%d service=%q error=%v", len(conns), d.name, err)
	}
	_, err = p.gatewayConns(context.Background(), "source")
	if err != nil || d.name != "gateway:grpc" {
		t.Fatalf("non-Kubernetes target changed: %q %v", d.name, err)
	}
}
