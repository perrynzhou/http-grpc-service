package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"nginx-service-tester/pb"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	srvHost = flag.String("s", "127.0.0.1", "define url for connected server")
	srvPort = flag.Int("p", 6666, "define port for connected server")
)

type CalculatorServer struct {
	Count    uint64
	Addr     string
	HttpPort int
	GrpcPort int
	stop     chan struct{}
	wg       *sync.WaitGroup
}

func NewCalculatorServer(addr string, port int) *CalculatorServer {
	return &CalculatorServer{
		Addr:     addr,
		HttpPort: port + 1,
		GrpcPort: port,
		wg:       &sync.WaitGroup{},
		stop:     make(chan struct{}),
	}
}
func (cs *CalculatorServer) Calculate(ctx context.Context, cReq *pb.CalulatorRequest) (*pb.CalulatorResponse, error) {
	x := cReq.GetX()
	y := cReq.GetY()

	resp := &pb.CalulatorResponse{
		Result: x + y,
	}
	log.Printf("Received request to calculate %d + %d =%d\n", x, y, resp.Result)
	atomic.AddUint64(&cs.Count, 1)
	return resp, nil
}

func (cs *CalculatorServer) Stop() {
	for i := 0; i < 2; i++ {
		cs.stop <- struct{}{}
	}
}
func (cs *CalculatorServer) Run() {
	cs.wg.Add(2)

	go cs.runGrpc()
	go cs.runHttp()
}
func (cs *CalculatorServer) runHttp() {
	defer cs.wg.Done()
	//http gateway
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	dialOptions := []grpc.DialOption{grpc.WithInsecure()}
	// register grpc port reflection http
	grpcServerEndpoint := fmt.Sprintf("%s:%d", cs.Addr, cs.GrpcPort)
	if err := pb.RegisterCalulatorHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, dialOptions); err != nil {
		log.Fatal("register http service failed:", err)
	}
	log.Infoln("start httpService on ", cs.HttpPort)
	go func(port int) {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			log.Errorf("start  %s:%d failed,err:%v\n", cs.Addr, cs.HttpPort, err)
		}
	}(cs.HttpPort)
	for {
		select {
		case <-cs.stop:
			log.Infof("stop http service on %s:%d success\n", cs.Addr, cs.HttpPort)
			return
		}
	}
}
func (cs *CalculatorServer) runGrpc() {
	defer cs.wg.Done()
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", cs.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen on %d,err: %v", cs.GrpcPort, err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCalulatorServer(grpcServer, cs)
	go func(srv *grpc.Server) {
		if err := srv.Serve(listen); err != nil {
			//log.Fatal("start  grpc service on %s:%d failed:%v ", s.addr, s.grpcPort, err)
		}
	}(grpcServer)
	log.Infoln("start grpcService on ", cs.GrpcPort)
	for {
		select {
		case <-cs.stop:
			grpcServer.Stop()
			log.Infof("stop grpc service on %s:%d success", cs.Addr, cs.GrpcPort)
			return
		}
	}
}

func main() {
	flag.Parse()
	s := NewCalculatorServer(*srvHost, *srvPort)
	s.Run()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()
	var count uint64
	for {
		select {
		case <-sigs:
			fmt.Printf("server %s handle %d request\n", s.Addr, s.Count)
			fmt.Printf("server stoped\n")
			s.Stop()
			return
		case <-ticker.C:
			if count != s.Count {
				fmt.Printf("server %s handle %d request\n", s.Addr, s.Count)
			}
			break
		}
	}
}
