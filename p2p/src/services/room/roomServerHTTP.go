package room

import (
	"context"
	"main/infrastructure"
	"main/services/audit"
	"main/services/cryptography"
	"slices"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
 * Basic messaging service on HTTP.
 * This implementation uses permanent persistence of room instances.
 * Whereas in P2P implementation, a room can only be run by a peer.
 * The peers must be part of the room members to be able to host the
 * room and send messages.
 */

type RoomServerHTTP interface {
	ReceiveMessages(id string, userId string, size string) []Message
	SendAMessage(id string, userId string, message string) Message
	DeleteAMessage(id string, userId string, message string) Message
}

var messageRepository = infrastructure.NewRepo("messaging")
var cachedRoom Room

func ReceiveMessages(id string, size string, userId string) []Message {
	var messages []Message = []Message{}
	var room Room = fetchTargetRoom(id)
	// Check if the user is in the room:
	if !validateUserRoomAuth(room, userId) {
		return messages
	}

	// Retrieve messages in the room:
	options := options.Find()
	var limit int64
	if size == "" {
		limit = 5
	} else {
		limit, _ = strconv.ParseInt(size, 10, 64)
	}
	options.SetLimit(limit)
	options.SetSort(bson.M{"$natural": -1})
	filter := bson.D{{Key: "roomId", Value: id}}
	list, err := messageRepository.Find(context.TODO(), filter, options)
	if err != nil && err != mongo.ErrNoDocuments {
		panic(err)
	} else {
		for list.Next(context.TODO()) {
			var currentMessage Message
			err := list.Decode(&currentMessage)
			if err != nil {
				panic(err)
			}
			currentMessage.Text = cryptography.Decrypt(currentMessage.Text, room.RoomMasterKey)
			messages = append(messages, currentMessage)
		}
	}
	return messages
}

func SendAMessage(id string, userId string, message Message) Message {
	// Check if the user is in the room:
	var room Room = fetchTargetRoom(id)
	if !validateUserRoomAuth(room, userId) {
		return CreateDefaultMessage()
	}

	// Build and send the message:
	var builtMessage Message = buildAMessage(room, userId, message)
	messageRepository.InsertOne(context.TODO(), builtMessage)
	return builtMessage
}

func fetchTargetRoom(id string) Room {
	if id == cachedRoom.Id {
		return cachedRoom
	}
	filter := bson.D{{Key: "id", Value: id}}
	cur, err := repository.FindOne(context.TODO(), filter, nil)
	if cur != nil && err == nil {
		cur.Decode(&cachedRoom)
	}
	return cachedRoom
}

func validateUserRoomAuth(room Room, userId string) bool {
	return slices.Contains(room.Members, userId)
}

// TODO change
func buildAMessage(room Room, userId string, message Message) Message {
	// Generate an id for message:
	var lastRecord Message = Message{}
	var newMessageId int
	options := options.FindOne().SetSort(bson.M{"$natural": -1})
	res, err := messageRepository.FindOne(context.TODO(), bson.M{}, options)
	if res == nil && err == nil {
		// No message is found in the DB,
		// Generate a default id:
		newMessageId = 100000
	} else {
		res.Decode(&lastRecord)
		newMessageId, _ = strconv.Atoi(lastRecord.Id)
	}
	return CreateMessage(
		WithMessageId(strconv.Itoa(newMessageId+1)),
		WithUserId(userId),
		WithRoomId(room.Id),
		WithText(cryptography.Encrypt(message.Text, room.RoomMasterKey)),
		WithMessageSignature(nil),
		WithMessageAudit(audit.CreateAuditForMessage()))
}
