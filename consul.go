package main

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Consul struct {
	client *api.Client
}

func NewConsul(consulHost string, consulPort int) (*Consul, error) {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%d", consulHost, consulPort)
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建consul客户端失败:%w", err)
	}
	return &Consul{client: client}, nil
}

// 注册consul
func (c *Consul) RegisterConsul(serviceName string, address string, port int, tags []string) error {
	if tags == nil {
		tags = []string{}
	}
	//参数校验
	if serviceName == "" || address == "" || port <= 0 {
		return fmt.Errorf("无效的注册参数")
	}
	// 生成唯一 ServiceID（格式：服务名-IP-端口）
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, address, port)
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Tags:    tags,
	}
	// 注册服务
	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("consul 注册失败: %v", err)
	}
	zap.S().Info("服务注册成功", "serviceName", serviceName, "serviceID", serviceID, "address", fmt.Sprintf("%s:%d", address, port))

	// 监听退出信号，优雅注销
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		if err := c.client.Agent().ServiceDeregister(serviceID); err != nil {
			zap.S().Error("服务注销失败", zap.Error(err))
		} else {
			zap.S().Info("服务注销成功", zap.String("serviceID", serviceID))
		}
		os.Exit(0)
	}()
	return nil
}

// 服务过滤 Filtration
func (c *Consul) FilterConsul(name string) (map[string]*api.AgentService, error) {

	sprintf := fmt.Sprintf(`Service == "%s"`, name)
	zap.S().Infof("filter %s", sprintf)
	filter, err := c.client.Agent().ServicesWithFilter(sprintf)
	if err != nil {
		zap.S().Error("error filter consul service")
		return nil, err
	}
	return filter, err
}

// GetServiceFromConsul 从 Consul 获取服务地址（无健康检查）
func (c *Consul) GetServiceFromConsul(serviceName string) (string, error) {
	// 获取所有注册实例（包括不健康的）
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("从 Consul 获取服务失败: %v", err)
	}
	if len(services) == 0 {
		return "", fmt.Errorf("服务 %s 无可用实例", serviceName)
	}

	// 随机负载均衡
	rand.Seed(time.Now().UnixNano())
	instance := services[rand.Intn(len(services))]
	address := fmt.Sprintf("%s:%d", instance.Service.Address, instance.Service.Port)

	zap.S().Debugw("发现服务实例", "serviceName", serviceName, "selectedAddress", address, "totalServices", len(services))
	return address, nil
}

// consul注销
func (c *Consul) ServiceDeregister(serviceID string) error {
	if err := c.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("服务注销失败: %w", err)
	}
	zap.S().Infow("服务注销成功", "serviceID", serviceID)
	return nil
}
