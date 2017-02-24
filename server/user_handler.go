package server

import (
	"fmt"

	"github.com/nporsche/gapidemo/userapi"
	"golang.org/x/net/context"
)

type UserHandler struct {
}

func (u *UserHandler) GetUser(ctx context.Context, req *userapi.GetUserRequest) (resp *userapi.GetUserResponse, err error) {
	fmt.Println("request comming")
	return &userapi.GetUserResponse{
		User: &userapi.User{
			Id:   req.Id,
			Name: "hangchen",
		},
	}, nil
}
