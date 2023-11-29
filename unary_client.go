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

func (uci *UnaryClientInterceptors) AddOnMethods(methods []string, interceptors ...grpc.UnaryClientInterceptor) *UnaryClientInterceptors {
	if len(methods) > 0 {
		uci.s = append(uci.s, unaryClientInterceptorGroup{
			ics:    interceptors,
			domain: newWhiteDomain(methods),
		})
	}
	return uci
}

func (uci *UnaryClientInterceptors) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		if len(uci.s) == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var cursor handleCursor
		var chainHandler grpc.UnaryInvoker

		chainHandler = func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, opts ...grpc.CallOption) error {

			for cursor.segment < len(uci.s) {
				group := uci.s[cursor.segment]
				if group.domain.isOnMethod(method) && cursor.offset < len(group.ics) {
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
