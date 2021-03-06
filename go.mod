module github.com/apigee/apigee-remote-service-envoy

go 1.13

replace github.com/apigee/apigee-remote-service-golib => github.com/theganyo/apigee-remote-service-golib v0.0.0-20200413223918-ced0c794923a

// replace github.com/apigee/apigee-remote-service-golib => ../apigee-remote-service-golib

replace github.com/apigee/apigee-remote-service-envoy => ./

require (
	github.com/apigee/apigee-remote-service-golib v0.0.0-00010101000000-000000000000
	github.com/envoyproxy/go-control-plane v0.9.4
	github.com/gogo/googleapis v1.3.2
	github.com/golang/protobuf v1.3.3
	github.com/hashicorp/go-multierror v1.1.0
	github.com/spf13/cobra v1.0.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	google.golang.org/genproto v0.0.0-20200225123651-fc8f55426688
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)
