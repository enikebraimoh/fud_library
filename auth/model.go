package auth

import "time"

const (
	UserCollectionName = "user"
)

type User struct {
	ID           string    `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName    string    `json:"first_name" bson:"first_name"`
	LastName     string    `json:"last_name" bson:"last_name"`
	PhoneNumber  string    `json:"phone_number" bson:"phone_number"`
	Reference_id string    `json:"reference_id" bson:"reference_id"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	EditedAt     time.Time `json:"edited_at" bson:"edited_at"`
	DeletedAt    time.Time `json:"deleted_at" bson:"deleted_at"`
}

type OTP struct {
	PhoneNumber string `json:"phone_number" bson:"phone_number" validate:"required"`
}

type Verify struct {
	PhoneNumber string `json:"phone_number" bson:"phone_number" validate:"required"`
	OtpCode     string `json:"otp_code" bson:"otp_code"`
}

type OtpError struct {
	Errors string `json:"error" bson:"error"`
}
