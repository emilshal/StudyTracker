package user

// Repository describes persistence behavior for users.
type Repository interface {
	Create(user User) (User, error)
	Update(user User) (User, error)
	GetByEmail(email string) (User, error)
	GetByID(id string) (User, error)
	GetByProvider(provider, providerID string) (User, error)
}
