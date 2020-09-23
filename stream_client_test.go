package grpc_interceptor

import (
    "context"
    "fmt"
    "google.golang.org/grpc"
    "testing"
)

func TestStreamClient(t *testing.T) {
    streamClientMd := &StreamClientInterceptors{}
    
    streamClientMd.AddSpecialHandler("xx.TestService/TestStream", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer) (stream grpc.ClientStream, err error) {
        fmt.Println("special --- 0")
        stream, err = streamer(ctx, desc, cc, srvMethod)
        fmt.Println("special --- 0 --- end")
        return
    })
    
    streamClientMd.AddGlobalHandlerGroup("test1", []StreamClientHandler{
        func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer) (stream grpc.ClientStream, err error) {
            fmt.Println("global --- 0")
            stream, err = streamer(ctx, desc, cc, srvMethod)
            fmt.Println("global --- 0 --- end")
            return
        },
    })
    
    streamClientMd.AddGlobalHandlerGroup("test2", []StreamClientHandler{
        func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer) (stream grpc.ClientStream, err error) {
            fmt.Println("global --- -1")
            stream, err = streamer(ctx, desc, cc, srvMethod)
            fmt.Println("global --- -1 --- end")
            return
        },
    }, "xx.TestService/TestStream")
    
    streamClientMd.AddGlobalHandler(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string, streamer grpc.Streamer) (stream grpc.ClientStream, err error) {
        fmt.Println("global --- 1")
        stream, err = streamer(ctx, desc, cc, srvMethod)
        fmt.Println("global --- 1 --- end")
        return
    })
    
    funcs := streamClientMd.StreamClientInterceptor()
    _, err := funcs(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "xx.TestService/TestStream", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
        fmt.Println("stream send ...")
        return nil, nil
    })
    fmt.Println(err)
}
