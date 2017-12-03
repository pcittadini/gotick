package taskHandlers

import (
	"github.com/gin-gonic/gin"
	"encoding/json"
	"github.com/pcittadini/gotick/lib/tasks"
)

func New(c *gin.Context)()  {

	t := new(tasks.Task)

	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&t)
	if err != nil {
		c.JSON(500, err.Error())
		c.Abort()
		return
	}

	go t.Scheduler()

	c.JSON(200, "created")
	return
}