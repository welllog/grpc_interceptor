package grpc_interceptor

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc"
)

func TestStreamClient(t *testing.T) {
	interceptors := &StreamClientInterceptors{}
	method := "/xx.TestService/TestStream"

	interceptors.AddOnMethod(method, streamClientInterceptorAddOne)
	interceptors.Add(streamClientInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{method}, streamClientInterceptorAddOne, streamClientInterceptorAddOne)
	interceptors.AddWithoutMethods(nil, streamClientInterceptorAddOne)
	interceptors.AddWithoutMethods([]string{"test"}, streamClientInterceptorAddOne)

	handle := interceptors.StreamClientInterceptor()
	_, _ = handle(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, method, func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		id := ctx.Value("id").(int)
		if id != 3 {
			t.Fatal("StreamClientInterceptors is not work")
		}

		return nil, nil
	})
}

func streamClientInterceptorAddOne(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
	val := ctx.Value("id")
	var id int
	if val != nil {
		id = val.(int) + 1
	}
	ctx = context.WithValue(ctx, "id", id)

	fmt.Println(strings.Repeat("-", id), id)
	stream, err = streamer(ctx, desc, cc, srvMethod, opts...)
	fmt.Println(strings.Repeat("-", id), id)

	return stream, err
}
