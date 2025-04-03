package gateway

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lckrugel/discord-bot/internal/gateway/events"
)

/* Listen for events on the gateway connection and send them in a channel */
func listener(c *Client) error {
	log.Println("[listener] starting listener...")
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Print("[listener] connection closed")
				return nil
			}
			errMsg := fmt.Sprintf("[listener] unexpected error reading gateway message: %v", err)
			c.reconnect_signal <- struct{}{}
			return errors.New(errMsg)
		}

		event, err := events.NewEvent(msg)
		if err != nil {
			log.Printf("[listener] error parsing gateway message: %v", err)
			continue
		}

		if event.Sequence != nil {
			c.last_sequence = event.Sequence
		}

		if event.Operation == events.Reconnect {
			c.reconnect_signal <- struct{}{}
			return errors.New("[listener] received reconnect event")
		}

		c.events <- *event
	}
}
