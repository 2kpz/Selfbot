package client

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Activity represents the structure for setting a user's activity.
type Activity struct {
	Name       string `json:"name"`
	Type       int    `json:"type"` // 0 = Playing, 4 = Custom Status
	Details    string `json:"details"`
	State      string `json:"state"`
	LargeImage string `json:"large_image"` // Predefined image key
	LargeText  string `json:"large_text"`
}

// DiscordRPC manages the connection to Discord's Gateway.
type DiscordRPC struct {
	wsConn *websocket.Conn
	appID  string
	token  string
}

// New creates a new DiscordRPC instance.
func New(appID string) (*DiscordRPC, error) {
	if appID == "" {
		return nil, fmt.Errorf("appID cannot be empty")
	}
	return &DiscordRPC{
		appID: appID,
	}, nil
}

// Connect establishes a WebSocket connection to Discord's Gateway.
func (d *DiscordRPC) Connect(token string) error {
	if d.wsConn != nil {
		return fmt.Errorf("already connected to Discord Gateway")
	}
	d.token = token

	// No token in header
	header := http.Header{}

	var err error
	d.wsConn, _, err = websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=9&encoding=json", header)
	if err != nil {
		return fmt.Errorf("failed to connect to Discord Gateway: %v", err)
	}
	log.Println("Connected to Discord Gateway!")

	// Handle Hello event
	var hello struct {
		Op int `json:"op"`
		D  struct {
			HeartbeatInterval int `json:"heartbeat_interval"`
		} `json:"d"`
	}
	err = d.wsConn.ReadJSON(&hello)
	if err != nil {
		d.Close()
		return fmt.Errorf("failed to read Hello event: %v", err)
	}

	// Send Identify event (OP 2)
	identify := map[string]interface{}{
		"op": 2,
		"d": map[string]interface{}{
			"token": d.token,
			"properties": map[string]interface{}{
				"$os":      "linux",
				"$browser": "selfbot",
				"$device":  "selfbot",
			},
			"presence": map[string]interface{}{
				"activities": []interface{}{},
				"status":     "online",
				"since":      nil,
				"afk":        false,
			},
			"intents": 0,
		},
	}
	err = d.wsConn.WriteJSON(identify)
	if err != nil {
		d.Close()
		return fmt.Errorf("failed to send Identify event: %v", err)
	}

	// Start heartbeat loop
	if hello.D.HeartbeatInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(hello.D.HeartbeatInterval) * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				err := d.wsConn.WriteJSON(map[string]interface{}{"op": 1, "d": nil})
				if err != nil {
					log.Printf("Error sending heartbeat: %v", err)
					break
				}
			}
		}()
	}

	return nil
}

// SetActivity updates the user's activity on Discord.
func (d *DiscordRPC) SetActivity(activity Activity) error {
	if d.wsConn == nil {
		return fmt.Errorf("not connected to Discord Gateway")
	}
	if activity == (Activity{}) {
		return fmt.Errorf("empty activity")
	}

	payload := map[string]interface{}{
		"op": 3, // Opcode 3 for setting presence
		"d": map[string]interface{}{
			"since": nil,
			"activities": []map[string]interface{}{
				{
					"name":    activity.Name,
					"type":    activity.Type,
					"details": activity.Details,
					"state":   activity.State,
					"assets": map[string]interface{}{
						"large_image": activity.LargeImage,
						"large_text":  activity.LargeText,
					},
					"timestamps": map[string]interface{}{
						"start": time.Now().Unix(),
					},
				},
			},
			"status": "online",
			"afk":    false,
		},
	}

	err := d.wsConn.WriteJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to set activity: %v", err)
	}
	return nil
}

// Close closes the WebSocket connection.
func (d *DiscordRPC) Close() {
	if d.wsConn != nil {
		d.wsConn.Close()
		d.wsConn = nil
	}
}

// ConnectWithReconnect attempts to reconnect if the connection drops
func (d *DiscordRPC) ConnectWithReconnect(token string) {
	for {
		err := d.Connect(token)
		if err != nil {
			log.Printf("Connection failed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println("Connection closed. Reconnecting...")
		time.Sleep(5 * time.Second)
	}
}
