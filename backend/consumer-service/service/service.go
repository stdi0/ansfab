package service

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"consumer.local/config"
	pb "consumer.local/proto"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/streadway/amqp"
)

const (
	ServiceName       = "consumer-service"
	GrpcCallTimeoutMs = 15000
)

type ConsumerService struct {
	tokenService pb.TokenServiceClient
	config       *config.ConsumerServiceConfig
}

func NewConsumerService(cfg *config.ConsumerServiceConfig) *ConsumerService {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.FabricServiceHost, cfg.FabricServicePort), opts...)
	if err != nil {
		handleError(err, "Failed to dial TokenService")
	}

	tokenService := pb.NewTokenServiceClient(conn)

	return &ConsumerService{
		tokenService: tokenService,
		config:       cfg,
	}
}

func registerServiceWithConsul() {
	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = ServiceName   //replace with service id
	registration.Name = ServiceName //replace with service name
	address := hostname()
	registration.Address = address
	port, err := strconv.Atoi(port()[1:len(port())])
	if err != nil {
		log.Fatalln(err)
	}
	registration.Port = port
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%v/healthcheck",
		address, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	consul.Agent().ServiceRegister(registration)
}

func port() string {
	p := os.Getenv("SERVICE_PORT")
	if len(strings.TrimSpace(p)) == 0 {
		return ":9994"
	}
	return fmt.Sprintf(":%s", p)
}

func hostname() string {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	return hn
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s service is good", ServiceName)
}

func (a *ConsumerService) Start() {
	//registerServiceWithConsul()
	//http.HandleFunc("/healthcheck", healthcheck)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("%s: %s", "Failed to connect to RabbitMQ", err)
		return
	}

	defer conn.Close()
	ch, err := conn.Channel()
	handleError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(a.config.QueueName, true, false, false, false, nil)
	handleError(err, "Failed to declare a queue")

	err = ch.Qos(1, 0, false)
	handleError(err, "Failed to configure QoS")

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	handleError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			token := &Token{}

			err := json.Unmarshal(d.Body, token)
			if err != nil {
				log.Printf("Error decoding JSON: %s", err)
			}

			if token != nil {
				pbToken := &pb.Token{
					Id:         token.Id,
					ObjectType: token.ObjectType,
					Payload:    token.Payload,
				}

				clientDeadline := time.Now().Add(time.Duration(GrpcCallTimeoutMs) * time.Millisecond)
				ctx, _ := context.WithDeadline(context.Background(), clientDeadline)

				_, err := a.tokenService.Create(ctx, pbToken)
				handleError(err, "Failed to create token")
			}

			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message: %s", err)
			} else {
				log.Printf("Acknowledged message")
			}
		}
	}()

	//http.ListenAndServe(port(), nil)

	<-forever
}
