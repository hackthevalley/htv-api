package htv_api

//go:generate go run github.com/99designs/gqlgen

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/hackthevalley/htv-api/database"
	"github.com/hackthevalley/htv-api/utils"
	"github.com/mitchellh/mapstructure"
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
func MapDate(dateInp *DateInput) *Date {
	return &Date{
		Day:   dateInp.Day,
		Month: dateInp.Month,
		Year:  dateInp.Year,
	}
}
func MapQuestions(quesInp []*QuestionInput) []*Question {
	var questionArr []*Question
	for _, elem := range quesInp {
		ques := Question{
			ID:       elem.ID,
			Title:    elem.Title,
			Info:     elem.Info,
			Options:  elem.Options,
			Default:  elem.Default,
			Type:     elem.Type,
			Required: elem.Required,
		}
		questionArr = append(questionArr, &ques)
	}
	return questionArr
}

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(dbctx context.Context, input CreateUser) (*User, error) {
	user := User{}
	gc, err := utils.GinContextFromContext(dbctx)
	if err != nil {
		log.Printf("Error extracting context: %s", err)
		return &user, Errorf("Error extracting context")
	}
	s := sessions.Default(gc)
	authToken := s.Get("htv-token")
	userFilter := &bson.M{"email": input.Email, "sessionID": authToken}
	var userMap bson.M
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = database.DbClient.Collection("users").FindOne(dbctx, userFilter).Decode(&userMap)
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
	res := database.DbClient.Collection("users").FindOneAndUpdate(dbctx, userFilter, bson.M{
		"$set": bson.M{
			"profile": &user,
		},
	})
	if res.Err() != nil {
		log.Printf("Could not insert user into database: %v", res.Err().Error())
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
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userFilter := &bson.M{"email": input.Email, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(dbctx, userFilter).Decode(&userMap)
	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &user, Errorf("Error decoding user map")
	}
	var u User
	err = mapstructure.Decode(userMap["profile"], &u)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &user, Errorf("Error converting user mongo document to struct")
	}
	log.Printf("Retrieved profile for user: %v", u.Email)

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
		CreatedAt: u.CreatedAt,
	}
	res := database.DbClient.Collection("users").FindOneAndUpdate(dbctx, userFilter, bson.M{
		"$set": bson.M{
			"profile": &user,
		},
	})
	if res.Err() != nil {
		log.Printf("Could not update user into database: %v", res.Err().Error())
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
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userFilter := &bson.M{"_id": searchID, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(dbctx, userFilter).Decode(&userMap)

	if err != nil {
		log.Printf("Error decoding user map or search id is wrong: %s", err)
		return &user, Errorf("Error decoding user map")
	}
	var u User
	err = mapstructure.Decode(userMap["profile"], &u)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &user, Errorf("Error converting user mongo document to struct")
	}
	log.Printf("Retrieved profile for user: %v", u.Email)

	res := database.DbClient.Collection("users").FindOneAndDelete(dbctx, userFilter)
	if res.Err() != nil {
		log.Printf("Could not delete user from database: %v", res.Err().Error())
		return &user, Errorf("Error deleting user in database")
	}
	log.Printf("Deleted user %s from database: %v", u.Email, res)
	return &u, err
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
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	createdAt := time.Now()
	newForm := Form{
		ID:        "",
		Title:     input.Title,
		Questions: []*Question{},
		Open:      input.Open,
		EndsAt:    MapDate(input.EndsAt),
		CreatedAt: &Date{
			Day:   createdAt.Day(),
			Month: int(createdAt.Month()),
			Year:  createdAt.Year(),
		},
	}
	res, err := database.DbClient.Collection("forms").InsertOne(dbctx, newForm)
	if err != nil {
		log.Printf("Could not insert form in database: %v", err)
		return &Form{}, Errorf("Error inserting form into database")
	}
	insertedID := res.InsertedID.(primitive.ObjectID).Hex()
	log.Printf("Inserted form: %v", insertedID)
	formFilter := &bson.M{"_id": res.InsertedID}
	upres := database.DbClient.Collection("forms").FindOneAndUpdate(dbctx, formFilter, bson.M{
		"$set": bson.M{
			"id": res.InsertedID.(primitive.ObjectID).Hex(),
		},
	})
	if upres.Err() != nil {
		log.Printf("Could not update form id with object id in database: %v", upres.Err().Error())
		return &Form{}, Errorf("Error inserting form into database")
	}
	newForm.ID = insertedID
	return &newForm, err
}
func (r *mutationResolver) UpdateForm(ctx context.Context, input UpdateForm) (*Form, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searchID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		log.Printf("Error decoding user id: %s", err)
		return &Form{}, Errorf("Error decoding user id")
	}

	formFilter := &bson.M{"_id": searchID}
	var formMap bson.M
	err = database.DbClient.Collection("forms").FindOne(dbctx, formFilter).Decode(&formMap)

	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &Form{}, Errorf("Error decoding user map")
	}
	var form Form
	err = mapstructure.Decode(formMap, &form)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &Form{}, Errorf("Error converting user mongo document to struct")
	}
	formUpdate := Form{
		ID:        form.ID,
		Title:     input.Title,
		Questions: MapQuestions(input.Questions),
		Open:      input.Open,
		EndsAt:    MapDate(input.EndsAt),
		CreatedAt: form.CreatedAt,
	}
	res := database.DbClient.Collection("forms").FindOneAndUpdate(dbctx, formFilter, bson.M{
		"$set": &formUpdate,
	})
	if res.Err() != nil {
		log.Printf("Could not update form in database: %v", res.Err().Error())
		return &Form{}, Errorf("Could not update form in database")
	}
	return &formUpdate, err
}
func (r *mutationResolver) DeleteForm(ctx context.Context, id string) (*Form, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searchID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Error decoding user id: %s", err)
		return &Form{}, Errorf("Error decoding user id")
	}

	formFilter := &bson.M{"_id": searchID}
	var formMap bson.M
	err = database.DbClient.Collection("forms").FindOneAndDelete(dbctx, formFilter).Decode(&formMap)

	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &Form{}, Errorf("Error decoding user map")
	}
	var deletedForm Form
	err = mapstructure.Decode(formMap, &deletedForm)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &Form{}, Errorf("Error converting user mongo document to struct")
	}
	return &deletedForm, err
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) ReadUser(ctx context.Context, email *string, id *string) (*User, error) {
	user := User{}
	gc, err := utils.GinContextFromContext(ctx)
	if err != nil {
		log.Printf("Error extracting context: %s", err)
		return &user, Errorf("Error extracting context")
	}
	s := sessions.Default(gc)
	authToken := s.Get("htv-token")
	searchID, err := primitive.ObjectIDFromHex(*id)
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	userFilter := &bson.M{"email": *email, "_id": searchID, "sessionID": authToken}
	var userMap bson.M
	err = database.DbClient.Collection("users").FindOne(dbctx, userFilter).Decode(&userMap)

	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &user, Errorf("Error decoding user map")
	}
	var u User
	err = mapstructure.Decode(userMap["profile"], &u)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &user, Errorf("Error converting user mongo document to struct")
	}
	log.Printf("Retrieved profile for user: %v", u.Email)

	return &u, err
}
func (r *queryResolver) ReadApp(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *queryResolver) ReadForm(ctx context.Context, id string) (*Form, error) {
	dbctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searchID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Error decoding user id: %s", err)
		return &Form{}, Errorf("Error decoding user id")
	}

	formFilter := &bson.M{"_id": searchID}
	var formMap bson.M
	err = database.DbClient.Collection("forms").FindOne(dbctx, formFilter).Decode(&formMap)

	if err != nil {
		log.Printf("Error decoding user map: %s", err)
		return &Form{}, Errorf("Error decoding user map")
	}
	var retrievedForm Form
	err = mapstructure.Decode(formMap, &retrievedForm)
	if err != nil {
		log.Printf("Error converting user mongo document to struct: %s", err)
		return &Form{}, Errorf("Error converting user mongo document to struct")
	}
	return &retrievedForm, err
}
