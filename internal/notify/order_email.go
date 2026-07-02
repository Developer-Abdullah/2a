package notify

import (
	"context"
	"fmt"
	"strings"

	"github.com/hibiken/asynq"

	"platform/internal/domain"
)

// OrderEmailer enqueues an activation-code email when a storefront order is fulfilled. It implements
// shop's fulfillment-notifier seam. Delivery itself happens in the notifier worker (handleEmail),
// which sends via SMTP when configured and otherwise logs — so the pipeline works in dev.
type OrderEmailer struct {
	queue *asynq.Client
}

func NewOrderEmailer(queue *asynq.Client) *OrderEmailer {
	return &OrderEmailer{queue: queue}
}

// NotifyFulfilled queues the code-delivery email. It is best-effort: a missing email, no codes, or an
// unconfigured queue is a silent no-op, and enqueue errors are returned for the caller to log (they
// must NOT fail the payment webhook — the code is still shown on the order page).
func (e *OrderEmailer) NotifyFulfilled(ctx context.Context, order *domain.Order) error {
	if e == nil || e.queue == nil || order == nil {
		return nil
	}
	if strings.TrimSpace(order.Email) == "" || len(order.Codes) == 0 {
		return nil
	}
	task, err := NewEmailTask(EmailPayload{
		To:      order.Email,
		Subject: "أكواد التفعيل الخاصة بطلبك",
		Body:    buildOrderEmailBody(order),
	})
	if err != nil {
		return err
	}
	_, err = e.queue.EnqueueContext(ctx, task, asynq.Queue("notifications"))
	return err
}

// buildOrderEmailBody renders a plain-text Arabic email listing the order's codes.
func buildOrderEmailBody(order *domain.Order) string {
	var b strings.Builder
	b.WriteString("شكرًا لشرائك من متجر Double A!\n\n")
	b.WriteString(fmt.Sprintf("رقم الطلب: %s\n\n", order.ID))
	if len(order.Codes) == 1 {
		b.WriteString("كود التفعيل الخاص بك:\n")
	} else {
		b.WriteString("أكواد التفعيل الخاصة بك:\n")
	}
	for _, code := range order.Codes {
		b.WriteString("  " + code + "\n")
	}
	b.WriteString("\nاحتفظ بالكود في مكان آمن. لأي استفسار تواصل مع الدعم.\n")
	return b.String()
}
