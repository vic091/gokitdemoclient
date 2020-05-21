package utils

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	consulapi "github.com/hashicorp/consul/api"
	. "gomicro2/Services"
	"io"
	"net/url"
	"os"
	"time"
)

func GetUser() (string, error) {
	{
		config := consulapi.DefaultConfig()
		config.Address = "192.168.137.129:8500"
		api_client, _ := consulapi.NewClient(config)
		client := consul.NewClient(api_client)
		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(os.Stdout)
		}
		{
			tags := []string{"primary"}
			// 可查询服务的状态
			instancer := consul.NewInstancer(client, logger, "userservice", tags, true)
			{
				factory := func(service_url string) (endpoint.Endpoint, io.Closer, error) {

					tgt, _ := url.Parse("http://" + service_url)
					return httptransport.NewClient("GET", tgt, GetUserInfo_Request, GetUserInfo_Response).Endpoint(), nil, nil
				}
				endpointer := sd.NewEndpointer(instancer, factory, logger)
				//endpoiints, _ := endpointer.Endpoints()
				//
				//fmt.Println("服务有：", len(endpoiints))
				//mylb := lb.NewRoundRobin(endpointer) // 轮询
				mylb := lb.NewRandom(endpointer, time.Now().UnixNano()) // 随机
				//getUserInfo := endpoiints[0]
				getUserInfo, _ := mylb.Endpoint() //
				// 第三，四步，创建context上下文，并执行
				res, err := getUserInfo(context.Background(), UserRequest{Uid: 102})
				if err != nil {
					fmt.Println(err)
					return "", err
				}
				// 第五步 断言，得到响应
				userInfo := res.(UserResponse)
				fmt.Println(userInfo.Result)
				return userInfo.Result, nil
			}
		}
	}
}
