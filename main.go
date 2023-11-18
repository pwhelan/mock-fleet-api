package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type search struct {
	ProjectID string `form:"project_id"`
	Resource  string `form:"resource"`
	Term      string `form:"term"`
}

type config struct {
	Format string `form:"format"`
}

type fleet struct {
	ID     string `uri:"fleetID"`
	Name   string `form:"name"`
	Config fleetConfig
}

type fleetConfig struct {
	LastModified time.Time
	RawConfig    string
}

var dummyConfig = `
[INPUT]
    Name dummy
    Tag dummy
[OUTPUT]
    Name   stdout
    Match  *
    Format json_lines
`

var fleetConfigs = map[string]fleet{
	"0BDF9CD3-1D31-4D3E-A47D-6967A3B6A53D": {
		ID:   "0BDF9CD3-1D31-4D3E-A47D-6967A3B6A53D",
		Name: "fleetbar",
		Config: fleetConfig{
			LastModified: time.Now(),
			RawConfig:    dummyConfig,
		},
	},
}

type patchConfig struct {
	ConfigFormat string `json:"configFormat"`
	Name         string `json:"name"`
	RawConfig    string `json:"rawConfig"`
}

func main() {
	r := gin.Default()
	r.GET("/v1/search", func(c *gin.Context) {
		var s search
		if err := c.ShouldBind(&s); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
		}
		for _, fleet := range fleetConfigs {
			if fleet.Name == s.Term {
				c.JSON(200, []gin.H{{
					"ID": fleet.ID,
				}})
				return
			}
		}
		c.JSON(404, []gin.H{{
			"status":  "error",
			"message": "Not Found",
		}})
	})
	r.GET("/v1/projects/:project_id/fleets", func(c *gin.Context) {
		var fp fleet
		if err := c.ShouldBind(&fp); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		for _, f := range fleetConfigs {
			if f.Name == fp.Name {
				c.JSON(200, []fleet{f})
				return
			}
		}
		c.JSON(404, []gin.H{{
			"status":  "error",
			"message": "Not Found",
			"name":    fp.Name,
		}})
	})
	r.PATCH("/v1/fleets/:fleetID", func(c *gin.Context) {
		var f fleet
		var p patchConfig
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if f, ok := fleetConfigs[f.ID]; ok {
			if err := c.ShouldBindJSON(&p); err != nil {
				c.JSON(500, gin.H{
					"status": "error",
					"error":  err.Error(),
				})
				return
			}
			fleetConfigs[f.ID] = fleet{
				Name: f.Name,
				ID:   f.ID,
				Config: fleetConfig{
					LastModified: time.Now(),
					RawConfig:    p.RawConfig,
				},
			}
			c.JSON(200, f)
			return
		}
		c.String(404, "Not Found")
	})
	r.GET("/v1/fleets/:fleetID/config", func(c *gin.Context) {
		var cfg config
		var f fleet
		if err := c.ShouldBind(&cfg); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if fleet, ok := fleetConfigs[f.ID]; ok {
			c.Header("Last-Modified",
				fleet.Config.LastModified.Format(http.TimeFormat))
			c.String(200, fleet.Config.RawConfig)
			return
		}
		c.String(404, "Not Found")
	})
	r.GET("/v1/fleets/:fleetID/files", func(c *gin.Context) {
		var f fleet
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
		}
		c.JSON(200, []map[string]string{})
	})
	r.POST("/v1/agents", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id":    "457DA5EB-027F-47C4-B71A-D7FA270AC337",
			"token": "E071B9E1-1E71-4F7E-BADF-D52E45CD9E88",
		})
	})
	r.POST("/v1/agents/:agent_id", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id":    "457DA5EB-027F-47C4-B71A-D7FA270AC337",
			"token": "E071B9E1-1E71-4F7E-BADF-D52E45CD9E88",
		})
	})
	r.POST("/v1/agents/:agent_id/metrics", func(c *gin.Context) {
	})
	r.Run("127.0.0.1:8080")
}
