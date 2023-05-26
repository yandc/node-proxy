package jobs

import (
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

// ProviderSet is jobs providers.
var ProviderSet = wire.NewSet(NewChainListGetNodeUrlJob, NewNodeUrlHeightJob, NewTopCoinJob, NewJobManager)

func NewJobManager(job1 *ChainListGetNodeUrlJob, job2 *NodeUrlHeightJob, job3 *TopCoinJob) *cron.Cron {
	jobManager := cron.New()
	_, err := jobManager.AddJob(job1.execTime, job1)
	_, err = jobManager.AddJob(job2.execTime, job2)
	_, err = jobManager.AddJob(job3.execTime, job3)
	if err != nil {
		panic(err.Error())
	}

	return jobManager
}
