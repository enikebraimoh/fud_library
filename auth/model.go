package auth

import (
	"time"
)

const (
	UserCollectionName = "user"
)

type User struct {
	ID           string    `json:"_id,omitempty" bson:"id,omitempty"`
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

type SendOTPRequest struct {
	SenderID    string `json:"sender_id"`
	Destination string `json:"destination"`
	Channel     string `json:"channel"`
	Length      int32  `json:"length"`
}

type SuccessOTPResponse struct {
	Entity []Entity `json:"entity"`
}

type Entity struct {
	ReferenceID string `json:"reference_id"`
	Destination string `json:"destination"`
	StatusID    string `json:"status_id"`
	Status      string `json:"status"`
}

type VerifiedOTPResponse struct {
	Entity SuccessEntity `json:"entity"`
}

type SuccessEntity struct {
	Valid bool `json:"valid"`
}
