package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
)

/********************
* PROGRAM FUNCTIONS	*
*********************/

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

func analyseGames(games []Match, user User) {
	//var totalMatchStats Match

	game_stats := initializeMatchesTotalStats()

	for i := 0; i < len(games); i++ {
		fmt.Println("\rAnalysing game " + games[i].Id)

		participantIndex := findUserIndex(games[i], user)
		if participantIndex == -1 {
			continue
		}

		gameMode := games[i].Info.GameMode + "_" + games[i].Info.GameType
		if _, exists := game_stats.Gamemode[gameMode]; !exists {
			game_stats.Gamemode[gameMode] = GameModeStats{
				Champions: make(map[string]*GameModeStats),
				Roles:     make(map[string]*GameModeStats),
			}
		}

		championName := games[i].Info.Participants[participantIndex].ChampionName
		if _, exists := game_stats.Gamemode[gameMode].Champions[championName]; !exists {
			game_stats.Gamemode[gameMode].Champions[championName] = &GameModeStats{}
		}

		role := games[i].Info.Participants[participantIndex].Role
		if _, exists := game_stats.Gamemode[gameMode].Roles[role]; !exists {
			game_stats.Gamemode[gameMode].Roles[role] = &GameModeStats{}
		}

		gameInfo := games[i].Info.Participants[participantIndex]
		gameDuration := games[i].Info.GameDuration

		gs := game_stats.Gamemode[gameMode]
		gr := game_stats.Gamemode[gameMode].Roles[role]
		gc := game_stats.Gamemode[gameMode].Champions[championName]

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

		updateIntIfHigher(&gs.HSDeaths, &gc.HSDeaths, &gr.HSDeaths, gameDeaths)
		updateIntIfHigher(&gs.HSKills, &gc.HSKills, &gr.HSKills, gameKills)
		updateIntIfHigher(&gs.HSAssists, &gc.HSAssists, &gr.HSAssists, gameAssists)
		updateIntIfHigher(&gs.HSGoldEarned, &gc.HSGoldEarned, &gr.HSGoldEarned, gameGoldEarned)
		updateIntIfHigher(&gs.HSVisionScore, &gc.HSVisionScore, &gr.HSVisionScore, gameVisionScore)
		updateIntIfHigher(&gs.HSCreepsFarmed, &gc.HSCreepsFarmed, &gr.HSCreepsFarmed, gameTotalMinions)
		updateIntIfHigher(&gs.HSDamageDealt, &gc.HSDamageDealt, &gr.HSDamageDealt, gameDamageDealt)
		updateIntIfHigher(&gs.HSGameDuration, &gc.HSGameDuration, &gr.HSGameDuration, gameDuration)
		updateIntIfHigher(&gs.HSTimeDead, &gc.HSTimeDead, &gr.HSTimeDead, gameTimeDead)
		updateFloatIfHigher(&gs.HSMinionsPerMinute, &gc.HSMinionsPerMinute, &gr.HSMinionsPerMinute, gameMinionsPerMinute)

		updateInt(&gs.TotalDeaths, gameDeaths)
		updateInt(&gs.TotalKills, gameKills)
		updateInt(&gs.TotalAssists, gameAssists)
		updateInt(&gs.TotalGoldEarned, gameGoldEarned)
		updateInt(&gs.TotalDoubleKills, gameDoubleKills)
		updateInt(&gs.TotalTripleKills, gameTripleKills)
		updateInt(&gs.TotalQuadraKills, gameQuadraKills)
		updateInt(&gs.TotalPentaKills, gamePentaKills)
		updateInt(&gs.TotalSpellQCast, gameSpellQCast)
		updateInt(&gs.TotalSpellWCast, gameSpellWCast)
		updateInt(&gs.TotalSpellECast, gameSpellECast)
		updateInt(&gs.TotalSpellRCast, gameSpellRCast)
		updateInt(&gs.TotalMinions, gameTotalMinions)
		updateInt(&gs.TotalGameDuration, gameDuration)
		updateInt(&gs.TotalDeadTime, gameTimeDead)

		updateInt(&gs.TotalGamesPlayed, gs.TotalGamesPlayed+1)
		if gameWin {
			updateInt(&gs.TotalWins, gs.TotalWins+1)
		}

		game_stats.Gamemode[gameMode] = gs
	}

	_printMatchesTotalStats(game_stats)
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
	queue.ProcessAllMatches()

	_printArray_match(queue.Processed)
	analyseGames(queue.Processed, userData)

}
