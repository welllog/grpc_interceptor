package grpc_interceptor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestUnaryServer(t *testing.T) {
	interceptors := &UnaryServerInterceptors{}
	method := "/test_package.TestService/Test"

	interceptors.Add(unaryServerInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{method}, unaryServerInterceptorAddOne, unaryServerInterceptorAddOne)
	interceptors.AddOnMethods([]string{method}, unaryServerInterceptorAddOne, unaryServerInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{}, unaryServerInterceptorAddOne, unaryServerInterceptorAddOne)

	handle := interceptors.UnaryServerInterceptor()
	_, _ = handle(context.Background(), 0, &grpc.UnaryServerInfo{FullMethod: "/test_package.TestService/Test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		id := ctx.Value("id").(int)
		if id != 4 {
			t.Fatal("UnaryServerInterceptors is not work")
		}
		return req, nil
	})
}

func unaryServerInterceptorAddOne(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
	val := ctx.Value("id")
	var id int
	if val != nil {
		id = val.(int) + 1
	}
	ctx = context.WithValue(ctx, "id", id)

	fmt.Println(strings.Repeat("-", id), id)
	i, err = handler(ctx, req)
	fmt.Println(strings.Repeat("-", id), id)

	return i, err
}
