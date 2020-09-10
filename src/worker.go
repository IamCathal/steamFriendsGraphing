package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type jobsStruct struct {
	level   int
	steamID string
	// APIKey  string
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

func Worker(jobs <-chan jobsStruct, results chan<- jobsStruct, cfg workerConfig, activeJobs *int64) {
	for {
		cfg.jobsMutex.Lock()
		job := <-jobs
		cfg.jobsMutex.Unlock()

		// Get friends list for this current user
		friendsObj, err := GetFriends(job.steamID, os.Getenv("APIKEY"), cfg.wg)
		CheckErr(err)

		numFriends := len(friendsObj.FriendsList.Friends)

		// For each friend we'll add them to the jobs queue if
		// their level is within our range
		for i := 0; i < numFriends; i++ {

			indivFriends := jobsStruct{
				level:   job.level + 1,
				steamID: friendsObj.FriendsList.Friends[i].Steamid,
			}

			fmt.Printf("New friend [Level %d][From %s][SteamID %s][Len %d]\n", indivFriends.level, job.steamID, indivFriends.steamID, len(jobs))

			// If their level is within range, we'll scrape them in the future
			// and therefore we up the counter of activeJobs
			if indivFriends.level <= cfg.levelCap {
				atomic.AddInt64(activeJobs, 1)
			}

			cfg.resMutex.Lock()
			results <- indivFriends
			cfg.resMutex.Unlock()
		}

		cfg.activeJobsMutex.Lock()
		atomic.AddInt64(activeJobs, -1)
		cfg.activeJobsMutex.Unlock()

		cfg.wg.Done()
	}
}
