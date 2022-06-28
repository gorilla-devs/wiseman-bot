package services

import (
	"fmt"
	"math"
	"sort"
	"wiseman/internal/db"
	"wiseman/internal/entities"

	"github.com/labstack/gommon/log"
)

func GetNextLevelMinExperience(u *entities.UserType) uint {
	fLevel := float64(u.CurrentLevel + 1)

	return uint(50 * (math.Pow(fLevel, 3) - 6*math.Pow(fLevel, 2) + 17*fLevel - 12) / 3)
}

func IncreaseExperience(u *entities.UserType, v uint, guildID string) uint {
	// Get original object using ComplexID to avoid injecting other mutated data
	serverMultiplier := db.GetServerMultiplierByGuildId(guildID)

	user := *db.GetUserByID(u.UserID, guildID)
	for {
		levelExp := GetNextLevelMinExperience(u)
		if user.CurrentLevelExperience+v < levelExp {
			user.CurrentLevelExperience += v * uint(serverMultiplier)
			break
		}

		v -= levelExp - user.CurrentLevelExperience
		user.CurrentLevelExperience = 0
		user.CurrentLevel += 1

		customRank := db.GetCustomRanksByGuildId(u.ServerID)
		sort.Slice(customRank, func(i, j int) bool {
			return customRank[i].MinLevel < customRank[j].MinLevel
		})

		for i, v := range customRank {
			if user.CurrentLevel >= v.MinLevel && user.CurrentLevel < v.MaxLevel {

				if i > 0 {
					err := RemoveRole(u.UserID, u.ServerID, v.Id, customRank[i-1].Id)
					if err != nil {
						log.Error("Error removing role", err)
					}
				}

				err := SetRole(u.UserID, u.ServerID, v.Id)
				if err != nil {
					log.Error("Error setting role", err)
				}
			}
		}
	}

	// TODO: UpsertUserByID should load from the cache inside the database all data
	// every n minutes (maybe with a cron or a goroutine?)
	db.UpsertUserByID(user.ComplexID, &user)

	return user.CurrentLevelExperience
}

func UpdateUsersRoles(serverID string, customRank entities.CustomRanks) {

	users := db.RetrieveUsersByServerID(serverID)
	for _, u := range users {
		if !u.Bot {
			if u.CurrentLevel >= customRank.MinLevel && u.CurrentLevel < customRank.MaxLevel {
				err := SetRole(u.UserID, u.ServerID, customRank.Id)
				fmt.Println("role set")
				if err != nil {
					continue
				}
			}
		}
	}
}
