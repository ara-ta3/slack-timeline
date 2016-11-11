package timeline

import "fmt"

type UserRepositoryOnMemory struct {
	data map[string]User
}

func (r UserRepositoryOnMemory) Get(userID string) (User, error) {
	u, found := r.data[userID]
	if found {
		return u, nil
	}
	return User{}, fmt.Errorf("not found")
}

func (r UserRepositoryOnMemory) GetAll() ([]User, error) {
	vs := []User{}
	for _, v := range r.data {
		vs = append(vs, v)
	}
	return vs, nil
}

func (r UserRepositoryOnMemory) Clear() error {
	r.data = map[string]User{}
	return nil
}
