package push

import (
	"github.com/openimsdk/protocol/sdkws"
	"testing"
)

func TestOfflinePushMetadataBothConsumers(t *testing.T) {
	msg := &sdkws.MsgData{ClientMsgID: "client-1", SendID: "sender", GroupID: "group-1", SessionType: 3, OfflinePushInfo: &sdkws.OfflinePushInfo{Title: "title", Desc: "body", Ex: "extra"}}
	consumer := &ConsumerHandler{}
	separate := &OfflinePushConsumerHandler{}
	_, _, first, err := consumer.getOfflinePushInfos(msg)
	if err != nil {
		t.Fatal(err)
	}
	_, _, second, err := separate.getOfflinePushInfos(msg)
	if err != nil {
		t.Fatal(err)
	}
	if first.SendID != msg.SendID || first.GroupID != msg.GroupID || first.SessionType != msg.SessionType || first.Signal.ClientMsgID != msg.ClientMsgID || first.Ex != "extra" {
		t.Fatalf("main consumer lost routing metadata: %+v", first)
	}
	if second.SendID != first.SendID || second.GroupID != first.GroupID || second.SessionType != first.SessionType || second.Signal.ClientMsgID != first.Signal.ClientMsgID || second.Ex != first.Ex {
		t.Fatalf("offline consumer lost routing metadata: %+v", second)
	}
}
