package sub

import (
	"encoding/base64"
	"fmt"
	"strings"

	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type SUBController struct {
	subPath        string
	subJsonPath    string
	subEncrypt     bool
	updateInterval string

	subService     *SubService
	subJsonService *SubJsonService
}

func NewSUBController(
	g *gin.RouterGroup,
	subPath string,
	jsonPath string,
	encrypt bool,
	showInfo bool,
	rModel string,
	update string,
	jsonFragment string) *SUBController {

	a := &SUBController{
		subPath:        subPath,
		subJsonPath:    jsonPath,
		subEncrypt:     encrypt,
		updateInterval: update,

		subService:     NewSubService(showInfo, rModel),
		subJsonService: NewSubJsonService(jsonFragment),
	}
	a.initRouter(g)
	return a
}

func (a *SUBController) initRouter(g *gin.RouterGroup) {
	gLink := g.Group(a.subPath)
	gJson := g.Group(a.subJsonPath)

	gLink.GET(":subid", a.subs)

	gJson.GET(":subid", a.subJsons)
}

func (a *SUBController) subs(c *gin.Context) {
	subId := c.Param("subid")
	var myproxySub = "1ba393cfe9ecf4c2b545d60aa21c60840a41aacc899231bb21d2d0613bdb9a97"
	var myproxyTitle = "myproxy++"
	var lanSub = "1f7e4c7913c53240acc1bdc8a3e48f260120faf5c504ea9f2aa7b51d8916497d"
	var lanTitle = "LAN"
	if subId == myproxySub || subId == lanSub {
		inboundService := service.InboundService{}
		inbounds, err := inboundService.GetAllInbounds()

		if err != nil {
			c.String(400, "Can not use inbound service")
		}
		var upload, download, total int64
		for _, inbound := range inbounds {
			upload += inbound.Up
			download += inbound.Down
			total += inbound.Total
		}
		header := fmt.Sprintf("upload=%d; download=%d; total=%d; expire=%d", upload, download, total, 0)
		c.Writer.Header().Set("Subscription-Userinfo", header)

		if subId == myproxySub {
			c.Writer.Header().Set("Profile-Title", myproxyTitle)
			c.Writer.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+myproxyTitle)
			c.File("clash" + "/" + myproxyTitle + ".yaml")
		} else {
			c.Writer.Header().Set("Profile-Title", lanTitle)
			c.Writer.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+lanTitle)
			c.File("clash" + "/" + lanTitle + ".yaml")
		}

	} else {
		host := strings.Split(c.Request.Host, ":")[0]
		subs, header, err := a.subService.GetSubs(subId, host)
		if err != nil || len(subs) == 0 {
			c.String(400, "Error!")
		} else {
			result := ""
			for _, sub := range subs {
				result += sub + "\n"
			}

			// Add headers
			c.Writer.Header().Set("Subscription-Userinfo", header)
			c.Writer.Header().Set("Profile-Update-Interval", a.updateInterval)
			c.Writer.Header().Set("Profile-Title", subId)

			if a.subEncrypt {
				c.String(200, base64.StdEncoding.EncodeToString([]byte(result)))
			} else {
				c.String(200, result)
			}
		}
	}
}

func (a *SUBController) subJsons(c *gin.Context) {
	subId := c.Param("subid")
	host := strings.Split(c.Request.Host, ":")[0]
	jsonSub, header, err := a.subJsonService.GetJson(subId, host)
	if err != nil || len(jsonSub) == 0 {
		c.String(400, "Error!")
	} else {

		// Add headers
		c.Writer.Header().Set("Subscription-Userinfo", header)
		c.Writer.Header().Set("Profile-Update-Interval", a.updateInterval)
		c.Writer.Header().Set("Profile-Title", subId)

		c.String(200, jsonSub)
	}
}
