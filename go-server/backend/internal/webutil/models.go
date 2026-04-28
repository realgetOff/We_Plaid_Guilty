package webutil

const USR_ID = " user ID = "
const JWT_ERROR = "Couldn't sign / generate JWT for user."

// FUNCTIONS //

type loginInfo struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email,omitempty"`
}
