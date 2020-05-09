// Copyright 2017 Xiaomi, Inc.
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

package g

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RpcConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type RpcWithTLSConfig struct {         //带有tls加密的RPC配置,可以在不同网络间供transfer互相传输数据时使用
	Enabled bool   `json:"enabled"`    //是否启用带有tls加密的RPC端口监听
	Listen  string `json:"listen"`     //监听的地址
	CrtFile string `json:"crtFile"`    //证书路径
	KeyFile string `json:"keyFile"`    //秘钥路径
}                                      //普通的RPC监听(不带tls加密)可以在内网供agent连接使用

type SocketConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
	Timeout int    `json:"timeout"`
}

type JudgeConfig struct {
	Enabled     bool                    `json:"enabled"`
	Batch       int                     `json:"batch"`
	ConnTimeout int                     `json:"connTimeout"`
	CallTimeout int                     `json:"callTimeout"`
	MaxConns    int                     `json:"maxConns"`
	MaxIdle     int                     `json:"maxIdle"`
	Replicas    int                     `json:"replicas"`
	Cluster     map[string]string       `json:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type GraphConfig struct {
	Enabled     bool                    `json:"enabled"`
	Batch       int                     `json:"batch"`
	ConnTimeout int                     `json:"connTimeout"`
	CallTimeout int                     `json:"callTimeout"`
	MaxConns    int                     `json:"maxConns"`
	MaxIdle     int                     `json:"maxIdle"`
	Replicas    int                     `json:"replicas"`
	Cluster     map[string]string       `json:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type TsdbConfig struct {
	Enabled     bool   `json:"enabled"`
	Batch       int    `json:"batch"`
	ConnTimeout int    `json:"connTimeout"`
	CallTimeout int    `json:"callTimeout"`
	MaxConns    int    `json:"maxConns"`
	MaxIdle     int    `json:"maxIdle"`
	MaxRetry    int    `json:"retry"`
	Address     string `json:"address"`
}

type TransferConfig struct {
	Enabled     bool              `json:"enabled"`
	UseTLS      bool              `json:"useTLS"`     //转发数据到其他的transfer时是否使用tls加密
	Batch       int               `json:"batch"`
	ConnTimeout int               `json:"connTimeout"`
	CallTimeout int               `json:"callTimeout"`
	MaxConns    int               `json:"maxConns"`
	MaxIdle     int               `json:"maxIdle"`
	MaxRetry    int               `json:"retry"`
	Cluster     map[string]string `json:"cluster"`
}

type InfluxdbConfig struct {
	Enabled   bool   `json:"enabled"`
	Batch     int    `json:"batch"`
	MaxRetry  int    `json:"retry"`
	MaxConns  int    `json:"maxConns"`
	Timeout   int    `json:"timeout"`
	Address   string `json:"address"`
	Database  string `json:"db"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Precision string `json:"precision"`
}

type GlobalConfig struct {
	Debug    bool            `json:"debug"`
	MinStep  int             `json:"minStep"` //最小周期,单位sec
	Http     *HttpConfig     `json:"http"`
	Rpc      *RpcConfig      `json:"rpc"`
	RpcWithTLS     *RpcWithTLSConfig  `json:"rpcWithTLS"`
	Socket   *SocketConfig   `json:"socket"`
	Judge    *JudgeConfig    `json:"judge"`
	Graph    *GraphConfig    `json:"graph"`
	Tsdb     *TsdbConfig     `json:"tsdb"`
	Transfer *TransferConfig `json:"transfer"`
	Influxdb *InfluxdbConfig `json:"influxdb"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	// split cluster config
	c.Judge.ClusterList = formatClusterItems(c.Judge.Cluster)
	c.Graph.ClusterList = formatClusterItems(c.Graph.Cluster)

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("g.ParseConfig ok, file ", cfg)
}

// CLUSTER NODE
type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}

// map["node"]="host1,host2" --> map["node"]=["host1", "host2"]
func formatClusterItems(cluster map[string]string) map[string]*ClusterNode {
	ret := make(map[string]*ClusterNode)
	for node, clusterStr := range cluster {
		items := strings.Split(clusterStr, ",")
		nitems := make([]string, 0)
		for _, item := range items {
			nitems = append(nitems, strings.TrimSpace(item))
		}
		ret[node] = NewClusterNode(nitems)
	}

	return ret
}
