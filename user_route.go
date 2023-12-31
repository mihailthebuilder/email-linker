package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

func (r *Controller) SetAuthenticatedUser(c *gin.Context) {
	token, err := getTokenFromRequest(c)
	if err != nil {
		log.Println("error to fetch token:", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := r.TokenClient.GetUserId(token)
	if err != nil {
		log.Println("error to fetch user id from token:", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	exists, err := r.Database.UserIdExists(c, userId)
	if err != nil {
		log.Printf("error to check if user id %s exists: %s", userId, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Printf("user id %s does not exist", userId)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("userId", userId)
	c.Next()
}

func getTokenFromRequest(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if lengthOfString(header) == 0 {
		return "", fmt.Errorf("no Authorization header")
	}

	bearerSlice := strings.Split(header, " ")
	if len(bearerSlice) != 2 || bearerSlice[0] != "Bearer" || lengthOfString(bearerSlice[1]) == 0 {
		return "", fmt.Errorf("invalid Authorization header: %s", header)
	}

	return bearerSlice[1], nil
}

func (r *Controller) TrackLink(c *gin.Context) {
	userId, err := getUserIdFromContext(c)
	if err != nil {
		log.Println("error getting email from context: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	apiRequest := TrackLinkRequest{}
	err = c.ShouldBindJSON(&apiRequest)
	if err != nil {
		log.Println("error parsing API request: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	path := uniuri.NewLen(r.RedirectPathLength)

	url := addHttpsToUrlIfNotIncludedAlready(apiRequest.Url)

	redirectRequest := AddRedirectRequest{
		UserId: userId,
		Url:    url,
		Path:   path,
		Tag:    apiRequest.Tag,
	}

	err = r.Database.AddRedirect(c, redirectRequest)
	if err != nil {
		log.Printf("error adding redirect url path %s for email %s: %s", path, userId, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusCreated, fmt.Sprintf("%s/%s", r.RedirectUri, path))
}

func getUserIdFromContext(c *gin.Context) (string, error) {
	idContext, exists := c.Get("userId")

	if !exists {
		return "", fmt.Errorf("userId key not set in context")
	}

	id := idContext.(string)
	if id == "" {
		return "", fmt.Errorf("no value set for userId key in context")
	}

	return id, nil
}

type TrackLinkRequest struct {
	Url string `json:"url" binding:"required"`
	Tag string `json:"tag"`
}

const HTTP_PREFIX = "http://"
const HTTP_PREFIX_LENGTH = 7
const HTTPS_PREFIX = "https://"
const HTTPS_PREFIX_LENGTH = 8

func addHttpsToUrlIfNotIncludedAlready(url string) string {
	lengthOfUrl := lengthOfString(url)

	if lengthOfUrl <= HTTPS_PREFIX_LENGTH {
		return "https://" + url
	}

	if url[0:HTTPS_PREFIX_LENGTH] == HTTPS_PREFIX || url[0:HTTP_PREFIX_LENGTH] == HTTP_PREFIX {
		return url
	}

	return "https://" + url
}
