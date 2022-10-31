package internal

import (
	"context"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// StreamMessages connects to a websocket and sends the received messages to the provided channel.
func StreamMessages(ctx context.Context, u url.URL, messageChan chan<- []byte) error {
	log.WithFields(log.Fields{
		"url": u.String(),
	}).Printf("Connecting")

	c, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case messageChan <- message:
		}
	}
}

// StreamMessagesWithRetry keeps a connection open to a websocket and calls the handler for every received message
func StreamMessagesWithRetry(ctx context.Context, url url.URL, handler func([]byte) error) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	streamRateLimiter := rate.NewLimiter(rate.Every(1*time.Second), 1)
	messageChan := make(chan []byte)

	// goroutine to handle reconnecting to the websocket
	go func() {
		for {
			if err := StreamMessages(ctx, url, messageChan); err != nil {
				log.WithFields(log.Fields{
					"url": url.String(),
				}).WithError(err).Println("Non-fatal error caught while streaming messages.")
			}

			// return early if context is done
			if ctx.Err() != nil {
				return
			}

			// wait for retry
			if err := streamRateLimiter.Wait(ctx); err != nil {
				log.WithFields(log.Fields{
					"url": url.String(),
				}).WithError(err).Println("Error caught while waiting for stream retry.")
			}

			// return if context is done
			if ctx.Err() != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case message := <-messageChan:
			if err := handler(message); err != nil {
				return err
			}
		}
	}
}
