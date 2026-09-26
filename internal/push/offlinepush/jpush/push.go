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

package jpush

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/jpush/body"
	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
	"github.com/openimsdk/open-im-server/v3/pkg/common/config"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/httputil"
)

type JPush struct {
	pushConf   *config.Push
	httpClient *httputil.HTTPClient
}

func NewClient(pushConf *config.Push) *JPush {
	return &JPush{pushConf: pushConf,
		httpClient: httputil.NewHTTPClient(httputil.NewClientConfig()),
	}
}

func (j *JPush) Auth(apiKey, secretKey string, timeStamp int64) (token string, err error) {
	return token, nil
}

func (j *JPush) SetAlias(cid, alias string) (resp string, err error) {
	return resp, nil
}

func (j *JPush) getAuthorization(appKey string, masterSecret string) string {
	str := fmt.Sprintf("%s:%s", appKey, masterSecret)
	buf := []byte(str)
	Authorization := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(buf))
	return Authorization
}

func (j *JPush) Push(ctx context.Context, userIDs []string, title, content string, opts *options.Opts) error {
	var au body.Audience
	au.SetAlias(userIDs)
	extras := make(map[string]string)
	extras["ex"] = opts.Ex
	extras["openim_message"] = "true"
	if opts.Signal.ClientMsgID != "" {
		extras["ClientMsgID"] = opts.Signal.ClientMsgID
	}

	androidPushObj := j.buildAndroidPushObj(&au, title, content, extras, opts)
	var androidResp map[string]any
	androidErr := j.request(ctx, androidPushObj, &androidResp, 5)
	j.logPlatformResult(ctx, "android", len(userIDs), len(content), androidErr)

	iosPushObj := j.buildIOSPushObj(&au, title, content, extras, opts)
	var iosResp map[string]any
	iosErr := j.request(ctx, iosPushObj, &iosResp, 5)
	j.logPlatformResult(ctx, "ios", len(userIDs), len(content), iosErr)
	if androidErr != nil && iosErr != nil {
		return fmt.Errorf("jpush android push failed: %w; ios push failed: %v", androidErr, iosErr)
	}
	return nil
}

func (j *JPush) logPlatformResult(ctx context.Context, platform string, userCount int, contentLength int, err error) {
	fields := []any{
		"targetPlatform", platform,
		"userCount", userCount,
		"contentLength", contentLength,
	}
	if err != nil {
		log.ZWarn(ctx, "jpush platform push failed", err, fields...)
		return
	}
	log.ZInfo(ctx, "jpush platform push accepted", fields...)
}

func (j *JPush) buildAndroidPushObj(au *body.Audience, title, content string, extras map[string]string, opts *options.Opts) body.PushObj {
	var pf body.Platform
	_ = pf.SetAndroid()

	var no body.Notification
	no.SetAlert(content, title, opts)
	no.SetExtras(extras)
	no.SetAndroidIntent(j.pushConf)

	var opt body.Options
	opt.SetApnsProduction(j.pushConf.IOSPush.Production)

	var pushObj body.PushObj
	pushObj.SetPlatform(&pf)
	pushObj.SetAudience(au)
	pushObj.SetNotification(&no)
	pushObj.SetOptions(&opt)
	return pushObj
}

func (j *JPush) buildIOSPushObj(au *body.Audience, title, content string, extras map[string]string, opts *options.Opts) body.PushObj {
	var pf body.Platform
	_ = pf.SetIOS()

	var no body.Notification
	no.IOSEnableMutableContent()
	no.SetExtras(extras)
	no.SetAlert(content, title, opts)
	no.SetAndroidIntent(j.pushConf)

	var msg body.Message
	msg.SetMsgContent(content)
	msg.SetTitle(title)
	for key, value := range extras {
		msg.SetExtras(key, value)
	}

	var opt body.Options
	opt.SetApnsProduction(j.pushConf.IOSPush.Production)

	var pushObj body.PushObj
	pushObj.SetPlatform(&pf)
	pushObj.SetAudience(au)
	pushObj.SetNotification(&no)
	pushObj.SetMessage(&msg)
	pushObj.SetOptions(&opt)
	return pushObj
}

func (j *JPush) request(ctx context.Context, po body.PushObj, resp *map[string]any, timeout int) error {
	err := j.httpClient.PostReturn(
		ctx,
		j.pushConf.JPush.PushURL,
		map[string]string{
			"Authorization": j.getAuthorization(j.pushConf.JPush.AppKey, j.pushConf.JPush.MasterSecret),
		},
		po,
		resp,
		timeout,
	)
	if err != nil {
		return err
	}
	if (*resp)["sendno"] != "0" {
		return jpushResponseError(*resp)
	}
	return nil
}

func jpushResponseError(resp map[string]any) error {
	errValue, _ := resp["error"].(map[string]any)
	if len(errValue) == 0 {
		return fmt.Errorf("jpush push failed")
	}
	return fmt.Errorf(
		"jpush push failed code=%v message=%v",
		errValue["code"],
		errValue["message"],
	)
}
