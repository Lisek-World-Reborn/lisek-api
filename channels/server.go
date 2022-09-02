package channels

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Lisek-World-Reborn/lisek-api/logger"
)

type MinecraftRequest struct {
	UUID      string   `json:"uuid"`
	Target    string   `json:"target"`
	Arguments []string `json:"arguments"`
	Hash      string   `json:"hash"`
}

type ServerAddedRequest struct {
	ServerId int `json:"server_id"`
}

func (m MinecraftRequest) IsValid() bool {
	stringToHash := m.UUID + m.Target + strings.Join(m.Arguments, " ")

	return m.Hash == GenerateHash(stringToHash)
}

func GenerateHash(stringToHash string) string {
	md5hash := md5.New()

	md5hash.Write([]byte(stringToHash))

	return fmt.Sprintf("%x", md5hash.Sum(nil))
}

func (m MinecraftRequest) SendToServer(serverId string) (bool, error) {

	if !m.IsValid() {
		return false, fmt.Errorf("invalid request (hash mismatch)")
	}
	ctx := context.Background()

	body, err := json.Marshal(m)

	if err != nil {
		return false, err
	}

	cmd := RedisConnection.Publish(ctx, fmt.Sprintf("servers:%s:request", serverId), body)

	if cmd.Err() != nil {
		return false, cmd.Err()
	}

	return true, nil
}

func StartAcceptingRequests() {
	ctx := context.Background()

	pubsub := RedisConnection.Subscribe(ctx, "servers:*:request")

	defer pubsub.Close()
	for {

		msg, err := pubsub.ReceiveMessage(ctx)

		if err != nil {
			logger.Error(err.Error() + " - Channel listening ")
			continue
		}

		var m MinecraftRequest

		err = json.Unmarshal([]byte(msg.Payload), &m)

		if err != nil {
			logger.Error(err.Error())
			continue
		}

		if !m.IsValid() {
			logger.Error("invalid request (hash mismatch)")
			continue
		}
		logger.Info(fmt.Sprintf("received request from %s: %s", msg.Channel, msg.Payload))
	}
}
