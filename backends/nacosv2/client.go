package nacosv2

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/confd/log"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var replacer = strings.NewReplacer("/", ".")

type Client struct {
	configClient config_client.IConfigClient
	group string
	namespace string
	channel chan int
	count int
}

func NewNacosClient(group string, sConfig constant.ServerConfig, cConfig constant.ClientConfig) (client *Client, err error) {
	
	if len(strings.TrimSpace(group)) == 0 {
		group = "DEFAULT_GROUP"
	}

	log.Info(fmt.Sprintf("serverConfig=%v, namespace=%s, group=%s", sConfig, cConfig.NamespaceId, group))
	
	clientConfig := cConfig

	serverConfigs := []constant.ServerConfig{
		sConfig,
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	client = &Client{configClient, group, cConfig.NamespaceId, make(chan int, 10), 0}

	return
}

func (client *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)

	for _, key := range keys {
		k := strings.TrimPrefix(key, "/")
		k = replacer.Replace(k)

		resp, err := client.configClient.GetConfig(vo.ConfigParam{
			DataId: k,
			Group: client.group,
		})

		log.Info(fmt.Sprintf("key=%s, value=%s", key, resp))

		if err == nil {
			vars[key] = resp
		}
	}

	return vars, nil
}


func (client *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		client.count++
		for _, key := range keys {
			k := strings.TrimPrefix(key, "/")
			k = replacer.Replace(k)

			err := client.configClient.ListenConfig(vo.ConfigParam{
				DataId: k,
				Group: client.group,
				OnChange: func(namespace, group, dataId, data string) {
					log.Info(fmt.Sprintf("config namespace=%s, dataId=%s, group=%s has changed", namespace, dataId, group))
					for i := 0; i < client.count; i++ {
						client.channel <- 1	
					}
				},
			})
			if err != nil {
				return 0,err
			}
		}
		return 1, nil
	}
	select {
		case <- client.channel:
			return waitIndex,nil
	}

	return waitIndex, nil
}