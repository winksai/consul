# Consul 服务治理模块

![Consul Logo](https://www.consul.io/_next/static/media/consul-community_on-light.e8e68a5f.svg)

> 基于 HashiCorp Consul 的轻量级服务注册与发现模块  
> 专为 Go 微服务架构设计 | 生产级可用 | 零外部依赖

## 特性亮点

- 🚀 **即插即用** - 3 行代码完成服务注册/发现
- ⚡ **高性能** - 基于官方 API 客户端优化
- 🔒 **可靠性** - 内置优雅下线处理
- ⚖️ **负载均衡** - 随机轮询算法开箱即用
- 📊 **可观测性** - 深度集成 Zap 日志

## 快速入门

### 安装
```bash
go get github.com/hashicorp/consul/api

基础用法
        // 初始化客户端
        consul, err := NewConsul("consul.example.com", 8500)
        if err != nil {
            log.Fatal(err)
        }
        
        // 注册Web服务
        err = consul.RegisterConsul(
         "web-api",                // 服务名称
         "192.168.1.100",          // 服务 IP
          8080,                     // 服务端口
          []string{"http", "v1.2"}, // 标签
         )
         if err != nil {
           log.Fatal(err)
         }
        
        // 发现用户服务
        userServiceAddr, err := consul.GetServiceFromConsul("user-service")

整体架构
[客户端] → [API 网关 (Gin HTTP)] → [Consul 服务发现] → [gRPC 微服务] → [返回结果]