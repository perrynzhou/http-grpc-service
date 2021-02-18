package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"nginx-service-tester/pb"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

var (
	numGoRoutine = flag.Int("g", 1, "define number of  goroutine")
	srvUrl       = flag.String("s", "127.0.0.1:6666", "define url for connected server")
	serviceType  = flag.String("t", "grpc", "default is grpc,option is grpc/http")
)

func runGRPC(addr string, opSuccess, opFailed *uint64) {

	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer cc.Close()
	first := math.MaxInt32 / (rand.Int31n(4096) + 1)
	second := math.MaxInt32 / (rand.Int31n(2048) + 1)

	req := &pb.CalulatorRequest{
		X: first,
		Y: second,
	}
	calulatorClient := pb.NewCalulatorClient(cc)
	_, err = calulatorClient.Calculate(context.Background(), req)
	if err != nil {
		atomic.AddUint64(opFailed, 1)
		return
	}
	atomic.AddUint64(opSuccess, 1)
}
func runHTTP(addr string, opSuccess, opFailed *uint64) {
	first := math.MaxInt32 / (rand.Int31n(4096) + 1)
	second := math.MaxInt32 / (rand.Int31n(2048) + 1)
	postUrl := fmt.Sprintf("http://%s/%s", addr, uuid.NewV4().String())
	reqData := pb.CalulatorRequest{
		X: first,
		Y: second,
	}
	fmt.Println("post url:", postUrl)
	value, _ := json.Marshal(reqData)
	fmt.Println("   request  data:",string(value))

	body := strings.NewReader(string(value))

	req, err := http.NewRequest("POST", postUrl, body)
	req.Header.Set("Content-Type", "application/json")
	clt := http.Client{}
	resp, err := clt.Do(req)
	if err != nil {
		atomic.AddUint64(opFailed, 1)
		return
	}
	defer resp.Body.Close()
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		atomic.AddUint64(opFailed, 1)
		return
	}
	fmt.Println("   response data:",string(res))

	atomic.AddUint64(opSuccess, 1)
}
func main() {

	flag.Parse()
	wg := &sync.WaitGroup{}
	wg.Add(*numGoRoutine)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	var opSuccess uint64
	var opFailed uint64
	ticker := time.NewTicker(time.Second * time.Duration(3))
	interval := time.NewTicker(time.Millisecond * time.Duration(5))
	defer ticker.Stop()
	defer interval.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Printf("success :%d,	failed:%d\n", opSuccess, opFailed)
			break
		case <-sigs:
			fmt.Printf("success :%d,	failed:%d\n", opSuccess, opFailed)
			fmt.Printf("stop nginx-service-tester\n")
			return
		case <-interval.C:
			if *serviceType == "grpc" {
				runGRPC(*srvUrl, &opSuccess, &opFailed)

			} else if *serviceType == "http" {
				runHTTP(*srvUrl, &opSuccess, &opFailed)
			}
			break
		}
	}

}
