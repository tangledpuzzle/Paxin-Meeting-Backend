package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"hyperpage/initializers"

	"github.com/gofiber/contrib/websocket"
)

// Map to keep track of connected clients
var Clients = make(map[string]*websocket.Conn)

type UserActivityMessage struct {
	Command string `json:"command"`
	UserID  string `json:"userID"`
}

func UserActivity(command string, userId string) error {
	var err error
	for _, c := range Clients {
		userActivityMessage := UserActivityMessage{
			Command: command,
			UserID:  userId,
		}
		jsonData, e := json.Marshal(userActivityMessage)
		if e != nil {
			err = fmt.Errorf("failed to marshal JSON: %v", e)
			continue
		}
		if e := c.WriteMessage(websocket.TextMessage, jsonData); e != nil {
			err = fmt.Errorf("failed to send message to client: %v", e)
		}
	}
	return err
}

func SendBlogMessageToClients(message string, userName string) error {

	fmt.Println("Sending message to clients:", message)

	// // Cut session id
	// messageWithSessionID := message + " Session ID:" + userName

	// Log the list of clients
	fmt.Println("List of clients:")
	for _, c := range Clients {
		fmt.Println(c)
	}

	if message == "newblog" {

		var err error
		for _, c := range Clients {
			if e := c.WriteMessage(websocket.TextMessage, []byte(message)); e != nil {
				err = fmt.Errorf("failed to send message to client: %v", e)
			}
		}
		return err

	}

	return nil

}

func SendPersonalMessageToClient(clientID, message string) error {

	type ClientMessage struct {
		Command string `json:"command"`
	}

	clientMessage := ClientMessage{
		Command: message,
	}

	jsonData, err := json.Marshal(clientMessage)
	if err != nil {
		return err
	}

	// Get client connection from Redis
	conn, err := GetClientConnFromRedis(clientID)
	if err != nil {
		return err
	}

	// Check if the conn variable is nil
	if conn == nil {
		return errors.New("connection is nil")
	}

	if message == "Activated" {

		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			return err
		}

	}

	if message == "newblog" {

		// Get the total count of records in the "blog" table
		var count int64
		if err := initializers.DB.Table("blogs").Count(&count).Error; err != nil {
			panic(err)
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(strconv.FormatInt(count, 10))); err != nil {
			return err
		}

		return nil

	}

	// Send message to client
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		return err
	}

	return nil
}

func GetClientConnFromRedis(clientID string) (*websocket.Conn, error) {
	if conn, ok := Clients[clientID]; ok {
		// Client connection found in map
		return conn, nil
	}

	// Initialize Redis client
	configPath := "./app.env"
	config, err := initializers.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	redisClient := initializers.ConnectRedis(&config)

	// Retrieve the byte slice representing the connection object from Redis
	var connBytes []byte
	connBytes, err = redisClient.HGet(context.Background(), "connected_clients", clientID).Bytes()
	if err != nil {
		fmt.Printf("Error retrieving value from Redis for key %s: %v\n", clientID, err)
		return nil, err
	}

	// Deserialize the byte slice back into a *websocket.Conn object
	var conn *websocket.Conn

	fmt.Println(conn)

	err = json.Unmarshal(connBytes, &conn)
	if err != nil {
		fmt.Printf("Error deserializing byte slice to websocket conn object: %v\n", err)
		return nil, err
	}

	Clients[clientID] = conn

	// Check if the conn variable is nil
	if conn == nil {
		return nil, errors.New("deserialized websocket conn object is nil")
	}

	//fmt.Println(conn)

	return conn, nil
}
