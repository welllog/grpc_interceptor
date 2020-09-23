package grpc_interceptor

import (
    "context"
    "errors"
    "fmt"
    "google.golang.org/grpc"
    "testing"
)

func TestUnaryServer(t *testing.T) {
    unaryServMd := &UnaryServerInterceptors{}
    unaryServMd.AddGlobalHandlerGroup("test1", []UnaryServerHandler{
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 1")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 1")
            return
        },
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 2")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 2")
            return
        },
    })
    
    unaryServMd.AddSpecialHandler("/test_package.TestService/Test", func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
        fmt.Println("exec ---- 9")
        i, err = handler(ctx, req)
        fmt.Println("complete ---- 9")
        return
    }, func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
        fmt.Println("exec ---- 10")
        i, err = handler(ctx, req)
        fmt.Println("complete ---- 10")
        return
    })
    
    unaryServMd.AddSpecialHandler("/test_package.TestService/Test", func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
        fmt.Println("exec ---- 11")
        i, err = handler(ctx, req)
        fmt.Println("complete ---- 11")
        return
    })
    
    unaryServMd.AddGlobalHandlerGroup("test2", []UnaryServerHandler{
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 3")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 3")
            return
        },
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 4")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 4")
            return
        },
    }, "/test_package.TestService/Test1")
    
    unaryServMd.AddGlobalHandlerGroup("test3", []UnaryServerHandler{
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 5")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 5")
            return
        },
        func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
            fmt.Println("exec ---- 6")
            i, err = handler(ctx, req)
            fmt.Println("complete ---- 6")
            return
        },
    }, "/test_package.TestService/Test")
    
    unaryServMd.AddGlobalHandler(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
        fmt.Println("exec ---- 7")
        if val, ok := req.(int); ok && val == 1 {
            return nil, errors.New("val is 1")
        }
        i, err = handler(ctx, req)
        fmt.Println("complete ---- 7")
        return
    }, "/test_package.TestService/Test1")
    
    unaryServMd.AddGlobalHandler(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (i interface{}, err error) {
        fmt.Println("exec ---- 8")
        i, err = handler(ctx, req)
        fmt.Println("complete ---- 8")
        return
    }, "/test_package.TestService/Test1")
    
    
    funcs := unaryServMd.UnaryServerInterceptor()
    rsp, err := funcs(context.Background(), 2, &grpc.UnaryServerInfo{FullMethod: "/test_package.TestService/Test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
        fmt.Println("in the end echo: ", req)
        return req, nil
    })
    fmt.Println("-------------")
    fmt.Println(rsp)
    fmt.Println(err)
    
    for _, v := range unaryServMd.global {
        fmt.Println(v.name)
    }
    fmt.Println(unaryServMd.count)
}
