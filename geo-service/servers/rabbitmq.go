package servers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	"igaku/geo-service/services"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQServer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	service services.GeoService
}

func NewRabbitMQServer(
	amqpURI string,
	service services.GeoService,
) (*RabbitMQServer, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to connect: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Printf("[RabbitMQ] Failed to open a channel: %v", err)
		return nil, &commonsErrors.MessageBrokerError{}
	}

	return &RabbitMQServer{conn: conn, ch: ch, service: service}, nil
}

func (s *RabbitMQServer) Start() error {
	if err := s.startReverseGeocodeListener(); err != nil {
		log.Printf(
			"[RabbitMQ] Failed to start 'ReverseGeocodeListener': %v",
			err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	return nil
}

func (s *RabbitMQServer) Shutdown() {
	if s.ch != nil {
		s.ch.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *RabbitMQServer) startReverseGeocodeListener() error {
	queueName := "reverse_geocode"

	q, err := s.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Printf(
			"[RabbitMQ] Failed to declare a queue '%s': %v",
			queueName, err,
		)
		return &commonsErrors.MessageBrokerError{}
	}

	if err = s.ch.Qos(1, 0, false); err != nil {
		log.Printf("[RabbitMQ] Failed to set QoS: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	msgs, err := s.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Printf("[RabbitMQ] Failed to register a consumer: %v", err)
		return &commonsErrors.MessageBrokerError{}
	}

	go func() {
		log.Printf(" [*] Awaiting RPC requests on queue '%s'", q.Name)
		for d := range msgs {
			var req dtos.ReverseGeocodeRequest
			log.Printf(
				"Received RPC request for reverse geocoding, ID: %s",
				d.CorrelationId,
			)

			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Printf(
					"Failed to unmarshal an RPC request: %v",
					err,
				)
				s.sendErrorResponse(
					d, "INVALID_REQUEST", err.Error(),
				)
				continue
			}

			location, err := s.service.Reverse(req.Lat, req.Lon)
			if err != nil {
				s.sendErrorResponse(
					d,
					"INTERNAL",
					"Failed to perform reverse geocoding",
				)
				continue
			}

			locationBytes, err := json.Marshal(location)
			if err != nil {
				s.sendErrorResponse(
					d,
					"INTERNAL",
					"Failed to marshal location data",
				)
				continue
			}

			s.sendResponse(d, locationBytes)
		}
	}()

	return nil
}

func (s *RabbitMQServer) sendResponse(d amqp.Delivery, data []byte) {
	resp := dtos.RPCResponse{Data: data}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal an RPC response: %v", err)
		s.sendErrorResponse(d, "INTERNAL", "Failed to create a response")
		return
	}
	s.publish(d, respBytes)
}

func (s *RabbitMQServer) sendErrorResponse(
	d amqp.Delivery, code, message string,
) {
	resp := dtos.RPCResponse{
		Error: &dtos.RPCError{Code: code, Message: message},
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal an RPC error response: %v", err)
		return
	}
	s.publish(d, respBytes)
}

func (s *RabbitMQServer) publish(d amqp.Delivery, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	err := s.ch.PublishWithContext(
		ctx,
		"",
		d.ReplyTo,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		},
	)
	if err != nil {
		log.Printf(
			"Failed to publish reply for ID %s: %v",
			d.CorrelationId, err,
		)
		return
	}
	d.Ack(false)
}
