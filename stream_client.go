package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
    "strconv"
)

type StreamClientHandler func(
    ctx context.Context,
    desc *grpc.StreamDesc,
    cc *grpc.ClientConn,
    srvMethod string,
    streamer grpc.Streamer,
) (grpc.ClientStream, error)

type streamClientHandlerGroup struct {
    name string
    handlers []StreamClientHandler
    skip map[string]struct{}
}

type StreamClientInterceptors struct {
    global []*streamClientHandlerGroup
    special  map[string][]StreamClientHandler
    count int
}

func (interceptors *StreamClientInterceptors) AddGlobalHandlerGroup(groupName string, handlers []StreamClientHandler,
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
    interceptors.global = append(interceptors.global, &streamClientHandlerGroup{
        name: groupName,
        handlers: handlers,
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *StreamClientInterceptors) AddGlobalHandler(handler StreamClientHandler, skip ...string) {
    skipm := make(map[string]struct{}, len(skip))
    for _, s := range skip {
        skipm[s] = struct{}{}
    }
    interceptors.global = append(interceptors.global, &streamClientHandlerGroup{
        name: "group@" + strconv.Itoa(interceptors.count),
        handlers: []StreamClientHandler{handler},
        skip: skipm,
    })
    interceptors.count++
}

func (interceptors *StreamClientInterceptors) AddSpecialHandler(srvMethod string, handlers ...StreamClientHandler) {
    if interceptors.special == nil {
        interceptors.special = make(map[string][]StreamClientHandler)
    }
    if _, ok := interceptors.special[srvMethod]; !ok {
        interceptors.special[srvMethod] = handlers
        return
    }
    interceptors.special[srvMethod] = append(interceptors.special[srvMethod], handlers...)
}

func (interceptors *StreamClientInterceptors) StreamClientInterceptor() grpc.StreamClientInterceptor {
    return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string,
        streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
        
        globalCount := len(interceptors.global)
        specialCount := len(interceptors.special[srvMethod])
        
        if globalCount + specialCount == 0 {
            return streamer(ctx, desc, cc, srvMethod, opts...)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.Streamer
        )
        chainHandler = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, srvMethod string,
            opts ...grpc.CallOption) (grpc.ClientStream, error) {
            
            if groupI < globalCount {
                for {
                    group := interceptors.global[groupI]
                    if _, ok := group.skip[srvMethod]; !ok {
                        index := handlerI
                        handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](ctx, desc, cc, srvMethod, chainHandler)
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
                return special[index](ctx, desc, cc, srvMethod, chainHandler)
            }
            
            return streamer(ctx, desc, cc, srvMethod, opts...)
        }
        
        return chainHandler(ctx, desc, cc, srvMethod, opts...)
    }
}
