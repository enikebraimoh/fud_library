package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fud_library/utils"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func SendOTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	request.Header.Add("Accept", "text/plain")
	request.Header.Add("AppId", "62082f3c7e7b1300341e0a27")
	request.Header.Add("Authorization", "test_pk_edG8flS7COjYXkdsaTBdhoQAZ")

	// get request URL
	//regURL, _ := url.Parse("https://sandbox.dojah.io/api/v1/messaging/otp")

	//var error OtpError
	// create request body

	var otpRequest SendOTPRequest

	err := utils.ParseJSONFromRequest(request, &otpRequest)

	jsonValue, _ := json.Marshal(otpRequest)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	otpRequest.Channel = "sms"
	otpRequest.Length = 4

	req, _ := http.NewRequest("POST", "https://sandbox.dojah.io/api/v1/messaging/otp", bytes.NewBuffer(jsonValue))

	req.Header.Set("content-type", "application/json")
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Authorization", "test_sk_cGXPTa7okHGOr6NIR35qh3ntB")
	req.Header.Set("AppId", "62082f3c7e7b1300341e0a27")

	client := &http.Client{}
	res, err := client.Do(req)

	var errorr OtpError

	// check for response error
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
	} else {

		data, _ := ioutil.ReadAll(res.Body)
		err := json.Unmarshal(data, &errorr)

		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}

		if res.StatusCode != 200 {
			utils.GetError(fmt.Errorf(errorr.Errors), res.StatusCode, response)
			return
		} else {

			var otpres SuccessOTPResponse
			var user User

			data, _ := ioutil.ReadAll(res.Body)
			err := json.Unmarshal(data, &otpres)

			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			user.PhoneNumber = otpRequest.Destination
			user.Reference_id = otpres.Entity[1].ReferenceID

			user.FirstName = ""
			user.LastName = ""
			user.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

			detail, _ := utils.StructToMap(user)

			res, err := utils.CreateMongoDBDoc(UserCollectionName, detail)
			_ = res
			if err != nil {
				utils.GetError(err, http.StatusInternalServerError, response)
				return
			}

			utils.GetSuccess("OTP verification sent", errorr, response)

		}

	}

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
