// Copyright 2020 Douyu
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

package invoker

import (
	"time"

	"github.com/douyu/juno/pkg/cfg"
	"github.com/douyu/jupiter/pkg/client/rocketmq"
	"github.com/douyu/jupiter/pkg/store/gorm"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
	clientv3 "go.etcd.io/etcd/client/v3"
	ggorm "gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	// JunoMysql mysql instance
	JunoMysql *gorm.DB
	// ConfgoEtcd for configuration
	ConfgoEtcd *clientv3.Client
	// Resty ..
	Resty *resty.Client

	// event producer
	EventProducer *rocketmq.Producer
)

// Init invoker
func Init() {
	if cfg.Cfg.Database.Enable {
		gormConfig := gorm.DefaultConfig()
		gormConfig.Name = "juno"
		gormConfig.DSN = cfg.Cfg.Database.DSN
		JunoMysql = gormConfig.WithGormConfig(ggorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
		}).MustSingleton()
	}

	var err error
	if cfg.Cfg.ServerProxy.DefaultEtcd.Enable {
		ConfgoEtcd, err = clientv3.New(clientv3.Config{
			Endpoints:   cfg.Cfg.ServerProxy.DefaultEtcd.Endpoints,
			DialTimeout: 2 * time.Second,
			Username:    cfg.Cfg.ServerProxy.DefaultEtcd.UserName,
			Password:    cfg.Cfg.ServerProxy.DefaultEtcd.Password,
		})
		if err != nil {
			panic(err.Error())
		}
	}

	Resty = resty.New().
		SetDebug(true).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cast.ToDuration("20s"))

	if cfg.Cfg.JunoEvent.Rocketmq.Enable {
		config := cfg.Cfg.JunoEvent.Rocketmq
		mqConfig := rocketmq.DefaultConfig().Producer
		mqConfig.Group = config.Group
		mqConfig.Retry = config.Retry
		mqConfig.Topic = config.Topic
		mqConfig.Addr = config.Addr
		mqConfig.DialTimeout = config.DialTimeout

		EventProducer = mqConfig.Build()
		if err := EventProducer.Start(); err != nil {
			panic("init junoevnet producer failed. err=" + err.Error())
		}
	}
}
