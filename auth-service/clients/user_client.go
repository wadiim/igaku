package clients

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"igaku/auth-service/utils"
	"igaku/commons/dtos"
	"igaku/commons/models"
)

type UserClient interface {
	FindByUsername(username string) (*models.User, error)
	Persist(user *models.User) error
}

type userClient struct {
	url string
}

// TODO: Use custom errors
// TODO: Reuse connection, channel, etc. between invocations.
func (c *userClient) FindByUsername(username string) (*models.User, error) {
	queueName := "find_by_username"

	conn, err := amqp.Dial(c.url)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return nil, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to create a channel")
		return nil, err
	}
	defer ch.Close()

        q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Println("Failed to create a queue")
		return nil, err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to register a consumer")
		return nil, err
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx, "", queueName, false, false,
		amqp.Publishing{
			ContentType:	"text/plain",
			CorrelationId:	corrId,
			ReplyTo:	q.Name,
			Body:		[]byte(username),
		},
	)

	for d := range msgs {
		if corrId != d.CorrelationId {
			continue
		}

		var rpcResp dtos.RPCResponse
		if err := json.Unmarshal(d.Body, &rpcResp); err != nil {
			log.Printf("Failed to unmarshal RPC response: %w\n", err)
			return nil, fmt.Errorf(
				"Failed to unmarshal RPC response: %w", err,
			)
		}

		if rpcResp.Error != nil {
			if rpcResp.Error.Code == "NOT_FOUND" {
				return nil, fmt.Errorf(
					"User not found: %s",
					rpcResp.Error.Message,
				)
			} else if rpcResp.Error.Code == "INTERNAL" {
				return nil, fmt.Errorf(
					"User service internal error: %s",
					rpcResp.Error.Message,
				)
			} else {
				return nil, fmt.Errorf(
					"Internal error: %s",
					rpcResp.Error.Message,
				)
			}
		}

		var user models.User
		if err := json.Unmarshal(rpcResp.Data, &user); err != nil {
			return nil, fmt.Errorf("Failed to unmarshal user: %w", err)
		}

		return &user, nil
	}

	return nil, fmt.Errorf("Failed to fetch the user")
}

// TODO: Use custom errors
// TODO: Reuse connection, channel, etc. between invocations.
// TODO: Consider modifying this function so that it takes only username and
// password as parameters.
func (c *userClient) Persist(user *models.User) error {
	queueName := "persist"

	conn, err := amqp.Dial(c.url)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to create a channel")
		return err
	}
	defer ch.Close()

        q, err := ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Println("Failed to create a queue")
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to register a consumer")
		return err
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Println("Failed to marshal the user")
		return err
	}

	err = ch.PublishWithContext(
		ctx, "", queueName, false, false,
		amqp.Publishing{
			ContentType:	"text/json",
			CorrelationId:	corrId,
			ReplyTo:	q.Name,
			Body:		userBytes,
		},
	)

	for d := range msgs {
		if corrId != d.CorrelationId {
			continue
		}

		var rpcResp dtos.RPCResponse
		if err := json.Unmarshal(d.Body, &rpcResp); err != nil {
			log.Printf("Failed to unmarshal RPC response: %w\n", err)
			return fmt.Errorf(
				"Failed to unmarshal RPC response: %w", err,
			)
		}

		if rpcResp.Error != nil {
			return fmt.Errorf(
				"Failed to persist the user: %s",
				rpcResp.Error.Message,
			)
		}

		return nil
	}

	return fmt.Errorf("Failed to persist the user")
}

func NewUserClient(url string) UserClient {
	return &userClient{url: url}
}
