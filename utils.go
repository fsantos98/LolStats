package main

import (
	"fmt"
	"log"
)

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

	}
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
