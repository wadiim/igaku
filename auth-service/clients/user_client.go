package clients

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"igaku/auth-service/errors"
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
		log.Printf("Failed to create a queue: %w", err)
		return nil, &errors.InternalError{}
	}

	msgs, err := c.ch.Consume(
		q.Name, "",
		true, false, false, false, nil,
	)
	if err != nil {
		log.Printf("Failed to register a consumer: %w", err)
		return nil, &errors.InternalError{}
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
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
			"[%s] Failed to publish request for username '%s': %w",
			corrId,
			username,
			err,
		)
		log.Println(errmsg)
		return nil, &errors.InternalError{}
	}

	// TODO: Fix waiting indifinitely for a response
	for d := range msgs {
		if corrId != d.CorrelationId {
			continue
		}

		var rpcResp dtos.RPCResponse
		if err := json.Unmarshal(d.Body, &rpcResp); err != nil {
			errmsg := fmt.Sprintf(
				"[%s] Failed to unmarshal RPC response: %w",
				corrId,
				err,
			)
			log.Println(errmsg)
			return nil, &errors.InternalError{}
		}

		if rpcResp.Error != nil {
			if rpcResp.Error.Code == "NOT_FOUND" {
				errmsg := fmt.Sprintf(
					"[%s] User not found: %s",
					corrId,
					rpcResp.Error.Message,
				)
				log.Println(errmsg)
				return nil, fmt.Errorf(errmsg)
			} else if rpcResp.Error.Code == "INTERNAL" {
				errmsg := fmt.Sprintf(
					"[%s] User service internal error: %s",
					corrId,
					rpcResp.Error.Message,
				)
				log.Println(errmsg)
				return nil, &errors.InternalError{}
			} else {
				errmsg := fmt.Sprintf(
					"[%s] Internal error: %s",
					corrId,
					rpcResp.Error.Message,
				)
				log.Println(errmsg)
				return nil, &errors.InternalError{}
			}
		}

		var user models.User
		if err := json.Unmarshal(rpcResp.Data, &user); err != nil {
				errmsg := fmt.Sprintf(
					"[%s] Internal error: %w", corrId, err,
				)
				log.Println(errmsg)
			return nil, &errors.InternalError{}
		}

		return &user, nil
	}

	errmsg := fmt.Sprintf("[%s] Failed to fetch the user", corrId)
	log.Println(errmsg)
	return nil, &errors.InternalError{}
}

// TODO: Use custom errors
// TODO: Consider modifying this function so that it takes only username and
// password as parameters.
func (c *userClient) Persist(user *models.User) error {
	queueName := "persist"

        q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
	if err != nil {
		log.Println("Failed to create a `persistReqQueue`: %w", err)
		return &errors.InternalError{}
	}

	msgs, err := c.ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to register a consumer: %w", err)
		return &errors.InternalError{}
	}

	corrId := utils.RandString(16)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("[%s] Failed to marshal the user", corrId)
		return &errors.InternalError{}
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
			"[%s] Failed to publish request to persist '%s': %w",
			corrId,
			user.Username,
			err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	// TODO: Fix waiting indifinitely for a response
	for d := range msgs {
		if corrId != d.CorrelationId {
			continue
		}

		var rpcResp dtos.RPCResponse
		if err := json.Unmarshal(d.Body, &rpcResp); err != nil {
			errmsg := fmt.Sprintf(
				"[%s] Failed to unmarshal RPC response: %w\n",
				corrId,
				err,
			)
			log.Println(errmsg)
			return &errors.InternalError{}
		}

		if rpcResp.Error != nil {
			errmsg := fmt.Sprintf(
				"[%s] Failed to persist the user: %s",
				corrId,
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return err
		}

		return nil
	}

	errmsg := fmt.Sprintf("[%s] Failed to persist the user", corrId)
	log.Println(errmsg)
	return &errors.InternalError{}
}
