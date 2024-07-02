package socket

import (
	"certification/database"
	"certification/logger"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type client struct {
	ID string
} // Add more data to this type if needed

type BroadcastMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

var clients = make(map[*websocket.Conn]client) // Note: although large maps with pointer-like types (e.g. strings) as keys are slow, using pointers themselves as keys is acceptable and fast
var broadcast = make(chan BroadcastMessage)
var unregister = make(chan *websocket.Conn)

func SendToBroadcast(message BroadcastMessage) {
	broadcast <- message
}

func runHub() {
	for {
		select {
		case message := <-broadcast:
			logger.Log.Infof("Broadcasting message: %v", message)
			// Example message format: "ID:message"

			// Send the message only to the targeted client
			for conn, cl := range clients {
				if cl.ID == message.ID {
					logger.Log.Infof("client: %v, message: %v\n", cl, message)
					if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Content)); err != nil {
						logger.Log.Error(err)
						unregister <- conn
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						conn.Close()
					}
				}
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)
		}
	}
}

func InitializeWebSocket(app *fiber.App, initializer *database.Initializer) {

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) { // Returns true if the client requested upgrade to the WebSocket protocol
			c.Next()
		}
		return nil
	})

	go runHub()

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// Extract ID from the URL
		id := c.Params("id")

		defer func() {
			unregister <- c
			c.Close()
		}()

		// Register the client with ID
		clients[c] = client{ID: id}

		// Existing read loop
		for {
			// Your message handling code here
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					// logger.Log.Error(err)
				}

				return // Calls the deferred function, i.e. closes the connection on error
			}
		}
	}))
}
