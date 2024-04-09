package user

import (
	"context"
	"main/infrastructure"
	"main/services/cryptography"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService interface {
	GetUsers(id string, size string) []User
	PostUser(user User) User
}

var repository = infrastructure.NewRepo("users")

func GetUsers(id string, name string, size string) []User {
	var users []User = []User{}
	if id != "" || name != "" {
		var user User = GetUser(id, name)
		if user.Id != "" {
			users = append(users, user)
		}
	} else {
		options := options.Find()
		var limit int64
		if size == "" {
			limit = 5
		} else {
			limit, _ = strconv.ParseInt(size, 10, 64)
		}
		options.SetLimit(limit)
		options.SetSort(bson.M{"$natural": -1})
		list, err := repository.Find(context.TODO(), bson.D{{}}, options)
		if err != nil && err != mongo.ErrNoDocuments {
			panic(err)
		} else {
			for list.Next(context.TODO()) {
				var currentUser User
				err := list.Decode(&currentUser)
				if err != nil {
					panic(err)
				}
				users = append(users, currentUser)
			}
		}
	}
	return users
}

func GetUser(id string, name string) User {
	var user User = User{}
	if id != "" {
		filterWithId := bson.D{{Key: "id", Value: id}}
		cur, err := repository.FindOne(context.TODO(), filterWithId, nil)
		if cur != nil && err == nil {
			cur.Decode(&user)
		}
	}
	if name != "" {
		filterWithName := bson.D{{Key: "name", Value: name}}
		cur, err := repository.FindOne(context.TODO(), filterWithName, nil)
		if cur != nil && err == nil {
			cur.Decode(&user)
		}
	}
	return user
}

func PostUser(user User) User {
	checkUserValidity(user)
	builtUser := buildUser(user)
	repository.InsertOne(context.TODO(), builtUser)
	return builtUser
}

// Pre user registration:
func checkUserValidity(user User) {
	filter := bson.D{{Key: "name", Value: user.Name}}
	cur, _ := repository.FindOne(context.TODO(), filter, nil)
	if cur != nil {
		var duplicateUser User
		cur.Decode(&duplicateUser)
		panic("A user with the same name exists.")
	}
}

func buildUser(user User) User {
	// User id generation
	var lastRecord User = User{}
	var newUserId int
	options := options.FindOne().SetSort(bson.M{"$natural": -1})
	res, err := repository.FindOne(context.TODO(), bson.M{}, options)
	if res == nil && err == nil {
		// No user is found in the DB,
		// Generate a default id:
		newUserId = 12345
	} else {
		res.Decode(&lastRecord)
		newUserId, _ = strconv.Atoi(lastRecord.Id)
	}
	return CreateUser(
		WithId(strconv.Itoa(newUserId+1)),
		WithName(user.Name),
		WithPassword(user.Password),
		WithRole(user.Role),
		WithSignature(cryptography.CreateDefaultCrypto(
			"keyPair",
			user.Name,
			user.Role)),
		WithActions(nil),
		WithAudit(user.Audit))
}

func PutUser(id string, user User) User {
	return user

}

func DeleteUser(id string) {

}
