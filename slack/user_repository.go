package slack

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

func NewUserRepository(s SlackClient) UserRepositoryOnSlack {
	c := cache.New(cache.NoExpiration, 30*time.Minute)
	return UserRepositoryOnSlack{
		s,
		*c,
	}
}

type UserRepositoryOnSlack struct {
	SlackClient SlackClient
	cache       cache.Cache
}

func (r UserRepositoryOnSlack) Get(userID string) (User, error) {
	u, found := r.cache.Get(userID)
	ret, ok := u.(User)
	if found && ok {
		return ret, nil
	}
	r.cache.Delete(userID)

	uu, err := r.SlackClient.getUser(userID)

	if err != nil {
		return User{}, err
	}

	r.cache.Set(userID, uu, cache.NoExpiration)
	return *uu, nil
}

func (r UserRepositoryOnSlack) Clear() error {
	r.cache.Flush()
	return nil
}
