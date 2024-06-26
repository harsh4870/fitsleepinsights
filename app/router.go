// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package client handles the client subdomain.
// The client subdomain is the subdomain that interacts
// with the Fitbit API

package app

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/joho/godotenv/autoload"
)

func NewRouter() (*echo.Echo, error) {
	router := echo.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Pre(middleware.RemoveTrailingSlashWithConfig(
		middleware.TrailingSlashConfig{
			RedirectCode: http.StatusMovedPermanently,
		},
	))

	router.Logger.SetLevel(log.DEBUG)

	// use echoview(goview) as simplified renderer
	viewConf := goview.DefaultConfig
	viewConf.Funcs["year"] = time.Now().UTC().Year
	viewConf.Funcs["contains"] = strings.Contains
	viewConf.Funcs["hasPrefix"] = strings.HasPrefix
	viewConf.Funcs["hasSuffix"] = strings.HasSuffix
	viewConf.Funcs["nl2br"] = func(text string) template.HTML {
		return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
	}
	viewConf.Funcs["md2html"] = func(md string) template.HTML {
		// create markdown parser with extensions
		extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
		p := parser.NewWithExtensions(extensions)
		doc := p.Parse([]byte(md))

		// create HTML renderer with extensions
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)

		return template.HTML(markdown.Render(doc, renderer))
	}
	viewConf.Funcs["min2ddhhmm"] = min2ddhhmm
	viewConf.Funcs["float64"] = func(n int64) float64 {
		return float64(n)
	}
	viewConf.Funcs["timeOnly"] = func(t time.Time) string {
		return t.Format(time.TimeOnly)
	}

	router.Renderer = echoview.New(viewConf)

	// OAuth2 routes
	router.GET("/auth", Auth())
	router.GET("/redirect", Redirect())

	// Login route is auth
	// Logout is the cookie removal
	router.GET("/login", Auth())
	router.GET("/logout", func(c echo.Context) (err error) {
		var cookie *http.Cookie
		if cookie, err = c.Cookie("token"); err == nil {
			cookie.Value = ""
			cookie.MaxAge = -1
			cookie.Expires = time.Now().Add(-time.Hour)
			c.SetCookie(cookie)
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	router.GET("/", Index())
	router.GET("/about", About())
	router.GET("/contact", Contact())
	router.GET("/privacy", Privacy())

	// The default dashboard is the monthly dashboard
	router.GET("/dashboard", MonthlyDashboard(), RequireFitbit())
	router.GET("/dashboard/:startYear/:startMonth/:startDay/:endYear/:endMonth/:endDay", CustomDashboard(), RequireFitbit())
	// All the dashboard routes below are defined but already superseded by the one above
	// decide with to do. Perhaps when clicking on some shortcut we can point to these, or link to them somehow...
	router.GET("/dashboard/week", WeeklyDashboard(), RequireFitbit())
	router.GET("/dashboard/month", MonthlyDashboard(), RequireFitbit())
	router.GET("/dashboard/year", YearlyDashboard(), RequireFitbit())
	router.GET("/dashboard/:year/:month/:day", WeeklyDashboard(), RequireFitbit())
	router.GET("/dashboard/:year/:month", MonthlyDashboard(), RequireFitbit())
	router.GET("/dashboard/:year", YearlyDashboard(), RequireFitbit())

	router.GET("/chat/:startYear/:startMonth/:startDay/:endYear/:endMonth/:endDay", ChatWithData(), RequireFitbit())

	router.Static("/static", "static")

	// INTERNAL routes used for:

	// Generating types and testing the Authorizer
	router.GET("/generate", GenerateTypes(), RequireFitbit())
	// Testing the API(*Authorizer)
	router.GET("/test", TestGET(), RequireFitbit())
	// Dump all data endpoint (INTERNAL)
	router.GET("/dump", Dump(), RequireFitbit())
	// Train and deploy sleep efficiency predictor
	router.GET("/train", TestTrainAndDeploy(), RequireFitbit())
	// Predict sleep efficiency
	router.GET("/predict", TestPredictSleepEfficiency(), RequireFitbit())
	// Fetch all data endpoint (INTERNAL)
	router.GET("/fetch", Fetch(), RequireFitbit())
	return router, nil
}
