package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func mhain() {

	response, err := http.Post("https://sandbox.dojah.io/api/v1/messaging/otp", "", nil)
	if err != nil {
		fmt.Printf("therer was an error", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Printf(string(data))
	}

}
