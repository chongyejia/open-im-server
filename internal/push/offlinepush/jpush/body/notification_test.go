package body

import (
	"encoding/json"
	"testing"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/stretchr/testify/require"
)

func TestSetAlertUsesAndroidFallbackTitleForContentTypeTitle(t *testing.T) {
	var notification Notification

	notification.SetAlert("😇😇😇", "[TEXT]", &options.Opts{})

	require.Equal(t, "😇😇😇", notification.Alert)
	require.Equal(t, "😇😇😇", notification.Android.Alert)
	require.Equal(t, "宠业家", notification.Android.Title)
	require.Equal(t, "😇😇😇", notification.IOS.Alert.Body)
	require.Empty(t, notification.IOS.Alert.Title)
}

func TestNotification3rdAndroidUsesFallbackTitleForContentTypeTitle(t *testing.T) {
	var notification Notification3rd
	pushConf := &config.Push{}
	pushConf.JPush.PushIntent = "chongyejia://app.com/message"

	notification.SetAndroid("😇😇😇", "[TEXT]", map[string]string{
		"openim_message": "true",
	}, pushConf)

	require.NotNil(t, notification.Android)
	require.Equal(t, "😇😇😇", notification.Android.Alert)
	require.Equal(t, "宠业家", notification.Android.Title)
	require.Equal(t, "chongyejia://app.com/message", notification.Android.Intent.URL)
	require.Equal(t, "true", notification.Android.Extras["openim_message"])
}

func TestPushObjSerializesNotification3rdWithoutNotification(t *testing.T) {
	var pushObj PushObj
	pushObj.Notification3rd = &Notification3rd{Android: &Android{Alert: "hello"}}
	pushObj.Message = &Message{MsgContent: "hello"}
	pushObj.Options = &Options{Notification3rdVer: "v2"}

	payload, err := json.Marshal(pushObj)
	require.NoError(t, err)

	require.Contains(t, string(payload), "notification_3rd")
	require.Contains(t, string(payload), "notification_3rd_ver")
	require.NotContains(t, string(payload), "\"notification\":")
}
