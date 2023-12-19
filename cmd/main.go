package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"

	"hyperpage/api"
	"hyperpage/controllers"
	"hyperpage/initializers"
	"hyperpage/models"

	"hyperpage/routes"
	"hyperpage/utils"

	"github.com/pion/webrtc/v3"
	uuid "github.com/satori/go.uuid"
	// "github.com/sideshow/apns2"
	// "github.com/sideshow/apns2/token"
)

type TimeEntry struct {
	Hour    int `json:"hour"`
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds"`
}

type TimeArray []TimeEntry

func GetDurationComponents(duration time.Duration) TimeArray {
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	timeEntry := TimeEntry{
		Hour:    hours,
		Minutes: minutes,
		Seconds: seconds,
	}

	return TimeArray{timeEntry}
}

var (
	// Map to keep track of connected clients
	clients = make(map[string]*websocket.Conn)
	w       = sync.WaitGroup{}
)

// Функция для поиска клиента по его идентификатору
func findClientByID(id string) (*websocket.Conn, bool) {
	client, ok := clients[id]
	return client, ok
}

type Peer struct {
	Conn           *websocket.Conn
	PeerConnection *webrtc.PeerConnection
}

var peers = make(map[string]*Peer)
var peersLock sync.RWMutex
var ClientsLock sync.RWMutex

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}

	initializers.ConnectDB(&config)
	initializers.ConnectRedis(&config)
	initializers.ConnectTelegram(&config)
	// initializers.ConnectRethinkDB()

}

func main() {

	// url := "https://api.development.push.apple.com/3/device/5334f3e850f3e06f5e3714344e4f6c5358751829290a64e65ed3afdeec085d1c"

	// payload := `{"uuid":"a582647b-7bf5-4bb4-a5da-98e6ef08eb5a", "action": "coming_call", "handle": "Arsen Beketov", "sdp":[{"type":"offer","sdp":"..."}]}`

	// certPath := "keys/voipCert.pem"
	// keyPath := "keys/key.pem"

	// // Load certificate and key
	// cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	// if err != nil {
	// 	fmt.Println("Error loading certificate:", err)
	// 	return
	// }

	// // Create a custom Transport with TLS configuration
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{
	// 		Certificates: []tls.Certificate{cert},
	// 	},
	// 	ForceAttemptHTTP2: true,
	// }

	// client := &http.Client{
	// 	Transport: tr,
	// }

	// req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	// if err != nil {
	// 	fmt.Println("Error creating request:", err)
	// 	return
	// }

	// req.Header.Set("Content-Type", "application/json")

	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error making request:", err)
	// 	return
	// }
	// defer resp.Body.Close()

	// PE
	// privateKeyPath := "keys/AuthKey_485K6P55G9.p8"
	// keyID := "485K6P55G9"        // Идентификатор ключа (Key ID) из Apple Developer Console
	// teamID := "DBJ8D3U6HY"       // Идентификатор команды (Team ID) из Apple Developer Console
	// bundleID := "dev.paxintrade" // Bundle ID вашего приложения

	// authKey, err := token.AuthKeyFromFile(privateKeyPath)
	// if err != nil {
	// 	fmt.Println("Ошибка загрузки AuthKey:", err)
	// 	return
	// }

	// tokenSource := &token.Token{
	// 	KeyID:   keyID,
	// 	TeamID:  teamID,
	// 	AuthKey: authKey,
	// }

	// client := apns2.NewTokenClient(tokenSource)

	// // Токен устройства, который вы получили после успешной регистрации на уведомления
	// deviceToken := "b2a6c3c409d0abdb0e245c5ce7cae00e0171b88372a386898e202dbcaa305e92"

	// notification := &apns2.Notification{}
	// notification.DeviceToken = deviceToken
	// notification.Topic = bundleID
	// notification.Payload = []byte(`{"aps":{"alert":"Hello, this is a push notification.", "sound":"default"}}`)

	// res, err := client.Push(notification)

	// if err != nil {
	// 	fmt.Println("Ошибка отправки уведомления:", err)
	// 	return
	// }

	// fmt.Println("Уведомление успешно отправлено:", res)

	engine := html.New("./views", ".html")

	// app := fiber.New(config)
	app := fiber.New(fiber.Config{
		ServerHeader: "paxintrade",
		Views:        engine,
		BodyLimit:    20 * 1024 * 1024, // 20 MB
	})

	micro := fiber.New()

	//VIEWS
	routes.Register(app)

	//API'S
	api.Register(micro)

	//REGISTER NEW ROUTES
	app.Mount("/api", micro)

	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Access-Control-Allow-Headers, Session, Mode",
		AllowMethods:     "GET, POST, PATCH, DELETE",
		AllowCredentials: true,
	}))

	app.Get("/socket.io/", websocket.New(func(c *websocket.Conn) {

		// Timeout client 5min.
		c.SetReadDeadline(time.Now().Add(5 * time.Hour))

		id := uuid.NewV4()
		idStr := base64.URLEncoding.EncodeToString(id[:])

		// Send the ID to the client

		// err := c.WriteMessage(websocket.TextMessage, []byte(idStr))
		// if err != nil {
		// 	fmt.Println("error writing message to client", idStr, ":", err)
		// 	return
		// }

		// session := initializers.GetSession()
		// res, err := r.Table("users").Nth(0).Run(session)
		// if err != nil {
		// 	log.Println("RethinkDB query error:", err)
		// 	return
		// }
		// defer res.Close()
		// type User struct {
		// 	ID    string `json:"id" rethinkdb:"id"`
		// 	UsersOnline string `json:"usersOnline" rethinkdb:"usersOnline"`
		// }
		// var user User
		// if res.Next(&user) {
		// 	// Access the data from the user struct
		// 	fmt.Println(user)
		// }

		// Initialize Redis client
		configPath := "./app.env"
		config, err := initializers.LoadConfig(configPath)
		if err != nil {
			fmt.Println("error loading config:", err)
			return
		}

		// Connect to Redis
		redisClient := initializers.ConnectRedis(&config)
		defer redisClient.Close()

		// Add client to clients map
		ClientsLock.Lock()
		utils.Clients[idStr] = c
		ClientsLock.Unlock()

		stunServer := webrtc.ICEServer{
			URLs: []string{"stun:stun.l.google.com:19302"}, // Use a publicly available STUN server
		}

		// Create a new WebRTC peer connection with STUN server configuration
		configrtc := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{stunServer},
		}

		peerConnection, err := webrtc.NewPeerConnection(configrtc)
		if err != nil {
			log.Fatal(err)
			return
		}

		peer := &Peer{
			Conn:           c,
			PeerConnection: peerConnection,
		}

		peersLock.Lock()
		peers[idStr] = peer
		peersLock.Unlock()

		//CHECK USER LOGIN OR NOT
		authToken := c.Cookies("access_token")

		refresh_token := c.Cookies("refresh_token")

		if authToken != "" {

			xconfig, _ := initializers.LoadConfig(".")

			tokenClaims, err := utils.ValidateToken(authToken, xconfig.AccessTokenPublicKey)
			if err != nil {
				fmt.Println("TOKEN DIE")

				if refresh_token != "" {

					fmt.Println("NEED MAKE NEW TOKEN HERE")

				}

			}
			// Check if tokenClaims is nil before accessing its fields
			if tokenClaims != nil {
				UserID := tokenClaims.UserID
				// Continue processing with UserID
				var user models.User
				if err := initializers.DB.Model(&user).Where("id = ?", UserID).First(&user).Error; err != nil {
					_ = err

					// Handle the error (e.g., user not found)
					// You might return an error response or take appropriate action.
				}
				userName := user.Name

				if UserID != "" {
					initializers.DB.Model(&user).Where("id = ?", UserID).Updates(map[string]interface{}{"online": true})
					utils.UserActivity("userOnline", userName)

				} else {
					fmt.Println("User is not logged in")
				}
			}
			// extract the TokenUuid field from tokenClaims

		}

		defer delete(utils.Clients, idStr)
		startTime := time.Now()

		// Setup dials time in min

		defer func() {
			peersLock.Lock()
			delete(peers, idStr)
			peersLock.Unlock()
			// Remove client from clients map
			// Calculate and log elapsed time for the client
			elapsedTime := time.Since(startTime)
			durationComponents := GetDurationComponents(elapsedTime)
			log.Printf("Client %s disconnected after %s", idStr, elapsedTime)

			///var clientsMap map[string]*websocket.Conn
			clients = utils.Clients

			fmt.Println(clients)

			// Remove client from clients map
			delete(clients, idStr)

			fmt.Println("client deleted from Redis:", idStr)

			// Close WebSocket connection
			err = c.Close()
			if err != nil {
				fmt.Println("error closing WebSocket connection:", err)
				return
			}

			// var user models.User
			// initializers.DB.Where("session = ?", idStr).First(&user)

			now := time.Now()

			//CHECK USER IS LOGIN OR NOT
			authToken := c.Cookies("access_token")
			// fmt.Println("authToken", authToken)

			if authToken != "" {

				xconfig, _ := initializers.LoadConfig(".")

				tokenClaims, err := utils.ValidateToken(authToken, xconfig.AccessTokenPublicKey)
				if err != nil {
					// handle error
					_ = err
				}

				if tokenClaims == nil {
					fmt.Println("Token is missing or invalid")
					return // Return if tokenClaims is nil
				}

				// extract the TokenUuid field from tokenClaims
				UserID := tokenClaims.UserID

				if UserID != "" {
					var user models.User
					// now := time.Now()

					initializers.DB.Where("id = ?", UserID).First(&user)
					existingHours := user.OnlineHours // Initialize existingHours as a TimeEntryScanner

					hours := int(durationComponents[0].Hour)
					minutes := int(durationComponents[0].Minutes)
					seconds := int(durationComponents[0].Seconds)

					if len(existingHours) > 0 {
						lastEntry := existingHours[len(existingHours)-1]
						totalSeconds := lastEntry.Seconds + seconds
						totalMinutes := lastEntry.Minutes + minutes + totalSeconds/60
						totalHours := lastEntry.Hour + hours + totalMinutes/60

						lastEntry.Seconds = totalSeconds % 60
						lastEntry.Minutes = totalMinutes % 60
						lastEntry.Hour = totalHours % 24

						existingHours[len(existingHours)-1] = lastEntry
					} else {
						// No existing hours, create a new entry
						timeEntry := models.TimeEntry{
							Hour:    hours,
							Minutes: minutes,
							Seconds: seconds,
						}
						existingHours = append(existingHours, timeEntry)
					}

					jsonBytes, err := json.Marshal(existingHours)
					if err != nil {
						// Handle the error
						_ = err

					}

					var updatedHours []models.TimeEntry // Define a variable of type []models.TimeEntry

					err = json.Unmarshal(jsonBytes, &updatedHours) // Unmarshal jsonBytes into updatedHours
					if err != nil {
						// Handle the error
						_ = err

					}

					user.OnlineHours = updatedHours // Assign updatedHours to user.OnlineHours
					formattedTime := string(jsonBytes)

					// Lost connection
					initializers.DB.Model(&user).Updates(map[string]interface{}{"online": false, "last_online": now, "online_hours": formattedTime})

					// Access the user's ID with `user.ID`
					// userID := user.ID.String()
					userName := user.Name
					//CHECK USER LOGIN OR NOT
					// authToken := c.Cookies("access_token")

					utils.UserActivity("userOffline", userName)
					// initializers.DB.Model(&user).Where("ID = ?", UserID).Updates(map[string]interface{}{"online": true})
				} else {
					fmt.Println("User is not logged in")
				}

			}

			// if authToken != "" {

			// xconfig, _ := initializers.LoadConfig(".")

			// tokenClaims, err := utils.ValidateToken(authToken, xconfig.AccessTokenPublicKey)
			// if err != nil {
			// 	// handle error
			// }

			// // extract the TokenUuid field from tokenClaims
			// UserID := tokenClaims.UserID

			// if UserID != "" {
			// 	utils.UserActivity("userOffline", UserID)

			// } else {
			// 	 fmt.Println("User is not logged in")
			// }

			// }
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// Set the clients JSON string as the value in a Redis hash
			redisKey := "connected_clients"
			err := redisClient.HDel(ctx, redisKey, idStr).Err()
			if err != nil {
				fmt.Println("error deleting client info from Redis:", err)
				return
			}

			fmt.Println("WebSocket client disconnected:", idStr)

		}()

		// // Convert the Clients map to a JSON string
		// clientsJson, err := json.Marshal(utils.Clients)
		// if err != nil {
		// 	fmt.Println("error marshaling clients to JSON:", err)
		// 	return
		// }

		// Create a context with a timeout of 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Set the clients JSON string as the value in a Redis hash
		redisKey := "connected_clients"
		clientInfo, err := json.Marshal(utils.Clients)
		err = redisClient.HSet(ctx, redisKey, idStr, clientInfo).Err()
		if err != nil {
			fmt.Println("error setting client info in Redis:", err)
			return
		}

		type messageSocket struct {
			MessageType string           `json:"messageType"`
			Data        []map[string]int `json:"data"`
		}

		// Wait for messages from the client
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("error reading message from client", idStr, ":", err)
				break
			}

			var Message messageSocket
			if err := json.Unmarshal(message, &Message); err != nil {
				fmt.Println("error unmarshalling JSON:", err)
				continue
			}

			if Message.MessageType == "getMySessionId" {
				fmt.Println("WebSocket client connected with ID:", idStr)
				w.Add(1)

				type SessionIDMessage struct {
					SessionID string `json:"session"`
				}

				sessionIDMessage := SessionIDMessage{
					SessionID: idStr,
				}

				jsonData, err := json.Marshal(sessionIDMessage)
				if err != nil {
					fmt.Println("Ошибка при преобразовании в JSON:", err)
					return
				}

				go func() {

					err := c.WriteMessage(websocket.TextMessage, jsonData)

					// err := utils.UserActivity("sessionId", idStr)
					if err != nil {
						fmt.Println("error writing message to client", idStr, ":", err)
						return
					}
					// w.Wait()
				}()
			}

			if Message.MessageType == "reject" {
				type RejectMessage struct {
					Command string `json:"command"`
				}

				type YourMessageType struct {
					MessageType string `json:"MessageType"`
					Data        []struct {
						ID string `json:"id"`
					} `json:"data"`
				}

				var data YourMessageType

				if err := json.Unmarshal([]byte(message), &data); err != nil {
					fmt.Println("Ошибка при разборе JSON:", err)
					return
				}

				for _, data := range data.Data {
					id := data.ID

					targetClientConn, ok := findClientByID(id)
					if !ok {
						fmt.Printf("Клиент с идентификатором %s не найден\n", id)
						return
					}

					rejectMessage := RejectMessage{
						Command: "endc",
					}

					jsonData, err := json.Marshal(rejectMessage)
					if err != nil {
						fmt.Println("Ошибка при преобразовании в JSON:", err)
						return
					}

					err = targetClientConn.WriteMessage(websocket.TextMessage, jsonData)
					if err != nil {
						fmt.Printf("Ошибка отправки запроса: %v\n", err)
						return
					}
				}

			}

			strMessage := string(message)

			// if strings.Contains(strMessage, "reject") {
			// 	type RejectMessage struct {
			// 		Command string `json:"command"`
			// 	}

			// 	var data map[string]interface{}
			// 	if err := json.Unmarshal([]byte(message), &data); err != nil {
			// 		fmt.Println("Ошибка при разборе JSON:", err)
			// 		return
			// 	}

			// 	id, ok := data["id"].(string)
			// 	if !ok {
			// 		fmt.Println("Не удалось получить значение id или тип не является строкой")
			// 		return
			// 	}

			// 	targetClientConn, ok := findClientByID(id)
			// 	if !ok {
			// 		fmt.Printf("Клиент с идентификатором %s не найден\n", id)
			// 		return
			// 	}

			// 	rejectMessage := RejectMessage{
			// 		Command: "endc",
			// 	}

			// 	jsonData, err := json.Marshal(rejectMessage)
			// 	if err != nil {
			// 		fmt.Println("Ошибка при преобразовании в JSON:", err)
			// 		return
			// 	}

			// 	err = targetClientConn.WriteMessage(websocket.TextMessage, jsonData)
			// 	if err != nil {
			// 		fmt.Printf("Ошибка отправки запроса: %v\n", err)
			// 		return
			// 	}
			// }

			if strings.Contains(strMessage, "sdpAnswer") {
				type Message struct {
					Command string `json:"command"`
					UserB   string `json:"userb"`
					SDP     string `json:"sdp"`
					UserA   string `json:"usera"`
				}

				var data map[string]interface{}
				if err := json.Unmarshal([]byte(message), &data); err != nil {
					fmt.Println("Ошибка при разборе JSON:", err)
					return
				}

				id, ok := data["sessionID"].(string)
				if !ok {
					fmt.Println("Не удалось получить значение id или тип не является строкой")
					return
				}

				sdp, ok := data["sdpAnswer"].(string)
				if !ok {
					fmt.Println("Не удалось получить значение sdpAnswer или тип не является строкой")
					return
				}

				message := Message{
					Command: "sdpAnswer",
					UserB:   id,
					SDP:     sdp,
					UserA:   idStr,
				}

				jsonData, err := json.Marshal(message)
				if err != nil {
					fmt.Println("Ошибка при преобразовании в JSON:", err)
					return
				}

				targetClientConn, ok := findClientByID(id)
				if !ok {
					fmt.Printf("Клиент с идентификатором %s не найден\n", id)
					return
				}

				err = targetClientConn.WriteMessage(websocket.TextMessage, jsonData)
				if err != nil {
					fmt.Printf("Ошибка отправки запроса: %v\n", err)
					return
				}
			}

			if strings.Contains(strMessage, "call") {

				id := strings.Split(strMessage, "call")

				targetClientID := id[1]
				var x = string(targetClientID)

				targetClientConn, ok := findClientByID(x)
				if !ok {
					fmt.Printf("Клиент с идентификатором %s не найден\n", x)
					continue
				}

				targetClientConn2, ok := findClientByID(idStr)
				if !ok {
					fmt.Printf("Клиент с идентификатором %s не найден\n", x)
					continue
				}

				var user models.User
				initializers.DB.Where("session = ?", idStr).First(&user)

				err := targetClientConn.WriteMessage(websocket.TextMessage, []byte("coming_call,"+user.Name+","+idStr))
				if err != nil {
					fmt.Printf("Ошибка отправки запроса на вызов: %v\n", err)
					continue
				}

				var user2 models.User
				initializers.DB.Where("session = ?", x).First(&user2)

				err = targetClientConn2.WriteMessage(websocket.TextMessage, []byte("incoming_call,"+user2.Name+","+x))
				if err != nil {
					fmt.Printf("Ошибка отправки запроса на вызов: %v\n", err)
					continue
				}
			}

			// if string(message) == "kickMe" {

			// 	elapsedTime := time.Since(startTime)
			// 	durationComponents := GetDurationComponents(elapsedTime)

			// 	log.Printf("Client %s disconnected after %s", idStr, elapsedTime)

			// 	// Remove client from Redis hash if it was added
			// 	redisKey := "connected_clients"
			// 	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			// 	defer cancel()

			// 	// Check if key exists
			// 	exists, err := redisClient.Exists(ctx, redisKey).Result()
			// 	if err != nil {
			// 		fmt.Println("error checking if key exists in Redis:", err)
			// 		return
			// 	}
			// 	if exists == 0 {
			// 		fmt.Println("key does not exist in Redis")
			// 		return
			// 	}

			// 	///var clientsMap map[string]*websocket.Conn
			// 	clients = utils.Clients

			// 	fmt.Println(clients)

			// 	// Remove client from clients map
			// 	delete(clients, idStr)

			// 	// Delete client from Redis hash
			// 	result := redisClient.HDel(ctx, redisKey, idStr)
			// 	if result.Err() != nil {
			// 		fmt.Println("error deleting client from Redis:", result.Err())
			// 		return
			// 	}

			// 	fmt.Println("client deleted from Redis:", idStr)

			// 	// Marshal the updated clients map
			// 	updatedClientsJson, err := json.Marshal(clients)
			// 	if err != nil {
			// 		fmt.Println("error marshaling updated clients to JSON:", err)
			// 		return
			// 	}

			// 	// Set the updated clients map in Redis
			// 	err = redisClient.HSet(ctx, redisKey, idStr, updatedClientsJson).Err()
			// 	if err != nil {
			// 		fmt.Println("error setting connected clients in Redis:", err)
			// 		return
			// 	}

			// 	// Close WebSocket connection
			// 	err = c.Close()
			// 	if err != nil {
			// 		fmt.Println("error closing WebSocket connection:", err)
			// 		return
			// 	}

			// 	var user models.User
			// 	now := time.Now()

			// 	initializers.DB.Where("session = ?", idStr).First(&user)

			// 	existingHours := user.OnlineHours // Initialize existingHours as a TimeEntryScanner

			// 	hours := int(durationComponents[0].Hour)
			// 	minutes := int(durationComponents[0].Minutes)
			// 	seconds := int(durationComponents[0].Seconds)

			// 	if len(existingHours) > 0 {
			// 		lastEntry := existingHours[len(existingHours)-1]
			// 		totalSeconds := lastEntry.Seconds + seconds
			// 		totalMinutes := lastEntry.Minutes + minutes + totalSeconds/60
			// 		totalHours := lastEntry.Hour + hours + totalMinutes/60

			// 		lastEntry.Seconds = totalSeconds % 60
			// 		lastEntry.Minutes = totalMinutes % 60
			// 		lastEntry.Hour = totalHours % 24

			// 		existingHours[len(existingHours)-1] = lastEntry
			// 	} else {
			// 		// No existing hours, create a new entry
			// 		timeEntry := models.TimeEntry{
			// 			Hour:    hours,
			// 			Minutes: minutes,
			// 			Seconds: seconds,
			// 		}
			// 		existingHours = append(existingHours, timeEntry)
			// 	}

			// 	jsonBytes, err := json.Marshal(existingHours)
			// 	if err != nil {
			// 		// Handle the error
			// 	}

			// 	var updatedHours []models.TimeEntry // Define a variable of type []models.TimeEntry

			// 	err = json.Unmarshal(jsonBytes, &updatedHours) // Unmarshal jsonBytes into updatedHours
			// 	if err != nil {
			// 		// Handle the error
			// 	}

			// 	user.OnlineHours = updatedHours // Assign updatedHours to user.OnlineHours
			// 	formattedTime := string(jsonBytes)

			// 	// Lost connection
			// 	initializers.DB.Model(&user).Updates(map[string]interface{}{"online": false, "last_online": now, "online_hours": formattedTime})

			// 	// Access the user's ID with `user.ID`
			// 	userID := user.ID.String()

			// 	//CHECK USER LOGIN OR NOT
			// 	// authToken := c.Cookies("access_token")

			// 	utils.UserActivity("userOffline", userID)

			// 	fmt.Println("WebSocket client disconnected:", idStr)

			// }

		}

		// Show log of all clients currently connected
		fmt.Println("Currently connected clients:")
		for id := range utils.Clients {
			fmt.Println(id)
		}
	}))

	// uncomment for test reset .... witouth wait

	// resetAndSaveOnlineData()
	// Emulate the current time

	// currentTime := time.Date(2027, 1, 31, 23, 59, 59, 0, time.UTC) // Set the desired date and time for testing

	currentTime := time.Now().UTC()

	// Calculate the end of the current month
	year, month, _ := currentTime.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, currentTime.Location())
	lastDayOfMonth := nextMonth.Add(-24 * time.Hour)

	// Check if the current day is the last day of the month
	if currentTime.Day() == lastDayOfMonth.Day() && currentTime.Hour() == 23 && currentTime.Minute() == 59 {
		resetAndSaveOnlineData()
	} else {
		// Calculate the desired execution time (23:59:00) on the last day of the month
		desiredTime := time.Date(year, month, lastDayOfMonth.Day(), 23, 59, 0, 0, currentTime.Location())

		// Calculate the duration until the desired execution time
		durationUntilDesiredTime := desiredTime.Sub(currentTime)

		// Create a timer with the duration until the desired execution time
		timer := time.NewTimer(durationUntilDesiredTime)

		go func() {
			<-timer.C
			resetAndSaveOnlineData()

			// Calculate the start of the next day
			nextDay := currentTime.Truncate(24 * time.Hour).Add(24 * time.Hour)

			// Calculate the duration until the start of the next day
			durationUntilNextDay := nextDay.Sub(time.Now().UTC())

			// Create a ticker with the duration until the start of the next day
			if durationUntilNextDay > 0 {
				ticker := time.NewTicker(durationUntilNextDay)

				// Start the ticker loop
				for range ticker.C {
					currentTime = time.Now().UTC()
					if currentTime.Hour() == 23 && currentTime.Minute() == 59 {
						resetAndSaveOnlineData()
					}
				}
			}
		}()
	}

	//Check blog Expired
	ticker := time.NewTicker(24 * time.Hour)
	config2, _ := initializers.LoadConfig(".")

	cfg := &initializers.Config{
		TELEGRAM_TOKEN: config2.TELEGRAM_TOKEN,
	}

	bot, err := initializers.ConnectTelegram(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Get a channel that continuously receives updates from the chat.
	updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{})
	if err != nil {
		log.Fatal(err)
	}
	// Get a channel that continuously receives updates from the chat.
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			// utils.CheckExpiration(bot)
			utils.MoveToArch(bot)
			utils.CheckPlan(bot)
			utils.CheckSite(bot)
			utils.CheckSiteTime(bot)

		}

	}()

	// Create a channel to receive messages that contain the desired words.

	// Define the words to filter for.
	allMsgs := make(chan *tgbotapi.Message)
	conn, ch := initializers.ConnectRabbitMQ()

	// Start a goroutine to send all messages to the allMsgs channel.
	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			msg := update.Message

			if strings.Contains(strings.ToLower(msg.Text), "активность") {
				// Declare a queue
				queueName := "profile_activity"            // Replace with your desired queue name
				conn, ch := initializers.ConnectRabbitMQ() // Create a new connection and channel for each request

				_, err := ch.QueueDeclare(
					queueName,
					false, // durable
					false, // autoDelete
					false, // exclusive
					false, // noWait
					nil,   // args
				)
				if err != nil {
					log.Printf("Failed to declare a queue: %s", err)
					// Handle the error if needed
				}

				// Publish a message to the declared queue
				message := "Hello, RabbitMQ!" // Replace with your desired message
				err = utils.PublishMessage(ch, queueName, message)
				if err != nil {
					log.Printf("Failed to publish message: %s", err)
					// Handle the error if needed
				}

				// Consume messages from the declared queue
				err = utils.ConsumeMessages(ch, conn, queueName)
				if err != nil {
					log.Printf("Failed to consume messages: %s", err)
					// Handle the error if needed
				}

				// Start consuming messages in a separate goroutine
				go func() {
					// Call the controller to process the message
					controllers.ProfileActivity(bot, msg)
				}()

			}

			if strings.Contains(strings.ToLower(msg.Text), "code") {

				words := strings.Split(msg.Text, " ")
				if len(words) > 1 {

					// Use the 'afterSpace' variable as needed
				} else {
					// Convert the message to lowercase for case-insensitive matching
					lowerText := strings.ToLower(msg.Text)

					// Check if the message contains the word "code" (without space)
					if strings.Contains(lowerText, "code") {
						// Replace "code" with an empty string to remove it from the message
						value := strings.ReplaceAll(msg.Text, "code", "")

						// Trim any leading or trailing spaces
						value = strings.TrimSpace(value)

						// Now, the 'value' variable will contain only the value after "code"
						// You can use this value as needed in your program
						// For example, you can print it:
						queueName := "profile_activated"           // Replace with your desired queue name
						conn, ch := initializers.ConnectRabbitMQ() // Create a new connection and channel for each request

						_, err := ch.QueueDeclare(
							queueName,
							false, // durable
							false, // autoDelete
							false, // exclusive
							false, // noWait
							nil,   // args
						)
						if err != nil {
							log.Printf("Failed to declare a queue: %s", err)
							// Handle the error if needed
						}

						// Publish a message to the declared queue
						message := "Hello, user!" // Replace with your desired message
						err = utils.PublishMessage(ch, queueName, message)
						if err != nil {
							log.Printf("Failed to publish message: %s", err)
							// Handle the error if needed
						}

						// Consume messages from the declared queue
						err = utils.ConsumeMessages(ch, conn, queueName)
						if err != nil {
							log.Printf("Failed to consume messages: %s", err)
							// Handle the error if needed
						}

						// Start consuming messages in a separate goroutine
						go func() {
							// Call the controller to process the message
							controllers.TryActivated(bot, msg, value)
						}()
					} else {
						// The message does not contain the word "code"
						fmt.Println("The message does not contain 'code'")
					}
				}

			}

			if strings.Contains(strings.ToLower(msg.Text), "биллинг") {

				words := strings.Split(msg.Text, " ")
				if len(words) > 1 {
					afterSpace := strings.Join(words[1:], " ")
					// Declare a queue
					queueName := "make_balance"                // Replace with your desired queue name
					conn, ch := initializers.ConnectRabbitMQ() // Create a new connection and channel for each request

					_, err := ch.QueueDeclare(
						queueName,
						false, // durable
						false, // autoDelete
						false, // exclusive
						false, // noWait
						nil,   // args
					)
					if err != nil {
						log.Printf("Failed to declare a queue: %s", err)
						// Handle the error if needed
					}

					// Publish a message to the declared queue
					message := "Hello, admin!" // Replace with your desired message
					err = utils.PublishMessage(ch, queueName, message)
					if err != nil {
						log.Printf("Failed to publish message: %s", err)
						// Handle the error if needed
					}

					// Consume messages from the declared queue
					err = utils.ConsumeMessages(ch, conn, queueName)
					if err != nil {
						log.Printf("Failed to consume messages: %s", err)
						// Handle the error if needed
					}

					// Start consuming messages in a separate goroutine
					go func() {
						// Call the controller to process the message
						controllers.MakeCodes(bot, msg, afterSpace)
					}()
					// Use the 'afterSpace' variable as needed
				} else {
					fmt.Println("No text after the first space")
				}

			}

			if strings.Contains(strings.ToLower(msg.Text), "баланс") {
				go controllers.BalanceProfile(bot, msg)

			}

			// Check if the received message is the /start command
			if update.Message.Command() == "start" {
				userLanguage := update.Message.From.LanguageCode

				// Default message in case the language is not supported
				var welcomeMessage string

				// Check the user's language and set the appropriate welcome message
				switch userLanguage {
				case "ru":
					welcomeMessage = "Добро пожаловать в Paxintrade, пришлите ваш код активации."
				case "ka":
					welcomeMessage = "კეთილი იყოს თქვენი მობრძანება Paxintrade-ში, გაგზავნეთ თქვენი აქტივაციის კოდი."
				// Add more cases for other languages as needed
				default:
					welcomeMessage = "Welcome to Paxintrade platform, send your activation code."
				}

				// Create a new message config
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
				// Send the message
				if _, err := bot.Send(msg); err != nil {
					log.Println(err)
				}
			}

			// if update.Message.IsCommand() && update.Message.Command() == "invite" {
			// 	// Process the invite command
			// 	// Replace "YOUR_CHANNEL_OR_GROUP_USERNAME" with the target channel/group username
			// 	channelUsername := "YOUR_CHANNEL_OR_GROUP_USERNAME"
			// 	inviteLink := getInviteLink(channelUsername)

			// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, inviteLink)
			// 	bot.Send(msg)
			// }
			if update.Message.IsCommand() {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

				switch update.Message.Command() {
				case "shop":
					msg.Text = "Добро пожаловать в paxintrade:\n\n" +
						"[Карта с балансом на 5.000](https://paxintrade.com/cth5kWSupR4/moy-tovar-v-telegram) - 5.000 ₽  /buy1\n" +
						"[Карта с балансом на 10.000](https://paxintrade.com/cth5kWSupR4/moy-tovar-v-telegram) - 10.000 ₽ /buy2\n" +
						"[Карта с балансом на 50.000](https://paxintrade.com/cth5kWSupR4/moy-tovar-v-telegram) - 50.000 ₽ /buy3\n"
					msg.ParseMode = tgbotapi.ModeMarkdown
				case "buy1":
					// Обработка покупки товара 1 через ЮKassa
					msg.Text = "Покупка карты с предоплаченным балансом 5.000 ₽ успешно завершена! Ваш код: , используйте его в вашем личном кабинете"
				case "buy2":
					// Обработка покупки товара 1 через ЮKassa
					msg.Text = "Покупка карты с предоплаченным балансом 10.000 ₽ успешно завершена! Ваш код: , используйте его в вашем личном кабинете"
				case "buy3":
					// Обработка покупки товара 1 через ЮKassa
					msg.Text = "Покупка карты с предоплаченным балансом 50.000 ₽ успешно завершена! Ваш код: , используйте его в вашем личном кабинете"
				}

				// Send the message
				_, err := bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
			}

			allMsgs <- update.Message
		}
	}()

	// Use the all messages channel and filter the messages.
	filterWords := []string{"link"}
	filteredMsgs := make(chan [2]string, 10)

	go func() {
		for msg := range allMsgs {
			// Extract the username from the message.
			username := ""
			if msg.From != nil && msg.From.UserName != "" {
				username = msg.From.UserName
			}
			tId := msg.From.ID
			// Check if the message contains any of the filter words.
			for _, word := range filterWords {
				if strings.Contains(strings.ToLower(msg.Text), word) {
					matches := strings.Fields(strings.Replace(msg.Text, "link", "", -1))

					fileURL := "https://example.com/default.png" // Set a default profile photo URL

					if msg.Chat.Type != "private" {
						config := tgbotapi.UserProfilePhotosConfig{
							UserID: msg.From.ID,
							Limit:  1,
						}

						userProfilePhotos, err := bot.GetUserProfilePhotos(config)
						if err != nil {
							log.Panic(err)
						}

						if len(userProfilePhotos.Photos) > 0 {
							photo := userProfilePhotos.Photos[0][0]

							// Call the GetFileDirectURL method to get the direct URL of the file on Telegram's servers
							fileURL, err = bot.GetFileDirectURL(photo.FileID)
							if err != nil {
								log.Panic(err)
							}
						}
					}

					for _, match := range matches {
						filteredMsgs <- [2]string{match, username}

						_, err := controllers.GetMeH(match, username, fileURL, tId)
						if err != nil {
							return
						}
						queueName := "profile_activity" // Replace with your desired queue name
						_, err = ch.QueueDeclare(
							queueName,
							false, // durable
							false, // autoDelete
							false, // exclusive
							false, // noWait
							nil,   // args
						)
						if err != nil {
							log.Printf("Failed to declare a queue: %s", err)
							// Handle the error if needed
						}

						// Publish a message to the declared queue
						message := "Hello, RabbitMQ!" // Replace with your desired message
						err = utils.PublishMessage(ch, queueName, message)
						if err != nil {
							log.Printf("Failed to publish message: %s", err)
							// Handle the error if needed
						}

						// Consume messages from the declared queue
						err = utils.ConsumeMessages(ch, conn, queueName)
						if err != nil {
							log.Printf("Failed to consume messages: %s", err)
							// Handle the error if needed
						}

						// Start consuming messages in a separate goroutine
						go func() {
							// Call the controller to process the message
							// controllers.TryActivated(bot, msg)
						}()
						// bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Спасибо @" + user.Name + " аккаунт активирован!"))
					}

					continue
				}
			}
		}
	}()

	// // Use the filtered messages channel.
	// for msg := range filteredMsgs {
	// 	controllers.GetMeH(msg[0], msg[1])
	// }

	// log.Fatal(app.Listen(":8000"))
	log.Fatal(app.ListenTLS(":8000", "./selfsigned.crt", "./selfsigned.key"))

}

// currentTime := time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC) // Set the desired date and time for testing
// if currentDay == time.Date(currentYear, currentMonth, time.Date(currentYear, currentMonth+1, 0, 0, 0, 0, 0, time.UTC).Day(), 23, 59, 0, 0, time.UTC).Day() {
// WORKING RESET TIMER FIRST EXP
func resetAndSaveOnlineData() {
	fmt.Println("ok?")
	// Get the current month and year
	currentTime := time.Now()
	// currentTime := time.Date(2023, 3, 25, 23, 59, 59, 0, time.UTC) // Set the desired date and time for testing
	currentYear, currentMonth, currentDay := currentTime.Date()

	// Find all users
	var users []models.User
	if err := initializers.DB.Find(&users).Error; err != nil {
		fmt.Println("Error finding users:", err)
		return
	}

	// Iterate over each user
	for _, user := range users {
		// Retrieve the current online hours data from the user
		originalOnlineHours := user.OnlineHours
		originalTotalBlogs := user.TotalBlogs

		// Retrieve the current total online hours data from the user
		totalOnlineHours := user.TotalOnlineHours
		totalRestBlogs := user.TotalRestBlogs

		// Sum the current online hours and add them to the total online hours
		totalOnlineHours[0].Hour += originalOnlineHours[0].Hour
		totalOnlineHours[0].Minutes += originalOnlineHours[0].Minutes
		totalOnlineHours[0].Seconds += originalOnlineHours[0].Seconds

		// Perform carry-over to adjust the time units
		totalOnlineHours[0].Minutes += totalOnlineHours[0].Seconds / 60
		totalOnlineHours[0].Seconds = totalOnlineHours[0].Seconds % 60

		totalOnlineHours[0].Hour += totalOnlineHours[0].Minutes / 60
		totalOnlineHours[0].Minutes = totalOnlineHours[0].Minutes % 60

		totalRestBlogs += originalTotalBlogs
		// Reset the online hours for each user
		user.OnlineHours = models.TimeEntryScanner{
			models.TimeEntry{
				Hour:    0,
				Minutes: 0,
				Seconds: 0,
			},
		}

		user.TotalBlogs = 0

		// Update the total online hours for the user
		user.TotalOnlineHours = totalOnlineHours
		user.TotalRestBlogs = totalRestBlogs
		// Save the updated user in the database
		if err := initializers.DB.Save(&user).Error; err != nil {
			fmt.Println("Error saving user:", err)
			continue
		}

		// Find the corresponding online storage entry for the current year and user
		var onlineStorage models.OnlineStorage
		if err := initializers.DB.Where("user_id = ? AND year = ?", user.ID, currentYear).First(&onlineStorage).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				fmt.Println("Error finding online storage:", err)
				continue
			}

			// If the online storage entry for the current year does not exist, create a new one
			onlineStorage = models.OnlineStorage{
				UserID: user.ID,
				Year:   currentYear,
				Data:   []byte("[]"), // Initialize with empty array for month data
			}
		}

		// Parse the existing online storage data from JSON
		var monthData []models.MonthData
		if err := json.Unmarshal(onlineStorage.Data, &monthData); err != nil {
			fmt.Println("Error unmarshaling online storage data:", err)
			continue
		}

		// Find the current month's data in the online storage data
		found := false
		for i, data := range monthData {
			if data.Month == currentMonth.String() {
				// Update the hours with the retrieved online hours data
				monthData[i].Hours = []models.TimeEntry(originalOnlineHours)
				monthData[i].PostsCount = originalTotalBlogs
				found = true
				break
			}
		}

		// If the current month's data was not found, create a new entry with the retrieved online hours data
		if !found {
			newMonthData := models.MonthData{
				Month:      currentMonth.String(),
				Hours:      []models.TimeEntry(originalOnlineHours),
				PostsCount: originalTotalBlogs,
				// ...
			}
			monthData = append(monthData, newMonthData)
		}

		// Serialize the updated online storage data to JSON
		updatedData, err := json.Marshal(monthData)
		if err != nil {
			fmt.Println("Error marshaling updated month data:", err)
			continue
		}

		// Update the online storage data in the database
		onlineStorage.Data = updatedData
		if err := initializers.DB.Save(&onlineStorage).Error; err != nil {
			fmt.Println("Error saving online storage:", err)
			continue
		}

		fmt.Printf("Online hours reset and online storage updated for user ID: %s\n", user.ID)
	}

	// Check if it's the last day of December
	if (currentMonth == time.December && currentDay == time.Date(currentYear, 12, 31, 0, 0, 0, 0, time.UTC).Day()) ||
		(currentMonth == time.February && currentDay == time.Date(currentYear, 2, getLastDayOfFebruary(currentYear), 0, 0, 0, 0, time.UTC).Day()) {
		// Find all users again
		var users []models.User
		if err := initializers.DB.Find(&users).Error; err != nil {
			fmt.Println("Error finding users:", err)
			return
		}

		// Iterate over each user
		for _, user := range users {
			// Check if the next year's online storage entry already exists for the user
			var nextYearOnlineStorage models.OnlineStorage
			err := initializers.DB.Where("user_id = ? AND year = ?", user.ID, currentYear+1).First(&nextYearOnlineStorage).Error
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					fmt.Println("Error finding next year's online storage:", err)
					continue
				}

				// Create a new row in the "online_storages" table for the next year
				nextYear := currentYear + 1
				newOnlineStorage := models.OnlineStorage{
					UserID: user.ID,
					Year:   nextYear,
					Data:   []byte("[]"), // Initialize with empty array for month data
				}
				if err := initializers.DB.Create(&newOnlineStorage).Error; err != nil {
					fmt.Println("Error creating online storage for the next year:", err)
				}
			}
		}
	}

}

// getLastDayOfFebruary returns the last day of February for the given year
func getLastDayOfFebruary(year int) int {
	if isLeapYear(year) {
		return 29
	}
	return 28
}

// Function to check if a year is a leap year
func isLeapYear(year int) bool {
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return true
	}
	return false
}
