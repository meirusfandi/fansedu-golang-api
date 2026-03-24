package cache

// Konvensi key Redis (FansEdu).
const (
	KeyProvinceList = "province:list"
	// KeyCityListPrefix + province_id → kab/kota
	KeyCityListPrefix = "city:list:"

	KeySchoolList = "school:list"

	// KeyPackagesList — GET /api/v1/packages (landing)
	KeyPackagesList = "packages:list"

	// KeyLeaderboardZPrefix + tryout_id — sorted set (score = nilai tryout)
	KeyLeaderboardZPrefix = "leaderboard:"
)

func CityListKey(provinceID string) string {
	return KeyCityListPrefix + provinceID
}

func LeaderboardZKey(tryoutID string) string {
	return KeyLeaderboardZPrefix + tryoutID
}
