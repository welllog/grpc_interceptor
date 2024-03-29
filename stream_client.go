package grpc_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type streamClientInterceptorGroup struct {
	ics    []grpc.StreamClientInterceptor
	domain domain
}

type StreamClientInterceptors struct {
	s []streamClientInterceptorGroup
}

func (sci *StreamClientInterceptors) Add(interceptors ...grpc.StreamClientInterceptor) *StreamClientInterceptors {
	sci.s = append(sci.s, streamClientInterceptorGroup{
		ics:    interceptors,
		domain: newDomain(),
	})
	return sci
}

func (sci *StreamClientInterceptors) AddWithoutMethods(methods []string, interceptors ...grpc.StreamClientInterceptor) *StreamClientInterceptors {
	sci.s = append(sci.s, streamClientInterceptorGroup{
		ics:    interceptors,
		domain: newBlackDomain(methods),
	})
	return sci
}

func (sci *StreamClientInterceptors) AddOnMethods(methods []string, interceptors ...grpc.StreamClientInterceptor) *StreamClientInterceptors {
	if len(methods) > 0 {
		sci.s = append(sci.s, streamClientInterceptorGroup{
			ics:    interceptors,
			domain: newWhiteDomain(methods),
		})
	}
	return sci
}

func (sci *StreamClientInterceptors) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

		if len(sci.s) == 0 {
			return streamer(ctx, desc, cc, method, opts...)
		}

		var cursor handleCursor
		var chainHandler grpc.Streamer

		chainHandler = func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
			opts ...grpc.CallOption) (grpc.ClientStream, error) {

			for cursor.segment < len(sci.s) {
				group := sci.s[cursor.segment]
				if group.domain.isOnMethod(method) && cursor.offset < len(group.ics) {
					ic := group.ics[cursor.offset]
					cursor.offset++
					return ic(ctx, desc, cc, method, chainHandler, opts...)
				}

				cursor.offset = 0
				cursor.segment++
			}

			return streamer(ctx, desc, cc, method, opts...)
		}

		return chainHandler(ctx, desc, cc, method, opts...)
	}
}
