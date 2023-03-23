package controller

import (
	"crypto/rand"
	"fmt"
	"github.com/its-me-debk007/QWait_backend/util"
	"log"
	"math"
	"math/big"
	"net/http"
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

	c.Redirect(http.StatusPermanentRedirect, fmt.Sprintf("https://qwait.netlify.app?token=%v", accessToken))
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

	//intPhoneNo, _ := strconv.Atoi(phoneNo)
	//int64PhoneNo := int64(intPhoneNo)

	flag := false
	idx, minLen := -1, math.MaxInt64
	for i, counter := range store.Customers {

		if minLen > len(counter) {
			flag = true
			minLen = len(counter)
			idx = i
		}

		splitArray := strings.Split(counter, ",")

		for _, no := range splitArray {
			if no == phoneNo {
				flag = true
				break
			}
		}

		if flag {
			break
		}
	}

	if flag {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.Message{Message: "already joined"})
		return
	}

	store.Customers[idx] += "," + phoneNo
	//store.WaitingTime = store.AvgTimePerPerson * len(store.Customers)
	store.WaitingTime = -1

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

	flag, idxI, idxJ := false, -1, -1
	for i := range store.Customers {
		splitArray := strings.Split(store.Customers[i], ",")

		for j := range splitArray {
			if splitArray[j] == phoneNo {
				flag = true
				idxI = i
				idxJ = j
				break
			}
		}

		if flag {
			break
		}
	}

	if !flag {
		c.AbortWithStatusJSON(http.StatusNotFound, model.Message{Message: "not joined queue"})
		return
	}

	//store.Customers[idxI] = append(store.Customers[idxI][:idxJ], store.Customers[idxI][idxJ+1:]...)
	newSplitArray := strings.Split(store.Customers[idxI], ",")
	newSplitArray = append(newSplitArray[:idxJ], newSplitArray[idxJ+1:]...)

	store.Customers[idxI] = strings.Join(newSplitArray, ",")

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

	var stores, joinedStores, hospitalStores, bankStores, regCampStores, govtOfficeStores, ticketSystemStores []model.Store
	database.DB.Find(&stores)

	for _, store := range stores {
		if dist := util.GetDistanceInKm(store.Latitude, store.Longitude, input.Latitude, input.Longitude); dist < util.NEAR_BY_DISTANCE {

			switch store.Category {

			case "hospital":
				hospitalStores = append(hospitalStores, store)

			case "bank":
				bankStores = append(bankStores, store)

			case "registration_camp":
				regCampStores = append(regCampStores, store)

			case "govt_office":
				govtOfficeStores = append(govtOfficeStores, store)

			case "ticketing_system":
				ticketSystemStores = append(ticketSystemStores, store)
			}
		}

		flag := false
		for _, counter := range store.Customers {
			splitArray := strings.Split(counter, ",")

			for _, no := range splitArray {
				if no == phoneNo {
					flag = true
					break
				}
			}

			if flag {
				break
			}
		}

		if flag {
			joinedStores = append(joinedStores, store)
		}
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"joined_stores":        joinedStores,
		"hospital_stores":      hospitalStores,
		"bank_stores":          bankStores,
		"reg_camp_stores":      regCampStores,
		"govt_office_stores":   govtOfficeStores,
		"ticket_system_stores": ticketSystemStores,
	})
}
