package scheduler

import (
	"decoration/models"
	"decoration/store"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AlertScheduler struct {
	store     store.PhaseStore
	interval  time.Duration
	notified  map[string]time.Time
}

func NewAlertScheduler(s store.PhaseStore, interval time.Duration) *AlertScheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &AlertScheduler{
		store:    s,
		interval: interval,
		notified: make(map[string]time.Time),
	}
}

func (a *AlertScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	a.checkAndNotify(time.Now())

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			a.checkAndNotify(t)
		}
	}
}

func (a *AlertScheduler) checkAndNotify(now time.Time) {
	overdue := a.store.FindOverduePhases(now)
	for _, o := range overdue {
		key := fmt.Sprintf("%s:%s", o.ProjectID, o.PhaseType)

		lastNotified, exists := a.notified[key]
		if exists && now.Sub(lastNotified) < 24*time.Hour {
			continue
		}

		severity := models.SeverityWarning
		if o.OverdueDays >= 3 {
			severity = models.SeverityCritical
		}

		alert := &models.PhaseAlert{
			ID:          uuid.NewString(),
			ProjectID:   o.ProjectID,
			PhaseType:   o.PhaseType,
			PhaseName:   o.PhaseName,
			Severity:    severity,
			Message:     fmt.Sprintf("阶段[%s]已逾期%d天，请尽快安排完成", o.PhaseName, o.OverdueDays),
			OverdueDays: o.OverdueDays,
			NotifiedAt:  now,
		}

		a.store.RecordAlert(alert)
		a.notified[key] = now
	}
}
