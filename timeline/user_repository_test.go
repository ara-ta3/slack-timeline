package timeline

import (
	"fmt"

	"../slack"
)

type UserRepositoryOnMemory struct {
	data map[string]slack.User
}

func (r UserRepositoryOnMemory) Get(userID string) (slack.User, error) {
	u, found := r.data[userID]
	if found {
		return u, nil
	}
	return slack.User{}, fmt.Errorf("not found")
}

func (r UserRepositoryOnMemory) GetAll() ([]slack.User, error) {
	vs := []slack.User{}
	for _, v := range r.data {
		vs = append(vs, v)
	}
	return vs, nil
}

func (r UserRepositoryOnMemory) Clear() error {
	r.data = map[string]slack.User{}
	return nil
}
