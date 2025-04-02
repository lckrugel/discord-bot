package gateway

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/lckrugel/discord-bot/internal/gateway/events"
)

/* Handle sending and receiving periodic heartbeat exchange */
func handleHeartbeat(client *Client) {
	// Send first heartbeat with a random jitter
	log.Println("[heartbeat] start sending heartbeats...")
	jitter := rand.Float64()
	interval := time.Duration(time.Millisecond * time.Duration(client.heartbeat_interval))

	time.Sleep(time.Duration(interval.Milliseconds() * int64(jitter)))

	client.last_sequence = nil
	heartbeat := events.NewHeartbeatEvent(client.last_sequence)
	err := events.SendEvent(client.conn, heartbeat)
	if err != nil {
		log.Print("[heartbeat] failed to send heartbeat: ", err)
		return
	}
	last_sent_at := time.Now()

	for lastEvent := range client.events {
		if time.Since(last_sent_at) > interval {
			client.Reconnect()
		}
		client.last_sequence = lastEvent.Sequence

		switch lastEvent.Operation {
		case events.Heartbeat_ACK:
			time.Sleep(interval)

			heartbeat = events.NewHeartbeatEvent(client.last_sequence)
			err = events.SendEvent(client.conn, heartbeat)
			last_sent_at = time.Now()
			if err != nil {
				log.Print("[heartbeat] failed to send heartbeat: ", err)
				return
			}

		case events.Heartbeat:
			log.Print("[heartbeat] received heartbeat, responding immediately")
			heartbeat = events.NewHeartbeatEvent(client.last_sequence)
			err = events.SendEvent(client.conn, heartbeat)
			last_sent_at = time.Now()
			if err != nil {
				log.Print("[heartbeat] failed to send heartbeat: ", err)
				return
			}
		}
	}
}
