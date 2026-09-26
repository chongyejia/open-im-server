// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package body

import (
	"strings"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
)

type Notification struct {
	Alert   string  `json:"alert,omitempty"`
	Android Android `json:"android,omitempty"`
	IOS     Ios     `json:"ios,omitempty"`
}

type Android struct {
	Alert  string `json:"alert,omitempty"`
	Title  string `json:"title,omitempty"`
	Intent struct {
		URL string `json:"url,omitempty"`
	} `json:"intent,omitempty"`
	Extras map[string]string `json:"extras,omitempty"`
}
type Ios struct {
	Alert          IosAlert          `json:"alert,omitempty"`
	Sound          string            `json:"sound,omitempty"`
	Badge          string            `json:"badge,omitempty"`
	Extras         map[string]string `json:"extras,omitempty"`
	MutableContent bool              `json:"mutable-content"`
}

type IosAlert struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type Notification3rd struct {
	Android *Android `json:"android,omitempty"`
	IOS     *Ios     `json:"ios,omitempty"`
}

func (n *Notification) SetAlert(alert string, title string, opts *options.Opts) {
	n.Alert = alert
	n.Android.Alert = alert
	n.Android.Title = normalizeAndroidNotificationTitle(title, alert)
	n.IOS.Alert.Body = alert
	n.IOS.Sound = opts.IOSPushSound
	if opts.IOSBadgeCount {
		n.IOS.Badge = "+1"
	}
}

func (n *Notification) SetExtras(extras map[string]string) {
	n.IOS.Extras = extras
	n.Android.Extras = extras
}

func (n *Notification) SetAndroidIntent(pushConf *config.Push) {
	n.Android.Intent.URL = pushConf.JPush.PushIntent
}

func (n *Notification) IOSEnableMutableContent() {
	n.IOS.MutableContent = true
}

func (n *Notification3rd) SetAndroid(alert string, title string, extras map[string]string, pushConf *config.Push) {
	android := Android{
		Alert:  alert,
		Title:  normalizeAndroidNotificationTitle(title, alert),
		Extras: extras,
	}
	android.Intent.URL = pushConf.JPush.PushIntent
	n.Android = &android
}

func normalizeAndroidNotificationTitle(title string, alert string) string {
	normalized := normalizeNotificationTitle(title, alert)
	if normalized == "" {
		return "宠业家"
	}
	return normalized
}

func normalizeNotificationTitle(title string, alert string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" || trimmed == strings.TrimSpace(alert) {
		return ""
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return ""
	}
	return trimmed
}
