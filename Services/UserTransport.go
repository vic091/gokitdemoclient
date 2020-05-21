package Services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func GetUserInfo_Request(ctx context.Context, request *http.Request, r interface{}) error {
	user_request := r.(UserRequest)
	request.URL.Path += "/user/" + strconv.Itoa(user_request.Uid)
	return nil
}

func GetUserInfo_Response(ctx context.Context, res *http.Response) (response interface{}, err error) {
	if res.StatusCode > 400 {
		return nil, errors.New("not data")
	}
	var user_response UserResponse
	err = json.NewDecoder(res.Body).Decode(&user_response)
	if err != nil {
		return nil, err
	}

	return user_response, nil
}
