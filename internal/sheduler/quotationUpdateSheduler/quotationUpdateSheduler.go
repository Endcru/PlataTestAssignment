package quotationUpdateSheduler

import (
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"

	quotationService "github.com/Endcru/PlataTestAssignment/internal/service/quotation"
)

type QuotationUpdateSheduler struct {
	quotationService quotationService.QuotationService
	log *slog.Logger
	scheduler gocron.Scheduler
}

func NewQuotationUpdateSheduler(qs quotationService.QuotationService, log *slog.Logger) *QuotationUpdateSheduler {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	_, _ = scheduler.NewJob(
		gocron.DurationJob(1*time.Minute),
		gocron.NewTask(func() {
			log.Info("Updating quotations")
			err := qs.UpdateQuotations()
			if err != nil {
				log.Error("Failed to update quotations", slog.String("error", err.Error()))
			} else {
				log.Info("Quotations updated")
			}
		}),
	)
	return &QuotationUpdateSheduler{
		quotationService: qs,
		log: log,
		scheduler: scheduler,
	}
}

func (s *QuotationUpdateSheduler) Start() {
	s.scheduler.Start()
	s.log.Info("Quotation update scheduler started")
}

func (s *QuotationUpdateSheduler) Stop() {
	_ = s.scheduler.Shutdown()
	s.log.Info("Quotation update scheduler stopped")
}
