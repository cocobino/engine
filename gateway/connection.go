package gateway

import (
	"errors"
	"fmt"

	"github.com/it-chain/it-chain-Engine/core/eventstore"
	"github.com/it-chain/midgard"
)

type Connection struct {
	midgard.AggregateModel
	Address string
}

func (c *Connection) On(event midgard.Event) error {
	switch v := event.(type) {

	case *ConnectionCreatedEvent:
		c.ID = v.ID
		c.Address = v.Address

	case *ConnectionClosedEvent:
		c.ID = ""
		c.Address = ""

	default:
		return errors.New(fmt.Sprintf("unhandled event [%s]", v))
	}

	return nil
}

func NewConnection(connection Connection) (Connection, error) {

	c := Connection{}

	event := ConnectionCreatedEvent{
		EventModel: midgard.EventModel{
			ID:   connection.ID,
			Type: "connection.created",
		},
		Address: connection.Address,
	}

	c.On(event)

	return c, eventstore.Save(connection.ID, event)
}

func CloseConnection(connectionID string) error {

	return eventstore.Save(connectionID, ConnectionClosedEvent{
		EventModel: midgard.EventModel{
			ID:   connectionID,
			Type: "connection.closed",
		},
	})
}
