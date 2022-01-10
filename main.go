package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/syssecfsu/web_terminal/term_conn"
)

// command line options
var cmdToExec = []string{"bash"}

var host *string = nil

// simple function to check origin
func checkOrigin(r *http.Request) bool {
	org := r.Header.Get("Origin")
	h, err := url.Parse(org)

	if err != nil {
		return false
	}

	if (host == nil) || (*host != h.Host) {
		log.Println("Failed origin check of ", org)
	}

	return (host != nil) && (*host == h.Host)
}

func main() {
	fp, err := os.OpenFile("web_term.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err == nil {
		defer fp.Close()
		log.SetOutput(fp)
		gin.DefaultWriter = fp
	}

	// parse the arguments. User can pass the command to execute
	// by default, we use bash, but macos users might want to use zsh
	// you can also run single program, such as pstree, htop...
	// but program might misbehave (htop seems to be fine)
	args := os.Args

	if len(args) > 1 {
		cmdToExec = args[1:]
		log.Println(cmdToExec)
	}

	rt := gin.Default()

	rt.SetTrustedProxies(nil)
	rt.LoadHTMLGlob("./assets/*.html")

	rt.GET("/view/*sname", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Viewer terminal",
			"path":  "/ws_view",
		})
	})

	rt.GET("/ws_do", func(c *gin.Context) {
		term_conn.ConnectTerm(c.Writer, c.Request, false, cmdToExec)
	})

	rt.GET("/ws_view", func(c *gin.Context) {
		term_conn.ConnectTerm(c.Writer, c.Request, true, nil)
	})

	// handle static files
	rt.Static("/assets", "./assets")

	rt.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Interactive terminal",
			"path":  "/ws_do",
		})

		host = &c.Request.Host
	})

	term_conn.Init(checkOrigin)

	rt.RunTLS(":8080", "./tls/cert.pem", "./tls/private-key.pem")
}
