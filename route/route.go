package route

import (
	"github.com/gin-gonic/gin"
	"github.com/its-me-debk007/QWait_backend/controller"
)

func SetupRoutes(app *gin.Engine) {
	api := app.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", controller.Signup)
			authGroup.GET("/verify", controller.Verify)
			// authGroup.POST("/send-otp", controller.SendOtp)
			// authGroup.POST("/verify", controller.VerifyOtp)
			// authGroup.POST("/reset", controller.ResetPassword)
		}

		api.POST("/join/:id", controller.JoinQueue)
		api.POST("/leave/:id", controller.LeaveQueue)
	}
}
