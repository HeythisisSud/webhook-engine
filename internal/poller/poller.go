package poller

import (
	"context"
	"time"
    "log"

	"github.com/heythisissud/webhook-engine/internal/db/generated"
	"github.com/heythisissud/webhook-engine/internal/worker"
	"github.com/hibiken/asynq"
)
type Poller struct {
    queries     *db.Queries
    asynqClient *asynq.Client
}

func NewPoller(queries *db.Queries, asynqClient *asynq.Client) *Poller {
    return &Poller{queries: queries, asynqClient: asynqClient}
}

func (p *Poller) Start(ctx context.Context) {
	ticker:=time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for{
        select{
        case <- ctx.Done():
            return
        case <-ticker.C:
            payload, err:=p.queries.GetPendingOutboxEntries(ctx)
            if err!=nil{
                return
            }

            for _, entries := range payload {

                webhook, err:= p.queries.GetWebhook(ctx, entries.WebhookID)
                if err!=nil{
                    continue
                }

                event, errs := p.queries.GetEvent(ctx, entries.EventID)
                if errs != nil {
                    continue
                }

                task, errr:=worker.NewWebhookDeliveryTask(entries.ID.String(),webhook.TargetUrl,event.Payload)
                
                if errr!=nil{
                    continue

                }


                info,err:=p.asynqClient.Enqueue(task)
                if err != nil {
                    log.Println("error enqueuing task:", err)
                    continue
                }
                log.Println("enqueued task:",info.ID)

                // update status to enqueued
                p.queries.UpdateOutboxStatus(ctx, db.UpdateOutboxStatusParams{
                    ID:     entries.ID,
                    Status: "enqueued",
                })
            




                // webhook is one item
                // _ is the index (0, 1, 2...) — we ignore it with _
            }




        }
    }


    // loop every 1 second
    // fetch pending rows
    // enqueue each as asynq job
    // update status to enqueued
}