package htv_api

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/hackthevalley/htv-api/database"
	"github.com/hackthevalley/htv-api/utils"
	. "github.com/vektah/gqlparser/gqlerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

func MapLinks(links []*LinkInput) []*Link {
	var linkArr []*Link

	for _, elem := range links {
		link := Link{
			Label: elem.Label,
			URL:   elem.URL,
		}
		linkArr = append(linkArr, &link)
	}
	return linkArr
}

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input CreateUser) (*User, error) {
	user := User{}
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Printf("Error extracting context: %s", err)
		return &user, Errorf("Error extracting context")
	}
	s := sessions.Default(gc)
	authToken := s.Get("htv-token")
	userFilter := &bson.M{"email": input.Email, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(ctx, userFilter).Decode(&userMap)
	if err != nil {
		log.Printf("Error decoding user map or unauthorized action: %s", err)
		return &user, Errorf("Error decoding user map or unauthorized action")
	}
	log.Printf("UserMap: %v", userMap)
	timeStamp := time.Now()
	user = User{
		ID:        userMap["_id"].(primitive.ObjectID).Hex(),
		Links:     MapLinks(input.Links),
		Status:    "",
		Email:     input.Email,
		Firstname: *input.Firstname,
		Lastname:  *input.Lastname,
		Gender:    *input.Gender,
		School:    *input.School,
		Bio:       *input.Bio,
		Photo:     *input.Photo,
		CreatedAt: &Date{
			Day:   timeStamp.Day(),
			Month: int(timeStamp.Month()),
			Year:  timeStamp.Year(),
		},
	}
	res := database.DbClient.Collection("users").FindOneAndUpdate(ctx, userFilter, bson.M{
		"$set": bson.M{
			"profile": &user,
		},
	})
	if res.Err() != nil {
		log.Printf("Could not insert user into database: %v", err)
		return &user, Errorf("Error inserting user into database")
	}
	log.Printf("Inserted user to database: %v", res)
	return &user, err
}
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUser) (*User, error) {
	user := User{}
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Printf("Error extracting context: %s", err)
		return &user, Errorf("Error extracting context")
	}
	s := sessions.Default(gc)
	authToken := s.Get("htv-token")
	userFilter := &bson.M{"email": input.Email, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(ctx, userFilter).Decode(&userMap)
	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &user, Errorf("Error decoding user map")
	}
	log.Printf("UserMap: %v", userMap)
	//timestamp := userMap[""]
	timeStamp := time.Now()
	user = User{
		ID:        userMap["_id"].(primitive.ObjectID).Hex(),
		Links:     MapLinks(input.Links),
		Status:    *input.Status,
		Email:     *input.Email,
		Firstname: *input.Firstname,
		Lastname:  *input.Lastname,
		Gender:    *input.Gender,
		School:    *input.School,
		Bio:       *input.Bio,
		Photo:     *input.Photo,
		CreatedAt: &Date{
			Day:   timeStamp.Day(),
			Month: int(timeStamp.Month()),
			Year:  timeStamp.Year(),
		},
	}
	res := database.DbClient.Collection("users").FindOneAndUpdate(ctx, userFilter, bson.M{
		"$set": bson.M{
			"profile": &user,
		},
	})
	if res.Err() != nil {
		log.Printf("Could not update user into database: %v", err)
		return &user, Errorf("Error updating user into database")
	}
	log.Printf("Updated user %s to database: %v", *input.Email, res)
	return &user, err
}
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (*User, error) {
	user := User{}
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Printf("Error extracting context: %s", err)
		return &user, Errorf("Error extracting context")
	}
	s := sessions.Default(gc)
	authToken := s.Get("htv-token")
	searchID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Error decoding user id: %s", err)
		return &user, Errorf("Error decoding user id")
	}
	userFilter := &bson.M{"_id": searchID, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(ctx, userFilter).Decode(&userMap)
	if err != nil {
		log.Printf("Error decoding user map or search id is wrong: %s", err)
		return &user, Errorf("Error decoding user map")
	}
	log.Printf("UserMap: %v", userMap)
	//timestamp := userMap[""]
	timeStamp := time.Now()
	user = User{
		ID:        userMap["_id"].(primitive.ObjectID).Hex(),
		Links:     []*Link{},
		Status:    "",
		Email:     "",
		Firstname: "",
		Lastname:  "",
		Gender:    "",
		School:    "",
		Bio:       "",
		Photo:     "",
		CreatedAt: &Date{
			Day:   timeStamp.Day(),
			Month: int(timeStamp.Month()),
			Year:  timeStamp.Year(),
		},
	}
	res := database.DbClient.Collection("users").FindOneAndDelete(ctx, userFilter)
	if res.Err() != nil {
		log.Printf("Could not delete user from database: %v", err)
		return &user, Errorf("Error deleting user in database")
	}
	log.Printf("Deleted user %s to database: %v", userMap["email"].(string), res)
	return &user, err
}
func (r *mutationResolver) CreateApp(ctx context.Context, form string, user string) (*Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateApp(ctx context.Context, id string, responses []*ResponseInput) (*Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApp(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateForm(ctx context.Context, input CreateForm) (*Form, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateForm(ctx context.Context, input UpdateForm) (*Form, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteForm(ctx context.Context, id string) (*Form, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) ReadUser(ctx context.Context, email *string, id *string) (*User, error) {
	//user := User{}
	//userFilter := &bson.M{"email": *email, "_id": primitive.ObjectIDFromHex(*id)}
	//var userMap bson.M
	//err := database.DbClient.Collection("users").FindOne(ctx, userFilter).Decode(&userMap)
	//
	//if err != nil {
	//	log.Printf("Error decoding user map: %s", err)
	//	return &user, Errorf("Error decoding user map")
	//}
	//log.Printf("UserMap: %v", userMap)
	//
	////timestamp := userMap[""]
	//timeStamp := time.Now()
	//user = User{
	//	ID:        userMap["_id"].(primitive.ObjectID).Hex(),
	//	Links:     MapLinks(input.Links),
	//	Status:    *input.Status,
	//	Email:     *input.Email,
	//	Firstname: *input.Firstname,
	//	Lastname:  *input.Lastname,
	//	Gender:    *input.Gender,
	//	School:    *input.School,
	//	Bio:       *input.Bio,
	//	Photo:     *input.Photo,
	//	CreatedAt: &Date{
	//		Day:   timeStamp.Day(),
	//		Month: int(timeStamp.Month()),
	//		Year:  timeStamp.Year(),
	//	},
	//}
	//res := database.DbClient.Collection("users").FindOneAndUpdate(ctx, userFilter, bson.M{
	//	"$set": bson.M{
	//		"profile": &user,
	//	},
	//})
	//if res.Err() != nil {
	//	log.Printf("Could not update user into database: %v", err)
	//	return &user, Errorf("Error updating user into database")
	//}
	//log.Printf("Updated user %s to database: %v", *input.Email, res)
	//return &user, err
	panic("not implemented")
}
func (r *queryResolver) ReadApp(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *queryResolver) ReadForm(ctx context.Context, id string) (*Form, error) {
	panic("not implemented")
}
