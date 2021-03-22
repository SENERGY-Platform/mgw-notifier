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
	"context"
	"github.com/SENERGY-Platform/mgw-notifier/pkg/configuration"
	"log"
	"net/http"
	"time"
)

func Start(config configuration.Config, ctx context.Context) (err error) {
	log.Println("start proxy on " + config.Port)

	var handler http.Handler
	handler, err = NewRelay(config)
	if err != nil {
		return err
	}
	if config.Debug {
		handler = NewLogger(handler)
	}

	server := &http.Server{Addr: ":" + config.Port, Handler: handler, WriteTimeout: 10 * time.Second, ReadTimeout: 2 * time.Second, ReadHeaderTimeout: 2 * time.Second}
	go func() {
		log.Println("listening on ", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("ERROR: server error", err)
		}
	}()
	go func() {
		<-ctx.Done()
		err = server.Shutdown(context.Background())
		if config.Debug {
			log.Println("DEBUG: proxy shutdown", err)
		}
	}()
	return nil
}
