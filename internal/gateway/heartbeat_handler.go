package gateway

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/lckrugel/discord-bot/internal/gateway/events"
)

/* Handle sending and receiving periodic heartbeat exchange */
func handleHeartbeat(c *Client) error {
	// Send first heartbeat with a random jitter
	log.Println("[heartbeat] start sending heartbeats...")
	jitter := rand.Float64()
	interval := time.Duration(time.Millisecond * time.Duration(c.heartbeat_interval))

	time.Sleep(time.Duration(interval.Milliseconds() * int64(jitter)))

	c.last_sequence = nil
	heartbeat := events.NewHeartbeatEvent(c.last_sequence)
	err := events.SendEvent(c.conn, heartbeat)
	if err != nil {
		errMsg := fmt.Sprint("[heartbeat] failed to send first heartbeat: ", err)
		return errors.New(errMsg)
	}
	last_sent_at := time.Now()

	for lastEvent := range c.events {
		if time.Since(last_sent_at) > interval {
			c.reconnect_signal <- struct{}{}
			return errors.New("[heartbeat] heartbeat timeout")
		}
		c.last_sequence = lastEvent.Sequence

		switch lastEvent.Operation {
		case events.Heartbeat_ACK:
			time.Sleep(interval)

			heartbeat = events.NewHeartbeatEvent(c.last_sequence)
			err = events.SendEvent(c.conn, heartbeat)
			last_sent_at = time.Now()
			if err != nil {
				log.Print("[heartbeat] failed to send heartbeat: ", err)
			}

		case events.Heartbeat:
			log.Print("[heartbeat] received heartbeat, responding immediately")
			heartbeat = events.NewHeartbeatEvent(c.last_sequence)
			err = events.SendEvent(c.conn, heartbeat)
			last_sent_at = time.Now()
			if err != nil {
				log.Print("[heartbeat] failed to send heartbeat: ", err)
			}
		}
	}
	return nil
}
