package service

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"

	pb "api.local/proto"

	"github.com/gin-gonic/gin"
)

func (a *ApiService) List(c *gin.Context) {
	r, err := a.tokenService.List(context.Background(), &pb.Empty{})
	if err != nil {
		c.JSON(500, err)
		return
	}

	c.JSON(200, r)
}

func (a *ApiService) Create(c *gin.Context) {
	if a.amqpChannel == nil {
		c.JSON(500, "Rabbit queue is not initialized")
		return
	}

	var token Token

	err := c.BindJSON(&token)
	if err != nil {
		c.JSON(500, err)
		return
	}

	body, err := json.Marshal(token)
	if err != nil {
		handleError(err, "Error encoding JSON")
	}

	err = a.amqpChannel.Publish("", "add", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})

	handleError(err, "Failed to publish a message")

	c.Status(201)
}

func (a *ApiService) Update(c *gin.Context) {
	var token Token

	err := c.BindJSON(&token)
	if err != nil {
		c.JSON(500, err)
		return
	}

	pbToken := &pb.Token{
		Id:         token.Id,
		ObjectType: token.ObjectType,
		Payload:    token.Payload,
	}

	r, err := a.tokenService.Update(context.Background(), pbToken)
	if err != nil {
		c.JSON(500, err)
		return
	}

	c.JSON(200, r)
}
