package timeline

type User struct {
	ID              string
	Name            string
	ProfileImageURL string
}

func NewUser(id, name, profileImageURL string) User {
	return User{
		ID:              id,
		Name:            name,
		ProfileImageURL: profileImageURL,
	}
}
