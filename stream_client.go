package grpc_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type streamClientInterceptorGroup struct {
	handlers []grpc.StreamClientInterceptor
	skip     map[string]struct{}
}

type StreamClientInterceptors struct {
	global  []*streamClientInterceptorGroup
	ggCount int
	part    map[string][]grpc.StreamClientInterceptor
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
	sci.ggCount++
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

		mCount := len(sci.part[method])

		if sci.ggCount+mCount == 0 {
			return streamer(ctx, desc, cc, method, opts...)
		}

		curI := handlerCurI{mCount: mCount}
		var chainHandler grpc.Streamer

		chainHandler = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
			opts ...grpc.CallOption) (grpc.ClientStream, error) {

			if curI.groupI < sci.ggCount {
				for {
					group := sci.global[curI.groupI]
					if _, ok := group.skip[method]; !ok {
						index := curI.handlerI
						curI.handlerI++
						if index < len(group.handlers) {
							return group.handlers[index](ctx, desc, cc, method, chainHandler, opts...)
						}
						curI.handlerI = 0
					}
					curI.groupI++
					if curI.groupI >= sci.ggCount {
						break
					}
				}
			}

			if curI.handlerI < curI.mCount {
				special := sci.part[method]
				index := curI.handlerI
				curI.handlerI++
				return special[index](ctx, desc, cc, method, chainHandler, opts...)
			}

			return streamer(ctx, desc, cc, method, opts...)
		}

		return chainHandler(ctx, desc, cc, method, opts...)
	}
}
