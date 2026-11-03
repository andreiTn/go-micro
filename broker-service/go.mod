module github.com/andreiTn/go-micro/broker-service

go 1.27.3

require (
	github.com/andreiTn/go-micro/contracts v0.0.0
	github.com/andreiTn/go-micro/pkg v0.0.0
	github.com/go-chi/chi/v5 v5.5.0
	github.com/rabbitmq/amqp091-go v1.16.0
	google.golang.org/grpc v1.85.0
)

replace (
	github.com/andreiTn/go-micro/contracts => ../contracts
	github.com/andreiTn/go-micro/pkg => ../pkg
)

require (
	github.com/go-chi/cors v1.2.2 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260817212433-ac3dfec99bb1 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
