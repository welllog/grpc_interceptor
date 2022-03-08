package grpc_interceptor

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"
)

func TestStreamClient(t *testing.T) {
	streamClientMd := &StreamClientInterceptors{}

	streamClientMd.UseMethod("/xx.TestService/TestStream", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		fmt.Println("part --- 0")
		stream, err = streamer(ctx, desc, cc, srvMethod, opts...)
		fmt.Println("part --- 0 --- end")
		return
	})

	streamClientMd.UseGlobal([]grpc.StreamClientInterceptor{
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
			fmt.Println("global --- 0")
			stream, err = streamer(ctx, desc, cc, srvMethod)
			fmt.Println("global --- 0 --- end")
			return
		},
	})

	streamClientMd.UseGlobal([]grpc.StreamClientInterceptor{
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
			fmt.Println("global --- -1")
			stream, err = streamer(ctx, desc, cc, srvMethod)
			fmt.Println("global --- -1 --- end")
			return
		},
	}, "/xx.TestService/TestStream")

	funcs := streamClientMd.StreamClientInterceptor()
	_, err := funcs(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "/xx.TestService/TestStream", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		fmt.Println("stream send ...")
		return nil, nil
	})
	fmt.Println(err)
}
