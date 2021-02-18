rm -rf *.go
#go grpc
protoc -I . \
  --go_out ./ --go_opt paths=source_relative \
  --go-grpc_out ./ --go-grpc_opt paths=source_relative calulator.proto
#gateway
protoc --grpc-gateway_out=logtostderr=true:. calulator.proto