package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
)

type unaryClientInterceptorGroup struct {
    handlers []grpc.UnaryClientInterceptor
    skip map[string]struct{}
}

type UnaryClientInterceptors struct {
    global []*unaryClientInterceptorGroup
    ggCount int
    part  map[string][]grpc.UnaryClientInterceptor
}

func (uci *UnaryClientInterceptors) UseGlobal(interceptors []grpc.UnaryClientInterceptor, skipMethods ...string) {
    skip := make(map[string]struct{}, len(skipMethods))
    for _, method := range skipMethods {
        skip[method] = struct{}{}
    }
    
    uci.global = append(uci.global, &unaryClientInterceptorGroup{
        handlers: interceptors,
        skip:     skip,
    })
    uci.ggCount++
}

func (uci *UnaryClientInterceptors) UseMethod(method string, interceptors ...grpc.UnaryClientInterceptor) {
    if uci.part == nil {
        uci.part = make(map[string][]grpc.UnaryClientInterceptor)
        uci.part[method] = interceptors
        return
    }
    
    if _, ok := uci.part[method]; !ok {
        uci.part[method] = interceptors
        return
    }
    
    uci.part[method] = append(uci.part[method], interceptors...)
}

func (uci *UnaryClientInterceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        
        mCount := len(uci.part[method])
        
        if uci.ggCount + mCount == 0 {
            return invoker(ctx, method, req, reply, cc, opts...)
        }
        
        curI := handlerCurI{mCount: mCount}
        var chainHandler grpc.UnaryInvoker
        
        chainHandler = func(ctx context.Context, method string, req, reply interface{},
            cc *grpc.ClientConn, opts ...grpc.CallOption) error {
            
            if curI.groupI < uci.ggCount {
                for {
                    group := uci.global[curI.groupI]
                    if _, ok := group.skip[method]; !ok {
                        index := curI.handlerI
                        curI.handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](ctx, method, req, reply, cc, chainHandler, opts...)
                        }
                        curI.handlerI = 0
                    }
                    curI.groupI++
                    if curI.groupI >= uci.ggCount {
                        break
                    }
                }
            }
            
            if curI.handlerI < curI.mCount {
                special := uci.part[method]
                index := curI.handlerI
                curI.handlerI++
                return special[index](ctx, method, req, reply, cc, chainHandler, opts...)
            }
            
            return invoker(ctx, method, req, reply, cc, opts...)
        }
        
        return chainHandler(ctx, method, req, reply, cc, opts...)
    }
}
