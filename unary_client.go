package grpc_interceptor

import (
	"context"

	"google.golang.org/grpc"
)

type unaryClientInterceptorGroup struct {
	ics    []grpc.UnaryClientInterceptor
	domain domain
}

type UnaryClientInterceptors struct {
	s []unaryClientInterceptorGroup
}

func (uci *UnaryClientInterceptors) Add(interceptors ...grpc.UnaryClientInterceptor) *UnaryClientInterceptors {
	uci.s = append(uci.s, unaryClientInterceptorGroup{
		ics:    interceptors,
		domain: newDomain(),
	})
	return uci
}

func (uci *UnaryClientInterceptors) AddWithoutMethods(methods []string, interceptors ...grpc.UnaryClientInterceptor) *UnaryClientInterceptors {
	uci.s = append(uci.s, unaryClientInterceptorGroup{
		ics:    interceptors,
		domain: newBlackDomain(methods),
	})
	return uci
}

func (uci *UnaryClientInterceptors) AddOnMethod(method string, interceptors ...grpc.UnaryClientInterceptor) *UnaryClientInterceptors {
	uci.s = append(uci.s, unaryClientInterceptorGroup{
		ics:    interceptors,
		domain: newSpecificDomain(method),
	})
	return uci
}

func (uci *UnaryClientInterceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		var cursor handleCursor
		for i := range uci.s {
			if uci.s[i].domain.isOnMethod(method) {
				cursor.ids = append(cursor.ids, i)
			}
		}

		if len(cursor.ids) == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var chainHandler grpc.UnaryInvoker

		chainHandler = func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, opts ...grpc.CallOption) error {

			for cursor.segment < len(cursor.ids) {
				group := uci.s[cursor.ids[cursor.segment]]
				if cursor.offset < len(group.ics) {
					ic := group.ics[cursor.offset]
					cursor.offset++
					return ic(ctx, method, req, reply, cc, chainHandler, opts...)
				}
				cursor.offset = 0
				cursor.segment++
			}

			return invoker(ctx, method, req, reply, cc, opts...)
		}

		return chainHandler(ctx, method, req, reply, cc, opts...)
	}
}
