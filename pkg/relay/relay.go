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
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewRelay(config configuration.Config) (result *Relay, err error) {
	target, err := url.Parse(config.NotificationUrl)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(request *http.Request) {
		origDirector(request)
		request.URL.RawPath = strings.TrimSuffix(request.URL.RawPath, "/")
		request.URL.Path = strings.TrimSuffix(request.URL.Path, "/")
	}

	return &Relay{
		target: target,
		config: config,
		auth:   &auth.OpenidToken{},
		proxy:  proxy,
	}, nil
}

type Relay struct {
	target *url.URL
	auth   *auth.OpenidToken
	config configuration.Config
	proxy  *httputil.ReverseProxy
}

func (this *Relay) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	token, err := this.auth.EnsureAccess(this.config)
	if err != nil {
		log.Println("ERROR: unable to get auth token:", err)
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if this.config.Debug {
		log.Println("forward", req.Method, req.URL.String(), "to", this.config.NotificationUrl)
	}
	req.Host = this.target.Host //proxy dos not replace host header
	req.Header.Set("Authorization", token)
	this.proxy.ServeHTTP(res, req)
}
