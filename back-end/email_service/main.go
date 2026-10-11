package main

import (
	"hoshiBmaTchi/back-end/email_service/internal/adapter"
	"hoshiBmaTchi/back-end/email_service/internal/core/service"
	emailpb "hoshiBmaTchi/back-end/proto"
	"log"
	"net"
	"os"
	"google.golang.org/grpc"
)

func main() {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	emailService := service.NewEmailService()
	if emailService == nil {
		log.Fatalf("Failed to create EmailService, check RabbitMQ/SMTP config")
	}

	emailGRPCServer := adapter.NewGRPCServer(emailService)

	s := grpc.NewServer()
	emailpb.RegisterEmailServiceServer(s, emailGRPCServer)


	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
	emailService.EmailWorker()

	forever := make(chan bool)
	<-forever
}