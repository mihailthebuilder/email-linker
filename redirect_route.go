package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Controller) Redirect(c *gin.Context) {
	path := c.Param("path")
	if lengthOfString(path) == 0 {
		log.Println("no path given")
		c.Redirect(http.StatusTemporaryRedirect, r.ErrorRedirectUrl)
		return
	}

	url, err := r.Database.GetRedirectUrl(c, path)
	if err != nil {
		log.Printf("fail to get redirect url for path %s: %s", path, err)
		c.Redirect(http.StatusTemporaryRedirect, r.ErrorRedirectUrl)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)

	record, err := r.Database.GetRedirectRecord(c, path)
	if err != nil {
		log.Printf("fail to get redirect data for path %s: %s", path, err)
		return
	}

	err = r.Database.AddLinkClick(c, record.LinkId)
	if err != nil {
		log.Printf("fail to add link click for link with id %s: %s", record.LinkId, err)
	}

	httpUserAgent := c.Request.Header.Get("User-Agent")
	log.Printf("HTTP user agent is %s", httpUserAgent)

	if record.NumberOfTimesClicked < r.MaxNumberOfEmailAlerts {
		subject_template := "LinkUp link id %s clicked"
		subject := fmt.Sprintf(subject_template, path)

		var content_template string
		if lengthOfString(record.Tag) > 0 {
			content_template = fmt.Sprintf("LinkUp link with tag '%s' has been clicked!", record.Tag)
			content_template = content_template + " The link's URL is %s and redirects to %s."
		} else {
			content_template = "LinkUp link %s that redirects to %s has been clicked!"
		}

		content := fmt.Sprintf(content_template, fmt.Sprintf("%s/%s", r.RedirectUri, path), record.RedirectUrl)

		if record.NumberOfTimesClicked == r.MaxNumberOfEmailAlerts-1 {
			content = content + " This is the last email alert you'll receive for this link. Please contact mihailthebuilder@gmail.com if you wish to receive more alerts."
		}

		ser := SendEmailRequest{Email: record.UserEmail, Subject: subject, Content: content}
		err = r.Emailer.SendEmail(ser)
		if err != nil {
			log.Printf("fail to notify that link with id %s was clicked: %s", record.LinkId, err)
		}

		log.Printf("sent notification email for link with id %s", record.LinkId)
	}
}
