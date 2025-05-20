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
	Shutdown()
}

type userClient struct {
	url	string
	conn	*amqp.Connection
	ch	*amqp.Channel
}

func NewUserClient(url string) (UserClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Println("Failed to connect to RabbitMQ")
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to create a channel")
		return nil, err
	}

	return &userClient{url: url, conn: conn, ch: ch}, nil
}

func (s *userClient) Shutdown() {
	if s.ch != nil { s.ch.Close() }
	if s.conn != nil { s.conn.Close() }
}

// TODO: Use custom errors
func (c *userClient) FindByUsername(username string) (*models.User, error) {
	queueName := "find_by_username"

        q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Println("Failed to create a queue")
		return nil, err
	}

	msgs, err := c.ch.Consume(
		q.Name, "",
		true, false, false, false, nil,
	)
	if err != nil {
		log.Println("Failed to register a consumer")
		return nil, err
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = c.ch.PublishWithContext(
		ctx, "", queueName, false, false,
		amqp.Publishing{
			ContentType:	"text/plain",
			CorrelationId:	corrId,
			ReplyTo:	q.Name,
			Body:		[]byte(username),
		},
	)
	if err != nil {
		errmsg := fmt.Sprintf(
			"Failed to publish request for username '%s': %w",
			username,
			err,
		)
		log.Println(errmsg)
		return nil, fmt.Errorf(errmsg)
	}

	// TODO: Fix waiting indifinitely for a response
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
// TODO: Consider modifying this function so that it takes only username and
// password as parameters.
func (c *userClient) Persist(user *models.User) error {
	queueName := "persist"

        q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Println("Failed to create a `persistReqQueue`")
		return err
	}

	msgs, err := c.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to register a consumer")
		return err
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Println("Failed to marshal the user")
		return err
	}

	err = c.ch.PublishWithContext(
		ctx, "", queueName, false, false,
		amqp.Publishing{
			ContentType:	"text/json",
			CorrelationId:	corrId,
			ReplyTo:	q.Name,
			Body:		userBytes,
		},
	)
	if err != nil {
		errmsg := fmt.Sprintf(
			"Failed to publish request to persist '%s': %w",
			user.Username,
			err,
		)
		log.Println(errmsg)
		return fmt.Errorf(errmsg)
	}

	// TODO: Fix waiting indifinitely for a response
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
