package auth

import (
	"fmt"
	"fud_library/utils"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SendOTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var otp OTP
	var user User

	err := utils.ParseJSONFromRequest(request, &otp)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	phone_number := otp.PhoneNumber

	user.PhoneNumber = phone_number
	user.FirstName = ""
	user.LastName = ""
	user.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDBDoc(UserCollectionName, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	insertedPostID := res.InsertedID.(primitive.ObjectID).Hex()

	_ = insertedPostID

	utils.GetSuccess("User created", bson.M{}, response)
}

func VerifyOTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var verifyotp Verify

	err := utils.ParseJSONFromRequest(request, &verifyotp)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	phone_number := verifyotp.PhoneNumber
	myotp := verifyotp.OtpCode

	_ = myotp

	save, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": phone_number})

	if save == nil {
		utils.GetError(fmt.Errorf("organization %s not found", phone_number), http.StatusNotFound, response)
		return
	}

	utils.GetSuccess("User Found!", bson.M{}, response)
}
