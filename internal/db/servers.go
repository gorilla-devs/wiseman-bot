package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"wiseman/internal/entities"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// var servers entities.ServersType
var SERVERS_DB *mongo.Collection

type Servers struct {
	cache  map[string]*entities.ServerType
	writes int
	lock   sync.RWMutex
}

var servers Servers = Servers{
	cache:  make(map[string]*entities.ServerType, 1000),
	writes: 0,
}

// StartServersDBUpdater updates the DB every
// 5 writes on the cache layer
func StartServersDBUpdater() {
	for {
		if GetServersWrites() > 5 {
			fmt.Println("updating server db")
			UpdateAllServersInDb()
		}
	}
}

func HydrateServers(d *discordgo.Session) (int, error) {
	var ns int
	var guilds []*discordgo.UserGuild
	var lastID string
	for {
		newGuilds, err := d.UserGuilds(100, "", lastID)
		if err != nil {
			return 0, err
		}

		if len(newGuilds) == 0 {
			break
		}

		lastID = newGuilds[len(newGuilds)-1].ID
		guilds = append(guilds, newGuilds...)
	}

	// TODO: Use InsertMany to optimize this
	for _, guild := range guilds {
		// Check if server is already in DB
		res := SERVERS_DB.FindOne(context.TODO(), bson.M{"serverid": guild.ID})

		if res.Err() != mongo.ErrNoDocuments {
			var server entities.ServerType
			err := res.Decode(&server)
			if err != nil {
				return 0, err
			}

			// TODO: FIX
			sort.SliceStable(server.CustomRanks, func(i, j int) bool {
				return server.CustomRanks[i].MinLevel > server.CustomRanks[j].MinLevel
			})

			UpsertServerByID(guild.ID, &server)
			continue
		}

		fmt.Println("Server not found in DB", guild.ID, guild.Name)
		ns += 1

		server := entities.ServerType{
			ServerID:            guild.ID,
			ServerPrefix:        "!",
			NotificationChannel: "",
			WelcomeChannel:      "",
			CustomRanks:         []entities.CustomRanks{},
			RankTime:            0,
			MsgExpMultiplier:    1.00,
			TimeExpMultiplier:   1.00,
			WelcomeMessage:      "",
			DefaultRole:         "",
		}
		UpsertServerByID(guild.ID, &server)

		SERVERS_DB.InsertOne(context.TODO(), server)
	}

	return ns, nil
}

// Get methods

// GetServerByID, given a serverID, returns
// the entity ServerType related to that
// serverID
func GetServerByID(serverID string) *entities.ServerType {
	servers.lock.RLock()
	s := servers.cache[serverID]
	servers.lock.RUnlock()

	return s
}

// GetCustomRanksByGuildId, give a guildId, returns the
// customRanks entity array related to that guildId
func GetCustomRanksByGuildId(guildId string) []entities.CustomRanks {
	servers.lock.Lock()
	cr := servers.cache[guildId].CustomRanks
	servers.lock.Unlock()

	return cr
}

// GetRankRoleByLevel, give the serverType and the level
// returns the customRanks related to that server and that level
func GetRankRoleByLevel(s entities.ServerType, level uint) entities.CustomRanks {
	for _, v := range s.CustomRanks {
		if level >= v.MinLevel {
			return v
		}
	}

	return entities.CustomRanks{
		Id:       "",
		MinLevel: 0,
	}
}

// GetServerMultiplierByGuildId, given a guildId
// returns the float value related to the msgMultiplier
// that is the multiplier for which every action
// is multiplied for
func GetServerMultiplierByGuildId(guildId string) float64 {
	servers.lock.RLock()
	mem := servers.cache[guildId].MsgExpMultiplier
	servers.lock.RUnlock()

	return mem

}

// GetServersWrites returns the number of
// writes in the cache
func GetServersWrites() int {
	users.lock.RLock()
	writes := users.writes
	users.lock.RUnlock()

	return writes
}

// Update methods
// If the server exists, update it, otherwise insert it.
func UpsertServerByID(serverID string, server *entities.ServerType) {
	servers.lock.Lock()
	servers.cache[serverID] = server
	servers.writes++
	servers.lock.Unlock()
}

// UpdateRoleServer takes a server ID and a custom rank,
// and adds the custom rank to the server's custom ranks
func UpdateRoleServer(serverID string, rank entities.CustomRanks) {
	servers.lock.Lock()
	servers.cache[serverID].CustomRanks = append(servers.cache[serverID].CustomRanks, rank)
	servers.writes++
	servers.lock.Unlock()
}

// UpdateServerByID updates a server in the database by its ID.
func UpdateServerByID(serverID string, server *entities.ServerType) {

	filter := bson.M{"serverid": serverID}
	replacement := bson.M{"$set": server}
	SERVERS_DB.ReplaceOne(context.TODO(), filter, replacement)
	SERVERS_DB.FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": bson.M{"customranks": server.CustomRanks}})
}

// Update all servers in the cache to the database.
func UpdateAllServersInDb() error {
	for k, v := range servers.cache {
		UpdateServerByID(k, v)
	}
	users.lock.Lock()
	users.writes = 0
	users.lock.Unlock()
	return nil
}

// Delete methods
// It takes a server ID and a role ID,
// and removes the role from the server's custom ranks
func DeleteRoleServer(serverID, roleId string) {
	servers.lock.Lock()
	s := servers.cache[serverID]
	for i, c := range s.CustomRanks {
		if c.Id == roleId {
			s.CustomRanks = append(s.CustomRanks[:i], s.CustomRanks[i+1:]...)
			servers.writes++
			break
		}
	}
	servers.lock.Unlock()
}
