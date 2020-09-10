package main

import (
	"fmt"
	"sync"
)

type jobsStruct struct {
	level   int
	steamID int
}

type workerConfig struct {
	jobsMutex       *sync.Mutex
	resMutex        *sync.Mutex
	activeJobsMutex *sync.Mutex
	wg              *sync.WaitGroup
	levelCap        int
}

func InitWorkerConfig(levelCap int) (*workerConfig, error) {

	if levelCap < 1 || levelCap > 4 {
		temp := &workerConfig{}
		return temp, fmt.Errorf("invalid levelCap %d given. levelCap must be in range 1-4 (inclusive)", levelCap)
	}

	var wg sync.WaitGroup
	var resMutex sync.Mutex
	var activeJobsMutex sync.Mutex
	var jobMutex sync.Mutex

	workConfig := &workerConfig{
		jobsMutex:       &jobMutex,
		resMutex:        &resMutex,
		activeJobsMutex: &activeJobsMutex,
		wg:              &wg,
		levelCap:        levelCap,
	}
	return workConfig, nil
}

func worker(jobs <-chan jobsStruct, results chan<- jobsStruct, workerConfig, activeJobs *int64) {

}
