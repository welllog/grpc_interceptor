# grpc_interceptor

grpc interceptor

## usage

##### server

```go
func ServerDemo1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, 
handler grpc.UnaryHandler) (interface{}, error) {

    log.Printf("before handling. Info: %+v", info)
    resp, err := handler(ctx, req)
    log.Printf("after handling. resp: %+v", resp)
    return rsp, err
}

func ServerDemo2(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
handler grpc.StreamHandler) error {

    log.Printf("before handling. Info: %+v", info)
    err := handler(srv, ss)
    log.Printf("after handling. err: %v", err)
    return err
}

...



// Load interceptor ServerDemo1, ignore on /package.Service/Do1 method
// Same set of interceptors using a common white list
unaryMd := &grpc_interceptor.UnaryServerInterceptors{}
unaryMd.UseGlobal([]grpc.UnaryServerInterceptor{
    ServerDemo1,
    ...
}, "/package.Service/Do1")

// Load interceptor ServerDemo2, ignore on /package.Service/Do2 method
streamMd := &grpc_interceptor.StreamServerInterceptors{}
streamMd.UseGlobal([]grpc.StreamServerInterceptor{ServerDemo2}, "/package.Service/Do2")


s := grpc.NewServer(
        grpc.StreamInterceptor(streamMd.StreamServerInterceptor()),
        grpc.UnaryInterceptor(unaryMd.UnaryServerInterceptor()),
)
```

##### client
```go
func clientDemo1(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	log.Printf("before invoker. method: %+v, request:%+v", method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("after invoker. reply: %+v", reply)
	return err
}

func clientDemo2(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, 
streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	log.Printf("before invoker. method: %+v, StreamDesc:%+v", method, desc)
	clientStream, err := streamer(ctx, desc, cc, method, opts...)
	log.Printf("before invoker. method: %+v", method)
	return clientStream, err
}

...


unaryMd := &grpc_interceptor.UnaryClientInterceptors{}
unaryMd.UseGlobal([]grpc.UnaryClientInterceptor{
    clientDemo1,
    ...
})

// Load the interceptor clientDemo2 for /package.Service/Do2 methods only
streamMd := &grpc_interceptor.StreamClientInterceptors{}
streamMd.UseMethod("/package.Service/Do2", clientDemo2)

grpc.Dial(*address, grpc.WithInsecure(), 
    grpc.WithUnaryInterceptor(unaryMd.UnaryClientInterceptor()),
	grpc.WithStreamInterceptor(streamMd.StreamClientInterceptor()),
)

```
