package grpc_interceptor

import (
    "context"
    "errors"
    "fmt"
    "google.golang.org/grpc"
    "testing"
)

func TestUnaryClient(t *testing.T) {
    unaryClientMd := &UnaryClientInterceptors{}
    
    unaryClientMd.AddSpecialHandler("/xx.TestService/Test", func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
        fmt.Println("special --- 0")
        if val, ok := req.(int); ok {
            if val == 1 {
                return errors.New("request param error")
            }
        }
        err := invoker(ctx, srvMethod, req, reply, cc)
        fmt.Println("special --- 0 --- end")
        return err
    })
    
    unaryClientMd.AddSpecialHandler("/xx.TestService/Test", func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
        fmt.Println("special --- 1")
        err := invoker(ctx, srvMethod, req, reply, cc)
        fmt.Println("special --- 1 --- end")
        return err
    })
    
    unaryClientMd.AddGlobalHandlerGroup("test1", []UnaryClientHandler{
        func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
            fmt.Println("global --- -2")
            err := invoker(ctx, srvMethod, req, reply, cc)
            fmt.Println("global --- -2 --- end")
            return err
        },
        func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
            fmt.Println("global --- -1")
            err := invoker(ctx, srvMethod, req, reply, cc)
            fmt.Println("global --- -1 --- end")
            return err
        },
    }, "/xx.TestService/Test")
    
    unaryClientMd.AddGlobalHandlerGroup("test2", []UnaryClientHandler{
        func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
            fmt.Println("global --- 0")
            err := invoker(ctx, srvMethod, req, reply, cc)
            fmt.Println("global --- 0 --- end")
            return err
        },
        func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
            fmt.Println("global --- 1")
            err := invoker(ctx, srvMethod, req, reply, cc)
            fmt.Println("global --- 1 --- end")
            return err
        },
    })
    
    unaryClientMd.AddGlobalHandler(func(ctx context.Context, srvMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker) error {
        fmt.Println("global --- 2")
        err := invoker(ctx, srvMethod, req, reply, cc)
        fmt.Println("global --- 2 --- end")
        return err
    })
    
    funcs := unaryClientMd.UnaryClientInterceptor()
    err := funcs(context.Background(), "/xx.TestService/Test", 2, 1, &grpc.ClientConn{}, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
        fmt.Println("send request ...")
        return nil
    })
    fmt.Println(err)
}
