package worker

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"concept-tracker/internal/repository"
)

type Worker struct {
	Reminder        repository.ReminderRepository
	UserPreferences repository.UserPreferencesRepository
	Cron            *cron.Cron
}

func NewWorker(reminder repository.ReminderRepository, userPreferences repository.UserPreferencesRepository) Worker {
	c := cron.New()

	return Worker{
		Reminder:        reminder,
		UserPreferences: userPreferences,
		Cron:            c,
	}
}

func (w *Worker) Start() error {
	_, err := w.Cron.AddFunc("* * * * *", w.Run)
	if err != nil {
		return err
	}

	w.Cron.Start()

	return nil
}

func (w *Worker) Stop() error {
	_ = w.Cron.Stop()

	return nil
}

func (w *Worker) Run() {
	ctx := context.Background()
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	a, err := w.Reminder.GetActiveReminders(ctx)
	if err != nil {
		log.Printf("worker: failed to get active reminders: %v", err)
		return
	}

	for _, v := range a {
		var nextFireAt *time.Time

		if v.IsRecurring {
			if v.CronExpr == nil {
				log.Printf("worker: recurring reminder %s has no cron expression, skipping", v.ID)
				continue
			}

			sched, err := specParser.Parse(*v.CronExpr)
			if err != nil {
				log.Printf("worker: failed to parse cron expression: %v", err)
				continue
			}

			g, err := w.UserPreferences.GetUserPreferences(ctx, v.UserID)
			if err != nil {
				log.Printf("worker: user preferences must include a timezone: %v", err)
				continue
			}

			userLocation, err := time.LoadLocation(g.Timezone)
			if err != nil {
				log.Printf("worker: could not properly parse timezone: %v", err)
				continue
			}

			n := sched.Next(time.Now().In(userLocation))

			nextFireAt = &n
		} else {
			v.IsActive = false
		}

		now := time.Now()
		advance := w.Reminder.AdvanceSchedule(ctx, v.ID, nextFireAt, &now, v.IsActive)
		if advance != nil {
			log.Printf("worker: error advancing schedule after last job run: %v", advance)
			continue
		}
	}
}
