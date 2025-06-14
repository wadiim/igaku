package clients

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"

	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"igaku/auth-service/errors"
	"igaku/commons/dtos"
	"igaku/commons/models"
)

type UserClient interface {
	FindByUsername(username string) (*models.User, error)
	Persist(user *models.User) error
	Shutdown()
}

type userClient struct {
	url		string
	conn		*amqp.Connection
	ch		*amqp.Channel
	replyMsgs	<-chan amqp.Delivery
	pendingCalls	sync.Map
}

type responseChan struct {
	ch	chan []byte
	err	chan error
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

	replyMsgs, err := ch.Consume(
		"amq.rabbitmq.reply-to", "",
		true, true, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf(
			"Failed to consume reply-to queue: %w", err,
		)
	}

	client := &userClient{
		url: url, conn: conn, ch: ch, replyMsgs: replyMsgs,
	}

	go client.listen()

	return client, nil
}

func (c *userClient) listen() {
	for msg := range c.replyMsgs {
		if val, ok := c.pendingCalls.Load(msg.CorrelationId); ok {
			res := val.(*responseChan)
			select {
			case res.ch <- msg.Body:
			default:
			}
			c.pendingCalls.Delete(msg.CorrelationId)
		}
	}
}

func (c *userClient) Shutdown() {
	if c.ch != nil { c.ch.Close() }
	if c.conn != nil { c.conn.Close() }
}

// TODO: Use custom errors
func (c *userClient) call(routingKey string, body []byte) ([]byte, error) {
	corrID := uuid.New().String()
	res := &responseChan{
		ch:	make(chan []byte, 1),
		err:	make(chan error, 1),
	}

	c.pendingCalls.Store(corrID, res)

	err := c.ch.Publish(
		"",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:	"application/json",
			CorrelationId:	corrID,
			ReplyTo:	"amq.rabbitmq.reply-to",
			Body:		body,
		},
	)
	if err != nil {
		c.pendingCalls.Delete(corrID)
		return nil, fmt.Errorf("Failed to publish message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case reply := <-res.ch:
		return reply, nil
	case <-ctx.Done():
		c.pendingCalls.Delete(corrID)
		return nil, fmt.Errorf("Timeout waiting for RPC response")
	}
}

func (c *userClient) FindByUsername(username string) (*models.User, error) {
	reply, err := c.call("find_by_username", []byte(username))
	if err != nil {
		errmsg := fmt.Sprintf(
			"Failed to publish request for username '%s': %w",
			username, err,
		)
		log.Println(errmsg)
		return nil, &errors.InternalError{}
	}

	var rpcResp dtos.RPCResponse
	if err := json.Unmarshal(reply, &rpcResp); err != nil {
		errmsg := fmt.Sprintf(
			"Failed to unmarshal RPC response: %w", err,
		)
		log.Println(errmsg)
		return nil, &errors.InternalError{}
	}

	if rpcResp.Error != nil {
		if rpcResp.Error.Code == "NOT_FOUND" {
			errmsg := fmt.Sprintf(
				"User not found: %s", rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return nil, fmt.Errorf(errmsg)
		} else if rpcResp.Error.Code == "INTERNAL" {
			errmsg := fmt.Sprintf(
				"User service internal error: %s",
				rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return nil, &errors.InternalError{}
		} else {
			errmsg := fmt.Sprintf(
				"Internal error: %s", rpcResp.Error.Message,
			)
			log.Println(errmsg)
			return nil, &errors.InternalError{}
		}
	}

	var user models.User
	if err := json.Unmarshal(rpcResp.Data, &user); err != nil {
		errmsg := fmt.Sprintf("Internal error: %w", err)
		log.Println(errmsg)
		return nil, &errors.InternalError{}
	}

	return &user, nil
}

// TODO: Consider modifying this function so that it takes only username and
// password hash as parameters.
func (c *userClient) Persist(user *models.User) error {
	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("Failed to marshal the user: %w", err)
		return &errors.InternalError{}
	}

	reply, err := c.call("persist", userBytes)
	if err != nil {
		errmsg := fmt.Sprintf(
			"Failed to publish request to persist '%s': %w",
			user.Username, err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	var rpcResp dtos.RPCResponse
	if err := json.Unmarshal(reply, &rpcResp); err != nil {
		errmsg := fmt.Sprintf(
			"Failed to unmarshal RPC response: %w\n", err,
		)
		log.Println(errmsg)
		return &errors.InternalError{}
	}

	if rpcResp.Error != nil {
		errmsg := fmt.Sprintf(
			"Failed to persist the user: %s",
			rpcResp.Error.Message,
		)
		log.Println(errmsg)
		return err
	}

	return nil
}
