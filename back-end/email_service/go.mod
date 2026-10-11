module hoshiBmaTchi/back-end/email_service

go 1.25.3

require (
	github.com/streadway/amqp v1.1.0
	google.golang.org/grpc v1.65.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	hoshiBmaTchi/back-end/proto v0.0.0
)

replace hoshiBmaTchi/back-end/proto => ../proto

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)
