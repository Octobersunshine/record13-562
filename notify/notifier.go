package notify

import (
	"decoration/models"
	"fmt"
	"log"
)

type Notifier interface {
	Send(foreman *models.ProjectInfo, alert *models.PhaseAlert) error
}

type LogNotifier struct{}

func NewLogNotifier() Notifier {
	return &LogNotifier{}
}

func (n *LogNotifier) Send(foreman *models.ProjectInfo, alert *models.PhaseAlert) error {
	if foreman == nil {
		log.Printf("[ALERT] 项目[%s] 阶段[%s]逾期%d天 - 未配置工长信息",
			alert.ProjectID, alert.PhaseName, alert.OverdueDays)
		return nil
	}
	msg := fmt.Sprintf("[ALERT PUSH to 工长%s(%s)] 项目[%s]的[%s]阶段已逾期%d天，严重程度：%s，提醒：%s",
		foreman.ForemanName, foreman.ForemanPhone,
		alert.ProjectID, alert.PhaseName, alert.OverdueDays, alert.Severity, alert.Message)
	log.Println(msg)
	return nil
}
