package util

import (
	"strconv"
	"time"
)

func UnixNanoTimestamp() int64 {
	return time.Now().UnixNano()
}

func UnixTimestamp() int64 {
	return time.Now().Unix()
}

// UnixTimeToTime converts a Unix timestamp to a time.Time object.
func UnixTimeToTime(unixTimestamp int64) time.Time {
	return time.Unix(unixTimestamp, 0)
}

func UnixNanoToTime(unixNanoTimestamp int64) time.Time {
	return time.Unix(0, unixNanoTimestamp)
}

func TimeToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

func UnixMilliTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// The format or neccessity here is not totally decided upon.
// UserSessionIdValue generates a session ID for a user based on the user's ID and the current time.
// It's the same as the timestamp used when logging in a user.
func UserSessionIdValue(userId uint64, timestamp time.Time) string {
	return strconv.FormatUint(userId, 10) + "-" + strconv.FormatInt(timestamp.UnixNano(), 10)[:5]
}

func TimestampToTime(timestamp string) time.Time {
	i, _ := strconv.ParseInt(timestamp, 10, 64)
	return time.Unix(i, 0)
}
