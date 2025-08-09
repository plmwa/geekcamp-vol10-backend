package end

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Bye(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Goodbye, World!",
	})
}
