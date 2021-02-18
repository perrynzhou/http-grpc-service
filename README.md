# nginx-tester-service


### 用途
- 用于测试nginx的相关功能，比如http和grpc服务

## 功能点

- 内置一个http的服务,并发访问，有对应的统计请求处理的信息输出
- 内置一个grpc服务，和http功能一样，并发访问，统计请求处理的信息
## 使用

- 下载源代码
```
git clone https://github.com/perrynzhou/nginx-tester-service.git
```

- 编译客户端和服务端
```
cd nginx-tester-service/server && go build server.go
cd nginx-tester-service/client && go build client.go
```

- 运行
```
// server 运行会启动两个端口，一个用于http，一个用于grpc
// server -p 指定grpc端口，http端口是在grpc端口+1
// server -s 指定服务运行的ip地址
//服务端主要是处理客户端的请求2个数字的和
Usage of ./server:
  -p int
        define port for connected server (default 6666)
  -s string
        define url for connected server (default "127.0.0.1")

[perrynzhou@linuxzhou ~/data/Source/perrynzhou/nginx-tester-service/server]$ ./server  -p 6666 -s 127.0.0.1
INFO[0000] start httpService on  6667                   
INFO[0000] start grpcService on  6666  


// 客户端是请求2个数到服务端，目前仅仅支持post请求
// client -g 指定协程的并发数
// client -s 指定服务端的端口，如果是grpc请求就指定服务端的grpc端口；否则指定服务端的http端口
// client -t 指定请求类型，有2个类型，一个是grpc,另外一个是http
[perrynzhou@linuxzhou ~/data/Source/perrynzhou/nginx-tester-service/client]$ ./client  -h
Usage of ./client:
  -g int
        define number of  goroutine (default 1)
  -s string
        define url for connected server (default "127.0.0.1:6666")
  -t string
        default is grpc,option is grpc/http (default "grpc")
```