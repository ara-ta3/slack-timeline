package slack

import (
	"time"

	"github.com/ara-ta3/slack-timeline/timeline"
	cache "github.com/patrickmn/go-cache"
)

func NewUserRepository(s SlackClient) UserRepositoryOnSlack {
	c := cache.New(cache.NoExpiration, 24*time.Hour)
	return UserRepositoryOnSlack{
		s,
		*c,
	}
}

type UserRepositoryOnSlack struct {
	SlackClient SlackClient
	cache       cache.Cache
}

func (r UserRepositoryOnSlack) GetAll() ([]timeline.User, error) {
	us, err := r.SlackClient.getAllUsers()
	if err != nil {
		return nil, err
	}
	users := []timeline.User{}
	for _, u := range us {
		r.cache.Set(u.ID, u, cache.NoExpiration)
		users = append(users, u.ToInternal())
	}

	return users, nil
}

func (r UserRepositoryOnSlack) Get(userID string) (timeline.User, error) {
	u, found := r.cache.Get(userID)
	ret, ok := u.(User)
	if found && ok {
		return ret.ToInternal(), nil
	}
	r.cache.Delete(userID)

	uu, err := r.SlackClient.getUser(userID)

	if err != nil {
		return timeline.User{}, err
	}

	r.cache.Set(userID, uu, cache.NoExpiration)
	return uu.ToInternal(), nil
}

func (r UserRepositoryOnSlack) Clear() error {
	r.cache.Flush()
	return nil
}
