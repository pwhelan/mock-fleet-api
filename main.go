package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type search struct {
	ProjectID string `form:"project_id"`
	Resource  string `form:"resource"`
	FleetID   string `form:"fleet_id"`
	Term      string `form:"term"`
}

type config struct {
	Format string `form:"format"`
}

type fleet struct {
	ID          string `uri:"fleetID"`
	Name        string `form:"name"`
	Config      fleetConfig
	AgentsCount struct {
		Active   int `json:"active"`
		Inactive int `json:"inactive"`
	} `json:"agentsCount"`
}

type fleetFile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Contents  string    `json:"contents"`
	CreatedAt time.Time `json:"createdAt"`
}

type fleetConfig struct {
	LastModified time.Time
	RawConfig    string
	Files        map[string]fleetFile
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

var fleetConfigs = map[string]*fleet{
	strings.ToLower("0BDF9CD3-1D31-4D3E-A47D-6967A3B6A53D"): {
		ID:   strings.ToLower("0BDF9CD3-1D31-4D3E-A47D-6967A3B6A53D"),
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

type agent struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Edition            string    `json:"edition"`
	Type               string    `json:"type"`
	Version            string    `json:"version"`
	Token              string    `json:"token"`
	MachineID          string    `json:"machineID"`
	FleetID            string    `json:"fleetID"`
	LastMetricsAddedAt time.Time `json:"lastMetricsAddedAt"`
}

var fleetagents map[string]map[string]*agent = map[string]map[string]*agent{}
var agents map[string]*agent = map[string]*agent{}

func main() {
	r := gin.Default()
	r.GET("/v1/search", func(c *gin.Context) {
		var s search
		if err := c.ShouldBind(&s); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
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
	r.GET("/v1/projects/:project_id/agents", func(c *gin.Context) {
		var s search
		agents := make([]agent, 0)

		if err := c.ShouldBind(&s); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		if fagents, ok := fleetagents[strings.ToLower(s.FleetID)]; ok {
			for _, agent := range fagents {
				agents = append(agents, *agent)
			}
			c.JSON(200, agents)
			return
		}

		for _, fleet := range fleetagents {
			for _, agent := range fleet {
				agents = append(agents, *agent)
			}
		}
		c.JSON(200, agents)
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
			f.AgentsCount.Active = 0
			f.AgentsCount.Inactive = 0
			for _, agent := range fleetagents[strings.ToLower(f.ID)] {
				if time.Since(agent.LastMetricsAddedAt) < time.Minute*5 {
					f.AgentsCount.Active++
				} else {
					f.AgentsCount.Inactive++
				}
			}
		}
		if fp.Name != "" {
			for _, f := range fleetConfigs {
				if f.Name == fp.Name {
					f.AgentsCount.Active = 0
					f.AgentsCount.Inactive = 0
					for _, agent := range fleetagents[strings.ToLower(f.ID)] {
						if time.Since(agent.LastMetricsAddedAt) < time.Minute*5 {
							f.AgentsCount.Active++
						} else {
							f.AgentsCount.Inactive++
						}
					}
					c.JSON(200, []fleet{*f})
					return
				}
			}
			c.JSON(404, []gin.H{{
				"status":  "error",
				"message": "Not Found",
				"name":    fp.Name,
			}})
			return
		}
		fleets := make([]fleet, 0)
		for _, f := range fleetConfigs {
			fleets = append(fleets, *f)
		}
		c.JSON(200, fleets)
	})
	r.POST("/v1/projects/:project_id/fleets", func(c *gin.Context) {
		var fp patchConfig
		if err := c.ShouldBindJSON(&fp); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		id := uuid.New()
		fleetConfigs[strings.ToLower(id.String())] = &fleet{
			ID:   strings.ToLower(id.String()),
			Name: fp.Name,
			Config: fleetConfig{
				LastModified: time.Now(),
				RawConfig:    fp.RawConfig,
			},
		}
		c.JSON(200, fleetConfigs[strings.ToLower(id.String())])
	})
	r.GET("/v1/fleets/:fleetID", func(c *gin.Context) {
		var f fleet
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if fleet, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			c.JSON(200, fleet)
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
	})
	r.DELETE("/v1/fleets/:fleetID", func(c *gin.Context) {
		var f fleet
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if _, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			delete(fleetConfigs, strings.ToLower(f.ID))
			c.JSON(204, nil)
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
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
		if f, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			if err := c.ShouldBindJSON(&p); err != nil {
				c.JSON(500, gin.H{
					"status": "error",
					"error":  err.Error(),
				})
				return
			}
			fleetConfigs[strings.ToLower(f.ID)] = &fleet{
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
		if fleet, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			c.Header("Last-Modified",
				fleet.Config.LastModified.Format(http.TimeFormat))
			c.String(200, fleet.Config.RawConfig)
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
	})
	r.GET("/v1/fleets/:fleetID/files", func(c *gin.Context) {
		var f fleet
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
		}
		if fleet, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			files := make([]fleetFile, 0)
			for _, file := range fleet.Config.Files {
				files = append(files, file)
			}
			c.JSON(200, files)
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
	})
	r.POST("/v1/fleets/:fleetID/files", func(c *gin.Context) {
		var f fleet
		var file fleetFile
		if err := c.ShouldBindUri(&f); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if err := c.ShouldBindJSON(&file); err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}
		if fleet, ok := fleetConfigs[strings.ToLower(f.ID)]; ok {
			if fleet.Config.Files == nil {
				fleet.Config.Files = make(map[string]fleetFile)
			}
			file.CreatedAt = time.Now()
			file.ID = strings.ToLower(uuid.New().String())
			fleet.Config.Files[file.Name] = file
			c.JSON(200, file)
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
	})
	r.POST("/v1/agents", func(c *gin.Context) {
		var ag agent
		if err := c.ShouldBindJSON(&ag); err != nil {
			c.JSON(500, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		if _, ok := fleetagents[strings.ToLower(ag.FleetID)]; !ok {
			fleetagents[strings.ToLower(ag.FleetID)] = make(map[string]*agent)
		}

		ag.ID = strings.ToLower(uuid.New().String())
		ag.Token = strings.ToLower(uuid.New().String())

		if agent, ok := fleetagents[strings.ToLower(ag.FleetID)][ag.MachineID]; ok {
			c.JSON(200, agent)
			return
		}
		fleetagents[strings.ToLower(ag.FleetID)][ag.MachineID] = &ag
		agents[ag.ID] = &ag

		c.JSON(200, ag)
	})
	r.POST("/v1/agents/:agent_id", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"id":    "457DA5EB-027F-47C4-B71A-D7FA270AC337",
			"token": "E071B9E1-1E71-4F7E-BADF-D52E45CD9E88",
		})
	})
	r.POST("/v1/agents/:agent_id/metrics", func(c *gin.Context) {
		var parm struct {
			AgentID string `form:"agent_id" uri:"agent_id"`
		}
		c.ShouldBindUri(&parm)
		if agent, ok := agents[parm.AgentID]; ok {
			agent.LastMetricsAddedAt = time.Now()
			c.JSON(200, gin.H{
				"status": "OK",
			})
			return
		}
		c.JSON(404, gin.H{
			"status":  "error",
			"message": "Not Found",
		})
	})
	r.Run("127.0.0.1:8080")
}
