package util

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/its-me-debk007/QWait_backend/database"
	"github.com/its-me-debk007/QWait_backend/model"
	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"
	"log"
	"math"
	"os"
	"time"
)

func SendSms(receiver string, url string) {

	client := twilio.NewRestClient()

	params := &api.CreateMessageParams{}
	params.SetBody(fmt.Sprintf("\nHi,\nWelcome to QWait!\nHere is your link for verification:\n%s", url))
	params.SetFrom("+15074364286")
	params.SetTo(receiver)

	resp, err := client.Api.CreateMessage(params)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		if resp.Sid != nil {
			fmt.Println(*resp.Sid)
		} else {
			fmt.Println(resp.Sid)
		}
	}
}

func GenerateToken(username string, subject string, expirationTime time.Duration) (string, error) {
	registeredClaims := jwt.RegisteredClaims{
		Issuer:  username,
		Subject: subject,
		ExpiresAt: &jwt.NumericDate{
			Time: time.Now().Add(time.Hour * expirationTime),
		},
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims)

	secretKey := os.Getenv("SECRET_KEY")

	token, err := claims.SignedString([]byte(secretKey))

	if err != nil {
		return token, err
	}

	log.Printf("Access Token is %s \n", token)

	return token, nil
}

func ParseToken(tokenString string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")

	registeredClaims := jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(tokenString, &registeredClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return "", errors.New("invalid token")
	}

	if db := database.DB.First(&model.User{}, "phone_no = ?", registeredClaims.Issuer); db.Error != nil {
		return "", errors.New("user not signed up")
	}

	if time.Since(registeredClaims.ExpiresAt.Time) >= 0 {
		return "", errors.New("token expired")
	}

	return registeredClaims.Issuer, nil
}

func GetDistanceInKm(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	var earthRadius = float64(6371) // Radius of the earth in km
	var dLat = deg2rad(lat2 - lat1) // deg2rad below
	var dLon = deg2rad(lon2 - lon1)
	var a = math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(deg2rad(lat1))*math.Cos(deg2rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := earthRadius * c // Distance in km

	return d
}

func deg2rad(deg float64) float64 {
	return deg * (math.Pi / 180)
}
