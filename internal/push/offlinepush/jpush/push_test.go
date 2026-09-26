package jpush

import (
	"encoding/json"
	"testing"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/jpush/body"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/stretchr/testify/require"
)

func TestBuildAndroidPushObjUsesStandardNotification(t *testing.T) {
	client := NewClient(&config.Push{})
	client.pushConf.JPush.PushIntent = "chongyejia://app.com/message"
	var audience body.Audience
	audience.SetAlias([]string{"user-a"})

	pushObj := client.buildAndroidPushObj(&audience, "[TEXT]", "😇😇😇", map[string]string{
		"ClientMsgID":    "client-msg-id",
		"ex":             "😇😇😇",
		"openim_message": "true",
	}, &options.Opts{})

	payload, err := json.Marshal(pushObj)
	require.NoError(t, err)
	payloadString := string(payload)

	require.Contains(t, payloadString, `"platform":["android"]`)
	require.Contains(t, payloadString, `"notification"`)
	require.Contains(t, payloadString, `"openim_message":"true"`)
	require.Contains(t, payloadString, `"url":"chongyejia://app.com/message"`)
	require.Contains(t, payloadString, `"title":"宠业家"`)
	require.NotContains(t, payloadString, `"notification_3rd"`)
	require.NotContains(t, payloadString, `"notification_3rd_ver"`)
	require.NotContains(t, payloadString, `"message":`)
	require.NotContains(t, payloadString, `"title":"[TEXT]"`)
}

func TestBuildIOSPushObjKeepsNotificationPath(t *testing.T) {
	client := NewClient(&config.Push{})
	var audience body.Audience
	audience.SetAlias([]string{"user-a"})

	pushObj := client.buildIOSPushObj(&audience, "[TEXT]", "😇😇😇", map[string]string{
		"ClientMsgID":    "client-msg-id",
		"ex":             "😇😇😇",
		"openim_message": "true",
	}, &options.Opts{})

	payload, err := json.Marshal(pushObj)
	require.NoError(t, err)
	payloadString := string(payload)

	require.Contains(t, payloadString, `"platform":["ios"]`)
	require.Contains(t, payloadString, `"notification"`)
	require.NotContains(t, payloadString, `"notification_3rd"`)
}

func TestJPushResponseErrorRedactsMsgID(t *testing.T) {
	err := jpushResponseError(map[string]any{
		"msg_id": "18103083292408631",
		"error": map[string]any{
			"code":    float64(1011),
			"message": "cannot find user by this audience or has been inactive for more than 255 days",
		},
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "code=1011")
	require.Contains(t, err.Error(), "cannot find user")
	require.NotContains(t, err.Error(), "18103083292408631")
	require.NotContains(t, err.Error(), "msg_id")
}
