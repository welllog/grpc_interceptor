package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
    "strconv"
)

type UnaryClientHandler func(
    ctx context.Context,
    srvMethod string,
    req, reply interface{},
    cc *grpc.ClientConn,
    invoker grpc.UnaryInvoker,
) error

type unaryClientHandlerGroup struct {
    name string
    handlers []UnaryClientHandler
    skip map[string]struct{}
}

type UnaryClientInterceptors struct {
    global []*unaryClientHandlerGroup
    special  map[string][]UnaryClientHandler
    count int
}

func (interceptors *UnaryClientInterceptors) AddGlobalHandlerGroup(groupName string, handlers []UnaryClientHandler,
    skip ...string) {
    
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    for i := range interceptors.global {
        if interceptors.global[i].name == groupName {
            interceptors.global[i].handlers = handlers
            interceptors.global[i].skip = skipm
            return
        }
    }
    interceptors.global = append(interceptors.global, &unaryClientHandlerGroup{
        name: groupName,
        handlers: handlers,
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *UnaryClientInterceptors) AddGlobalHandler(handler UnaryClientHandler, skip ...string) {
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    interceptors.global = append(interceptors.global, &unaryClientHandlerGroup{
        name: "group@" + strconv.Itoa(interceptors.count),
        handlers: []UnaryClientHandler{handler},
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *UnaryClientInterceptors) AddSpecialHandler(srvMethod string, handlers ...UnaryClientHandler) {
    if interceptors.special == nil {
        interceptors.special = make(map[string][]UnaryClientHandler)
    }
    if _, ok := interceptors.special[srvMethod]; !ok {
        interceptors.special[srvMethod] = handlers
        return
    }
    interceptors.special[srvMethod] = append(interceptors.special[srvMethod], handlers...)
}

func (interceptors *UnaryClientInterceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, srvMethod string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        
        globalCount := len(interceptors.global)
        specialCount := len(interceptors.special[srvMethod])
        
        if globalCount + specialCount == 0 {
            return invoker(ctx, srvMethod, req, reply, cc, opts...)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.UnaryInvoker
        )
        chainHandler = func(ctx context.Context, srvMethod string, req, reply interface{},
            cc *grpc.ClientConn, opts ...grpc.CallOption) error {
            
            if groupI < globalCount {
                for {
                    group := interceptors.global[groupI]
                    if _, ok := group.skip[srvMethod]; !ok {
                        index := handlerI
                        handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](ctx, srvMethod, req, reply, cc, chainHandler)
                        }
                        handlerI = 0
                    }
                    groupI++
                    if groupI >= globalCount {
                        break
                    }
                }
            }
            
            if handlerI < specialCount {
                special := interceptors.special[srvMethod]
                index := handlerI
                handlerI++
                return special[index](ctx, srvMethod, req, reply, cc, chainHandler)
            }
            
            return invoker(ctx, srvMethod, req, reply, cc, opts...)
        }
        
        return chainHandler(ctx, srvMethod, req, reply, cc, opts...)
    }
}
