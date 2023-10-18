package jobs

import (
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

// ProviderSet is jobs providers.
var ProviderSet = wire.NewSet(NewChainListGetNodeUrlJob, NewTopCoinJob, NewJobManager, NewUpdateTokenListJob)

func NewJobManager(job1 *ChainListGetNodeUrlJob, job2 *TopCoinJob, job3 *UpdateTokenListJob) *cron.Cron {
	jobManager := cron.New()

	go job1.Run()

	_, err := jobManager.AddJob(job1.execTime, job1)
	_, err = jobManager.AddJob(job2.execTime, job2)
	_, err = jobManager.AddJob(job3.execTime, job3)
	if err != nil {
		panic(err.Error())
	}

	return jobManager
}
