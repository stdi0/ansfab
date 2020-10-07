package service

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sync"

	"api.local/config"
	pb "api.local/proto"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type ApiService struct {
	router       *gin.Engine
	amqpChannel  *amqp.Channel
	tokenService pb.TokenServiceClient
	config       *config.ApiServiceConfig
}

func NewApiService(cfg *config.ApiServiceConfig) *ApiService {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn2, err := grpc.Dial(fmt.Sprintf("%s:%d", cfg.FabricServiceHost, cfg.FabricServicePort), opts...)
	if err != nil {
		handleError(err, "Failed to dial TokenService")
	}

	tokenService := pb.NewTokenServiceClient(conn2)

	r := gin.Default()

	return &ApiService{
		router:       r,
		amqpChannel:  nil,
		tokenService: tokenService,
		config:       cfg,
	}
}

func (a *ApiService) Start() {
	var wg sync.WaitGroup

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("%s: %s", "Failed to connect to RabbitMQ", err)
	} else {
		defer conn.Close()
		ch, err := conn.Channel()
		handleError(err, "Failed to open a channel")
		defer ch.Close()

		_, err = ch.QueueDeclare("add", true, false, false, false, nil)
		handleError(err, "Failed to declare a queue")

		err = ch.Qos(1, 0, false)
		handleError(err, "Failed to configure QoS")

		a.amqpChannel = ch
	}

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{a.config.WebDomain}
	a.router.Use(cors.New(config))
	//a.router.Use(cors.Default())

	InitializeRoutes(a)
	listenAddr := fmt.Sprintf("0.0.0.0:%d", a.config.WebListenPort)
	wg.Add(1)
	go a.router.Run(listenAddr)

	wg.Wait()
}
