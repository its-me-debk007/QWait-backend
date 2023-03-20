package controller

import (
	"crypto/rand"
	"fmt"
	"github.com/its-me-debk007/QWait_backend/util"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/its-me-debk007/QWait_backend/database"
	"github.com/its-me-debk007/QWait_backend/model"
)

func Signup(c *gin.Context) {
	input := new(struct {
		PhoneNo string `json:"phone_no"    binding:"required"`
	})

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	bigIntToken, _ := rand.Int(rand.Reader, big.NewInt(9000000000000))
	bigIntToken.Add(bigIntToken, big.NewInt(1000000000000))

	token := bigIntToken.Int64()

	url := fmt.Sprintf("%s/api/v1/auth/verify?token=%v", c.Request.Host, token)
	log.Println(fmt.Sprintf("URL Is %s", url))

	util.SendSms(input.PhoneNo, url)

	if input.PhoneNo[0] == '+' {
		input.PhoneNo = input.PhoneNo[3:]
	} else if input.PhoneNo[0] == '0' {
		input.PhoneNo = input.PhoneNo[1:]
	}

	user := model.User{
		PhoneNo: input.PhoneNo,
		VerCode: token,
	}

	if err := database.DB.Save(&user); err.Error != nil {
		log.Println("\n" + err.Error.Error() + "\n")

		// c.AbortWithStatusJSON(http.StatusBadRequest, model.Message{"Error"})
		// return
	}

	c.JSON(http.StatusOK, model.Message{Message: "link sent"})
}

func Verify(c *gin.Context) {

	query := c.Query("token")
	query = strings.TrimSpace(query)

	var user model.User

	if db := database.DB.First(&user, "ver_code = ?", query); db.Error != nil {
		c.String(http.StatusNotFound, "Invalid Link :(")
		return
	}

	accessToken, _ := util.GenerateToken(user.PhoneNo, "User", 96)

	log.Printf("Access Token generated:- %s \n", accessToken)

	// TODO: send access token to website!

	//location := fmt.Sprintf("?access=%s", accessToken)

	c.Redirect(http.StatusPermanentRedirect, "https://www.google.co.in")
}

func JoinQueue(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: "no token provided"})
		return
	}

	token := header[7:]

	phoneNo, err := util.ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: err.Error()})
		return
	}

	input := new(struct {
		Latitude  float64 `json:"latitude"    binding:"required"`
		Longitude float64 `json:"longitude"   binding:"required"`
	})

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var store model.Store

	id := c.Param("id")
	if db := database.DB.First(&store, "id = ?", id); db.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, model.Message{Message: "no store found"})
		return
	}

	dist := util.GetDistanceInKm(store.Latitude, store.Longitude, input.Latitude, input.Longitude)

	log.Println(dist)

	if dist > util.NEAR_BY_DISTANCE {
		c.AbortWithStatusJSON(http.StatusForbidden, model.Message{Message: "customer not around 2.5 km from counter"})
		return
	}

	intPhoneNo, _ := strconv.Atoi(phoneNo)
	int64PhoneNo := int64(intPhoneNo)

	flag := false
	for _, v := range store.Customers {
		if v == int64PhoneNo {
			flag = true
			break
		}
	}

	if flag {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: "already joined"})
		return
	}

	store.Customers = append(store.Customers, int64PhoneNo)
	store.WaitingTime = store.AvgTimePerPerson * len(store.Customers)

	if err := database.DB.Save(&store); err.Error != nil {
		log.Println("\n" + err.Error.Error() + "\n")

		// c.AbortWithStatusJSON(http.StatusBadRequest, model.Message{"Error"})
		// return
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"avg_time_per_person": store.AvgTimePerPerson,
		"customers":           store.Customers,
	})
}

func LeaveQueue(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: "no token provided"})
		return
	}

	token := header[7:]

	phoneNo, err := util.ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: err.Error()})
		return
	}

	var store model.Store

	id := c.Param("id")
	if db := database.DB.First(&store, "id = ?", id); db.Error != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, model.Message{Message: "no store found"})
		return
	}

	flag, idx := false, -1
	for i, v := range store.Customers {
		no, _ := strconv.ParseInt(phoneNo, 10, 64)
		if v == no {
			flag = true
			idx = i
			break
		}
	}

	if !flag {
		c.AbortWithStatusJSON(http.StatusNotFound, model.Message{Message: "not joined queue"})
		return
	}

	store.Customers = append(store.Customers[:idx], store.Customers[idx+1:]...)
	if err := database.DB.Save(&store); err.Error != nil {
		log.Println("\n" + err.Error.Error() + "\n")

		// c.AbortWithStatusJSON(http.StatusBadRequest, model.Message{"Error"})
		// return
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"avg_time_per_person": store.AvgTimePerPerson,
		"customers":           store.Customers,
	})
}

func Home(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: "no token provided"})
		return
	}

	token := header[7:]

	phoneNo, err := util.ParseToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: err.Error()})
		return
	}

	input := new(struct {
		Latitude  float64 `json:"latitude"    binding:"required"`
		Longitude float64 `json:"longitude"   binding:"required"`
	})

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var stores, nearStores, joinedStores []model.Store
	database.DB.Find(&stores)

	for _, store := range stores {
		if dist := util.GetDistanceInKm(store.Latitude, store.Longitude, input.Latitude, input.Longitude); dist < util.NEAR_BY_DISTANCE {
			nearStores = append(nearStores, store)
		}

		flag := false
		for _, v := range store.Customers {
			no, _ := strconv.ParseInt(phoneNo, 10, 64)
			if v == no {
				flag = true
				break
			}
		}

		if flag {
			joinedStores = append(joinedStores, store)
		}
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"joined_stores": joinedStores,
		"near_stores":   nearStores,
	})
}
