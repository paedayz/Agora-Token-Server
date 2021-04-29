package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	"github.com/gin-gonic/gin"
)

var appID, appCertificate string

func main() {
	appIDEnv := ""
	appCertEnv := ""
	appIDEnv = os.Getenv("APP_ID")
	appCertEnv = os.Getenv("APP_CERTIFICATE")

	if appIDEnv != "" || appCertEnv != "" {
		log.Fatal("FATAL ERROR: ENV not properly configured, check APP_ID and APP_CERTIFICATE")
	} else {
		appID = appIDEnv
		appCertificate = appCertEnv
	}

	api := gin.Default()

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pae",
		})
	})

	api.GET("rtc/:channelName/:role/:tokenType/:uid/", getRtcToken)
	api.GET("rtm/:uid/", getRtmToken)
	api.GET("rte/:channelName/:role/:tokenType/:uid/", getBothTokens)

	api.Run(":8080")
}

func getRtcToken(c *gin.Context) {
	log.Printf("rtc token\n")

	// Get param values
	channelName, tokentype, uidStr, role, expireTimestamp, err := parseRtcParams(c)

	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"message": "Error Generating RTC token: " + err.Error(),
			"status":  400,
		})
		return
	}

	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokentype, role, expireTimestamp)

	if tokenErr != nil {
		log.Println(tokenErr) // token failed to generate
		c.Error(tokenErr)
		errMsg := "Error Generating RTC token - " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error":  errMsg,
		})
	} else {
		log.Println("RTC Token generated")
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
		})
	}
}

func getRtmToken(c *gin.Context) {

}

func getBothTokens(c *gin.Context) {

}

func parseRtcParams(c *gin.Context) (channelName, tokentype, uidStr string, role rtctokenbuilder.Role, expireTimestamp uint32, err error) {
	// Get param values
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokentype = c.Param("tokenType")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry", "3600")

	if roleStr == "publisher" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)

	if parseErr != nil {
		// if string conversion fails return an error
		err = fmt.Errorf("failed to parse expireTime: %s, causing error: %s", expireTime, parseErr)
	}

	// Set timestamps
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return channelName, tokentype, uidStr, role, expireTimestamp, err
}

// func parseRtmParams(c *gin.Context) (uidStr string, expireTimestamp uint32, err error) {

// }

func generateRtcToken(channelName, uidStr, tokentype string, role rtctokenbuilder.Role, expireTimestamp uint32) (rtcToken string, err error) {
	if tokentype == "userAccount" {
		log.Printf("Building Token with userAcoount: %s\n", uidStr)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUserAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
		return rtcToken, err
	} else if tokentype == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		// Check if conversion fails
		if parseErr != nil {
			err = fmt.Errorf("failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
			return "", err
		}

		uid := uint32(uid64) // convert uid from uint64 to uint32
		log.Printf("Building Token with uid: %d\n", uid)
		rtcToken, err = rtctokenbuilder.BuildTokenWithUserAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for Unknown Tokentype: %s", tokentype)
		log.Println(err)
		return "", err
	}
}
