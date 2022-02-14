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

	otpRequest.Channel = "sms"
	value := 4
	otpRequest.Length = int32(value)
	log.Println(int32(value))

	jsonValue, _ := json.Marshal(otpRequest)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

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
			//	err := json.NewDecoder(data).Decode(&jsonData)

			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				log.Println("trying to connvert otp response to struct")
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
				log.Println("trying to save to db")
				return
			}

			utils.GetSuccess("OTP verification sent", nil, response)
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

	save, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": phone})

	if save == nil {
		utils.GetError(fmt.Errorf("phone number %s not found", phone), http.StatusNotFound, response)
		return
	}

	// dbuser, err := utils.GetMongoDBDoc(UserCollectionName, bson.M{"phone_number": phone})

	// if err != nil {
	// 	utils.GetError(err, http.StatusInternalServerError, response)
	// 	return
	// }

	var currentUser User

	bsonBytes, _ := bson.Marshal(save)

	if err := bson.Unmarshal(bsonBytes, &currentUser); err != nil {
		utils.GetError(fmt.Errorf("errror converting from database to struct"), http.StatusNotFound, response)
		return
	}

	log.Println(currentUser.ID)
	//ValidatePhone(phone, currentUser.Reference_id)

	ref := currentUser.Reference_id

	log.Println(ref)
	url := fmt.Sprintf("https://sandbox.dojah.io/api/v1/messaging/otp/validate?code=%s&reference_id=%s", code, ref)

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
			log.Println("error converting error response body")
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println(string(data))

		if res.StatusCode == 200 {
			var otpres VerifiedOTPResponse
			//err := json.NewDecoder(string(data)).Decode(&otpres)
			//err := json.NewDecoder(data).Decode(&jsonData)
			err := json.Unmarshal(data, &otpres)

			if err != nil {
				log.Println("error convertinng success responnd")
				log.Println(string(data))
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}

			if otpres.Entity.Valid {
				//detail, _ := utils.StructToMap(currentUser)

				result, err := utils.UpdateOneMongoDBDoc(UserCollectionName, currentUser.ID, bson.M{"reference_id": ""})
				log.Println(result.UpsertedID)

				if err != nil {
					log.Println("error updated user to saving to db")
					utils.GetError(err, http.StatusInternalServerError, response)
					return
				}

				if result.ModifiedCount == 0 {
					utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, response)
					return
				}

				utils.GetSuccess("User verified", nil, response)
				return
			} else {
				utils.GetError(fmt.Errorf("otp has expired or something"), http.StatusNotFound, response)
				return
			}

		} else {
			log.Println("response is not 200")
			utils.GetError(fmt.Errorf(errorr.Errors), res.StatusCode, response)
			return

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
