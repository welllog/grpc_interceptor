package grpc_interceptor

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
)

func TestUnaryServer(t *testing.T) {
	unaryServMd := &UnaryServerInterceptors{}
	unaryServMd.UseGlobal([]grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
			fmt.Println("global start")
			return handler(ctx, req)
		},
	})

	unaryServMd.UseMethod("/test_package.TestService/Test", func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
		fmt.Println("part start")
		return handler(ctx, req)
	}, unaryServerPrint)

	unaryServMd.UseMethod("/test_package.TestService/Test", unaryServerPrint)

	unaryServMd.UseGlobal([]grpc.UnaryServerInterceptor{
		unaryServerPrint,
		unaryServerPrint,
	}, "/test_package.TestService/Skip")

	unaryServMd.UseGlobal([]grpc.UnaryServerInterceptor{
		unaryServerPrint,
		unaryServerPrint,
	}, "/test_package.TestService/Test")

	funcs := unaryServMd.UnaryServerInterceptor()
	rsp, err := funcs(context.Background(), 0, &grpc.UnaryServerInfo{FullMethod: "/test_package.TestService/Test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		fmt.Println("handle the request: ", req)
		return req, nil
	})
	fmt.Println("-------------")
	fmt.Println(rsp)
	fmt.Println(err)
}

func unaryServerPrint(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
	fmt.Println("exec ---- ", req)
	val, ok := req.(int)
	if ok {
		val++
	}
	i, err = handler(ctx, val)
	fmt.Println("complete ---- ", req)
	return
}
