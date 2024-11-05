package main

import (
	"fmt"
	"time"
)

/*************
* STRUCTURES *
*************/

type Participant struct {
	Deaths                 int    `json:"deaths"`
	Assists                int    `json:"assists"`
	Kills                  int    `json:"kills"`
	DoubleKills            int    `json:"doubleKills"`
	TripleKills            int    `json:"tripleKills"`
	QuadraKills            int    `json:"quadraKills"`
	PentaKills             int    `json:"pentaKills"`
	Puuid                  string `json:"puuid"`
	GoldEarned             int    `json:"goldEarned"`
	Role                   string `json:"role"`
	SpellQCast             int    `json:"spell1Casts"`
	SpellWCast             int    `json:"spell2Casts"`
	SpellECast             int    `json:"spell3Casts"`
	SpellRCast             int    `json:"spell4Casts"`
	TotalDMG               int    `json:"totalDamageDealtToChampions"`
	TimeDead               int    `json:"totalTimeSpentDead"`
	VisionScore            int    `json:"visionScore"`
	Win                    bool   `json:"win"`
	ChampionName           string `json:"championName"`
	Position               string `json:"individualPosition"`
	LargestCriticalStrike  int    `json:"largestCriticalStrike"`
	LargestKillingSpree    int    `json:"largestKillingSpree"`
	LongestTimeAlive       int    `json:"longestTimeSpentLiving"`
	TotalMinions           int    `json:"totalMinionsKilled"`
	TotalNeutralMinions    int    `json:"neutralMinionsKilled"`
	TotalAllyJungleMinions int    `json:"totalAllyJungleMinionsKilled"`
	TotalEnemJungleMinions int    `json:"totalEnemyJungleMinionsKilled"`
	WardsKilled            int    `json:"wardsKilled"`
	WardsPlaced            int    `json:"wardsPlaced"`
	Challenges             Challenges
}

type Challenges struct {
	GoldPerMinute     float64 `json:"goldPerMinute"`
	KillParticipation float64 `json:"killParticipation"`
}

type MatchInfo struct {
	GameMode     string        `json:"gameMode"`
	GameType     string        `json:"gameType"`
	GameId       int           `json:"gameId"`
	GameDuration int           `json:"gameDuration"`
	Participants []Participant `json:"participants"`
}

type MatchMetadata struct {
	DataVersion  string   `json:"dataVersion"`
	Participants []string `json:"participants"`
}

type Match struct {
	Id       string
	Metadata MatchMetadata `json:"metadata"`
	Info     MatchInfo     `json:"info"`
}

type User struct {
	GameName          string `json:"gameName"`
	TagLine           string `json:"tagLine"`
	Puuid             string `json:"puuid"`
	Matches           []Match
	TotalMatchesStats MatchesTotalStats
}

type MatchesProcessQueue struct {
	ToProcess []Match
	Processed []Match
}

type MatchesTotalStats struct {
	Gamemode map[string]GameModeStats
}

type GameModeStats struct {
	Champions          map[string]*GameModeStats
	Roles              map[string]*GameModeStats
	TotalGameDuration  int
	TotalGamesPlayed   int
	TotalDeaths        int
	HSDeaths           int
	HSGameDuration     int
	HSTimeDead         int
	HSMinionsPerMinute float64
	TotalAssists       int
	HSAssists          int
	TotalKills         int
	HSKills            int
	TotalMinions       int
	HSCreepsFarmed     int
	TotalsKeysPressed  map[string]int
	TotalGoldEarned    int
	HSGoldEarned       int
	TotalDamageDealt   int
	HSDamageDealt      int
	HSVisionScore      int
	TotalDoubleKills   int
	TotalTripleKills   int
	TotalQuadraKills   int
	TotalPentaKills    int
	TotalSpellQCast    int
	TotalSpellWCast    int
	TotalSpellECast    int
	TotalSpellRCast    int
	TotalWins          int
	TotalDeadTime      int
	MinionsPerMinute   float64
}

func initializeMatchesTotalStats() MatchesTotalStats {
	return MatchesTotalStats{
		Gamemode: make(map[string]GameModeStats),
	}
}

func initializeQueue(values []Match) MatchesProcessQueue {
	matchesQueue := MatchesProcessQueue{
		ToProcess: values,
		Processed: []Match{},
	}
	return matchesQueue
}

func (mpq *MatchesProcessQueue) AddMatches(matches []Match) {
	mpq.ToProcess = append(mpq.ToProcess, matches...)
}

func (mpq *MatchesProcessQueue) ProcessMatch() {
	if len(mpq.ToProcess) == 0 {
		fmt.Println("No match to process!")
		return
	}

	match := mpq.ToProcess[0]
	//fmt.Printf("\rProcessing match %s\n", match.Id)

	matchR, err := getMatchDetails(&match)
	if err != nil {
		fmt.Println("Failed to process match " + match.Id)
		fmt.Println(err)
		fmt.Println("Waiting 10s to retry again ... ")
		time.Sleep(10000 * time.Millisecond)
		return
	}

	mpq.Processed = append(mpq.Processed, *matchR)
	mpq.ToProcess = mpq.ToProcess[1:]

	fmt.Printf("\rProcessed match %s (%d remaning)", matchR.Id, len(mpq.ToProcess))
}

func (mpq *MatchesProcessQueue) ProcessAllMatches() {
	for len(mpq.ToProcess) > 0 {
		mpq.ProcessMatch()
		time.Sleep(600 * time.Millisecond)
	}
}
