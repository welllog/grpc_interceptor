package grpc_interceptor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestUnaryClient(t *testing.T) {
	interceptors := &UnaryClientInterceptors{}
	method := "/xx.TestService/Test"

	interceptors.Add(unaryClientInterceptorAddOne, unaryClientInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{method}, unaryClientInterceptorAddOne, unaryClientInterceptorAddOne)
	interceptors.AddWithoutMethods(nil, unaryClientInterceptorAddOne, unaryClientInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{"test"}, unaryClientInterceptorAddOne, unaryClientInterceptorAddOne)
	interceptors.AddOnMethods([]string{method}, unaryClientInterceptorAddOne, unaryClientInterceptorAddOne)

	handle := interceptors.UnaryClientInterceptor()
	err := handle(context.Background(), method, 0, 1, &grpc.ClientConn{}, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		id := ctx.Value("id").(int)
		if id != 7 {
			t.Fatal("UnaryClientInterceptors is not work")
		}
		return nil
	})
	fmt.Println(err)
}

func unaryClientInterceptorAddOne(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	val := ctx.Value("id")
	var id int
	if val != nil {
		id = val.(int) + 1
	}
	ctx = context.WithValue(ctx, "id", id)

	fmt.Println(strings.Repeat("-", id), id)
	err := invoker(ctx, srvMethod, req, reply, cc)
	fmt.Println(strings.Repeat("-", id), id)

	return err
}
