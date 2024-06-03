package paxcall

import (
	"sync"

	"encoding/base64"
	"time"

	"github.com/pion/webrtc/v3"
	uuid "github.com/satori/go.uuid"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var Clients = make(map[string]*websocket.Conn)
var ClientsLock sync.RWMutex

type Peer struct {
	Conn           *websocket.Conn
	PeerConnection *webrtc.PeerConnection
}

var peers = make(map[string]*Peer)
var peersLock sync.RWMutex

func Register(app *fiber.App) {

	app.Get("/", func(c *fiber.Ctx) error {
		//Set session in cookie
		id := uuid.NewV4()
		idSession := base64.URLEncoding.EncodeToString(id[:])

		c.Cookie(&fiber.Cookie{
			Name:     "paxcall_session",
			Value:    idSession,
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: false,
			SameSite: "lax",
			Secure:   false, // Ensure this is false for HTTP; true for HTTPS

		})

		return c.Render("index", fiber.Map{
			"Title":       "Powerful paxintrade/paxcall server",
			"Description": "server developed by paxintrade/paxcall",
		})
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Set up WebSocket route
	app.Get("/ws", func(c *fiber.Ctx) error {

		return c.JSON(fiber.Map{
			"status": "success",
			"data":   "123123132",
		})
	})
	// app.Get("/ws", websocket.New(func(c *websocket.Conn) {
	// 	var session = c.Cookies("paxcall_session")
	// 	ClientsLock.Lock()
	// 	Clients[session] = c
	// 	ClientsLock.Unlock()

	// 	defer func() {
	// 		ClientsLock.Lock()
	// 		delete(Clients, session)
	// 		ClientsLock.Unlock()
	// 	}()

	// 	err := c.WriteMessage(websocket.TextMessage, []byte(session))
	// 	if err != nil {
	// 		fmt.Println("error writing message to client", session, ":", err)
	// 		return
	// 	}

	// 	// Add STUN server configuration
	// 	stunServer := webrtc.ICEServer{
	// 		URLs: []string{"stun:stun.l.google.com:19302"}, // Use a publicly available STUN server
	// 	}

	// 	// Create a new WebRTC peer connection with STUN server configuration
	// 	config := webrtc.Configuration{
	// 		ICEServers: []webrtc.ICEServer{stunServer},
	// 	}

	// 	peerConnection, err := webrtc.NewPeerConnection(config)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 		return
	// 	}

	// 	// Add the peerConnection to the Peer struct
	// 	peer := &Peer{
	// 		Conn:           c,
	// 		PeerConnection: peerConnection,
	// 	}

	// 	peersLock.Lock()
	// 	peers[session] = peer
	// 	peersLock.Unlock()

	// 	defer func() {
	// 		peersLock.Lock()
	// 		delete(peers, session)
	// 		peersLock.Unlock()
	// 	}()

	// 	// Handle incoming SDP messages
	// 	for {
	// 		messageType, msg, err := c.ReadMessage()
	// 		if err != nil {
	// 			break
	// 		}

	// 		strMessage := string(msg)

	// 		// Check and find session in the map for creating SDP
	// 		if strings.Contains(strMessage, "call") {

	// 			var callMessage map[string]string
	// 			if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 				fmt.Println("Error decoding JSON:", err)
	// 				break
	// 			}

	// 			sessionA := callMessage["sessionA"]
	// 			sessionB := callMessage["sessionB"]
	// 			SDPOffer := callMessage["sdpOffer"]

	// 			ClientsLock.RLock()
	// 			_, foundA := Clients[sessionA]
	// 			clientB, foundB := Clients[sessionB]
	// 			ClientsLock.RUnlock()

	// 			if foundB && foundA {
	// 				// Send a message to sessionB
	// 				comingCallMessage := map[string]string{
	// 					"command":  "coming_call",
	// 					"sessionA": sessionA,
	// 					"sessionB": sessionB,
	// 					"sdpOffer": SDPOffer,
	// 					"message":  "Coming call from sessionA",
	// 				}

	// 				comingCallJSON, _ := json.Marshal(comingCallMessage)
	// 				clientB.WriteMessage(messageType, comingCallJSON)
	// 			} else {
	// 				fmt.Println("Session not found in map")
	// 			}
	// 		}

	// 		// if strings.Contains(strMessage, "video_answer") {

	// 		// 	var callMessage map[string]string
	// 		// 	if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 		// 		fmt.Println("Error decoding JSON:", err)
	// 		// 		break
	// 		// 	}

	// 		// 	fmt.Println(callMessage)

	// 		// }

	// 		if strings.Contains(strMessage, "video_started") {

	// 			var callMessage map[string]string
	// 			if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 				fmt.Println("Error decoding JSON:", err)
	// 				break
	// 			}

	// 			sessionB := callMessage["sessionB"]
	// 			clientB := Clients[sessionB]
	// 			SDPOffer := callMessage["sdpOffer"]

	// 			sdpAnswerJSON, _ := json.Marshal(map[string]interface{}{
	// 				"command":  "start_video",
	// 				"sdpOffer": SDPOffer,
	// 			})

	// 			clientB.WriteMessage(messageType, sdpAnswerJSON)

	// 		}

	// 		if strings.Contains(strMessage, "video_stoped") {
	// 			var callMessage map[string]string
	// 			if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 				fmt.Println("Error decoding JSON:", err)
	// 				break
	// 			}

	// 			sdpAnswerJSON, _ := json.Marshal(map[string]interface{}{
	// 				"command": "stoped_video",
	// 			})

	// 			sessionB := callMessage["sessionB"]
	// 			clientB := Clients[sessionB]

	// 			clientB.WriteMessage(messageType, sdpAnswerJSON)

	// 		}
	// 		if strings.Contains(strMessage, "sdpAnswer") {

	// 			type IceCandidate struct {
	// 				Candidate        string `json:"candidate"`
	// 				SdpMid           string `json:"sdpMid"`
	// 				SdpMLineIndex    int    `json:"sdpMLineIndex"`
	// 				UsernameFragment string `json:"usernameFragment"`
	// 			}

	// 			type SdpAnswerMessage struct {
	// 				Command       string         `json:"command"`
	// 				SessionA      string         `json:"sessionA"`
	// 				SessionB      string         `json:"sessionB"`
	// 				SdpAnswer     string         `json:"sdpAnswer"`
	// 				IceCandidates []IceCandidate `json:"iceCandidates"`
	// 			}

	// 			var callMessage SdpAnswerMessage
	// 			if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 				fmt.Println("Error decoding JSON:", err)
	// 				return
	// 			}

	// 			sessionA := callMessage.SessionA
	// 			sessionB := callMessage.SessionB

	// 			SDPOffer := callMessage.SdpAnswer
	// 			iceCandidates := callMessage.IceCandidates

	// 			ClientsLock.RLock()
	// 			clientA, foundA := Clients[sessionA]
	// 			clientB, foundB := Clients[sessionB]
	// 			ClientsLock.RUnlock()

	// 			if foundA {
	// 				// Отправить SDP Answer обоим клиентам
	// 				sdpAnswerJSON, _ := json.Marshal(map[string]interface{}{
	// 					"command":       "sdp_answer",
	// 					"sessionA":      sessionA,
	// 					"sdpAnswer":     SDPOffer,
	// 					"iceCandidates": iceCandidates,
	// 					"message-type":  "text",
	// 				})

	// 				clientA.WriteMessage(messageType, sdpAnswerJSON)
	// 			} else {
	// 				fmt.Println("Session not found in map")
	// 			}
	// 			if foundB {
	// 				// Отправить SDP Answer обоим клиентам
	// 				sdpAnswerJSON, _ := json.Marshal(map[string]interface{}{
	// 					"command":       "sdp_answer",
	// 					"sessionA":      sessionA,
	// 					"sdpAnswer":     SDPOffer,
	// 					"iceCandidates": iceCandidates,
	// 					"message-type":  "text",
	// 				})

	// 				clientB.WriteMessage(messageType, sdpAnswerJSON)
	// 			} else {
	// 				fmt.Println("Session not found in map")
	// 			}
	// 		}

	// 		if strings.Contains(strMessage, "ice_candidate") {
	// 			type IceCandidateMessage struct {
	// 				Command      string `json:"command"`
	// 				SessionA     string `json:"sessionA"`
	// 				SessionB     string `json:"sessionB"`
	// 				IceCandidate struct {
	// 					Candidate     string `json:"candidate"`
	// 					SdpMid        string `json:"sdpMid"`
	// 					SdpMLineIndex int    `json:"sdpMLineIndex"`
	// 					UsernameFrag  string `json:"usernameFragment"`
	// 				} `json:"iceCandidate"`
	// 			}
	// 			var iceCandidateMsg IceCandidateMessage
	// 			err := json.Unmarshal([]byte(strMessage), &iceCandidateMsg)
	// 			if err != nil {
	// 				fmt.Println("Ошибка декодирования JSON:", err)
	// 				return
	// 			}

	// 			if iceCandidateMsg.SessionB != "" {

	// 				sessionBID := iceCandidateMsg.SessionB
	// 				sessionB := Clients[sessionBID]
	// 				comingCallJSON, _ := json.Marshal(iceCandidateMsg.IceCandidate)
	// 				sessionB.WriteMessage(messageType, comingCallJSON)

	// 			}

	// 			if iceCandidateMsg.SessionA != "" {

	// 				// Отправить сообщение для SessionA
	// 				sessionAID := iceCandidateMsg.SessionA
	// 				sessionA := Clients[sessionAID]
	// 				comingCallJSON, _ := json.Marshal(iceCandidateMsg.IceCandidate)
	// 				sessionA.WriteMessage(messageType, comingCallJSON)
	// 			}

	// 			// sessionBID := iceCandidateMsg.SessionB
	// 			// sessionB := Clients[sessionBID]

	// 			// // iceCandidateMsg.Command = ""
	// 			// // iceCandidateMsg.SessionB = ""

	// 			// // fmt.Println(iceCandidateMsg)
	// 			// comingCallJSON, _ := json.Marshal(iceCandidateMsg.IceCandidate)
	// 			// sessionB.WriteMessage(messageType, comingCallJSON)

	// 		}

	// 		// End call
	// 		if strings.Contains(strMessage, "finish") {
	// 			var callMessage map[string]string
	// 			if err := json.Unmarshal([]byte(strMessage), &callMessage); err != nil {
	// 				fmt.Println("Error decoding JSON:", err)
	// 				break
	// 			}

	// 			sessionA := callMessage["sessionA"]
	// 			sessionB := callMessage["sessionB"]

	// 			ClientsLock.RLock()
	// 			clientA, foundA := Clients[sessionA]
	// 			clientB, foundB := Clients[sessionB]
	// 			ClientsLock.RUnlock()

	// 			if foundA && foundB {
	// 				// Send a message to sessionA
	// 				outgoingCall := map[string]string{
	// 					"command":  "finish",
	// 					"sessionA": sessionA,
	// 					"sessionB": sessionB,
	// 					"message":  "call ended",
	// 				}
	// 				comingCallJSON, _ := json.Marshal(outgoingCall)
	// 				clientA.WriteMessage(messageType, comingCallJSON)
	// 			} else {
	// 				fmt.Println("Session not found in map")
	// 			}

	// 			if foundB && foundA {
	// 				// Send a message to sessionB
	// 				comingCallMessage := map[string]string{
	// 					"command":  "finish",
	// 					"sessionA": sessionA,
	// 					"sessionB": sessionB,
	// 					"message":  "all ended",
	// 				}
	// 				comingCallJSON, _ := json.Marshal(comingCallMessage)
	// 				clientB.WriteMessage(messageType, comingCallJSON)
	// 			} else {
	// 				fmt.Println("Session not found in map")
	// 			}
	// 		}
	// 	}
	// }))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Access-Control-Allow-Headers, Session, Mode",
		AllowMethods:     "GET, POST, PATCH, DELETE",
		AllowCredentials: true,
	}))
}
