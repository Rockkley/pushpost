package dto

type RegisterUserDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponseDto struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
