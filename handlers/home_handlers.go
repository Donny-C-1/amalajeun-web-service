package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Amalajeun Web Service</title>
		</head>
		<body>
			<h1>This is the Amalajeun Web Service</h1>
			<p>A private service for Amalajeun hackathon project. Unauthorized access is prohibited.</p>
			<a href="https://amalajeun.vercel.app">Visit Amalajeun Main Page</a>
			<img src="https://res.cloudinary.com/dgmbzqk2p/image/upload/v1757713765/circle1_kix7j5.png" width="30" height="30" alt="Amalajeun Logo">
		</body>
		</html>
	`))
}
