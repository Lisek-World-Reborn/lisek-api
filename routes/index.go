package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
)

func Index(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Lisek world reborn api is running, no docs here",
	})
}

func GetSystemLoad(c *gin.Context) {

	cpuLoad, cpuErr := cpu.Get()
	mem, _ := memory.Get()

	totalMemory := mem.Total / 1024 / 1024
	usedMemory := mem.Used / 1024 / 1024
	freeMemory := mem.Free / 1024 / 1024

	notes := []string{}

	cpuSystem := 0
	cpuUser := 0
	cpuIdle := 0

	if cpuErr == nil {
		cpuSystem = int(cpuLoad.System)
		cpuUser = int(cpuLoad.User)
		cpuIdle = int(cpuLoad.Idle)
	} else {
		notes = append(notes, "Cpu load error: "+cpuErr.Error())
	}

	c.JSON(200, gin.H{
		"cpu": gin.H{
			"system": cpuSystem,
			"user":   cpuUser,
			"idle":   cpuIdle,
		},
		"memory": gin.H{
			"total": totalMemory,
			"used":  usedMemory,
			"free":  freeMemory,
		},
		"notes": notes,
	})
}
