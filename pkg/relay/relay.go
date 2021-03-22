/*
 * Copyright 2021 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package relay

import (
	"github.com/SENERGY-Platform/mgw-notifier/pkg/auth"
	"github.com/SENERGY-Platform/mgw-notifier/pkg/configuration"
	"net/http"
	"net/http/httputil"
	"net/url"
)


func NewRelay(config configuration.Config) (result *Relay, err error) {
	target, err := url.Parse(config.NotificationUrl)
	if err != nil {
		return nil, err
	}
	return &Relay{
		config: config,
		auth: &auth.OpenidToken{},
		proxy: httputil.NewSingleHostReverseProxy(target),
	}, nil
}

type Relay struct {
	auth   *auth.OpenidToken
	config configuration.Config
	proxy  *httputil.ReverseProxy
}

func (this *Relay) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	token, err := this.auth.EnsureAccess(this.config)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", token)
	this.proxy.ServeHTTP(res, req)
}
