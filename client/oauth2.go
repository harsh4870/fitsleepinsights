// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the endpoint for commnuicating with
// the Fitbit API.

package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/galeone/sleepbit/database"
	"github.com/galeone/sleepbit/fitbit"
	"github.com/galeone/sleepbit/fitbit/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var db = database.NewStorage()

// Auth redirects the user to the fitbit authorization page
// It sets a cookie the univocally identifies the user
// because the fitbitClient.Exchange (used in Redirect)
// needs to check the `code` and CSRF tokens - and these tokens
// are attributes of the fitbit client that needs to persist
// from Auth() to Redirect()
func Auth() func(echo.Context) error {
	return func(c echo.Context) (err error) {
		fitbitClient := fitbit.NewClient(db)

		authorizing := types.AuthorizingUser{
			CSRFToken: uuid.New().String(),
			// Code verifier for PKCE
			// https://dev.fitbit.com/build/reference/web-api/developer-guide/authorization/#Authorization-Code-Grant-Flow-with-PKCE
			Code: fmt.Sprintf("%s-%s", uuid.New().String(), uuid.New().String()),
		}

		fitbitClient.SetAuthorizing(&authorizing)

		c.SetCookie(&http.Cookie{
			Name: "authorizing",
			// Also used as primary key in db for retrieval (see middelware
			// RequireFitbitClient).
			Value: fitbitClient.CSRFToken().String(),
			// No Expires = Session cookie
			HttpOnly: true,
		})

		if err = db.InsertAuhorizingUser(&authorizing); err != nil {
			return err
		}

		var auth_url *url.URL
		if auth_url, err = fitbitClient.AuthorizationURL(); err != nil {
			return err
		}

		return c.Redirect(http.StatusTemporaryRedirect, auth_url.String())
	}
}

// Redirect handles the redirect from the Fitbit API to our redirect URI.
// Sets the "token" cookie for the whole domain, containing the access token
// The access token univocally identifies the user. The token expires when the
// access token expires.
func Redirect() func(echo.Context) error {
	return func(c echo.Context) error {
		// We can assume fitbitClient is present and valid
		// because this route is protected by the RequireFitbit middleware
		fitbitClient := c.Get("fitbit").(*fitbit.FitbitClient)

		state := c.QueryParam("state")
		if state != fitbitClient.CSRFToken().String() {
			c.Logger().Warnf("Invalid state in /redirect. Expected %s got %s", fitbitClient.CSRFToken().String(), state)
			return c.Redirect(http.StatusTemporaryRedirect, "/error?status=csrf")
		}

		code := c.QueryParam("code")
		var token *types.AuthorizedUser
		var err error
		if token, err = fitbitClient.ExchangeAuthorizationCode(code); err != nil {
			c.Logger().Warnf("ExchangeAuthorizationCode: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, "/error?status=exchange")
		}
		// Update the fitbitclient. Now it contains a valid token and HTTP can be used to query the API
		fitbitClient.SetToken(token)

		// Save token and redirect user to the application
		if err = db.UpsertAuthorizedUser(token); err != nil {
			return err
		}
		cookie := http.Cookie{
			Name:     "token",
			Value:    token.AccessToken,
			Domain:   os.Getenv("DOMAIN"),
			Expires:  time.Now().Add(time.Second * time.Duration(token.ExpiresIn)),
			HttpOnly: true,
		}
		c.SetCookie(&cookie)
		return c.Redirect(http.StatusTemporaryRedirect, "/app")
	}
}

func Error() func(echo.Context) error {
	return func(c echo.Context) error {
		status := c.QueryParam("status")

		type ErrorMessage struct {
			Error string `json:"error"`
		}
		switch status {
		case "csrf":
			return c.JSON(http.StatusBadRequest, ErrorMessage{
				Error: "CSRF attempt detected",
			})
		case "exchange":
			return c.JSON(http.StatusBadGateway, ErrorMessage{
				Error: "Error exchanging OAuht2 Authorization code for the tokens",
			})
		}
		return nil
	}
}
