package htv_api

import (
	"context"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input CreateUser) (*User, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateUser(ctx context.Context, input UpdateUser) (*User, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (*User, error) {
	panic("not implemented")
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
	panic("not implemented")
}
func (r *queryResolver) ReadApp(ctx context.Context, id string) (*Application, error) {
	panic("not implemented")
}
func (r *queryResolver) ReadForm(ctx context.Context, id string) (*Form, error) {
	panic("not implemented")
}
