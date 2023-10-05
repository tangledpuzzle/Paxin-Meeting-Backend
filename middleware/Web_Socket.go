package middleware

// import (
// 	"fmt"
// 	"hyperpage/initializers"

// 	"github.com/gofiber/websocket/v2"
// )

// func WebSocketHandler(c *websocket.Conn) {
// 	// Get the remote address of the client
// 	clientID := c.RemoteAddr().String()

// 	// Add the client to the clients map
// 	initializers.Clients[clientID] = c
// 	defer delete(initializers.Clients, clientID)

// 	// Read messages from the client
// 	for {
// 		_, msg, err := c.ReadMessage()
// 		if err != nil {
// 			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
// 				// Client has closed the connection normally
// 				return
// 			}

// 			// An error occurred while reading from the client
// 			fmt.Println("error reading message:", err)
// 			return
// 		}

// 		// Handle the message
// 		fmt.Printf("received message from client %s: %s\n", clientID, msg)
// 	}
// }