package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
)

type GeoClient interface {
	ReverseGeocode(lat, lon string) (*dtos.Location, error)
	Shutdown()
}

type geoClient struct {
	url		string
	conn		*amqp.Connection
	ch		*amqp.Channel
	replyMsgs	<-chan amqp.Delivery
	pendingCalls	sync.Map
}

type responseChan struct {
	ch  chan []byte
	err chan error
}

func NewGeoClient(url string) (GeoClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to connect: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[RabbitMQ] Failed to create a channel: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	replyMsgs, err := ch.Consume(
		"amq.rabbitmq.reply-to", "",
		true, true, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Printf(
			"[RabbitMQ] Failed to consume `reply-to` queue: %v",
			err,
		)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	client := &geoClient{
		url: url, conn: conn, ch: ch, replyMsgs: replyMsgs,
	}

	go client.listen()

	return client, nil
}

func (c *geoClient) listen() {
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

func (c *geoClient) Shutdown() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *geoClient) call(routingKey string, body []byte) ([]byte, error) {
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
		log.Printf("[RabbitMQ] Failed to publish a message: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case reply := <-res.ch:
		return reply, nil
	case <-ctx.Done():
		c.pendingCalls.Delete(corrID)
		log.Println("[RabbitMQ] Timeout waiting for RPC response")
		return nil, &commonsErrors.MessageBrokerError{}
	}
}

func (c *geoClient) ReverseGeocode(lat, lon string) (*dtos.Location, error) {
	req := dtos.ReverseGeocodeRequest{Lat: lat, Lon: lon}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf(
			"Failed to marshal a reverse geocode request: %v",
			err,
		)
		return nil, &commonsErrors.InternalError{}
	}

	reply, err := c.call("reverse_geocode", reqBytes)
	if err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to publish request for reverse geocoding: %v",
			err,
		)
		log.Println(errmsg)
		return nil, &commonsErrors.InternalError{}
	}

	var rpcResp dtos.RPCResponse
	if err := json.Unmarshal(reply, &rpcResp); err != nil {
		errmsg := fmt.Sprintf(
			"[RabbitMQ] Failed to unmarshal RPC response: %v",
			err,
		)
		log.Println(errmsg)
		return nil, &commonsErrors.InternalError{}
	}

	if rpcResp.Error != nil {
		errmsg := fmt.Sprintf(
			"Geo service internal error: %s",
			rpcResp.Error.Message,
		)
		log.Println(errmsg)
		return nil, &commonsErrors.InternalError{}
	}

	var location dtos.Location
	if err := json.Unmarshal(rpcResp.Data, &location); err != nil {
		errmsg := fmt.Sprintf("Failed to unmarshal a location: %v", err)
		log.Println(errmsg)
		return nil, &commonsErrors.InternalError{}
	}

	return &location, nil
}
