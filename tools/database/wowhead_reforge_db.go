package database

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/tailscale/hujson"
	"github.com/wowsims/mop/sim/core/proto"
)

func ParseWowheadReforgeStats(contents string) WowheadReforgeStats {
	var stats WowheadReforgeStats
	standardized, err := hujson.Standardize([]byte(contents)) // Removes invalid JSON, such as trailing commas
	if err != nil {
		log.Fatalf("Failed to standardize json %s\n\n%s\n\n%s", err, contents[0:30], contents[len(contents)-30:])
	}

	err = json.Unmarshal(standardized, &stats)
	if err != nil {
		log.Fatalf("failed to parse wowhead item db to json %s\n\n%s", err, contents[0:30])
	}
	fmt.Printf("\n--\nWowhead reforges loaded\n--\n")
	return stats
}

// statStringToEnum maps wowhead stat strings to their corresponding proto.Stat enum values
var statStringToEnum = map[string]proto.Stat{
	"spi":          proto.Stat_StatSpirit,
	"dodgertng":    proto.Stat_StatDodgeRating,
	"parryrtng":    proto.Stat_StatParryRating,
	"hitrtng":      proto.Stat_StatHitRating,
	"critstrkrtng": proto.Stat_StatCritRating,
	"hastertng":    proto.Stat_StatHasteRating,
	"exprtng":      proto.Stat_StatExpertiseRating,
	"mastrtng":     proto.Stat_StatMasteryRating,
}

func (reforgeStats WowheadReforgeStats) ToProto() map[int32]*proto.ReforgeStat {
	protoStatsMap := make(map[int32]*proto.ReforgeStat)
	for _, stat := range reforgeStats {
		protoStat := &proto.ReforgeStat{
			Id:         int32(stat.ReforgeID),
			FromStat:   statStringToEnum[stat.FromStat],
			ToStat:     statStringToEnum[stat.ToStat],
			Multiplier: stat.ReforgeMultiplier,
		}
		protoStatsMap[protoStat.Id] = protoStat
	}
	return protoStatsMap
}

type WowheadReforgeStat struct {
	ReforgeID         int     `json:"id"` // Reforge ID used by game
	FromID            int     `json:"i1"` // WH Stat ID to reforge from
	FromStat          string  `json:"s1"` // WH Stat string to reforge from
	ToID              int     `json:"i2"` // WH Stat ID to reforge to
	ToStat            string  `json:"s2"` // WH Stat string to reforge to
	ReforgeMultiplier float64 `json:"v"`  // Multiplier for reforge, always 0.4
}

type WowheadReforgeStats map[string]WowheadReforgeStat
