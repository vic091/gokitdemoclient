package main

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	consulapi "github.com/hashicorp/consul/api"
	. "gomicro2/Services"
	"gomicro2/utils"
	"io"
	"net/url"
	"os"
	"time"
)

// 直连方式
func main_2() {
	tgt, _ := url.Parse("http://localhost:8080")
	// 第一步：创建一个直连client，这里我们必须写两个func，一个是如何请求，一个是响应我们怎么处理
	client := httptransport.NewClient("GET", tgt, GetUserInfo_Request, GetUserInfo_Response)
	// 第二步，暴露endpoint 这货就是一个func，以便执行
	getUserInfo := client.Endpoint()
	// 第三，四步，创建context上下文，并执行
	res, err := getUserInfo(context.Background(), UserRequest{Uid: 101})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// 第五步 断言，得到响应
	userInfo := res.(UserResponse)
	fmt.Println(userInfo.Result)
}

func main3() {
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
				endpoiints, _ := endpointer.Endpoints()

				fmt.Println("服务有：", len(endpoiints))
				//mylb := lb.NewRoundRobin(endpointer) // 轮询
				mylb := lb.NewRandom(endpointer, time.Now().UnixNano()) // 随机
				for {
					//getUserInfo := endpoiints[0]
					getUserInfo, _ := mylb.Endpoint() //
					// 第三，四步，创建context上下文，并执行
					res, err := getUserInfo(context.Background(), UserRequest{Uid: 102})
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					// 第五步 断言，得到响应
					userInfo := res.(UserResponse)
					fmt.Println(userInfo.Result)
					time.Sleep(3 * time.Second)
				}
			}
		}
	}
}

func main() {
	configA := hystrix.CommandConfig{
		Timeout:                2000,
		MaxConcurrentRequests:  5,
		RequestVolumeThreshold: 3,
		ErrorPercentThreshold:  20,
		SleepWindow:            int(time.Second * 100),
	}
	hystrix.ConfigureCommand("getuser", configA)
	err := hystrix.Do("getuser", func() error {
		res, err := utils.GetUser()
		fmt.Println(res)
		return err
	}, func(e error) error {
		fmt.Println("降级用户")
		return e
	})

	if err != nil {

	}

	fmt.Println(err)
}
