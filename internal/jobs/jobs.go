package jobs

import (
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

// ProviderSet is jobs providers.
var ProviderSet = wire.NewSet(NewChainListGetNodeUrlJob, NewTopCoinJob, NewJobManager)

func NewJobManager(job1 *ChainListGetNodeUrlJob, job2 *TopCoinJob) *cron.Cron {
	jobManager := cron.New()

	go job1.Run()

	_, err := jobManager.AddJob(job1.execTime, job1)
	_, err = jobManager.AddJob(job2.execTime, job2)
	if err != nil {
		panic(err.Error())
	}

	return jobManager
}
