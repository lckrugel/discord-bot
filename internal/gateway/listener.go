package gateway

import (
	"log"
	"strings"

	"github.com/lckrugel/discord-bot/internal/gateway/events"
)

/* Listen for events on the gateway connection and send them in a channel */
func listener(client *Client) {
	log.Println("[listener] starting listener...")
	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("[listener] connection closed")
				return
			}
			log.Printf("[listener] unexpected error reading gateway message: %v", err)
			client.Reconnect()
			return
		}

		msgPayload, err := events.NewEvent(msg)
		if err != nil {
			log.Fatalf("[listener] error parsing gateway message: %v", err)
		}

		if msgPayload.Sequence != nil {
			client.last_sequence = msgPayload.Sequence
		}

		client.events <- *msgPayload
	}
}
