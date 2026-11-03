package data

// Models implements UserStore-style access via GetByEmail on the embedded User model.
func (m Models) GetByEmail(email string) (*User, error) {
	return m.User.GetByEmail(email)
}
