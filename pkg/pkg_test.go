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

package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/mgw-notifier/pkg/auth"
	"github.com/SENERGY-Platform/mgw-notifier/pkg/configuration"
	"github.com/SENERGY-Platform/mgw-notifier/pkg/relay"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestPkg(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auth := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		json.NewEncoder(writer).Encode(auth.OpenidToken{
			AccessToken:      "test",
			ExpiresIn:        10000000,
			RefreshExpiresIn: 10000000,
			RefreshToken:     "",
			TokenType:        "",
			RequestTime:      time.Now(),
		})
	}))

	go func() {
		<-ctx.Done()
		auth.Close()
	}()

	calls := map[string][]string{}
	backend := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		pl, _ := io.ReadAll(request.Body)
		calls[request.URL.Path] = append(calls[request.URL.Path], string(pl))
	}))

	go func() {
		<-ctx.Done()
		backend.Close()
	}()

	port, err := GetFreePortStr()
	if err != nil {
		t.Error(err)
		return
	}

	err = relay.Start(configuration.Config{
		NotificationUrl:          backend.URL + "/foo",
		Port:                     port,
		AuthExpirationTimeBuffer: 2,
		AuthEndpoint:             auth.URL,
		Debug:                    true,
	}, ctx)

	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(1 * time.Second)

	_, err = http.Post("http://localhost:"+port+"/bar", "", bytes.NewBufferString("batz"))
	if err != nil {
		t.Error(err)
		return
	}

	_, err = http.Post("http://localhost:"+port, "", bytes.NewBufferString("batz2"))
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(calls, map[string][]string{"/foo/bar": {"batz"}, "/foo": {"batz2"}}) {
		t.Error(calls)
	}
}

func GetFreePortStr() (string, error) {
	intPort, err := GetFreePort()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(intPort), nil
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
