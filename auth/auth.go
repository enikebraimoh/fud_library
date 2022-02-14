package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fud_library/utils"
	"io/ioutil"
	"log"
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
	//var newresp []interface{}

	err := utils.ParseJSONFromRequest(request, &otpRequest)

	jsonValue, _ := json.Marshal(otpRequest)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	otpRequest.Channel = "sms"
	otpRequest.Length = 5

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

		if res.StatusCode == 200 {
			log.Println("otp sent")

			//json.Unmarshal(newres, &otpres)
			var otpres SuccessOTPResponse
			var user User

			//data, _ := ioutil.ReadAll(res.Body)
			err := json.Unmarshal(data, &otpres)
			//err := json.NewDecoder(data).Decode(&jsonData)

			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				log.Println("tryinng to connvert otp response to struct")
				return
			}

			user.PhoneNumber = otpRequest.Destination
			user.Reference_id = otpres.Entity[0].ReferenceID

			user.FirstName = ""
			user.LastName = ""
			user.CreatedAt = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().UTC().Hour(), time.Now().Minute(), time.Now().Second(), 0, time.Local)

			detail, _ := utils.StructToMap(user)

			res, err := utils.CreateMongoDBDoc(UserCollectionName, detail)
			_ = res

			if err != nil {
				utils.GetError(err, http.StatusInternalServerError, response)
				log.Println("tryinng to save to db")
				return
			}

			utils.GetSuccess("OTP verification sent", otpres, response)
			return
		}

		utils.GetError(fmt.Errorf(errorr.Errors), res.StatusCode, response)

	}

}

func VerifyOTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	request.Header.Add("Accept", "text/plain")
	request.Header.Add("AppId", "62082f3c7e7b1300341e0a27")
	request.Header.Add("Authorization", "test_pk_edG8flS7COjYXkdsaTBdhoQAZ")

	//err := utils.ParseJSONFromRequest(request, &verifyotp)

	code := request.FormValue("code")
	phone := request.FormValue("phone_number")

	_ = code

	save, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": phone})

	if save == nil {
		utils.GetError(fmt.Errorf("phone number %s not found", phone), http.StatusNotFound, response)
		return
	}

	dbuser, err := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": phone})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	var currentUser User

	bsonBytes, _ := bson.Marshal(dbuser)
	if err := bson.Unmarshal(bsonBytes, &currentUser); err != nil {
		return
	}

	ValidatePhone(phone, currentUser.Reference_id)

	url := fmt.Sprintf("https://sandbox.dojah.io/api/v1/messaging/otp/validate?code=%s&reference_id=%s", code, currentUser.Reference_id)

	req, _ := http.NewRequest("GET", url, nil)

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

			var otpres VerifiedOTPResponse
			var user User

			data, _ := ioutil.ReadAll(res.Body)
			err := json.Unmarshal(data, &otpres)

			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			detail, _ := utils.StructToMap(user)

			res, err := utils.CreateMongoDBDoc(UserCollectionName, detail)
			_ = res
			if err != nil {
				utils.GetError(err, http.StatusInternalServerError, response)
				return
			}

			utils.GetSuccess("OTP verified", errorr, response)

		}

	}
}

// check that a member belongs in the an organization.
func ValidatePhone(PhoneNumber, reference_id string) error {

	// check that member exists
	memberDoc, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": PhoneNumber, "reference_id": reference_id})
	if memberDoc == nil {
		fmt.Printf("phone number %s doesn't exist!", PhoneNumber)
		return errors.New("phone number does not exist")
	}

	return nil
}
