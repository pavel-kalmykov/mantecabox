package logs

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	Log = logrus.New()
	DaoLog = Log.WithFields(logrus.Fields{"package": "dao"})
	ServicesLog = Log.WithFields(logrus.Fields{"package": "service"})
	ControllerLog = Log.WithFields(logrus.Fields{"package": "webservice"})
)


func init() {
	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Log.Error("Unable to open log file")
	} else {
		Log.Out = file
	}
}
