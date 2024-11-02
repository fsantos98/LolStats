package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const DEBUG_MODE = false
const DEBUG_MATCHES = true
const MATCHES_PER_REQUEST = 50
const BASE_API_URL = "https://europe.api.riotgames.com/"

// GET KEY HERE: https://developer.riotgames.com/
const API_KEY = "RGAPI-9e8e24b9-5f66-4331-8e00-c18810b02cb9"
const START_TIME_SECONDS_OLD = "1603230484"
const START_TIME_SECONDS_NEW = "1729800484"

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
	ChampionCounter    map[string]int
	RoleCounter        map[string]int
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

/******************
* UTILS FUNCTIONS	*
*******************/

func _l(context, msg string) {
	if DEBUG_MODE {
		log.Print("[" + context + "]$> " + msg)
	}
}

func _printArray_string(array []string) {
	for i := 0; i < len(array); i++ {
		fmt.Printf("A[%d] = %s \n", i, array[i])
	}
}

func _printArray_match(array []Match) {
	for i := 0; i < len(array); i++ {
		_printMatch(array[i])
	}
}

func _printMatch(match Match) {
	fmt.Println("MatchID: ", match.Id)
	fmt.Println("\tMode: ", match.Info.GameMode)
	fmt.Println("\tMatch duration: ", match.Info.GameDuration)
}

func _printMatchesTotalstats(matches MatchesTotalStats) {
	for k, v := range matches.Gamemode {

		fmt.Printf("==== MODE: %s ====\n", k)
		fmt.Printf("\tTotal games played: %d (Score: %d/%d)\n", v.TotalGamesPlayed, v.TotalWins, v.TotalGamesPlayed-v.TotalWins)

		kdaScore := float64(v.TotalKills+v.TotalAssists) / float64(v.TotalDeaths)
		fmt.Printf("\tTotal KDA: %d/%d/%d - %.2f\n", v.TotalKills, v.TotalDeaths, v.TotalAssists, kdaScore)
		fmt.Printf("\tTotal minions: %d\n", v.TotalMinions)
		fmt.Printf("\tTotal gold earned: %d\n", v.TotalGoldEarned)
		fmt.Printf("\tTotal damage dealt: %d\n", v.TotalDamageDealt)
		fmt.Printf("\tTotal game duration: %d\n", v.TotalGameDuration)
		fmt.Printf("\tTotal time dead: %d\n", v.TotalDeadTime)

		fmt.Printf("\tMost kills: %d\n", v.HSKills)
		fmt.Printf("\tMost assists: %d\n", v.HSAssists)
		fmt.Printf("\tMost deaths: %d\n", v.HSDeaths)
		fmt.Printf("\tMost gold earned: %d\n", v.HSGoldEarned)
		fmt.Printf("\tMost vision score: %d\n", v.HSVisionScore)
		fmt.Printf("\tMost minions farmed: %d\n", v.HSCreepsFarmed)
		fmt.Printf("\tMost minions per minute: %.1f\n", v.HSMinionsPerMinute)
		fmt.Printf("\tMost damage dealt: %d\n", v.HSDamageDealt)
		fmt.Printf("\tLongest game: %d\n", v.HSGameDuration)

		totalKeysPressed := v.TotalSpellECast + v.TotalSpellQCast + v.TotalSpellECast + v.TotalSpellRCast
		fmt.Printf("\tTotal keys pressed: %d\n", totalKeysPressed)
		fmt.Printf("\tTotal Q: %d\n", v.TotalSpellQCast)
		fmt.Printf("\tTotal W: %d\n", v.TotalSpellWCast)
		fmt.Printf("\tTotal E: %d\n", v.TotalSpellECast)
		fmt.Printf("\tTotal R: %d\n", v.TotalSpellRCast)

		fmt.Printf("\tTotal PentaKills:  %d\n", v.TotalPentaKills)
		fmt.Printf("\tTotal QuadraKills: %d\n", v.TotalQuadraKills)
		fmt.Printf("\tTotal TripleKills: %d\n", v.TotalTripleKills)
		fmt.Printf("\tTotal DoubleKills: %d\n", v.TotalDoubleKills)

		/*
			ChampionCounter   map[string]int
			RoleCounter       map[string]int
		*/
	}
}

/********************
* PROGRAM FUNCTIONS	*
*********************/

func getRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to send get request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to complete the request:\n\tStatus code: %d\n\tURL: %s\n\tResponse: %s", resp.StatusCode, url, body)
	}

	return body, nil
}

func getUserPUID(name string, tag string) (User, error) {
	_l("getUserPUID", "⏳Loading user puuid")

	url := BASE_API_URL + "riot/account/v1/accounts/by-riot-id/" + name + "/" + tag + "?api_key=" + API_KEY

	resp, err := getRequest(url)
	if err != nil {
		log.Fatalln(err)
		return User{}, err
	}

	_l("getUserPUID", "✅User loadead successfully")

	user := User{}
	err = json.Unmarshal(resp, &user)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	fmt.Println("User: " + user.Puuid)
	fmt.Println("GameName: " + user.GameName)
	fmt.Println("UserData: " + user.TagLine)

	return user, nil
}

func getUserMatches(userData User, start, count int) ([]Match, error) {
	_l("getUserMatches", "⏳Load User Matches ...")

	url := BASE_API_URL + "lol/match/v5/matches/by-puuid/" + userData.Puuid + "/ids?startTime=" + START_TIME_SECONDS_OLD + "&endTime=1760996884&start=" + strconv.Itoa(start) + "&count=" + strconv.Itoa(count) + "&api_key=" + API_KEY
	fmt.Println(url)

	body, err := getRequest(url)
	if err != nil {
		return nil, err
	}

	var newMatches []string
	err = json.Unmarshal(body, &newMatches)
	if err != nil {
		return nil, err
	}

	var matches []Match
	for i := 0; i < len(newMatches); i++ {
		matches = append(matches, Match{
			Id: newMatches[i],
		})
	}

	userData.Matches = matches

	_l("getUserMatches", "✅User matches loaded!")
	return userData.Matches, nil
}

func getAllUserMatches(userData User) ([]Match, error) {
	allMatches := []Match{}
	start, count := 0, MATCHES_PER_REQUEST

	matchesLen := -1
	for matchesLen != 0 {
		matches, err := getUserMatches(userData, start, count)
		if err != nil {
			fmt.Println("Error getting matches")
			return nil, err
		}

		allMatches = append(allMatches, matches...)
		matchesLen = len(matches)
		start += count

		if DEBUG_MATCHES {
			matchesLen = 0
		}
	}

	fmt.Printf("Total maches found: %d!\n", len(allMatches))

	return allMatches, nil
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

func getMatchDetails(match *Match) (*Match, error) {
	//_l("getMatchDetails", "\r⏳Load Match Details...")

	url := BASE_API_URL + "lol/match/v5/matches/" + match.Id + "?api_key=" + API_KEY
	body, err := getRequest(url)
	if err != nil {
		fmt.Println("error requesting match")
		return nil, err
	}

	err = json.Unmarshal(body, match)
	if err != nil {
		fmt.Println("Error unmarshalling match")
		return nil, err
	}

	if len(match.Metadata.Participants) == 0 {
		return nil, errors.New("Participants len equal to zero")
	}

	//_l("getMatchDetails", "\r✅Match details loaded!")

	return match, nil
}

func findUserIndex(game Match, user User) int {
	indexPart := -1
	for j := 0; j < len(game.Info.Participants); j++ {
		if game.Info.Participants[j].Puuid == user.Puuid {
			return j
		}
	}
	return indexPart
}

func updateIntIfHigher(curr_value *int, new_value int) {
	if new_value > *curr_value {
		*curr_value = new_value
	}
}

func updateFloatIfHigher(curr_value *float64, new_value float64) {
	if new_value > *curr_value {
		*curr_value = new_value
	}
	fmt.Printf("Old value: %f, New value %f\n", *curr_value, new_value)
}

func analyseGames(games []Match, user User) {
	//var totalMatchStats Match

	game_stats := initializeMatchesTotalStats()

	for i := 0; i < len(games); i++ {
		gameMode := games[i].Info.GameMode + "_" + games[i].Info.GameType
		if _, exists := game_stats.Gamemode[gameMode]; !exists {
			game_stats.Gamemode[gameMode] = GameModeStats{}
		}

		fmt.Println("\rAnalysing game " + games[i].Id)

		participantIndex := findUserIndex(games[i], user)
		if participantIndex == -1 {
			continue
		}
		gameInfo := games[i].Info.Participants[participantIndex]
		gameDuration := games[i].Info.GameDuration

		gs := game_stats.Gamemode[gameMode]

		gameDeaths := gameInfo.Deaths
		gameKills := gameInfo.Kills
		gameAssists := gameInfo.Assists
		gameGoldEarned := gameInfo.GoldEarned
		gameDamageDealt := gameInfo.TotalDMG
		gamePentaKills := gameInfo.PentaKills
		gameQuadraKills := gameInfo.QuadraKills
		gameTripleKills := gameInfo.TripleKills
		gameDoubleKills := gameInfo.DoubleKills
		gameSpellQCast := gameInfo.SpellQCast
		gameSpellWCast := gameInfo.SpellWCast
		gameSpellECast := gameInfo.SpellECast
		gameSpellRCast := gameInfo.SpellRCast
		gameTotalMinions := gameInfo.TotalMinions + gameInfo.TotalAllyJungleMinions + gameInfo.TotalEnemJungleMinions
		gameMinionsPerMinute := float64(gameTotalMinions) / (float64(gameDuration) / float64(60))
		gameVisionScore := gameInfo.VisionScore
		gameWin := gameInfo.Win
		gameTimeDead := gameInfo.TimeDead

		updateIntIfHigher(&gs.HSDeaths, gameDeaths)
		updateIntIfHigher(&gs.HSKills, gameKills)
		updateIntIfHigher(&gs.HSAssists, gameAssists)
		updateIntIfHigher(&gs.HSGoldEarned, gameGoldEarned)
		updateIntIfHigher(&gs.HSVisionScore, gameVisionScore)
		updateIntIfHigher(&gs.HSCreepsFarmed, gameTotalMinions)
		updateIntIfHigher(&gs.HSDamageDealt, gameDamageDealt)
		updateIntIfHigher(&gs.HSGameDuration, gameDuration)
		updateIntIfHigher(&gs.HSTimeDead, gameTimeDead)
		updateFloatIfHigher(&gs.HSMinionsPerMinute, gameMinionsPerMinute)

		gs.TotalDeaths += gameDeaths
		gs.TotalKills += gameKills
		gs.TotalAssists += gameAssists
		gs.TotalGoldEarned += gameGoldEarned
		gs.TotalDoubleKills += gameDoubleKills
		gs.TotalTripleKills += gameTripleKills
		gs.TotalQuadraKills += gameQuadraKills
		gs.TotalPentaKills += gamePentaKills
		gs.TotalSpellQCast += gameSpellQCast
		gs.TotalSpellWCast += gameSpellWCast
		gs.TotalSpellECast += gameSpellECast
		gs.TotalSpellRCast += gameSpellRCast
		gs.TotalMinions += gameTotalMinions
		gs.TotalGameDuration += gameDuration
		gs.TotalDeadTime += gameTimeDead

		gs.TotalGamesPlayed += 1
		if gameWin {
			gs.TotalWins += 1
		}
		game_stats.Gamemode[gameMode] = gs
	}

	_printMatchesTotalstats(game_stats)
}

func main() {

	fmt.Println("Starter")
	queue := initializeQueue([]Match{})

	userData, err := getUserPUID("xico", "000")
	if err != nil {
		log.Fatalln(err)
		return
	}

	userData.Matches, err = getAllUserMatches(userData)
	if err != nil {
		fmt.Println(err)
		return
	}
	queue.AddMatches(userData.Matches)
	//_printArray_match(queue.ToProcess)

	//queue.ProcessMatch()
	queue.ProcessAllMatches()
	_printArray_match(queue.Processed)
	analyseGames(queue.Processed, userData)

	/*
		for i := 1; i < 1000; i++ {
			fmt.Printf("\rLoading... %d%% complete", i)
			time.Sleep(50 * time.Millisecond) // Simulate work with a small delay
		}
	*/
}
