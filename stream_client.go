package grpc_interceptor

import (
    "context"
    "google.golang.org/grpc"
)

type streamClientInterceptorGroup struct {
    handlers []grpc.StreamClientInterceptor
    skip map[string]struct{}
}

type StreamClientInterceptors struct {
    global []*streamClientInterceptorGroup
    part  map[string][]grpc.StreamClientInterceptor
}

func (sci *StreamClientInterceptors) UseGlobal(interceptors []grpc.StreamClientInterceptor, skipMethods ...string) {
    skip := make(map[string]struct{}, len(skipMethods))
    for _, method := range skipMethods {
        skip[method] = struct{}{}
    }
    
    sci.global = append(sci.global, &streamClientInterceptorGroup{
        handlers: interceptors,
        skip:     skip,
    })
}

func (sci *StreamClientInterceptors) UseMethod(method string, interceptors ...grpc.StreamClientInterceptor) {
    if sci.part == nil {
        sci.part = make(map[string][]grpc.StreamClientInterceptor)
        sci.part[method] = interceptors
        return
    }
    
    if _, ok := sci.part[method]; !ok {
        sci.part[method] = interceptors
        return
    }
    
    sci.part[method] = append(sci.part[method], interceptors...)
}

func (sci *StreamClientInterceptors) StreamClientInterceptor() grpc.StreamClientInterceptor {
    return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
        streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
        
        globalCount := len(sci.global)
        methodCount := len(sci.part[method])
        
        if globalCount + methodCount == 0 {
            return streamer(ctx, desc, cc, method, opts...)
        }
        
        var (
            groupI, handlerI int
            chainHandler grpc.Streamer
        )
        chainHandler = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
            opts ...grpc.CallOption) (grpc.ClientStream, error) {
            
            if groupI < globalCount {
                for {
                    group := sci.global[groupI]
                    if _, ok := group.skip[method]; !ok {
                        index := handlerI
                        handlerI++
                        if index < len(group.handlers) {
                            return group.handlers[index](ctx, desc, cc, method, chainHandler, opts...)
                        }
                        handlerI = 0
                    }
                    groupI++
                    if groupI >= globalCount {
                        break
                    }
                }
            }
            
            if handlerI < methodCount {
                special := sci.part[method]
                index := handlerI
                handlerI++
                return special[index](ctx, desc, cc, method, chainHandler, opts...)
            }
            
            return streamer(ctx, desc, cc, method, opts...)
        }
        
        return chainHandler(ctx, desc, cc, method, opts...)
    }
}
