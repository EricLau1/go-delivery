package workflows

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func Run(temporalClient client.Client) {
	w := worker.New(temporalClient, TaskQueueName, worker.Options{})

	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(&Activities{})

	err := w.Run(worker.InterruptCh())
	if err != nil {
		log.Panicln(err)
	}
}
